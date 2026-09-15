package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// SetOrderLock 锁定/解锁待派工单；锁定后不参与自动派单与批量派单。
func SetOrderLock(ctx context.Context, conn sqlx.Session, orderID, locked int64) (bool, error) {
	if locked != 0 {
		locked = 1
	}
	result, err := conn.ExecCtx(ctx,
		"update orders set dispatch_locked = ?, updated_at = now() where id = ? and status = ?",
		locked, orderID, StatusPending)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// UpdateOrderPriority 调整工单优先级（1 最低 ~ 5 最高），待派队列按 priority desc 排序。
func UpdateOrderPriority(ctx context.Context, conn sqlx.Session, orderID, priority int64) (bool, error) {
	if priority < 1 {
		priority = 1
	}
	if priority > 5 {
		priority = 5
	}
	result, err := conn.ExecCtx(ctx,
		"update orders set priority = ?, updated_at = now() where id = ? and status in (?, ?)",
		priority, orderID, StatusPending, StatusDispatched)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}