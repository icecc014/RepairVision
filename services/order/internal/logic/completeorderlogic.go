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
	return &types.EmptyResponse{}, nil
}
