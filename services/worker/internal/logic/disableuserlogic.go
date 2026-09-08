package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type DisableUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableUserLogic {
	return &DisableUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DisableUserLogic) DisableUser(in *worker.IdRequest) (*worker.Response, error) {
	u, err := store.FindUserByID(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	if u.Role == 1 {
		return nil, errors.New("管理员账号不可停用")
	}
	if err := store.DisableUser(l.ctx, l.svcCtx.DB, in.Id); err != nil {
		return nil, err
	}
	return &worker.Response{Pong: "ok"}, nil
}
