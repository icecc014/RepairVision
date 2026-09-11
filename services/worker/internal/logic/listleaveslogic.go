package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type ListLeavesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLeavesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLeavesLogic {
	return &ListLeavesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListLeavesLogic) ListLeaves(in *worker.ListLeavesRequest) (*worker.ListLeavesResponse, error) {
	rows, total, err := store.ListLeaves(l.ctx, l.svcCtx.DB, in.WorkerId, in.Status, in.Page, in.Size, in.Days)
	if err != nil {
		return nil, err
	}
	resp := &worker.ListLeavesResponse{Total: total}
	for _, row := range rows {
		resp.Items = append(resp.Items, leaveToPb(row))
	}
	return resp, nil
}
