package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func slaSince(days int64) any {
	if days <= 0 {
		return nil
	}
	return time.Now().AddDate(0, 0, -int(days)+1)
}

func CountPendingOverdue(ctx context.Context, conn sqlx.Session, hours, days int64) (int64, error) {
	query := `select count(*) from orders
		 where status = 1 and timestampdiff(minute, created_at, now()) > ?`
	args := []any{hours * 60}
	if since := slaSince(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var cnt int64
	if err := conn.QueryRowCtx(ctx, &cnt, query, args...); err != nil {
		return 0, err
	}
	return cnt, nil
}

func CountDispatchedOverdue(ctx context.Context, conn sqlx.Session, hours, days int64) (int64, error) {
	query := `select count(*) from orders
		 where status = 2 and dispatched_at is not null
		   and timestampdiff(minute, dispatched_at, now()) > ?`
	args := []any{hours * 60}
	if since := slaSince(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var cnt int64
	if err := conn.QueryRowCtx(ctx, &cnt, query, args...); err != nil {
		return 0, err
	}
	return cnt, nil
}

func AvgDispatchMinutes(ctx context.Context, conn sqlx.Session, days int64) (float64, error) {
	query := `select coalesce(avg(timestampdiff(minute, created_at, dispatched_at)), 0)
		 from orders where dispatched_at is not null and status in (2,3,4)`
	var args []any
	if since := slaSince(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var avg float64
	if err := conn.QueryRowCtx(ctx, &avg, query, args...); err != nil {
		return 0, err
	}
	return avg, nil
}

func AvgRepairMinutes(ctx context.Context, conn sqlx.Session, days int64) (float64, error) {
	query := `select coalesce(avg(timestampdiff(minute, started_at, completed_at)), 0)
		 from orders where status = 4 and completed_at is not null and started_at is not null`
	var args []any
	if since := slaSince(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var avg float64
	if err := conn.QueryRowCtx(ctx, &avg, query, args...); err != nil {
		return 0, err
	}
	return avg, nil
}

// ListSlaOverdueOrders 返回待派/已派未开工两类超时工单（按时间升序，最多 20 条）。
func ListSlaOverdueOrders(ctx context.Context, conn sqlx.Session, pendingHours, dispatchedHours, days int64) ([]Order, error) {
	query := orderBase + `where ((status = 1 and timestampdiff(minute, created_at, now()) > ?)
		or (status = 2 and dispatched_at is not null and timestampdiff(minute, dispatched_at, now()) > ?))`
	args := []any{pendingHours * 60, dispatchedHours * 60}
	if since := slaSince(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	query += " order by created_at asc limit 20"
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}
