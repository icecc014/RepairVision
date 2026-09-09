package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminScheduleSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminScheduleSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminScheduleSaveLogic {
	return &AdminScheduleSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminScheduleSaveLogic) AdminScheduleSave(req *types.ScheduleSaveRequest) (resp *types.ScheduleListResponse, err error) {
	items := make([]*workerclient.ScheduleItem, 0, len(req.Items))
	for i := range req.Items {
		items = append(items, scheduleTypeToPb(req.Items[i]))
	}
	out, err := l.svcCtx.WorkerRpc.SaveSchedules(l.ctx, &workerclient.SaveSchedulesRequest{Items: items})
	if err != nil {
		return nil, rpcBizError(err)
	}
	return &types.ScheduleListResponse{List: schedulePbListToTypes(out.Items)}, nil
}
