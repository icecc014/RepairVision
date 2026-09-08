package store

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OperationLog struct {
	ID           int64          `db:"id"`
	OpID         string         `db:"op_id"`
	UserID       sql.NullInt64  `db:"user_id"`
	Username     sql.NullString `db:"username"`
	Role         sql.NullInt64  `db:"role"`
	Module       string         `db:"module"`
	Action       string         `db:"action"`
	Method       string         `db:"method"`
	Path         string         `db:"path"`
	RequestBody  string         `db:"request_body"`
	ResponseCode int64          `db:"response_code"`
	IP           string         `db:"ip"`
	CostMs       int64          `db:"cost_ms"`
	CreatedAt    time.Time      `db:"created_at"`
}

type LogCount struct {
	Cnt int64 `db:"cnt"`
}

func nullableString(s sql.NullString) any {
	if !s.Valid || s.String == "" {
		return nil
	}
	return s.String
}

func InsertOperationLog(ctx context.Context, conn sqlx.Session, l *OperationLog) error {
	_, err := conn.ExecCtx(ctx,
		`insert into operation_logs(op_id, user_id, username, role, module, action, method, path,
			request_body, response_code, ip, cost_ms)
		 values(?,?,?,?,?,?,?,?,?,?,?,?)`,
		l.OpID, nullableInt64(l.UserID), nullableString(l.Username), nullableInt64(l.Role),
		l.Module, l.Action, l.Method, l.Path, l.RequestBody, l.ResponseCode, l.IP, l.CostMs)
	return err
}

func ListOperationLogs(ctx context.Context, conn sqlx.Session, page, size int64, module, action, keyword string) ([]OperationLog, int64, error) {
	where := []string{"1 = 1"}
	var args []any
	if module != "" {
		where = append(where, "module = ?")
		args = append(args, module)
	}
	if action != "" {
		where = append(where, "action = ?")
		args = append(args, action)
	}
	if keyword != "" {
		where = append(where, "(username like ? or path like ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	condition := strings.Join(where, " and ")
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	var total LogCount
	if err := conn.QueryRowCtx(ctx, &total,
		"select count(*) cnt from operation_logs where "+condition, args...); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	queryArgs := append(append([]any{}, args...), size, offset)
	var logs []OperationLog
	if err := conn.QueryRowsCtx(ctx, &logs,
		"select * from operation_logs where "+condition+" order by id desc limit ? offset ?", queryArgs...); err != nil {
		return nil, 0, err
	}
	return logs, total.Cnt, nil
}
