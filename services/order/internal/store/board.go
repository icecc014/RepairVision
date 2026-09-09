package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CountCompletedTodayByWorkers 返回各工人今天(按 completed_at)已完成工单数。
func CountCompletedTodayByWorkers(ctx context.Context, conn sqlx.Session, workerIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(workerIDs))
	if len(workerIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, len(workerIDs))
	args := make([]any, len(workerIDs))
	for i, id := range workerIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `select worker_id, count(*) cnt from orders
		where status = 4 and worker_id in (` + join(placeholders) + `)
		and date(completed_at) = date(?)
		group by worker_id`
	args = append(args, time.Now().Format("2006-01-02"))
	var rows []struct {
		WorkerID int64 `db:"worker_id"`
		Cnt      int64 `db:"cnt"`
	}
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.WorkerID] = row.Cnt
	}
	return result, nil
}
