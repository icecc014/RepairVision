package store

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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
