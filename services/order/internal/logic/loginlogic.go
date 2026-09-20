package logic

import (
	"context"
	"strconv"
	"strings"
	"time"

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
	// V9.9 抗压：同一账号 10 分钟内失败 5 次后锁定，防在线爆破
	lockKey := "loginlock:" + strings.ToLower(strings.TrimSpace(req.Username))
	if l.loginLocked(lockKey) {
		return nil, errs.New(429, "登录尝试过于频繁，请 10 分钟后再试")
	}
	out, err := l.svcCtx.WorkerRpc.Login(l.ctx, &workerclient.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		l.recordLoginFailure(lockKey)
		return nil, errs.Unauthorized("用户名或密码错误")
	}
	l.clearLoginFailure(lockKey)
	u := out.User
	expire := l.svcCtx.Config.Auth.AccessExpire
	if expire <= 0 {
		expire = 86400
	}
	token, err := auth.Sign(l.svcCtx.Config.Auth.AccessSecret, expire, auth.Identity{
		UID:        u.Id,
		PwdVersion: auth.CurrentPwdVersion(l.ctx, l.svcCtx.PubCache, u.Id),
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

// ---- V9.9 抗压：登录失败限次（依赖 Redis 缓存，缓存不可用时放行）----

const (
	loginFailLimit  = 5
	loginFailWindow = 10 * time.Minute
)

func (l *LoginLogic) loginLocked(key string) bool {
	if l.svcCtx == nil || l.svcCtx.PubCache == nil {
		return false
	}
	raw, ok := l.svcCtx.PubCache.Get(l.ctx, key)
	if !ok {
		return false
	}
	n, convErr := strconv.Atoi(strings.TrimSpace(raw))
	return convErr == nil && n >= loginFailLimit
}

func (l *LoginLogic) recordLoginFailure(key string) {
	if l.svcCtx == nil || l.svcCtx.PubCache == nil {
		return
	}
	l.svcCtx.PubCache.Incr(l.ctx, key, loginFailWindow)
}

func (l *LoginLogic) clearLoginFailure(key string) {
	if l.svcCtx == nil || l.svcCtx.PubCache == nil {
		return
	}
	l.svcCtx.PubCache.Del(l.ctx, key)
}
