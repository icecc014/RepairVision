package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	StatusPending    = int64(1)
	StatusDispatched = int64(2)
	StatusWorking    = int64(3)
	StatusCompleted  = int64(4)
	StatusCanceled   = int64(5)
)

type Order struct {
	ID          int64         `db:"id"`
	OrderNo     string        `db:"order_no"`
	Title       string        `db:"title"`
	Description string        `db:"description"`
	BuildingID  int64         `db:"building_id"`
	Room        string        `db:"room"`
	Floor       int64         `db:"floor"`
	FaultType   string        `db:"fault_type"`
	Status      int64         `db:"status"`
	IsMerged    int64         `db:"is_merged"`
	MainOrderID sql.NullInt64 `db:"main_order_id"`
	WorkerID    sql.NullInt64 `db:"worker_id"`
	ReporterID  int64         `db:"reporter_id"`
	Source      string        `db:"source"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"`
}

type WorkerLoad struct {
	WorkerID int64 `db:"worker_id"`
	Cnt      int64 `db:"cnt"`
}

const orderColumns = `id, order_no, title, description, building_id, room, floor, fault_type,
	status, is_merged, main_order_id, worker_id, reporter_id, source, created_at, updated_at`

const orderBase = "select " + orderColumns + " from orders "

func FindOrder(ctx context.Context, conn sqlx.Session, id int64) (*Order, error) {
	var o Order
	if err := conn.QueryRowCtx(ctx, &o, orderBase+"where id = ?", id); err != nil {
		return nil, err
	}
	return &o, nil
}

func ListOrdersByBuilding(ctx context.Context, conn sqlx.Session, buildingID, status int64) ([]Order, error) {
	query := orderBase + "where building_id = ?"
	args := []any{buildingID}
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	query += " order by id desc"
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

func ListOrdersByBuildingIDs(ctx context.Context, conn sqlx.Session, buildingIDs []int64, activeOnly bool) ([]Order, error) {
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
	query += " order by id desc"
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}
func ListOrdersByWorker(ctx context.Context, conn sqlx.Session, workerID, status int64) ([]Order, error) {
	query := orderBase + "where worker_id = ?"
	args := []any{workerID}
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	query += " order by id desc"
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

func ListAllOrders(ctx context.Context, conn sqlx.Session, status, buildingID int64) ([]Order, error) {
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
	query += " order by id desc"
	var orders []Order
	if err := conn.QueryRowsCtx(ctx, &orders, query, args...); err != nil {
		return nil, err
	}
	return orders, nil
}

func InsertOrder(ctx context.Context, conn sqlx.Session, o *Order) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		`insert into orders(order_no, title, description, building_id, room, floor, fault_type,
			status, is_merged, main_order_id, worker_id, reporter_id, source)
		 values(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		o.OrderNo, o.Title, o.Description, o.BuildingID, o.Room, o.Floor, o.FaultType,
		o.Status, o.IsMerged, nullableInt64(o.MainOrderID), nullableInt64(o.WorkerID), o.ReporterID, o.Source)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func AssignOrder(ctx context.Context, conn sqlx.Session, orderID, workerID int64) error {
	_, err := conn.ExecCtx(ctx,
		"update orders set status = ?, worker_id = ? where id = ? and status = ?",
		StatusDispatched, workerID, orderID, StatusPending)
	return err
}

func InsertDispatchRecord(ctx context.Context, conn sqlx.Session, orderID, workerID int64, score, skillScore, distanceScore, loadScore float64) error {
	_, err := conn.ExecCtx(ctx,
		`insert into dispatch_records(order_id, worker_id, score, skill_score, distance_score, load_score, status)
		 values(?,?,?,?,?,?,1)`, orderID, workerID, score, skillScore, distanceScore, loadScore)
	return err
}

type DispatchSummary struct {
	OrderID       int64   `db:"order_id"`
	WorkerID      int64   `db:"worker_id"`
	Score         float64 `db:"score"`
	SkillScore    float64 `db:"skill_score"`
	DistanceScore float64 `db:"distance_score"`
	LoadScore     float64 `db:"load_score"`
}

func ListDispatchSummaries(ctx context.Context, conn sqlx.Session, orderIDs []int64) (map[int64]DispatchSummary, error) {
	result := make(map[int64]DispatchSummary)
	if len(orderIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, len(orderIDs))
	args := make([]any, len(orderIDs))
	for i, id := range orderIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `select d.order_id, d.worker_id,
		cast(d.score as decimal(10,4)) score,
		cast(coalesce(d.skill_score,0) as decimal(10,4)) skill_score,
		cast(coalesce(d.distance_score,0) as decimal(10,4)) distance_score,
		cast(coalesce(d.load_score,0) as decimal(10,4)) load_score
		from dispatch_records d
		join (select order_id, max(id) mid from dispatch_records where order_id in (` + join(placeholders) + `) group by order_id) m
		on d.id = m.mid`
	var rows []DispatchSummary
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.OrderID] = row
	}
	return result, nil
}
func FindRecentDuplicateOrder(ctx context.Context, conn sqlx.Session, buildingID, floor int64, room, faultType string, since time.Time) (*Order, error) {
	var o Order
	if err := conn.QueryRowCtx(ctx, &o,
		orderBase+`where building_id = ? and floor = ? and room = ? and fault_type = ?
			and status in (1,2,3) and created_at >= ? order by id limit 1`,
		buildingID, floor, room, faultType, since); err != nil {
		return nil, err
	}
	return &o, nil
}

func CountInProgressByWorkers(ctx context.Context, conn sqlx.Session, workerIDs []int64) (map[int64]int64, error) {
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
	query := `select worker_id, count(*) cnt from orders
		where worker_id in (` + join(placeholders) + `) and status in (2,3) group by worker_id`
	var rows []WorkerLoad
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.WorkerID] = row.Cnt
	}
	return result, nil
}

func CancelOrder(ctx context.Context, conn sqlx.Session, orderID, buildingID int64) (bool, error) {
	query := "update orders set status = ?, worker_id = null where id = ? and status in (?,?)"
	args := []any{StatusCanceled, orderID, StatusPending, StatusDispatched}
	if buildingID > 0 {
		query += " and building_id = ?"
		args = append(args, buildingID)
	}
	result, err := conn.ExecCtx(ctx, query, args...)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func StartOrder(ctx context.Context, conn sqlx.Session, orderID, workerID int64) (bool, error) {
	result, err := conn.ExecCtx(ctx,
		"update orders set status = ? where id = ? and worker_id = ? and status = ?",
		StatusWorking, orderID, workerID, StatusDispatched)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

type batchOrderRow struct {
	ID int64 `db:"id"`
}

func BatchCompleteOrders(ctx context.Context, conn sqlx.Session, workerID, buildingID int64, faultType string) ([]int64, error) {
	var rows []batchOrderRow
	if err := conn.QueryRowsCtx(ctx, &rows,
		`select id from orders where worker_id = ? and building_id = ? and fault_type = ? and status in (2,3)`,
		workerID, buildingID, faultType); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	if len(ids) == 0 {
		return ids, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	_, err := conn.ExecCtx(ctx,
		"update orders set status = ? where id in ("+join(placeholders)+") and worker_id = ?",
		append([]any{StatusCompleted}, append(args, workerID)...)...)
	if err != nil {
		return nil, err
	}
	return ids, nil
}
func CompleteOrder(ctx context.Context, conn sqlx.Session, orderID, workerID int64) (bool, error) {
	result, err := conn.ExecCtx(ctx,
		"update orders set status = ? where id = ? and worker_id = ? and status = ?",
		StatusCompleted, orderID, workerID, StatusWorking)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func nullableInt64(n sql.NullInt64) any {
	if !n.Valid || n.Int64 == 0 {
		return nil
	}
	return n.Int64
}

func join(items []string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += ","
		}
		out += item
	}
	return out
}
