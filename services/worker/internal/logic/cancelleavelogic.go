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
	// WorkerId<=0 表示管理员代为撤销（可撤销待审批与已通过），否则只能撤销本人待审批的申请
	admin := in.WorkerId <= 0
	if !admin && row.WorkerID != in.WorkerId {
		return nil, errors.New("只能撤销本人的请假申请")
	}
	if admin {
		if row.Status != store.LeavePending && row.Status != store.LeaveApproved {
			return nil, errors.New("该申请已撤销或已驳回")
		}
	} else if row.Status != store.LeavePending {
		return nil, errors.New("仅待审批的申请可撤销")
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if admin {
			ok, err := store.CancelLeaveByAdmin(txCtx, session, in.Id)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("撤销失败，申请状态已变化")
			}
			if row.Status == store.LeaveApproved {
				// 已通过的请假此前写入了 OFF 班次，撤销时一并清理
				return store.ClearWorkerScheduleRange(txCtx, session, row.WorkerID,
					row.StartDate.Format("2006-01-02"), row.EndDate.Format("2006-01-02"))
			}
			return nil
		}
		_, err := store.CancelLeave(txCtx, session, in.Id, row.WorkerID)
		return err
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
