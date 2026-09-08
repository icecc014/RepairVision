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

type AdminUserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserUpdateLogic {
	return &AdminUserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUserUpdateLogic) AdminUserUpdate(req *types.AdminUserUpdateRequest) (resp *types.EmptyResponse, err error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errs.BadRequest("姓名不能为空")
	}
	if req.Role < 1 || req.Role > 3 {
		return nil, errs.BadRequest("角色不合法")
	}
	if _, err := l.svcCtx.WorkerRpc.UpdateUser(l.ctx, &workerclient.UpdateUserRequest{
		Id:          req.Id,
		Name:        strings.TrimSpace(req.Name),
		Phone:       strings.TrimSpace(req.Phone),
		Role:        req.Role,
		Status:      req.Status,
		BuildingId:  req.BuildingId,
		BuildingIds: req.BuildingIds,
	}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
