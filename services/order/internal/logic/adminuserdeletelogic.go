package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminUserDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserDeleteLogic {
	return &AdminUserDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUserDeleteLogic) AdminUserDelete(req *types.AdminUserIdRequest) (resp *types.EmptyResponse, err error) {
	if _, err := l.svcCtx.WorkerRpc.DisableUser(l.ctx, &workerclient.IdRequest{Id: req.Id}); err != nil {
		return nil, rpcBizError(err)
	}
	// 人员变动后自动重算各工种的管辖楼栋（保证每栋楼每类工种都有人）
	triggerAutoAssign(l.ctx, l.svcCtx)
	return &types.EmptyResponse{}, nil
}
