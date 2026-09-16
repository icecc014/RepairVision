package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// V6.3 派单规则的写入接口（供"派单规则配置"可视化编辑使用）。

// UpsertDispatchRule 新增或更新一条派单规则（按 rule_key 唯一）。
func UpsertDispatchRule(ctx context.Context, conn sqlx.Session, key string, value float64, enabled int64, remark string) error {
	_, err := conn.ExecCtx(ctx,
		`insert into dispatch_rule_config(rule_key, rule_value, enabled, remark) values(?, ?, ?, ?)
		 on duplicate key update rule_value = values(rule_value), enabled = values(enabled), remark = values(remark)`,
		key, value, enabled, remark)
	return err
}

// DeleteDispatchRuleByKey 按 key 删除派单规则。
func DeleteDispatchRuleByKey(ctx context.Context, conn sqlx.Session, key string) error {
	_, err := conn.ExecCtx(ctx, "delete from dispatch_rule_config where rule_key = ?", key)
	return err
}
