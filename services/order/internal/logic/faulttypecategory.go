package logic

import "strings"

// normalizeFaultCategory 规范故障类型类别为 electric / water / other，未知或空值回退为 other。
func normalizeFaultCategory(raw string) string {

	v := strings.ToLower(strings.TrimSpace(raw))

	switch v {

	case "electric", "water", "other":

		return v

	default:

		return "other"

	}
}
