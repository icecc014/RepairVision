package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminFaultTypeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminFaultTypeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFaultTypeDeleteLogic {
	return &AdminFaultTypeDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminFaultTypeDeleteLogic) AdminFaultTypeDelete(req *types.AdminFaultTypeIdRequest) (resp *types.EmptyResponse, err error) {
	if _, err := store.FindFaultTypeByID(l.ctx, l.svcCtx.DB, req.Id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("维修类型不存在")
		}
		return nil, errs.Internal(err)
	}
	if err := store.DisableFaultType(l.ctx, l.svcCtx.DB, req.Id); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
