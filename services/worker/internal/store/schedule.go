package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ScheduleRow 对应 worker_schedules 单日班次。
type ScheduleRow struct {
	WorkerID  int64          `db:"worker_id"`
	WorkDate  time.Time      `db:"work_date"`
	ShiftType string         `db:"shift_type"`
	Note      sql.NullString `db:"note"`
}

// FindWorkerShift 返回某工人在指定日期的班次；未排班视为无约束（空字符串）。
func FindWorkerShift(ctx context.Context, conn sqlx.SqlConn, workerID int64, workDate string) (string, error) {
	var shiftType string
	if err := conn.QueryRowCtx(ctx, &shiftType,
		"select shift_type from worker_schedules where worker_id = ? and work_date = ? limit 1",
		workerID, workDate); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return shiftType, nil
}

// ListSchedules 查询排班；workerID<=0 时查全部工人，日期区间按 YYYY-MM-DD。
func ListSchedules(ctx context.Context, conn sqlx.SqlConn, workerID int64, startDate, endDate string) ([]ScheduleRow, error) {
	query := "select worker_id, work_date, shift_type, note from worker_schedules where work_date between ? and ?"
	args := []any{startDate, endDate}
	if workerID > 0 {
		query += " and worker_id = ?"
		args = append(args, workerID)
	}
	query += " order by worker_id, work_date"
	var rows []ScheduleRow
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// UpsertSchedule 单日班次 upsert（同一工人同一天仅一条）。
func UpsertSchedule(ctx context.Context, conn sqlx.Session, workerID int64, workDate, shiftType, note string) error {
	var noteArg any
	if note != "" {
		noteArg = note
	}
	_, err := conn.ExecCtx(ctx,
		`insert into worker_schedules(worker_id, work_date, shift_type, note) values(?,?,?,?)
		 on duplicate key update shift_type = values(shift_type), note = values(note)`,
		workerID, workDate, shiftType, noteArg)
	return err
}

// ClearScheduleRange 删除日期区间内全部排班（重新生成前调用）。
func ClearScheduleRange(ctx context.Context, conn sqlx.Session, startDate, endDate string) error {
	_, err := conn.ExecCtx(ctx,
		"delete from worker_schedules where work_date between ? and ?", startDate, endDate)
	return err
}

// ListWorkerIDs 返回全部启用工人的 ID（用于一键生成周排班）。
func ListWorkerIDs(ctx context.Context, conn sqlx.SqlConn) ([]int64, error) {
	var ids []int64
	if err := conn.QueryRowsCtx(ctx, &ids,
		"select id from users where role = 2 and status = 1 order by id"); err != nil {
		return nil, err
	}
	return ids, nil
}
