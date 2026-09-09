package logic

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/mapclient"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
	"worker/workerclient"
)

type AdminOrderReassignLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOrderReassignLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOrderReassignLogic {
	return &AdminOrderReassignLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminOrderReassign 管理员对待派/已派工单手动指派或改派工人。
func (l *AdminOrderReassignLogic) AdminOrderReassign(req *types.AdminOrderReassignRequest) (resp *types.EmptyResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if req.WorkerId <= 0 {
		return nil, errs.BadRequest("请选择目标工人")
	}

	order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.BadRequest("工单不存在")
		}
		return nil, errs.Internal(err)
	}
	if order.Status != store.StatusPending && order.Status != store.StatusDispatched {
		return nil, errs.Conflict("只有待派单或已派单（未开工）工单可以改派")
	}

	faultNames, err := faultTypeNameMap(l.ctx, l.svcCtx)
	if err != nil {
		return nil, errs.Internal(err)
	}
	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	var current *mapclient.Building
	for _, b := range buildingResp.Buildings {
		if b.Id == order.BuildingID {
			current = b
			break
		}
	}
	if current == nil {
		return nil, errs.BadRequest("工单所属楼栋不存在")
	}

	rules, err := store.ListDispatchRules(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	wSkill, wDistance, wLoad := dispatchWeights(rules)

	workerResp, err := l.svcCtx.WorkerRpc.ListWorkersByBuilding(l.ctx,
		&workerclient.BuildingWorkersRequest{
			BuildingId: order.BuildingID,
			WorkDate:   time.Now().Format("2006-01-02"),
		})
	if err != nil {
		return nil, errs.Upstream()
	}
	var target *workerclient.WorkerInfo
	workerIDs := make([]int64, 0, len(workerResp.Workers))
	for _, w := range workerResp.Workers {
		workerIDs = append(workerIDs, w.Id)
		if w.Id == req.WorkerId {
			target = w
		}
	}
	if target == nil {
		return nil, errs.BadRequest("目标工人不负责该楼栋")
	}
	countLoads, err := store.CountInProgressByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !workerCanTake(target, countLoads) {
		return nil, errs.Conflict("目标工人当天休息或已达最大并发，请改派其他工人")
	}
	minuteLoads, err := store.CountWorkloadByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}
	score := scoreWorkerForOrder(target, buildingResp.Buildings, current,
		faultNames[order.FaultType], minuteLoads, wSkill, wDistance, wLoad)

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if order.WorkerID.Valid && order.WorkerID.Int64 > 0 && order.WorkerID.Int64 != req.WorkerId {
			if err := store.RevokeActiveDispatch(txCtx, session, order.ID, order.WorkerID.Int64); err != nil {
				return err
			}
		}
		affected, err := store.ReassignOrderWorker(txCtx, session, order.ID, req.WorkerId)
		if err != nil {
			return err
		}
		if !affected {
			return errors.New("工单状态已变化，改派失败")
		}
		return store.InsertDispatchRecord(txCtx, session, order.ID, req.WorkerId,
			score.totalScore, score.skillScore, score.distanceScore, score.loadScore)
	})
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.Conflict(err.Error())
		}
		return nil, errs.Internal(err)
	}

	l.svcCtx.WS.PublishOrder(ws.OrderEvent{
		Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
		BuildingId: order.BuildingID, WorkerId: req.WorkerId, Status: store.StatusDispatched,
	})
	return &types.EmptyResponse{}, nil
}
