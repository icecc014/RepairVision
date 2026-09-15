package store

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// WorkSettings 排班与工作时段配置（work_settings 单行表）。
type WorkSettings struct {
	RestDaysPerWeek   int64  `db:"rest_days_per_week"`
	RestMode          string `db:"rest_mode"`
	FixedRestWeekdays string `db:"fixed_rest_weekdays"`
	MorningStart      string `db:"morning_start"`
	MorningEnd        string `db:"morning_end"`
	AfternoonStart    string `db:"afternoon_start"`
	AfternoonEnd      string `db:"afternoon_end"`
	AllowForceStart   int64  `db:"allow_force_start"`
}

const workSettingsColumns = "rest_days_per_week, rest_mode, fixed_rest_weekdays, morning_start, morning_end, afternoon_start, afternoon_end, allow_force_start"

// DefaultWorkSettings 内置默认：双休（错峰轮休）、白班 08:00-12:00 / 14:00-18:00。
func DefaultWorkSettings() *WorkSettings {
	return &WorkSettings{
		RestDaysPerWeek:   2,
		RestMode:          "staggered",
		FixedRestWeekdays: "6,7",
		MorningStart:      "08:00",
		MorningEnd:        "12:00",
		AfternoonStart:    "14:00",
		AfternoonEnd:      "18:00",
		AllowForceStart:   1,
	}
}

// GetWorkSettings 读取单行配置；记录不存在时返回内置默认值（保证功能可用）。
func GetWorkSettings(ctx context.Context, conn sqlx.SqlConn) (*WorkSettings, error) {
	var s WorkSettings
	if err := conn.QueryRowCtx(ctx, &s,
		"select "+workSettingsColumns+" from work_settings where id = 1"); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return DefaultWorkSettings(), nil
		}
		return nil, err
	}
	normalizeSettings(&s)
	return &s, nil
}

// SaveWorkSettings 写入单行配置（不存在则插入）。
func SaveWorkSettings(ctx context.Context, conn sqlx.Session, s *WorkSettings) error {
	_, err := conn.ExecCtx(ctx,
		`insert into work_settings(id, rest_days_per_week, rest_mode, fixed_rest_weekdays,
			morning_start, morning_end, afternoon_start, afternoon_end, allow_force_start)
		 values(1,?,?,?,?,?,?,?,?)
		 on duplicate key update rest_days_per_week = values(rest_days_per_week), rest_mode = values(rest_mode),
			fixed_rest_weekdays = values(fixed_rest_weekdays), morning_start = values(morning_start),
			morning_end = values(morning_end), afternoon_start = values(afternoon_start),
			afternoon_end = values(afternoon_end), allow_force_start = values(allow_force_start)`,
		s.RestDaysPerWeek, s.RestMode, s.FixedRestWeekdays,
		s.MorningStart, s.MorningEnd, s.AfternoonStart, s.AfternoonEnd, s.AllowForceStart)
	return err
}

// normalizeSettings 修正越界值，保证下游逻辑拿到的是合法配置。
func normalizeSettings(s *WorkSettings) {
	if s.RestDaysPerWeek < 0 || s.RestDaysPerWeek > 3 {
		s.RestDaysPerWeek = 2
	}
	if s.RestMode != "fixed" && s.RestMode != "staggered" {
		s.RestMode = "staggered"
	}
	if s.FixedRestWeekdays == "" {
		s.FixedRestWeekdays = "6,7"
	}
	if s.MorningStart == "" {
		s.MorningStart = "08:00"
	}
	if s.MorningEnd == "" {
		s.MorningEnd = "12:00"
	}
	if s.AfternoonStart == "" {
		s.AfternoonStart = "14:00"
	}
	if s.AfternoonEnd == "" {
		s.AfternoonEnd = "18:00"
	}
	if s.AllowForceStart != 0 {
		s.AllowForceStart = 1
	}
}