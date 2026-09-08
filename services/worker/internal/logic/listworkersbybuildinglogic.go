package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type ListWorkersByBuildingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWorkersByBuildingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWorkersByBuildingLogic {
	return &ListWorkersByBuildingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListWorkersByBuildingLogic) ListWorkersByBuilding(in *worker.BuildingWorkersRequest) (*worker.BuildingWorkersResponse, error) {
	users, err := store.ListWorkersByBuilding(l.ctx, l.svcCtx.DB, in.BuildingId)
	if err != nil {
		return nil, err
	}
	resp := &worker.BuildingWorkersResponse{}
	for _, u := range users {
		resp.Workers = append(resp.Workers, workerInfoToPb(u))
	}
	return resp, nil
}
