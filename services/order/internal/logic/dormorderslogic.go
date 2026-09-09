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
	total, err := store.CountOrdersByBuildingFilter(l.ctx, l.svcCtx.DB, identity.BuildingID, req.Status)
	if err != nil {
		return nil, errs.Internal(err)
	}
	var orders []store.Order
	if req.Page > 0 || req.Size > 0 {
		orders, err = store.ListOrdersByBuildingPage(l.ctx, l.svcCtx.DB, identity.BuildingID, req.Status, req.Page, req.Size)
	} else {
		orders, err = store.ListOrdersByBuilding(l.ctx, l.svcCtx.DB, identity.BuildingID, req.Status)
	}
	if err != nil {
		return nil, err
	}
	items, err := buildOrderItems(l.ctx, l.svcCtx, orders)
	if err != nil {
		return nil, err
	}
	return &types.OrderListResponse{Total: total, List: items}, nil
}
