package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, errs.BadRequest("用户名和密码不能为空")
	}
	out, err := l.svcCtx.WorkerRpc.Login(l.ctx, &workerclient.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, errs.Unauthorized("用户名或密码错误")
	}
	u := out.User
	expire := l.svcCtx.Config.Auth.AccessExpire
	if expire <= 0 {
		expire = 86400
	}
	token, err := auth.Sign(l.svcCtx.Config.Auth.AccessSecret, expire, auth.Identity{
		UID:        u.Id,
		Role:       u.Role,
		BuildingID: u.BuildingId,
		Username:   u.Username,
		Name:       u.Name,
	})
	if err != nil {
		return nil, errs.Internal(err)
	}
	return &types.LoginResponse{
		Token: token,
		User: types.UserInfo{
			Id:         u.Id,
			Username:   u.Username,
			Name:       u.Name,
			Role:       u.Role,
			BuildingId: u.BuildingId,
		},
	}, nil
}
