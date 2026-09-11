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

type AdminLeaveCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLeaveCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLeaveCreateLogic {
	return &AdminLeaveCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AdminLeaveCreate 管理员代工人登记请假：
// 先以该工人身份提交申请，再立即审批通过并标注「管理员代录」，通过时会同步把请假日期写成 OFF 班次。
func (l *AdminLeaveCreateLogic) AdminLeaveCreate(req *types.AdminLeaveCreateRequest) (*types.LeaveItem, error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if req.WorkerId <= 0 {
		return nil, errs.BadRequest("请选择要登记请假的工人")
	}
	created, err := l.svcCtx.WorkerRpc.SubmitLeave(l.ctx, &workerclient.SubmitLeaveRequest{
		WorkerId:  req.WorkerId,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	approved, err := l.svcCtx.WorkerRpc.ReviewLeave(l.ctx, &workerclient.ReviewLeaveRequest{
		Id:         created.Id,
		Status:     2,
		ReviewerId: identity.UID,
		ReviewNote: "管理员代录",
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	out := leavePbToType(approved)
	if users, err := l.svcCtx.WorkerRpc.GetUsers(l.ctx, &workerclient.UserIdsRequest{Ids: []int64{req.WorkerId}}); err == nil {
		for _, u := range users.Users {
			if u.Id == req.WorkerId {
				out.WorkerName = u.Name
			}
		}
	}
	notifyUsers(l.ctx, l.svcCtx, []int64{req.WorkerId}, "leave",
		"管理员已为你登记请假", out.StartDate+" 至 "+out.EndDate+"："+out.Reason, 0)
	return &out, nil
}
