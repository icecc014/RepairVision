package logic

import (
	auth "order/internal/auth"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminUserResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUserResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserResetPasswordLogic {
	return &AdminUserResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUserResetPasswordLogic) AdminUserResetPassword(req *types.AdminUserResetPasswordRequest) (resp *types.EmptyResponse, err error) {
	defer func() {
		// V9.8 方案A2：改密成功后递增版本号，使该账号在其它端的旧 token 失效
		if err == nil {
			auth.BumpPwdVersion(l.ctx, l.svcCtx.PubCache, req.Id)
		}
	}()
	if len(req.Password) < 6 {
		return nil, errs.BadRequest("新密码长度不能少于6位")
	}
	if _, err := l.svcCtx.WorkerRpc.ResetPassword(l.ctx, &workerclient.ResetPasswordRequest{
		Id:       req.Id,
		Password: req.Password,
	}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
