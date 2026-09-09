package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
)

type StartOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartOrderLogic {
	return &StartOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StartOrderLogic) StartOrder(req *types.OrderIdRequest) (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	affected, err := store.StartOrder(l.ctx, l.svcCtx.DB, req.Id, identity.UID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("工单不存在或当前状态不可开工")
	}
	if err := store.AcceptDispatchRecord(l.ctx, l.svcCtx.DB, req.Id, identity.UID); err != nil {
		logx.WithContext(l.ctx).Errorf("mark dispatch accepted failed order=%d worker=%d: %v", req.Id, identity.UID, err)
	}
	if order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id); err == nil {
		l.svcCtx.WS.PublishOrder(ws.OrderEvent{
			Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
			BuildingId: order.BuildingID, WorkerId: identity.UID, Status: order.Status,
		})
	}
	return &types.EmptyResponse{}, nil
}
