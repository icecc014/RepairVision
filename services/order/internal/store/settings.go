package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// V6.2 通用键值配置：sys_settings(setting_key, setting_value, remark)。

// GetSetting 读取配置；不存在时返回 sqlx.ErrNotFound。
func GetSetting(ctx context.Context, conn sqlx.Session, key string) (string, error) {
	var value string
	if err := conn.QueryRowCtx(ctx, &value,
		"select setting_value from sys_settings where setting_key = ?", key); err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting 写入/更新配置。
func SetSetting(ctx context.Context, conn sqlx.Session, key, value, remark string) error {
	if remark == "" {
		_, err := conn.ExecCtx(ctx,
			`insert into sys_settings(setting_key, setting_value) values(?, ?)
			 on duplicate key update setting_value = values(setting_value)`, key, value)
		return err
	}
	_, err := conn.ExecCtx(ctx,
		`insert into sys_settings(setting_key, setting_value, remark) values(?, ?, ?)
		 on duplicate key update setting_value = values(setting_value), remark = values(remark)`,
		key, value, remark)
	return err
}
