package logic

import (
	"context"
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

// GenerateWeekly 生成一周精细排班：
//  1. 每名工人每周固定休 1 天，轮休日按工人顺序 + 周序号错开；
//  2. 同一楼栋尽量不在同一天全员休息；
//  3. 每日每个楼栋按在岗人数分配 MORNING/AFTERNOON，多余人员为 DAY，保证午/晚班双覆盖。
func (l *GenerateWeeklyLogic) GenerateWeekly(in *worker.GenerateScheduleRequest) (*worker.ScheduleListResponse, error) {
	weekStart, err := time.ParseInLocation("2006-01-02", in.WeekStart, time.Local)
	if err != nil || weekStart.Weekday() != time.Monday {
		return nil, errDate("请提供周一的日期，格式 YYYY-MM-DD")
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

	weekSeed := int(weekStart.Unix()/86400) % 7
	if weekSeed < 0 {
		weekSeed += 7
	}
	offDay := assignRestDays(workerIDs, buildingsByWorker, weekSeed)

	items := make([]*worker.ScheduleItem, 0, len(workerIDs)*7)
	for day := 0; day < 7; day++ {
		shiftByWorker := assignShiftsForDay(workerIDs, buildingsByWorker, offDay, day)
		for _, workerID := range workerIDs {
			date := weekStart.AddDate(0, 0, day).Format("2006-01-02")
			shift := shiftByWorker[workerID]
			note := ""
			switch shift {
			case "OFF":
				note = "周模板轮休"
			case "MORNING":
				note = "午班 08:00-14:00"
			case "AFTERNOON":
				note = "晚班 14:00-20:00"
			default:
				note = "全天班"
			}
			items = append(items, &worker.ScheduleItem{
				WorkerId: workerID, WorkDate: date, ShiftType: shift, Note: note,
			})
		}
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if err := store.ClearScheduleRange(txCtx, session,
			weekStart.Format("2006-01-02"), weekStart.AddDate(0, 0, 6).Format("2006-01-02")); err != nil {
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
	return &worker.ScheduleListResponse{Items: items}, nil
}

// assignRestDays 为每名工人选轮休日：优先错峰，避免同楼栋同日全员休息。
func assignRestDays(workerIDs []int64, buildingsByWorker map[int64][]int64, weekSeed int) map[int64]int {
	offDay := make(map[int64]int)
	offCountByDay := make([]int, 7)
	buildingWorkers := make(map[int64][]int64)
	for _, id := range workerIDs {
		for _, bid := range buildingsByWorker[id] {
			buildingWorkers[bid] = append(buildingWorkers[bid], id)
		}
	}
	for i, id := range workerIDs {
		preferred := (i + weekSeed) % 7
		bestDay := preferred
		bestCost := int(1) << 30
		for step := 0; step < 7; step++ {
			day := (preferred + step) % 7
			cost := offCountByDay[day] * 10
			for _, bid := range buildingsByWorker[id] {
				group := buildingWorkers[bid]
				onDuty := 0
				for _, other := range group {
					if other == id {
						continue
					}
					if d, ok := offDay[other]; !ok || d != day {
						onDuty++
					}
				}
				if onDuty == 0 {
					cost += 1000 // 该楼栋当天将无人值守
				} else if onDuty == 1 && len(group) >= 2 {
					cost += 20 // 仅剩 1 人，无法午/晚双覆盖
				}
			}
			if cost < bestCost {
				bestCost = cost
				bestDay = day
			}
		}
		offDay[id] = bestDay
		offCountByDay[bestDay]++
	}
	return offDay
}

// assignShiftsForDay 计算某天各工人班次：每个楼栋按人数分配午班/晚班/全天。
func assignShiftsForDay(workerIDs []int64, buildingsByWorker map[int64][]int64, offDay map[int64]int, day int) map[int64]string {
	result := make(map[int64]string, len(workerIDs))
	for _, id := range workerIDs {
		if offDay[id] == day {
			result[id] = "OFF"
		} else {
			result[id] = "DAY"
		}
	}
	workersOfBuilding := make(map[int64][]int64)
	for _, id := range workerIDs {
		for _, bid := range buildingsByWorker[id] {
			workersOfBuilding[bid] = append(workersOfBuilding[bid], id)
		}
	}
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
	// 先处理人少的楼栋，保证只剩 1 人的楼栋能保持全天班（不被其它楼栋改写成半天班）。
	sort.Slice(groups, func(i, j int) bool { return len(groups[i]) < len(groups[j]) })
	for _, onDuty := range groups {
		if len(onDuty) < 2 {
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
