package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
)

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.OrderIdRequest) (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	prevOrder, _ := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id)
	affected, err := store.CancelOrder(l.ctx, l.svcCtx.DB, req.Id, identity.BuildingID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("工单不存在或已开工，无法取消")
	}
	if order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id); err == nil {
		if _, markerErr := l.svcCtx.MapRpc.RemoveFaultMarker(l.ctx, &mapclient.FaultMarkerOrderRequest{OrderId: order.ID}); markerErr != nil {
			logx.WithContext(l.ctx).Errorf("remove fault marker failed: %v", markerErr)
		}
		l.svcCtx.WS.PublishOrder(ws.OrderEvent{
			Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
			BuildingId: order.BuildingID, Status: order.Status,
		})
		recipients := adminUserIDs(l.ctx, l.svcCtx)
		if prevOrder != nil && prevOrder.WorkerID.Valid {
			recipients = append(recipients, prevOrder.WorkerID.Int64)
		}
		notifyUsers(l.ctx, l.svcCtx, recipients, "cancel",
			"工单已取消 "+order.OrderNo, order.Room+"室 工单已取消", order.ID)
	}
	return &types.EmptyResponse{}, nil
}
