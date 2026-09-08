package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUsersLogic) ListUsers(in *worker.ListUsersRequest) (*worker.ListUsersResponse, error) {
	users, err := store.ListUsers(l.ctx, l.svcCtx.DB, in.Role, in.Status, in.Keyword)
	if err != nil {
		return nil, err
	}
	resp := &worker.ListUsersResponse{}
	for _, u := range users {
		var buildingIDs []int64
		if u.Role == 2 {
			buildingIDs, err = store.ListWorkerBuildingIDs(l.ctx, l.svcCtx.DB, u.ID)
			if err != nil {
				return nil, err
			}
		}
		resp.Users = append(resp.Users, userToPbWithBuildings(u, buildingIDs))
	}
	return resp, nil
}
