package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// WorkerActiveOrder 工人在途工单（已派单/维修中），用于看板展示"当前所在楼栋"。
type WorkerActiveOrder struct {
	WorkerID   int64  `db:"worker_id"`
	OrderID    int64  `db:"id"`
	OrderNo    string `db:"order_no"`
	BuildingID int64  `db:"building_id"`
	Status     int64  `db:"status"`
}

// ListActiveOrdersByWorkers 查询指定工人的在途工单，按"维修中优先、派单时间倒序"返回。
func ListActiveOrdersByWorkers(ctx context.Context, conn sqlx.Session, workerIDs []int64) ([]WorkerActiveOrder, error) {
	if len(workerIDs) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(workerIDs))
	args := make([]any, 0, len(workerIDs)+1)
	for i, id := range workerIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, StatusDispatched, StatusWorking)
	var rows []WorkerActiveOrder
	query := fmt.Sprintf(
		"select worker_id, id, order_no, building_id, status from orders where worker_id in (%s) and status in (?, ?) order by (status = ?) desc, dispatched_at desc, id desc",
		strings.Join(placeholders, ","))
	args = append(args, StatusWorking)
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}