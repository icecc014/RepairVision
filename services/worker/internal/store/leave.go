package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	LeavePending  = int64(1)
	LeaveApproved = int64(2)
	LeaveRejected = int64(3)
	LeaveCanceled = int64(4)
)

type LeaveRequest struct {
	ID         int64          `db:"id"`
	WorkerID   int64          `db:"worker_id"`
	StartDate  time.Time      `db:"start_date"`
	EndDate    time.Time      `db:"end_date"`
	Reason     string         `db:"reason"`
	Status     int64          `db:"status"`
	ReviewerID sql.NullInt64  `db:"reviewer_id"`
	ReviewNote sql.NullString `db:"review_note"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

const leaveColumns = `id, worker_id, start_date, end_date, reason, status, reviewer_id, review_note, created_at, updated_at`

func InsertLeave(ctx context.Context, conn sqlx.Session, workerID int64, startDate, endDate, reason string) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into worker_leave_requests(worker_id, start_date, end_date, reason, status) values(?,?,?,?,1)",
		workerID, startDate, endDate, reason)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func FindLeaveByID(ctx context.Context, conn sqlx.Session, id int64) (*LeaveRequest, error) {
	var row LeaveRequest
	if err := conn.QueryRowCtx(ctx, &row,
		"select "+leaveColumns+" from worker_leave_requests where id = ?", id); err != nil {
		return nil, err
	}
	return &row, nil
}

func ListLeaves(ctx context.Context, conn sqlx.Session, workerID, status, page, size int64) ([]LeaveRequest, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	where := "where 1 = 1"
	args := []any{}
	if workerID > 0 {
		where += " and worker_id = ?"
		args = append(args, workerID)
	}
	if status > 0 {
		where += " and status = ?"
		args = append(args, status)
	}
	var total int64
	if err := conn.QueryRowCtx(ctx, &total, "select count(*) from worker_leave_requests "+where, args...); err != nil {
		return nil, 0, err
	}
	query := "select " + leaveColumns + " from worker_leave_requests " + where + " order by id desc limit ? offset ?"
	args = append(args, size, (page-1)*size)
	var rows []LeaveRequest
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func HasOverlappingLeave(ctx context.Context, conn sqlx.Session, workerID int64, startDate, endDate string) (bool, error) {
	var cnt int64
	if err := conn.QueryRowCtx(ctx, &cnt,
		`select count(*) from worker_leave_requests
		 where worker_id = ? and status in (1,2)
		   and not (end_date < ? or start_date > ?)`,
		workerID, startDate, endDate); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func UpdateLeaveStatus(ctx context.Context, conn sqlx.Session, id, status, reviewerID int64, note string) (bool, error) {
	var noteArg any
	if note != "" {
		noteArg = note
	}
	result, err := conn.ExecCtx(ctx,
		"update worker_leave_requests set status = ?, reviewer_id = ?, review_note = ? where id = ? and status = 1",
		status, reviewerID, noteArg, id)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func CancelLeave(ctx context.Context, conn sqlx.Session, id, workerID int64) (bool, error) {
	result, err := conn.ExecCtx(ctx,
		"update worker_leave_requests set status = 4 where id = ? and worker_id = ? and status = 1",
		id, workerID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

// FindApprovedLeaveForDate 判断某工人在指定日期是否处于已批准请假。
func FindApprovedLeaveForDate(ctx context.Context, conn sqlx.SqlConn, workerID int64, workDate string) (bool, error) {
	var cnt int64
	if err := conn.QueryRowCtx(ctx, &cnt,
		`select count(*) from worker_leave_requests
		 where worker_id = ? and status = 2 and ? between start_date and end_date`,
		workerID, workDate); err != nil {
		return false, err
	}
	return cnt > 0, nil
}
