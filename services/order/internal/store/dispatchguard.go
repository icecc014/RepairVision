package store

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// pendingBacklog 自动派单队列的积压指标。
type pendingBacklog struct {
	Count       int64         `db:"pending_count"`
	WaitMinutes sql.NullInt64 `db:"wait_minutes"`
}

// PendingBacklog 待派单且未标记"待管理员处置/外援"的数量与最长等待分钟数。
func PendingBacklog(ctx context.Context, conn sqlx.Session) (int64, int64, error) {
	var row pendingBacklog
	if err := conn.QueryRowCtx(ctx, &row,
		`select count(*) as pending_count,
		        coalesce(timestampdiff(minute, min(created_at), now()), 0) as wait_minutes
		   from orders
		  where status = ? and manual_review = 0 and external_mark = 0`,
		StatusPending); err != nil {
		return 0, 0, err
	}
	minutes := int64(0)
	if row.WaitMinutes.Valid && row.WaitMinutes.Int64 > 0 {
		minutes = row.WaitMinutes.Int64
	}
	return row.Count, minutes, nil
}

// ruleValueRow 派单规则数值行。
type ruleValueRow struct {
	RuleValue float64 `db:"rule_value"`
}

// GetRuleValue 读取派单规则配置数值；不存在时返回 def。
func GetRuleValue(ctx context.Context, conn sqlx.Session, key string, def float64) (float64, error) {
	var row ruleValueRow
	if err := conn.QueryRowCtx(ctx, &row,
		"select rule_value from dispatch_rule_config where rule_key = ? limit 1", key); err != nil {
		if sqlx.ErrNotFound == err {
			return def, nil
		}
		return def, nil // 规则缺失/读取异常时回落默认值，不阻塞派单
	}
	return row.RuleValue, nil
}

// SetRuleValue 写入/更新派单规则配置数值。
func SetRuleValue(ctx context.Context, conn sqlx.Session, key string, value float64) error {
	_, err := conn.ExecCtx(ctx,
		`insert into dispatch_rule_config(rule_key, rule_value, enabled, remark) values(?,?,1,'')
		 on duplicate key update rule_value = values(rule_value)`, key, value)
	return err
}