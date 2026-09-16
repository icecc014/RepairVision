package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type StatusCount struct {
	Status int64 `db:"status"`
	Cnt    int64 `db:"cnt"`
}

type BuildingCount struct {
	BuildingID int64 `db:"building_id"`
	Cnt        int64 `db:"cnt"`
}

type FaultCount struct {
	FaultType string `db:"fault_type"`
	Cnt       int64  `db:"cnt"`
}

type DayCount struct {
	Day string `db:"day"`
	Cnt int64  `db:"cnt"`
}

func CountOrdersByStatus(ctx context.Context, conn sqlx.Session) ([]StatusCount, error) {
	var rows []StatusCount
	if err := conn.QueryRowsCtx(ctx, &rows,
		"select status, count(*) cnt from orders group by status order by status"); err != nil {
		return nil, err
	}
	return rows, nil
}

func CountOrdersByBuilding(ctx context.Context, conn sqlx.Session) ([]BuildingCount, error) {
	var rows []BuildingCount
	if err := conn.QueryRowsCtx(ctx, &rows,
		"select building_id, count(*) cnt from orders group by building_id order by building_id"); err != nil {
		return nil, err
	}
	return rows, nil
}

func CountOrdersByFaultType(ctx context.Context, conn sqlx.Session) ([]FaultCount, error) {
	var rows []FaultCount
	if err := conn.QueryRowsCtx(ctx, &rows,
		"select fault_type, count(*) cnt from orders group by fault_type order by cnt desc"); err != nil {
		return nil, err
	}
	return rows, nil
}

func CountOrdersLastDays(ctx context.Context, conn sqlx.Session, days int) ([]DayCount, error) {
	since := time.Now().AddDate(0, 0, -days+1)
	var rows []DayCount
	if err := conn.QueryRowsCtx(ctx, &rows,
		`select date_format(created_at, '%Y-%m-%d') day, count(*) cnt
		 from orders where created_at >= ? group by day order by day`, since); err != nil {
		return nil, err
	}
	return rows, nil
}

// ---- V6：支持时间窗口（days>0 时仅统计近 N 天）与历史累计 ----

func CountOrdersByStatusWithinDays(ctx context.Context, conn sqlx.Session, days int64) ([]StatusCount, error) {
	query := "select status, count(*) cnt from orders"
	var args []any
	if days > 0 {
		query += " where created_at >= date_sub(now(), interval ? day)"
		args = append(args, days)
	}
	query += " group by status order by status"
	var rows []StatusCount
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func CountOrdersByBuildingWithinDays(ctx context.Context, conn sqlx.Session, days int64) ([]BuildingCount, error) {
	query := "select building_id, count(*) cnt from orders"
	var args []any
	if days > 0 {
		query += " where created_at >= date_sub(now(), interval ? day)"
		args = append(args, days)
	}
	query += " group by building_id order by building_id"
	var rows []BuildingCount
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func CountOrdersByFaultTypeWithinDays(ctx context.Context, conn sqlx.Session, days int64) ([]FaultCount, error) {
	query := "select fault_type, count(*) cnt from orders"
	var args []any
	if days > 0 {
		query += " where created_at >= date_sub(now(), interval ? day)"
		args = append(args, days)
	}
	query += " group by fault_type order by cnt desc"
	var rows []FaultCount
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

type ordersTotalsRow struct {
	Total int64 `db:"total"`
	Done  int64 `db:"done"`
}

// OrdersTotals 历史累计：全部工单数 + 已完成工单数（不受时间筛选影响）。
func OrdersTotals(ctx context.Context, conn sqlx.Session) (int64, int64, error) {
	var row ordersTotalsRow
	if err := conn.QueryRowCtx(ctx, &row,
		"select count(*) total, sum(case when status = 4 then 1 else 0 end) done from orders"); err != nil {
		return 0, 0, err
	}
	return row.Total, row.Done, nil
}
