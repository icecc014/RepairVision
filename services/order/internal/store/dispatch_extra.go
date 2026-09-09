package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CountWorkloadByWorkers 返回每个工人在途单的预计工时总量（分钟），
// 未填 expect_minutes 时按 30 分钟兜底，用于派单负载维度打分。
func CountWorkloadByWorkers(ctx context.Context, conn sqlx.Session, workerIDs []int64) (map[int64]int64, error) {
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
	query := `select worker_id, sum(coalesce(expect_minutes, 30)) minutes from orders
		where worker_id in (` + join(placeholders) + `) and status in (2,3) group by worker_id`
	var rows []struct {
		WorkerID int64 `db:"worker_id"`
		Minutes  int64 `db:"minutes"`
	}
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.WorkerID] = row.Minutes
	}
	return result, nil
}

// ListPendingOrdersByIDs 查询处于待派/指定状态的工单（批量派单入口）。
func ListPendingOrdersByIDs(ctx context.Context, conn sqlx.Session, ids []int64, status int64) ([]Order, error) {
	result := make([]Order, 0, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := orderBase + "where id in (" + join(placeholders) + ")"
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	query += " order by id"
	if err := conn.QueryRowsCtx(ctx, &result, query, args...); err != nil {
		return nil, err
	}
	return result, nil
}

// AcceptDispatchRecord 工人开工时把对应派单记录标记为“已接受”。
func AcceptDispatchRecord(ctx context.Context, conn sqlx.Session, orderID, workerID int64) error {
	_, err := conn.ExecCtx(ctx,
		"update dispatch_records set status = 2 where order_id = ? and worker_id = ? and status = 1",
		orderID, workerID)
	return err
}

// RevokeActiveDispatch 改派时撤销原工人的未接受派单记录。
func RevokeActiveDispatch(ctx context.Context, conn sqlx.Session, orderID, workerID int64) error {
	_, err := conn.ExecCtx(ctx,
		"update dispatch_records set status = 3 where order_id = ? and worker_id = ? and status in (1,2)",
		orderID, workerID)
	return err
}

// ReassignOrderWorker 更新订单为已派并绑定新工人（仅待派/已派可改派）。
func ReassignOrderWorker(ctx context.Context, conn sqlx.Session, orderID, workerID int64) (bool, error) {
	result, err := conn.ExecCtx(ctx,
		"update orders set status = ?, worker_id = ?, dispatched_at = ? where id = ? and status in (?,?)",
		StatusDispatched, workerID, time.Now(), orderID, StatusPending, StatusDispatched)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}
