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

type AdminOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOrdersLogic {
	return &AdminOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminOrdersLogic) AdminOrders(req *types.AdminOrderListRequest) (resp *types.OrderListResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	orders, err := store.ListAllOrders(l.ctx, l.svcCtx.DB, req.Status, req.BuildingId)
	if err != nil {
		return nil, err
	}
	items, err := buildOrderItems(l.ctx, l.svcCtx, orders)
	if err != nil {
		return nil, err
	}
	return &types.OrderListResponse{List: items}, nil
}
