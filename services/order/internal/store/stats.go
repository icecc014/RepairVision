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
