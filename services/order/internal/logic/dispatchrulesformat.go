package logic

import (
	"strconv"

	"order/internal/types"
	"order/internal/store"
)

// 规则值格式化小工具（避免在 meta 文件里引入多余依赖）。
func strconvFormatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func strconvFormatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// ruleRowToEntry 把数据库行转换成前端条目。
func ruleRowToEntry(row store.DispatchRule) types.DispatchRuleEntry {
	remark := ""
	if row.Remark.Valid {
		remark = row.Remark.String
	}
	return types.DispatchRuleEntry{
		Key:       row.RuleKey,
		Name:      row.RuleKey,
		Type:      "number",
		Value:     row.RuleValue,
		Enabled:   row.Enabled,
		Remark:    remark,
		Custom:    true,
		UpdatedAt: row.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
