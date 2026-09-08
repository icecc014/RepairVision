package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type GetUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsersLogic {
	return &GetUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUsersLogic) GetUsers(in *worker.UserIdsRequest) (*worker.UsersResponse, error) {
	users, err := store.FindUsersByIDs(l.ctx, l.svcCtx.DB, in.Ids)
	if err != nil {
		return nil, err
	}
	resp := &worker.UsersResponse{}
	for _, u := range users {
		resp.Users = append(resp.Users, userToPb(u))
	}
	return resp, nil
}
