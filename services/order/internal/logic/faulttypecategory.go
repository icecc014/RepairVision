package logic

import "strings"

// normalizeFaultCategory 规范故障类型类别：electric / water / masonry / wood / other。
func normalizeFaultCategory(raw string) string {
	switch v := strings.ToLower(strings.TrimSpace(raw)); v {
	case "electric", "water", "masonry", "wood", "other":
		return v
	default:
		return "other"
	}
}

// requiredJobTypeOf 类别 → 所需工种（V6 扩展泥瓦/木工）：
// electric→1 电工、water→2 水工、masonry→3 泥瓦工、wood→4 木工、other→0（由 manual_review 转人工处置）。
// 约定：工人 job_type = 0 为通用，可接任意类别（见 jobTypeAllowed）。
func requiredJobTypeOf(category string) int64 {
	switch normalizeFaultCategory(category) {
	case "electric":
		return 1
	case "water":
		return 2
	case "masonry":
		return 3
	case "wood":
		return 4
	default:
		return 0
	}
}