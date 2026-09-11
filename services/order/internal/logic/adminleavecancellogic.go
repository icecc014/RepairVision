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

type AdminLeaveCancelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLeaveCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLeaveCancelLogic {
	return &AdminLeaveCancelLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AdminLeaveCancel 管理员撤销请假（待审批或已通过都可撤销；已通过的会同时清掉对应的 OFF 班次）。
func (l *AdminLeaveCancelLogic) AdminLeaveCancel(req *types.LeaveIdRequest) (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	item, err := l.svcCtx.WorkerRpc.CancelLeave(l.ctx, &workerclient.LeaveIdRequest{Id: req.Id, WorkerId: 0})
	if err != nil {
		return nil, rpcBizError(err)
	}
	notifyUsers(l.ctx, l.svcCtx, []int64{item.WorkerId}, "leave",
		"请假已被管理员撤销", item.StartDate+" 至 "+item.EndDate, 0)
	return &types.EmptyResponse{}, nil
}
