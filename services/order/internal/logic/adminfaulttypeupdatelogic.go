package logic

import (
	"context"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminFaultTypeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminFaultTypeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFaultTypeUpdateLogic {
	return &AdminFaultTypeUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminFaultTypeUpdateLogic) AdminFaultTypeUpdate(req *types.AdminFaultTypeUpdateRequest) (resp *types.EmptyResponse, err error) {
	existing, err := store.FindFaultTypeByID(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("维修类型不存在")
		}
		return nil, errs.Internal(err)
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errs.BadRequest("名称不能为空")
	}
	status := existing.Status
	if req.Status == 1 {
		status = 1
	}
	if err := store.UpdateFaultType(l.ctx, l.svcCtx.DB, req.Id, req.Sort, status, name); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
