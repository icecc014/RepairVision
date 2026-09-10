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

type ReviewLeaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewLeaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewLeaveLogic {
	return &ReviewLeaveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReviewLeave 审批请假：通过时同步把请假日期写成 OFF 班次，确保派单不派给请假工人。
func (l *ReviewLeaveLogic) ReviewLeave(in *worker.ReviewLeaveRequest) (*worker.LeaveItem, error) {
	if in.Status != store.LeaveApproved && in.Status != store.LeaveRejected {
		return nil, errors.New("审批状态必须为通过或驳回")
	}
	row, err := store.FindLeaveByID(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("请假申请不存在")
		}
		return nil, err
	}
	if row.Status != store.LeavePending {
		return nil, errors.New("该申请已审批，无法重复操作")
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		ok, err := store.UpdateLeaveStatus(txCtx, session, in.Id, in.Status, in.ReviewerId, in.ReviewNote)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("审批失败，申请状态已变化")
		}
		if in.Status != store.LeaveApproved {
			return nil
		}
		for d := row.StartDate; !d.After(row.EndDate); d = d.AddDate(0, 0, 1) {
			if err := store.UpsertSchedule(txCtx, session, row.WorkerID,
				d.Format("2006-01-02"), "OFF", "请假已批准"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	updated, err := store.FindLeaveByID(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		return nil, err
	}
	return leaveToPb(*updated), nil
}
