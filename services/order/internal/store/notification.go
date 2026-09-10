package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Notification struct {
	ID        int64         `db:"id"`
	UserID    int64         `db:"user_id"`
	Role      int64         `db:"role"`
	Type      string        `db:"type"`
	Title     string        `db:"title"`
	Content   string        `db:"content"`
	OrderID   sql.NullInt64 `db:"order_id"`
	IsRead    int64         `db:"is_read"`
	CreatedAt time.Time     `db:"created_at"`
}

func InsertNotification(ctx context.Context, conn sqlx.Session, n *Notification) error {
	_, err := conn.ExecCtx(ctx,
		`insert into notifications(user_id, role, type, title, content, order_id, is_read)
		 values(?,?,?,?,?,?,0)`,
		n.UserID, n.Role, n.Type, n.Title, n.Content, nullableInt64(n.OrderID))
	return err
}

func ListNotifications(ctx context.Context, conn sqlx.Session, userID int64, unreadOnly bool, page, size int64) ([]Notification, int64, error) {
	page, size = normalizePage(page, size)
	where := "where user_id = ?"
	args := []any{userID}
	if unreadOnly {
		where += " and is_read = 0"
	}
	var total int64
	if err := conn.QueryRowCtx(ctx, &total, "select count(*) from notifications "+where, args...); err != nil {
		return nil, 0, err
	}
	query := `select id, user_id, role, type, title, content, order_id, is_read, created_at
		from notifications ` + where + " order by id desc limit ? offset ?"
	args = append(args, size, (page-1)*size)
	var rows []Notification
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func CountUnreadNotifications(ctx context.Context, conn sqlx.Session, userID int64) (int64, error) {
	var cnt int64
	if err := conn.QueryRowCtx(ctx, &cnt,
		"select count(*) from notifications where user_id = ? and is_read = 0", userID); err != nil {
		return 0, err
	}
	return cnt, nil
}

func MarkNotificationRead(ctx context.Context, conn sqlx.Session, userID, id int64) (bool, error) {
	result, err := conn.ExecCtx(ctx,
		"update notifications set is_read = 1 where id = ? and user_id = ? and is_read = 0", id, userID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func MarkAllNotificationsRead(ctx context.Context, conn sqlx.Session, userID int64) error {
	_, err := conn.ExecCtx(ctx,
		"update notifications set is_read = 1 where user_id = ? and is_read = 0", userID)
	return err
}
