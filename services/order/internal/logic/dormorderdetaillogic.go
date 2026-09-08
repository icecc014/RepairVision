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

type DormOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDormOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DormOrderDetailLogic {
	return &DormOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DormOrderDetailLogic) DormOrderDetail(req *types.OrderIdRequest) (resp *types.OrderDetailResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		return nil, errs.NotFound("工单不存在")
	}
	if identity.BuildingID != order.BuildingID {
		return nil, errs.NotFound("工单不存在")
	}
	items, err := buildOrderItems(l.ctx, l.svcCtx, []store.Order{*order})
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, errs.NotFound("工单不存在")
	}
	return &types.OrderDetailResponse{Order: items[0]}, nil
}
