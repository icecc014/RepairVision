package logic

import (
	"strings"
	"time"

	"worker/internal/store"
	"worker/worker"
)

// parseHHMM 解析 "HH:MM" 为当天的分钟数。
func parseHHMM(v string) (int, bool) {
	t, err := time.Parse("15:04", strings.TrimSpace(v))
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}

// inTimeRange 判断给定时刻是否落在 [start, end) 之间（同一天内的时段）。
func inTimeRange(at time.Time, startHHMM, endHHMM string) bool {
	start, ok1 := parseHHMM(startHHMM)
	end, ok2 := parseHHMM(endHHMM)
	if !ok1 || !ok2 || end <= start {
		return false
	}
	cur := at.Hour()*60 + at.Minute()
	return cur >= start && cur < end
}

// inWorkPeriod 判断时刻是否处于任一日间工作时段。
func inWorkPeriod(at time.Time, s *store.WorkSettings) bool {
	return inTimeRange(at, s.MorningStart, s.MorningEnd) || inTimeRange(at, s.AfternoonStart, s.AfternoonEnd)
}

// shiftIsWork 判断班次是否属于"白班"（历史 MORNING / AFTERNOON 视为白班兼容数据）。
func shiftIsWork(shift string) bool {
	switch strings.ToUpper(strings.TrimSpace(shift)) {
	case "DAY", "MORNING", "AFTERNOON":
		return true
	default:
		return false
	}
}

func settingsToPb(s *store.WorkSettings) *worker.WorkSettings {
	return &worker.WorkSettings{
		RestDaysPerWeek:   s.RestDaysPerWeek,
		RestMode:          s.RestMode,
		FixedRestWeekdays: s.FixedRestWeekdays,
		MorningStart:      s.MorningStart,
		MorningEnd:        s.MorningEnd,
		AfternoonStart:    s.AfternoonStart,
		AfternoonEnd:      s.AfternoonEnd,
		AllowForceStart:   s.AllowForceStart,
	}
}

func workPeriodText(s *store.WorkSettings) string {
	return s.MorningStart + "-" + s.MorningEnd + " / " + s.AfternoonStart + "-" + s.AfternoonEnd
}