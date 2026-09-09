package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type WorkerSchedulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerSchedulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerSchedulesLogic {
	return &WorkerSchedulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WorkerSchedules 工人查询“我的班次”，WorkerId 以登录身份为准。
func (l *WorkerSchedulesLogic) WorkerSchedules(req *types.ScheduleQueryRequest) (resp *types.ScheduleListResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	out, err := l.svcCtx.WorkerRpc.ListSchedules(l.ctx, &workerclient.ScheduleListRequest{
		WorkerId:  identity.UID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	return &types.ScheduleListResponse{List: schedulePbListToTypes(out.Items)}, nil
}
