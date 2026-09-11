package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ListOrdersByBuildingIDsWithinDays 按楼栋集合查询工单，并可限定报修时间窗。
// activeOnly=true 只返回未完成工单（待派单/已派单/维修中）；days<=0 表示不限制时间。
func ListOrdersByBuildingIDsWithinDays(ctx context.Context, conn sqlx.Session, buildingIDs []int64, activeOnly bool, days int64) ([]Order, error) {
	if len(buildingIDs) == 0 {
		return []Order{}, nil
	}
	placeholders := make([]string, len(buildingIDs))
	args := make([]any, len(buildingIDs))
	for i, id := range buildingIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := orderBase + "where building_id in (" + join(placeholders) + ")"
	if activeOnly {
		query += " and status in (1,2,3)"
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	query += " order by id desc"
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

// CountOrdersByBuildingIDsWithinDays 统计时间窗内的工单数（用于界面提示）。
func CountOrdersByBuildingIDsWithinDays(ctx context.Context, conn sqlx.Session, buildingIDs []int64, activeOnly bool, days int64) (int64, error) {
	if len(buildingIDs) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(buildingIDs))
	args := make([]any, len(buildingIDs))
	for i, id := range buildingIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := "select count(*) cnt from orders where building_id in (" + join(placeholders) + ")"
	if activeOnly {
		query += " and status in (1,2,3)"
	}
	if since := sinceForDays(days); since != nil {
		query += " and created_at >= ?"
		args = append(args, since)
	}
	var row LogCount
	if err := conn.QueryRowCtx(ctx, &row, query, args...); err != nil {
		return 0, err
	}
	return row.Cnt, nil
}
