package logic

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type GenerateWeeklyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateWeeklyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateWeeklyLogic {
	return &GenerateWeeklyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

const (
	maxRestDaysPerWeek = 3
	maxMinPerBuilding  = 5
)

// GenerateWeekly 生成一周精细排班：
//  1. 已批准请假先固定为 OFF，并占用「每周休息天数」配额（避免请假后又排轮休导致在岗天数过低）；
//  2. 剩余休息名额按「错峰 + 保证每栋楼最少在岗人数」的代价函数分配到非请假日；
//  3. 每日每栋楼按在岗人数分配 MORNING / AFTERNOON，人数不足时保持全天班优先保证覆盖；
//  4. 某栋楼当天在岗人数低于 minPerBuilding 时写入 warnings 返回给管理员。
func (l *GenerateWeeklyLogic) GenerateWeekly(in *worker.GenerateScheduleRequest) (*worker.ScheduleListResponse, error) {
	weekStart, err := time.ParseInLocation("2006-01-02", in.WeekStart, time.Local)
	if err != nil || weekStart.Weekday() != time.Monday {
		return nil, errDate("请提供周一的日期，格式 YYYY-MM-DD")
	}

	restDays := int(in.RestDaysPerWeek)
	if restDays <= 0 {
		restDays = 1
	}
	if restDays > maxRestDaysPerWeek {
		restDays = maxRestDaysPerWeek
	}
	minPerBuilding := int(in.MinPerBuilding)
	if minPerBuilding <= 0 {
		minPerBuilding = 1
	}
	if minPerBuilding > maxMinPerBuilding {
		minPerBuilding = maxMinPerBuilding
	}

	workerIDs, err := store.ListWorkerIDs(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	if len(workerIDs) == 0 {
		return &worker.ScheduleListResponse{}, nil
	}
	buildingsByWorker := make(map[int64][]int64, len(workerIDs))
	for _, id := range workerIDs {
		ids, err := store.ListWorkerBuildingIDs(l.ctx, l.svcCtx.DB, id)
		if err != nil {
			return nil, err
		}
		buildingsByWorker[id] = ids
	}

	weekStartStr := weekStart.Format("2006-01-02")
	weekEndStr := weekStart.AddDate(0, 0, 6).Format("2006-01-02")
	leaveDays, err := l.loadLeaveDays(workerIDs, weekStart, weekStartStr, weekEndStr)
	if err != nil {
		return nil, err
	}

	weekSeed := int(weekStart.Unix()/86400) % 7
	if weekSeed < 0 {
		weekSeed += 7
	}
	offSets := assignOffDays(workerIDs, buildingsByWorker, weekSeed, leaveDays, restDays, minPerBuilding)

	items := make([]*worker.ScheduleItem, 0, len(workerIDs)*7)
	warnings := make([]string, 0)
	for day := 0; day < 7; day++ {
		shiftByWorker := assignShiftsForDay(workerIDs, buildingsByWorker, offSets, day, minPerBuilding)
		date := weekStart.AddDate(0, 0, day).Format("2006-01-02")
		for _, workerID := range workerIDs {
			shift := shiftByWorker[workerID]
			items = append(items, &worker.ScheduleItem{
				WorkerId:  workerID,
				WorkDate:  date,
				ShiftType: shift,
				Note:      noteForShift(shift, offSets[workerID][day], leaveDays[workerID][day]),
			})
		}
		warnings = append(warnings, coverageWarnings(workerIDs, buildingsByWorker, shiftByWorker, day, weekStart, minPerBuilding)...)
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if err := store.ClearScheduleRange(txCtx, session, weekStartStr, weekEndStr); err != nil {
			return err
		}
		for _, item := range items {
			if err := store.UpsertSchedule(txCtx, session, item.WorkerId, item.WorkDate, item.ShiftType, item.Note); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &worker.ScheduleListResponse{Items: items, Warnings: warnings}, nil
}

// loadLeaveDays 把该周内「已批准」的请假展开成 worker -> 星期序号(0-6) 的集合。
func (l *GenerateWeeklyLogic) loadLeaveDays(workerIDs []int64, weekStart time.Time, startStr, endStr string) (map[int64]map[int]bool, error) {
	rows, err := store.ListApprovedLeavesBetween(l.ctx, l.svcCtx.DB, startStr, endStr)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]map[int]bool, len(workerIDs))
	for _, id := range workerIDs {
		result[id] = map[int]bool{}
	}
	for workerID, days := range rows {
		set, ok := result[workerID]
		if !ok {
			continue // 已停用或不在排班名单里的工人忽略
		}
		for _, day := range days {
			d, err := time.ParseInLocation("2006-01-02", day, time.Local)
			if err != nil {
				continue
			}
			index := int(d.Sub(weekStart).Hours() / 24)
			if index >= 0 && index < 7 {
				set[index] = true
			}
		}
	}
	return result, nil
}

// assignOffDays 分配休息日：请假先占名额，剩余名额用「错峰 + 楼栋最少在岗」代价函数挑选。
func assignOffDays(workerIDs []int64, buildingsByWorker map[int64][]int64, weekSeed int,
	leaveDays map[int64]map[int]bool, restDays, minPerBuilding int) map[int64]map[int]bool {
	off := make(map[int64]map[int]bool, len(workerIDs))
	for _, id := range workerIDs {
		set := make(map[int]bool, restDays+1)
		for day := range leaveDays[id] {
			set[day] = true
		}
		off[id] = set
	}
	buildingWorkers := buildBuildingWorkers(workerIDs, buildingsByWorker)
	offCountByDay := make([]int, 7)
	for _, id := range workerIDs {
		for day := range off[id] {
			offCountByDay[day]++
		}
	}

	for i, id := range workerIDs {
		need := restDays - len(off[id])
		for k := 0; k < need; k++ {
			preferred := (i + weekSeed + k*2) % 7
			bestDay, bestCost := -1, 1<<30
			for step := 0; step < 7; step++ {
				day := (preferred + step) % 7
				if off[id][day] {
					continue
				}
				cost := offCountByDay[day] * 10
				if off[id][(day+6)%7] || off[id][(day+1)%7] {
					cost += 5 // 尽量不连续休息
				}
				for _, bid := range buildingsByWorker[id] {
					onDuty := 0
					for _, other := range buildingWorkers[bid] {
						if other == id {
							continue
						}
						if !off[other][day] {
							onDuty++
						}
					}
					if onDuty == 0 {
						cost += 1000 // 该楼栋当天将无人值守
					} else if onDuty < minPerBuilding {
						cost += 100 // 低于最少在岗要求
					}
				}
				if cost < bestCost {
					bestCost, bestDay = cost, day
				}
			}
			if bestDay < 0 {
				break
			}
			off[id][bestDay] = true
			offCountByDay[bestDay]++
		}
	}
	return off
}

func buildBuildingWorkers(workerIDs []int64, buildingsByWorker map[int64][]int64) map[int64][]int64 {
	result := make(map[int64][]int64)
	for _, id := range workerIDs {
		for _, bid := range buildingsByWorker[id] {
			result[bid] = append(result[bid], id)
		}
	}
	return result
}

// assignShiftsForDay 计算某天各工人班次：每栋楼按在岗人数分配午班/晚班/全天。
func assignShiftsForDay(workerIDs []int64, buildingsByWorker map[int64][]int64,
	off map[int64]map[int]bool, day, minPerBuilding int) map[int64]string {
	result := make(map[int64]string, len(workerIDs))
	for _, id := range workerIDs {
		if off[id][day] {
			result[id] = "OFF"
		} else {
			result[id] = "DAY"
		}
	}
	workersOfBuilding := buildBuildingWorkers(workerIDs, buildingsByWorker)
	groups := make([][]int64, 0, len(workersOfBuilding))
	for _, group := range workersOfBuilding {
		onDuty := make([]int64, 0, len(group))
		for _, id := range group {
			if result[id] != "OFF" {
				onDuty = append(onDuty, id)
			}
		}
		groups = append(groups, onDuty)
	}
	// 先处理人少的楼栋，保证只剩 1~2 人的楼栋保持全天班（不被其它楼栋改写成半天班）。
	sort.Slice(groups, func(i, j int) bool { return len(groups[i]) < len(groups[j]) })
	for _, onDuty := range groups {
		// 在岗人数不足以拆分半天班时保持全天班，优先保证覆盖
		if len(onDuty) < max(2, minPerBuilding) {
			continue
		}
		candidates := make([]int64, 0, len(onDuty))
		for _, id := range onDuty {
			if result[id] == "DAY" {
				candidates = append(candidates, id)
			}
		}
		if len(candidates) < 2 {
			continue
		}
		offset := day % len(candidates)
		result[candidates[offset]] = "MORNING"
		result[candidates[(offset+1)%len(candidates)]] = "AFTERNOON"
	}
	return result
}

// coverageWarnings 统计各楼栋当天在岗人数是否满足最少在岗要求。
func coverageWarnings(workerIDs []int64, buildingsByWorker map[int64][]int64,
	shiftByWorker map[int64]string, day int, weekStart time.Time, minPerBuilding int) []string {
	workersOfBuilding := buildBuildingWorkers(workerIDs, buildingsByWorker)
	warnings := make([]string, 0)
	for bid, group := range workersOfBuilding {
		onDuty := 0
		for _, id := range group {
			if shiftByWorker[id] != "OFF" {
				onDuty++
			}
		}
		if onDuty < minPerBuilding {
			warnings = append(warnings, fmt.Sprintf("%s 楼栋#%d 仅 %d 人在岗（要求 %d 人）",
				weekStart.AddDate(0, 0, day).Format("01-02"), bid, onDuty, minPerBuilding))
		}
	}
	sort.Strings(warnings)
	return warnings
}

func noteForShift(shift string, isOff, isLeave bool) string {
	switch shift {
	case "OFF":
		if isLeave {
			return "已批准请假"
		}
		return "周模板轮休"
	case "MORNING":
		return "午班 08:00-14:00"
	case "AFTERNOON":
		return "晚班 14:00-20:00"
	default:
		return "全天班"
	}
}
