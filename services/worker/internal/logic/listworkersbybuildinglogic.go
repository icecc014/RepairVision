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
		base := int64(0)
		if u.BuildingID.Valid {
			base = u.BuildingID.Int64
		}
		skills, err := store.ListWorkerSkills(l.ctx, l.svcCtx.DB, u.ID)
		if err != nil {
			return nil, err
		}
		info := workerInfoToPb(u, base, skills)
		if in.WorkDate != "" {
			onLeave, err := store.FindApprovedLeaveForDate(l.ctx, l.svcCtx.DB, u.ID, in.WorkDate)
			if err != nil {
				return nil, err
			}
			if onLeave {
				info.TodayShift = "OFF"
			} else {
				shift, err := store.FindWorkerShift(l.ctx, l.svcCtx.DB, u.ID, in.WorkDate)
				if err != nil {
					return nil, err
				}
				info.TodayShift = shift
			}
		}

		resp.Workers = append(resp.Workers, info)
	}
	return resp, nil
}
