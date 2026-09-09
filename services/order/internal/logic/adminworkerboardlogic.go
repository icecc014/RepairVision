package logic

import (
	"context"
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
func (l *AdminWorkerBoardLogic) AdminWorkerBoard() (resp *types.AdminWorkerBoardResponse, err error) {
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
	buildingNames := make(map[int64]string, len(buildingResp.Buildings))
	for _, b := range buildingResp.Buildings {
		buildingNames[b.Id] = b.Code + " " + b.Name
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
	completedCounts, err := store.CountCompletedTodayByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
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
			Available:      shift != shiftOff && active < maxConcurrent,
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
