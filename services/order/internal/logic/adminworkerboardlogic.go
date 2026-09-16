package logic

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminWorkerBoardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminWorkerBoardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminWorkerBoardLogic {
	return &AdminWorkerBoardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminWorkerBoard 聚合今日班次、在途单、今日完工与最大并发，形成调度看板。
func (l *AdminWorkerBoardLogic) AdminWorkerBoard(req *types.AdminWorkerBoardRequest) (resp *types.AdminWorkerBoardResponse, err error) {
	usersResp, err := l.svcCtx.WorkerRpc.ListUsers(l.ctx, &workerclient.ListUsersRequest{Role: 2, Status: 1})
	if err != nil {
		return nil, errs.Upstream()
	}
	if len(usersResp.Users) == 0 {
		return &types.AdminWorkerBoardResponse{List: []types.AdminWorkerBoardItem{}}, nil
	}

	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	days := int64(1)
	if req != nil && req.Days > 0 {
		days = req.Days
	}
	buildingNames := make(map[int64]string, len(buildingResp.Buildings))
	for _, b := range buildingResp.Buildings {
		// 楼栋名去重：编码已包含在名称里（如 code=1 / name=1号宿舍楼）时只显示名称
		if b.Code != "" && !strings.HasPrefix(b.Name, b.Code) {
			buildingNames[b.Id] = b.Code + " " + b.Name
		} else {
			buildingNames[b.Id] = b.Name
		}
	}

	today := time.Now().Format("2006-01-02")
	scheduleResp, err := l.svcCtx.WorkerRpc.ListSchedules(l.ctx,
		&workerclient.ScheduleListRequest{StartDate: today, EndDate: today})
	if err != nil {
		return nil, errs.Upstream()
	}
	todayShift := make(map[int64]string)
	for _, item := range scheduleResp.Items {
		todayShift[item.WorkerId] = item.ShiftType
	}

	workerIDs := make([]int64, 0, len(usersResp.Users))
	for _, u := range usersResp.Users {
		workerIDs = append(workerIDs, u.Id)
	}
	activeCounts, err := store.CountInProgressByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}
	completedCounts, err := store.CountCompletedByWorkersWithinDays(l.ctx, l.svcCtx.DB, workerIDs, days)
	if err != nil {
		return nil, errs.Internal(err)
	}
	// V5.4：在途工时、当前所在工单/楼栋与在岗状态（白班 ∧ 时段内 ∧ 未请假 ∧ 启用）
	loadMinutes, err := store.CountWorkloadByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}
	activeOrders, err := store.ListActiveOrdersByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}
	currentOrder := make(map[int64]store.WorkerActiveOrder)
	for _, ao := range activeOrders {
		if _, ok := currentOrder[ao.WorkerID]; !ok {
			currentOrder[ao.WorkerID] = ao
		}
	}
	dutyStatus := make(map[int64]*workerclient.DutyStatusResponse, len(workerIDs))
	for _, id := range workerIDs {
		status, err := l.svcCtx.WorkerRpc.GetDutyStatus(l.ctx, &workerclient.DutyStatusRequest{WorkerId: id})
		if err != nil {
			continue
		}
		dutyStatus[id] = status
	}
	avgLoad := 0.0
	for _, u := range usersResp.Users {
		avgLoad += float64(loadMinutes[u.Id])
	}
	if len(usersResp.Users) > 0 {
		avgLoad /= float64(len(usersResp.Users))
	}

	resp = &types.AdminWorkerBoardResponse{List: make([]types.AdminWorkerBoardItem, 0, len(usersResp.Users))}
	for _, u := range usersResp.Users {
		maxConcurrent := u.MaxConcurrent
		if maxConcurrent <= 0 {
			maxConcurrent = defaultMaxConcurrent
		}
		shift := todayShift[u.Id]
		active := activeCounts[u.Id]
		item := types.AdminWorkerBoardItem{
			WorkerId:       u.Id,
			Name:           u.Name,
			Username:       u.Username,
			MaxConcurrent:  maxConcurrent,
			TodayShift:     shift,
			ActiveOrders:   active,
			TodayCompleted: completedCounts[u.Id],
			JobType:        u.JobType,
			JobTypeText:    jobTypeText(u.JobType),
			LoadMinutes:    loadMinutes[u.Id],
		}
		if avgLoad > 0 {
			item.LoadDeviation = math.Round((float64(loadMinutes[u.Id])-avgLoad)/avgLoad*10000) / 10000
		}
		if status := dutyStatus[u.Id]; status != nil {
			item.OnDuty = status.OnDuty
			item.DutyReason = status.Reason
		}
		// V5.4：取消硬性并发上限，可派 = 在岗（负载只影响优先级）
		item.Available = item.OnDuty
		if ao, ok := currentOrder[u.Id]; ok {
			item.CurrentBuildingId = ao.BuildingID
			item.CurrentBuildingName = buildingNames[ao.BuildingID]
			item.CurrentOrderNo = ao.OrderNo
		}
		for _, bid := range u.BuildingIds {
			if name := buildingNames[bid]; name != "" {
				item.BuildingNames = append(item.BuildingNames, name)
			}
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}
