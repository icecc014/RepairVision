package store

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// RepairRecordFilter 面向业务的报修记录筛选条件。
type RepairRecordFilter struct {
	BuildingID int64
	Status     int64
	FaultType  string
	Keyword    string
	Days       int64
	StartDate  string
	EndDate    string
}

const repairRecordColumns = `o.id, o.order_no, o.title, o.description, o.building_id, o.room, o.floor,
	o.fault_type, o.priority, o.expect_minutes, o.status, o.is_merged, o.main_order_id, o.worker_id,
	o.dispatched_at, o.started_at, o.completed_at, o.reporter_id, o.source, o.created_at, o.updated_at`

// repairRecordWhere 生成筛选条件；withStatus=false 时忽略状态条件（用于按状态汇总）。
func repairRecordWhere(f RepairRecordFilter, withStatus bool) (string, []any) {
	where := []string{"1 = 1"}
	var args []any
	if f.BuildingID > 0 {
		where = append(where, "o.building_id = ?")
		args = append(args, f.BuildingID)
	}
	if withStatus && f.Status > 0 {
		where = append(where, "o.status = ?")
		args = append(args, f.Status)
	}
	if fault := strings.TrimSpace(f.FaultType); fault != "" {
		where = append(where, "o.fault_type = ?")
		args = append(args, fault)
	}
	if f.StartDate != "" {
		where = append(where, "date(o.created_at) >= ?")
		args = append(args, f.StartDate)
	}
	if f.EndDate != "" {
		where = append(where, "date(o.created_at) <= ?")
		args = append(args, f.EndDate)
	}
	if f.StartDate == "" && f.EndDate == "" {
		if since := sinceForDays(f.Days); since != nil {
			where = append(where, "o.created_at >= ?")
			args = append(args, since)
		}
	}
	if keyword := strings.TrimSpace(f.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, `(o.room like ? or o.order_no like ? or o.description like ?
			or exists (select 1 from worker_db.users ru where ru.id = o.reporter_id and ru.name like ?)
			or exists (select 1 from worker_db.users wu where wu.id = o.worker_id and wu.name like ?))`)
		args = append(args, like, like, like, like, like)
	}
	return strings.Join(where, " and "), args
}

func CountRepairRecords(ctx context.Context, conn sqlx.Session, f RepairRecordFilter, withStatus bool) (int64, error) {
	condition, args := repairRecordWhere(f, withStatus)
	var row LogCount
	if err := conn.QueryRowCtx(ctx, &row,
		"select count(*) cnt from orders o where "+condition, args...); err != nil {
		return 0, err
	}
	return row.Cnt, nil
}

func ListRepairRecords(ctx context.Context, conn sqlx.Session, f RepairRecordFilter, page, size int64) ([]Order, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 1000 {
		size = 1000
	}
	condition, args := repairRecordWhere(f, true)
	query := "select " + repairRecordColumns + " from orders o where " + condition +
		" order by o.id desc limit ? offset ?"
	args = append(args, size, (page-1)*size)
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

// SummarizeRepairRecords 在当前筛选条件下（忽略状态条件）按状态分组计数。
func SummarizeRepairRecords(ctx context.Context, conn sqlx.Session, f RepairRecordFilter) (map[int64]int64, error) {
	condition, args := repairRecordWhere(f, false)
	var rows []struct {
		Status int64 `db:"status"`
		Cnt    int64 `db:"cnt"`
	}
	if err := conn.QueryRowsCtx(ctx, &rows,
		"select o.status status, count(*) cnt from orders o where "+condition+" group by o.status",
		args...); err != nil {
		return nil, err
	}
	result := make(map[int64]int64, len(rows))
	for _, row := range rows {
		result[row.Status] = row.Cnt
	}
	return result, nil
}
