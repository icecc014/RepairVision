package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func CountPendingOverdue(ctx context.Context, conn sqlx.Session, hours int64) (int64, error) {
	var cnt int64
	if err := conn.QueryRowCtx(ctx, &cnt,
		`select count(*) from orders
		 where status = 1 and timestampdiff(minute, created_at, now()) > ?`,
		hours*60); err != nil {
		return 0, err
	}
	return cnt, nil
}

func CountDispatchedOverdue(ctx context.Context, conn sqlx.Session, hours int64) (int64, error) {
	var cnt int64
	if err := conn.QueryRowCtx(ctx, &cnt,
		`select count(*) from orders
		 where status = 2 and dispatched_at is not null
		   and timestampdiff(minute, dispatched_at, now()) > ?`,
		hours*60); err != nil {
		return 0, err
	}
	return cnt, nil
}

func AvgDispatchMinutes(ctx context.Context, conn sqlx.Session) (float64, error) {
	var avg float64
	if err := conn.QueryRowCtx(ctx, &avg,
		`select coalesce(avg(timestampdiff(minute, created_at, dispatched_at)), 0)
		 from orders where dispatched_at is not null and status in (2,3,4)`); err != nil {
		return 0, err
	}
	return avg, nil
}

func AvgRepairMinutes(ctx context.Context, conn sqlx.Session) (float64, error) {
	var avg float64
	if err := conn.QueryRowCtx(ctx, &avg,
		`select coalesce(avg(timestampdiff(minute, started_at, completed_at)), 0)
		 from orders where status = 4 and completed_at is not null and started_at is not null`); err != nil {
		return 0, err
	}
	return avg, nil
}

// ListSlaOverdueOrders 返回待派/已派未开工两类超时工单（按时间升序，最多 20 条）。
func ListSlaOverdueOrders(ctx context.Context, conn sqlx.Session, pendingHours, dispatchedHours int64) ([]Order, error) {
	query := orderBase + `where (status = 1 and timestampdiff(minute, created_at, now()) > ?)
		or (status = 2 and dispatched_at is not null and timestampdiff(minute, dispatched_at, now()) > ?)
		order by created_at asc limit 20`
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, pendingHours*60, dispatchedHours*60); err != nil {
		return nil, err
	}
	return orders, nil
}
