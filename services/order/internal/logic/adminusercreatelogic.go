package logic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminUserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserCreateLogic {
	return &AdminUserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUserCreateLogic) AdminUserCreate(req *types.AdminUserCreateRequest) (resp *types.EmptyResponse, err error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, errs.BadRequest("用户名、密码和姓名不能为空")
	}
	if req.Role < 1 || req.Role > 3 {
		return nil, errs.BadRequest("角色不合法")
	}
	if _, err := l.svcCtx.WorkerRpc.CreateUser(l.ctx, &workerclient.CreateUserRequest{
		Username:      strings.TrimSpace(req.Username),
		Password:      req.Password,
		Name:          strings.TrimSpace(req.Name),
		Phone:         strings.TrimSpace(req.Phone),
		Role:          req.Role,
		BuildingId:    req.BuildingId,
		BuildingIds:   req.BuildingIds,
		MaxConcurrent: req.MaxConcurrent,
	}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
