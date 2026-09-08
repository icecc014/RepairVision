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

type ResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResetPasswordLogic) ResetPassword(in *worker.ResetPasswordRequest) (*worker.Response, error) {
	if len(in.Password) < 6 {
		return nil, errors.New("新密码长度不能少于6位")
	}
	if _, err := store.FindUserByID(l.ctx, l.svcCtx.DB, in.Id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := store.UpdateUserPassword(l.ctx, l.svcCtx.DB, in.Id, string(hash)); err != nil {
		return nil, err
	}
	return &worker.Response{Pong: "ok"}, nil
}
