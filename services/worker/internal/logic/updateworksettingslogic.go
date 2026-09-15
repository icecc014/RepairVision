package logic

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type UpdateWorkSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateWorkSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateWorkSettingsLogic {
	return &UpdateWorkSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateWorkSettings 更新排班与工作时段配置（双休模式、固定休息日、上午/下午时段、是否允许强制开工）。
func (l *UpdateWorkSettingsLogic) UpdateWorkSettings(in *worker.UpdateWorkSettingsRequest) (*worker.WorkSettingsResponse, error) {
	restDays := in.RestDaysPerWeek
	if restDays < 0 || restDays > 3 {
		return nil, errDate("每周休息天数需在 0~3 之间")
	}
	mode := strings.ToLower(strings.TrimSpace(in.RestMode))
	if mode == "" {
		mode = "staggered"
	}
	if mode != "staggered" && mode != "fixed" {
		return nil, errDate("轮休模式必须为 staggered（错峰轮休）或 fixed（固定休息日）")
	}
	weekdays := normalizeWeekdays(in.FixedRestWeekdays)
	if mode == "fixed" && weekdays == "" {
		return nil, errDate("固定休息日模式需至少选择一个休息日（1=周一 … 7=周日）")
	}
	morningStart := strings.TrimSpace(in.MorningStart)
	morningEnd := strings.TrimSpace(in.MorningEnd)
	afternoonStart := strings.TrimSpace(in.AfternoonStart)
	afternoonEnd := strings.TrimSpace(in.AfternoonEnd)
	for _, pair := range [][2]string{{morningStart, morningEnd}, {afternoonStart, afternoonEnd}} {
		s, ok1 := parseHHMM(pair[0])
		e, ok2 := parseHHMM(pair[1])
		if !ok1 || !ok2 {
			return nil, errDate("工作时段格式应为 HH:MM（如 08:00）")
		}
		if e <= s {
			return nil, errDate("工作时段的下班时间必须晚于上班时间")
		}
	}
	allowForce := int64(1)
	if in.AllowForceStart == 0 {
		allowForce = 0
	}
	settings := &store.WorkSettings{
		RestDaysPerWeek:   restDays,
		RestMode:          mode,
		FixedRestWeekdays: weekdays,
		MorningStart:      morningStart,
		MorningEnd:        morningEnd,
		AfternoonStart:    afternoonStart,
		AfternoonEnd:      afternoonEnd,
		AllowForceStart:   allowForce,
	}
	if err := store.SaveWorkSettings(l.ctx, l.svcCtx.DB, settings); err != nil {
		return nil, err
	}
	return &worker.WorkSettingsResponse{Settings: settingsToPb(settings)}, nil
}

// normalizeWeekdays 规范"固定休息日"字符串：仅保留 1~7，去重并升序。
func normalizeWeekdays(raw string) string {
	seen := make(map[int]bool)
	for _, part := range strings.Split(raw, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > 7 {
			continue
		}
		seen[n] = true
	}
	parts := make([]string, 0, len(seen))
	for d := 1; d <= 7; d++ {
		if seen[d] {
			parts = append(parts, strconv.Itoa(d))
		}
	}
	return strings.Join(parts, ",")
}