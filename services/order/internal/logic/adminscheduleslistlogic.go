package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminSchedulesListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminSchedulesListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSchedulesListLogic {
	return &AdminSchedulesListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminSchedulesListLogic) AdminSchedulesList(req *types.ScheduleQueryRequest) (resp *types.ScheduleListResponse, err error) {
	out, err := l.svcCtx.WorkerRpc.ListSchedules(l.ctx, &workerclient.ScheduleListRequest{
		WorkerId:  req.WorkerId,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	return &types.ScheduleListResponse{List: schedulePbListToTypes(out.Items)}, nil
}
