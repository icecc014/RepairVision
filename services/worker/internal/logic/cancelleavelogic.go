package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type CancelLeaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelLeaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelLeaveLogic {
	return &CancelLeaveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelLeaveLogic) CancelLeave(in *worker.LeaveIdRequest) (*worker.LeaveItem, error) {
	row, err := store.FindLeaveByID(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("请假申请不存在")
		}
		return nil, err
	}
	if in.WorkerId > 0 && row.WorkerID != in.WorkerId {
		return nil, errors.New("只能撤销本人的请假申请")
	}
	if row.Status != store.LeavePending {
		return nil, errors.New("仅待审批的申请可撤销")
	}
	if _, err := store.CancelLeave(l.ctx, l.svcCtx.DB, in.Id, row.WorkerID); err != nil {
		return nil, err
	}
	updated, err := store.FindLeaveByID(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		return nil, err
	}
	return leaveToPb(*updated), nil
}
