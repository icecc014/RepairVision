package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *worker.LoginRequest) (*worker.LoginResponse, error) {
	if in.Username == "" || in.Password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	u, err := store.FindUserByUsername(l.ctx, l.svcCtx.DB, in.Username)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return &worker.LoginResponse{User: userToPb(*u)}, nil
}
