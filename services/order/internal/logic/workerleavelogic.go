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

type WorkerLeaveSubmitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerLeaveSubmitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerLeaveSubmitLogic {
	return &WorkerLeaveSubmitLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *WorkerLeaveSubmitLogic) WorkerLeaveSubmit(req *types.WorkerLeaveSubmitRequest) (resp *types.LeaveItem, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	item, err := l.svcCtx.WorkerRpc.SubmitLeave(l.ctx, &workerclient.SubmitLeaveRequest{
		WorkerId:  identity.UID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	out := leavePbToType(item)
	notifyUsers(l.ctx, l.svcCtx, adminUserIDs(l.ctx, l.svcCtx), "leave",
		"新的请假申请", out.StartDate+" 至 "+out.EndDate+"："+out.Reason, 0)
	return &out, nil
}

type WorkerLeaveListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerLeaveListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerLeaveListLogic {
	return &WorkerLeaveListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *WorkerLeaveListLogic) WorkerLeaveList(req *types.WorkerLeaveListRequest) (resp *types.LeaveListResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	out, err := l.svcCtx.WorkerRpc.ListLeaves(l.ctx, &workerclient.ListLeavesRequest{
		WorkerId: identity.UID,
		Status:   req.Status,
		Page:     req.Page,
		Size:     req.Size,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	resp = &types.LeaveListResponse{Total: out.Total, List: make([]types.LeaveItem, 0, len(out.Items))}
	for _, item := range out.Items {
		resp.List = append(resp.List, leavePbToType(item))
	}
	return resp, nil
}

type WorkerLeaveCancelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerLeaveCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerLeaveCancelLogic {
	return &WorkerLeaveCancelLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *WorkerLeaveCancelLogic) WorkerLeaveCancel(req *types.LeaveIdRequest) (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if _, err := l.svcCtx.WorkerRpc.CancelLeave(l.ctx, &workerclient.LeaveIdRequest{
		Id: req.Id, WorkerId: identity.UID,
	}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
