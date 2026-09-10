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

type AdminLeaveListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLeaveListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLeaveListLogic {
	return &AdminLeaveListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminLeaveListLogic) AdminLeaveList(req *types.AdminLeaveListRequest) (resp *types.LeaveListResponse, err error) {
	out, err := l.svcCtx.WorkerRpc.ListLeaves(l.ctx, &workerclient.ListLeavesRequest{
		WorkerId: req.WorkerId,
		Status:   req.Status,
		Page:     req.Page,
		Size:     req.Size,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	workerIDs := make([]int64, 0, len(out.Items))
	for _, item := range out.Items {
		workerIDs = append(workerIDs, item.WorkerId)
	}
	names := make(map[int64]string)
	if len(workerIDs) > 0 {
		users, err := l.svcCtx.WorkerRpc.GetUsers(l.ctx, &workerclient.UserIdsRequest{Ids: workerIDs})
		if err != nil {
			return nil, errs.Upstream()
		}
		for _, u := range users.Users {
			names[u.Id] = u.Name
		}
	}
	resp = &types.LeaveListResponse{Total: out.Total, List: make([]types.LeaveItem, 0, len(out.Items))}
	for _, item := range out.Items {
		converted := leavePbToType(item)
		converted.WorkerName = names[item.WorkerId]
		resp.List = append(resp.List, converted)
	}
	return resp, nil
}

type AdminLeaveReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLeaveReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLeaveReviewLogic {
	return &AdminLeaveReviewLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminLeaveReviewLogic) AdminLeaveReview(req *types.LeaveReviewRequest) (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	item, err := l.svcCtx.WorkerRpc.ReviewLeave(l.ctx, &workerclient.ReviewLeaveRequest{
		Id:         req.Id,
		Status:     req.Status,
		ReviewerId: identity.UID,
		ReviewNote: req.ReviewNote,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	result := "已通过"
	if item.Status == 3 {
		result = "已驳回"
	}
	notifyUsers(l.ctx, l.svcCtx, []int64{item.WorkerId}, "leave",
		"请假审批结果："+result, item.StartDate+" 至 "+item.EndDate, 0)
	return &types.EmptyResponse{}, nil
}
