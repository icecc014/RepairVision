package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminScheduleGenerateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminScheduleGenerateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminScheduleGenerateLogic {
	return &AdminScheduleGenerateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminScheduleGenerateLogic) AdminScheduleGenerate(req *types.ScheduleGenerateRequest) (resp *types.ScheduleListResponse, err error) {
	out, err := l.svcCtx.WorkerRpc.GenerateWeekly(l.ctx,
		&workerclient.GenerateScheduleRequest{WeekStart: req.WeekStart})
	if err != nil {
		return nil, rpcBizError(err)
	}
	return &types.ScheduleListResponse{List: schedulePbListToTypes(out.Items)}, nil
}
