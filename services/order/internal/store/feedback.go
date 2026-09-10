package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Feedback struct {
	OrderID    int64          `db:"order_id"`
	BuildingID int64          `db:"building_id"`
	WorkerID   sql.NullInt64  `db:"worker_id"`
	Rating     int64          `db:"rating"`
	Comment    sql.NullString `db:"comment"`
	CreatedAt  time.Time      `db:"created_at"`
}

type FeedbackRatingRow struct {
	Rating int64 `db:"rating"`
	Cnt    int64 `db:"cnt"`
}

type FeedbackRecentRow struct {
	OrderNo    string         `db:"order_no"`
	BuildingID int64          `db:"building_id"`
	Room       string         `db:"room"`
	WorkerID   sql.NullInt64  `db:"worker_id"`
	Rating     int64          `db:"rating"`
	Comment    sql.NullString `db:"comment"`
	CreatedAt  time.Time      `db:"created_at"`
}

func InsertFeedback(ctx context.Context, conn sqlx.Session, orderID, buildingID int64, workerID sql.NullInt64, rating int64, comment string) error {
	var commentArg any
	if comment != "" {
		commentArg = comment
	}
	_, err := conn.ExecCtx(ctx,
		`insert into order_feedbacks(order_id, building_id, worker_id, rating, comment) values(?,?,?,?,?)`,
		orderID, buildingID, nullableInt64(workerID), rating, commentArg)
	return err
}

func FindFeedbackByOrderID(ctx context.Context, conn sqlx.Session, orderID int64) (*Feedback, error) {
	var f Feedback
	if err := conn.QueryRowCtx(ctx, &f,
		`select order_id, building_id, worker_id, rating, comment, created_at
		 from order_feedbacks where order_id = ?`, orderID); err != nil {
		return nil, err
	}
	return &f, nil
}

func ListFeedbackByOrderIDs(ctx context.Context, conn sqlx.Session, orderIDs []int64) (map[int64]Feedback, error) {
	result := make(map[int64]Feedback)
	if len(orderIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, len(orderIDs))
	args := make([]any, len(orderIDs))
	for i, id := range orderIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	var rows []Feedback
	if err := conn.QueryRowsCtx(ctx, &rows,
		`select order_id, building_id, worker_id, rating, comment, created_at
		 from order_feedbacks where order_id in (`+join(placeholders)+`)`, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.OrderID] = row
	}
	return result, nil
}

func FeedbackStats(ctx context.Context, conn sqlx.Session) (int64, float64, []FeedbackRatingRow, error) {
	var total int64
	if err := conn.QueryRowCtx(ctx, &total, "select count(*) from order_feedbacks"); err != nil {
		return 0, 0, nil, err
	}
	var avg float64
	if err := conn.QueryRowCtx(ctx, &avg,
		"select coalesce(avg(rating), 0) from order_feedbacks"); err != nil {
		return 0, 0, nil, err
	}
	var rows []FeedbackRatingRow
	if err := conn.QueryRowsCtx(ctx, &rows,
		"select rating, count(*) cnt from order_feedbacks group by rating order by rating"); err != nil {
		return 0, 0, nil, err
	}
	return total, avg, rows, nil
}

func ListRecentFeedback(ctx context.Context, conn sqlx.Session, limit int64) ([]FeedbackRecentRow, error) {
	var rows []FeedbackRecentRow
	if err := conn.QueryRowsCtx(ctx, &rows,
		`select o.order_no, f.building_id, o.room, f.worker_id, f.rating, f.comment, f.created_at
		 from order_feedbacks f join orders o on o.id = f.order_id
		 order by f.id desc limit ?`, limit); err != nil {
		return nil, err
	}
	return rows, nil
}
