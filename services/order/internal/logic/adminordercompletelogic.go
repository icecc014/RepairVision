package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/mapclient"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
)

// adminCompleteWorkerID 取工单当前承接工人（可空）。
func adminCompleteWorkerID(order *store.Order) int64 {
	if order != nil && order.WorkerID.Valid {
		return order.WorkerID.Int64
	}
	return 0
}
type AdminOrderCompleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOrderCompleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOrderCompleteLogic {
	return &AdminOrderCompleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AdminOrderComplete 管理员代为完工：用于演示/救急（工人无法登录时把工单标记完成）。
// 语义与工人端完工一致：写 completed_at、清故障标记、通知相关人、触发派单重算。
func (l *AdminOrderCompleteLogic) AdminOrderComplete(req *types.OrderIdRequest) (*types.EmptyResponse, error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("工单不存在")
		}
		return nil, errs.Internal(err)
	}
	if order.Status != store.StatusDispatched && order.Status != store.StatusWorking {
		return nil, errs.Conflict("只有「已派单 / 维修中」的工单可以代为完工")
	}
	affected, err := store.AdminCompleteOrder(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("工单状态已变化，请刷新后重试")
	}
	logx.WithContext(l.ctx).Infof("order %s completed by admin %d", order.OrderNo, identity.UID)
	if _, markerErr := l.svcCtx.MapRpc.RemoveFaultMarker(l.ctx, &mapclient.FaultMarkerOrderRequest{OrderId: order.ID}); markerErr != nil {
		logx.WithContext(l.ctx).Errorf("remove fault marker failed: %v", markerErr)
	}
	l.svcCtx.WS.PublishOrder(ws.OrderEvent{
		Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
		BuildingId: order.BuildingID, WorkerId: adminCompleteWorkerID(order), Status: store.StatusCompleted,
	})
	recipients := append([]int64{order.ReporterID}, adminUserIDs(l.ctx, l.svcCtx)...)
	if adminCompleteWorkerID(order) > 0 {
		recipients = append(recipients, adminCompleteWorkerID(order))
	}
	notifyUsers(l.ctx, l.svcCtx, recipients, "complete",
		"工单已完成 "+order.OrderNo, "管理员已将 "+order.Room+"室工单标记为完成", order.ID)
	// 完工后工人空闲出来，立即参与待派队列重算
	TriggerDispatchRecheck(l.svcCtx)
	return &types.EmptyResponse{}, nil
}
