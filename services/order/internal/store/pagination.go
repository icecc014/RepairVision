package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func normalizePage(page, size int64) (int64, int64) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return page, size
}

// sinceForDays 返回时间窗起点；days<=0 表示不限制。
func sinceForDays(days int64) any {
	if days <= 0 {
		return nil
	}
	return time.Now().AddDate(0, 0, -int(days)+1)
}

func CountAllOrdersFilter(ctx context.Context, conn sqlx.Session, status, buildingID, days int64) (int64, error) {
	query := "select count(*) from orders where 1 = 1"
	var args []any
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	if buildingID > 0 {
		query += " and building_id = ?"
		args = append(args, buildingID)
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var total int64
	if err := conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func ListAllOrdersPage(ctx context.Context, conn sqlx.Session, status, buildingID, page, size, days int64) ([]Order, error) {
	page, size = normalizePage(page, size)
	query := orderBase + "where 1 = 1"
	var args []any
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	if buildingID > 0 {
		query += " and building_id = ?"
		args = append(args, buildingID)
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	query += " order by id desc limit ? offset ?"
	args = append(args, size, (page-1)*size)
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

func CountOrdersByBuildingFilter(ctx context.Context, conn sqlx.Session, buildingID, status, days int64) (int64, error) {
	query := "select count(*) from orders where building_id = ?"
	args := []any{buildingID}
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var total int64
	if err := conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func ListOrdersByBuildingPage(ctx context.Context, conn sqlx.Session, buildingID, status, page, size, days int64) ([]Order, error) {
	page, size = normalizePage(page, size)
	query := orderBase + "where building_id = ?"
	args := []any{buildingID}
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	query += " order by id desc limit ? offset ?"
	args = append(args, size, (page-1)*size)
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

func CountOrdersByWorkerFilter(ctx context.Context, conn sqlx.Session, workerID, status, days int64) (int64, error) {
	query := "select count(*) from orders where worker_id = ?"
	args := []any{workerID}
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var total int64
	if err := conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func ListOrdersByWorkerPage(ctx context.Context, conn sqlx.Session, workerID, status, page, size, days int64) ([]Order, error) {
	page, size = normalizePage(page, size)
	query := orderBase + "where worker_id = ?"
	args := []any{workerID}
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	query += " order by id desc limit ? offset ?"
	args = append(args, size, (page-1)*size)
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}
