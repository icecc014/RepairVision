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

type DormOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDormOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DormOrdersLogic {
	return &DormOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DormOrdersLogic) DormOrders(req *types.OrderListRequest) (resp *types.OrderListResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if identity.BuildingID <= 0 {
		return nil, errs.Forbidden("宿管账号未绑定楼栋")
	}
	orders, err := store.ListOrdersByBuilding(l.ctx, l.svcCtx.DB, identity.BuildingID, req.Status)
	if err != nil {
		return nil, err
	}
	items, err := buildOrderItems(l.ctx, l.svcCtx, orders)
	if err != nil {
		return nil, err
	}
	return &types.OrderListResponse{List: items}, nil
}
