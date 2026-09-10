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

type CompleteOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompleteOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteOrderLogic {
	return &CompleteOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompleteOrderLogic) CompleteOrder(req *types.OrderIdRequest) (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	affected, err := store.CompleteOrder(l.ctx, l.svcCtx.DB, req.Id, identity.UID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("工单不存在或当前状态不可完工")
	}
	if order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id); err == nil {
		if _, markerErr := l.svcCtx.MapRpc.RemoveFaultMarker(l.ctx, &mapclient.FaultMarkerOrderRequest{OrderId: order.ID}); markerErr != nil {
			logx.WithContext(l.ctx).Errorf("remove fault marker failed: %v", markerErr)
		}
		l.svcCtx.WS.PublishOrder(ws.OrderEvent{
			Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
			BuildingId: order.BuildingID, WorkerId: identity.UID, Status: order.Status,
		})
		recipients := append([]int64{order.ReporterID}, adminUserIDs(l.ctx, l.svcCtx)...)
		notifyUsers(l.ctx, l.svcCtx, recipients, "complete",
			"工单已完成 "+order.OrderNo, order.Room+"室维修已完成", order.ID)
	}
	return &types.EmptyResponse{}, nil
}
