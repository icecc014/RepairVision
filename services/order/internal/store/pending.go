package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ListPendingOrders 返回待派单队列，limit<=0 时不限制。
func ListPendingOrders(ctx context.Context, conn sqlx.Session, limit int64) ([]Order, error) {
	query := orderBase + "where status = ? order by priority desc, id asc"
	args := []any{StatusPending}
	if limit > 0 {
		query += " limit ?"
		args = append(args, limit)
	}
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

// ListPendingOrdersByBuilding 返回指定楼栋的待派队列。
func ListPendingOrdersByBuilding(ctx context.Context, conn sqlx.Session, buildingID, limit int64) ([]Order, error) {
	query := orderBase + "where building_id = ? and status = ? order by priority desc, id asc"
	args := []any{buildingID, StatusPending}
	if limit > 0 {
		query += " limit ?"
		args = append(args, limit)
	}
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

// TryAssignOrder 仅在工单仍为待派时完成指派，返回是否成功（防止并发重复派单）。
func TryAssignOrder(ctx context.Context, conn sqlx.Session, orderID, workerID int64) (bool, error) {
	result, err := conn.ExecCtx(ctx,
		"update orders set status = ?, worker_id = ?, dispatched_at = ? where id = ? and status = ?",
		StatusDispatched, workerID, time.Now(), orderID, StatusPending)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}
