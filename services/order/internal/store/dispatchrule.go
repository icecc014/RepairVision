package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DispatchRule struct {
	ID        int64          `db:"id"`
	RuleKey   string         `db:"rule_key"`
	RuleValue float64        `db:"rule_value"`
	Enabled   int64          `db:"enabled"`
	Remark    sql.NullString `db:"remark"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func ListDispatchRules(ctx context.Context, conn sqlx.Session) ([]DispatchRule, error) {
	var list []DispatchRule
	if err := conn.QueryRowsCtx(ctx, &list,
		`select id, rule_key, cast(rule_value as decimal(10,4)) rule_value, enabled, remark, updated_at
		 from dispatch_rule_config order by id`); err != nil {
		return nil, err
	}
	return list, nil
}

func FindDispatchRuleByID(ctx context.Context, conn sqlx.Session, id int64) (*DispatchRule, error) {
	var r DispatchRule
	if err := conn.QueryRowCtx(ctx, &r,
		`select id, rule_key, cast(rule_value as decimal(10,4)) rule_value, enabled, remark, updated_at
		 from dispatch_rule_config where id = ?`, id); err != nil {
		return nil, err
	}
	return &r, nil
}

func FindDispatchRuleByKey(ctx context.Context, conn sqlx.Session, key string) (*DispatchRule, error) {
	var r DispatchRule
	if err := conn.QueryRowCtx(ctx, &r,
		`select id, rule_key, cast(rule_value as decimal(10,4)) rule_value, enabled, remark, updated_at
		 from dispatch_rule_config where rule_key = ?`, key); err != nil {
		return nil, err
	}
	return &r, nil
}

func InsertDispatchRuleIfAbsent(ctx context.Context, conn sqlx.Session, key string, value float64, remark string) error {
	_, err := conn.ExecCtx(ctx,
		"insert ignore into dispatch_rule_config(rule_key, rule_value, enabled, remark) values(?,?,1,?)",
		key, value, remark)
	return err
}

func CreateDispatchRule(ctx context.Context, conn sqlx.Session, key string, value float64, enabled int64, remark string) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into dispatch_rule_config(rule_key, rule_value, enabled, remark) values(?,?,?,?)",
		key, value, enabled, remark)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateDispatchRule(ctx context.Context, conn sqlx.Session, id int64, value float64, enabled int64, remark string) error {
	_, err := conn.ExecCtx(ctx,
		"update dispatch_rule_config set rule_value = ?, enabled = ?, remark = ? where id = ?",
		value, enabled, remark, id)
	return err
}
