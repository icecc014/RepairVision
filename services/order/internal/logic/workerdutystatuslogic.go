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

type WorkerDutyStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerDutyStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerDutyStatusLogic {
	return &WorkerDutyStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WorkerDutyStatus 工人端查询本人当前是否在岗（今日白班 ∧ 时段内 ∧ 未请假 ∧ 启用）。
func (l *WorkerDutyStatusLogic) WorkerDutyStatus() (resp *types.DutyStatusItem, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	r, err := l.svcCtx.WorkerRpc.GetDutyStatus(l.ctx, &workerclient.DutyStatusRequest{WorkerId: identity.UID})
	if err != nil {
		return nil, rpcBizError(err)
	}
	return dutyStatusToItem(identity.UID, "", r), nil
}

func dutyStatusToItem(workerID int64, name string, r *workerclient.DutyStatusResponse) *types.DutyStatusItem {
	item := &types.DutyStatusItem{WorkerId: workerID, Name: name}
	if r == nil {
		return item
	}
	item.OnDuty = r.OnDuty
	item.ShiftType = r.ShiftType
	item.InWorkPeriod = r.InWorkPeriod
	item.OnLeave = r.OnLeave
	item.Enabled = r.Enabled
	item.Reason = r.Reason
	item.Morning = r.Morning
	item.Afternoon = r.Afternoon
	return item
}