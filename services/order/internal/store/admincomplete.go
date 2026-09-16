package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// AdminCompleteOrder 管理员代为完工（演示/救急用）：
// 不要求 worker_id 匹配当前登录人，允许 status = 2(已派单) / 3(维修中)。
func AdminCompleteOrder(ctx context.Context, conn sqlx.Session, orderID int64) (bool, error) {
	now := time.Now()
	result, err := conn.ExecCtx(ctx,
		`update orders set status = ?, completed_at = ?, started_at = coalesce(started_at, ?)
		  where id = ? and status in (?, ?)`,
		StatusCompleted, now, now, orderID, StatusDispatched, StatusWorking)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}
