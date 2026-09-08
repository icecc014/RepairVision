package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type WorkerOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerOrdersLogic {
	return &WorkerOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkerOrdersLogic) WorkerOrders(req *types.OrderListRequest) (resp *types.OrderListResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	orders, err := store.ListOrdersByWorker(l.ctx, l.svcCtx.DB, identity.UID, req.Status)
	if err != nil {
		return nil, err
	}
	items, err := buildOrderItems(l.ctx, l.svcCtx, orders)
	if err != nil {
		return nil, err
	}
	return &types.OrderListResponse{List: items}, nil
}
