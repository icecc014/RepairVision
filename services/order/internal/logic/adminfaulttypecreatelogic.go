package logic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminFaultTypeCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminFaultTypeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFaultTypeCreateLogic {
	return &AdminFaultTypeCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminFaultTypeCreateLogic) AdminFaultTypeCreate(req *types.AdminFaultTypeCreateRequest) (resp *types.EmptyResponse, err error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, errs.BadRequest("编码和名称不能为空")
	}
	if _, err := store.CreateFaultType(l.ctx, l.svcCtx.DB, code, name, req.Sort); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errs.Conflict("该编码已存在")
		}
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
