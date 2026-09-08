package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type ListManagedBuildingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListManagedBuildingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListManagedBuildingsLogic {
	return &ListManagedBuildingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListManagedBuildingsLogic) ListManagedBuildings(in *worker.IdRequest) (*worker.BuildingIdsResponse, error) {
	ids, err := store.ListWorkerBuildingIDs(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		return nil, err
	}
	return &worker.BuildingIdsResponse{Ids: ids}, nil
}
