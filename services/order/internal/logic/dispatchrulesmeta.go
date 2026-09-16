package logic

import (
	"errors"
	"math"
	"regexp"
)

// V6.3 派单规则注册表：让"派单规则配置"页面变成可视化编辑。
// 引擎只认注册表里的 key，所以新增规则时优先从注册表里选（补录），
// 自定义 key 会标注"引擎暂不支持"，只作记录用。

type dispatchRuleMeta struct {
	Key         string
	Name        string
	Group       string
	Type        string // number | bool
	Unit        string
	Min         float64
	Max         float64
	Step        float64
	Precision   int
	Default     float64
	Description string
	RuntimeOnly bool // 运行态：由系统维护，界面只读
}

var dispatchRuleGroupMeta = []struct {
	Key   string
	Label string
}{
	{"weights", "评分权重（技能 / 距离 / 负载）"},
	{"auto", "自动派单开关"},
	{"timeout", "超时提醒"},
	{"backlog", "积压保护线（预警 / 保护）"},
	{"runtime", "运行状态（只读）"},
	{"custom", "自定义规则（引擎暂不支持）"},
}

var dispatchRuleRegistry = []dispatchRuleMeta{
	{Key: "skill_weight", Name: "技能匹配权重", Group: "weights", Type: "number", Unit: "", Min: 0, Max: 1, Step: 0.05, Precision: 2, Default: 0.4,
		Description: "工单故障类型与工人技能的匹配度在总分中的占比，建议与距离/负载三者合计为 1"},
	{Key: "distance_weight", Name: "距离权重", Group: "weights", Type: "number", Unit: "", Min: 0, Max: 1, Step: 0.05, Precision: 2, Default: 0.3,
		Description: "按路网最短路计算的建筑间距离占比（无路网数据时回退欧氏距离）"},
	{Key: "load_weight", Name: "负载权重", Group: "weights", Type: "number", Unit: "", Min: 0, Max: 1, Step: 0.05, Precision: 2, Default: 0.3,
		Description: "工人在途负载占比；高于人均 1.2 倍会额外施加动态惩罚"},

	{Key: "auto_dispatch_enabled", Name: "启用自动派单", Group: "auto", Type: "bool", Min: 0, Max: 1, Step: 1, Default: 1,
		Description: "关闭后新工单只入队，全部由管理员手动指派"},

	{Key: "pending_timeout_hours", Name: "待派超时阈值", Group: "timeout", Type: "number", Unit: "小时", Min: 0.5, Max: 72, Step: 0.5, Precision: 1, Default: 2,
		Description: "工单处于待派状态超过该时长会在看板上标红提醒"},
	{Key: "dispatched_timeout_hours", Name: "已派未开工超时阈值", Group: "timeout", Type: "number", Unit: "小时", Min: 0.5, Max: 72, Step: 0.5, Precision: 1, Default: 4,
		Description: "已派单但迟迟未开工超过该时长会在看板上标红提醒"},

	{Key: "backlog_warn_min_orders", Name: "预警线：绝对单量", Group: "backlog", Type: "number", Unit: "单", Min: 1, Max: 500, Step: 1, Precision: 0, Default: 10,
		Description: "待派工单达到该数量即预警（与「在岗 × 倍数」取较大者作为门槛）"},
	{Key: "backlog_warn_ratio", Name: "预警线：人均倍数", Group: "backlog", Type: "number", Unit: "倍", Min: 1, Max: 100, Step: 1, Precision: 0, Default: 5,
		Description: "待派 ≥ 在岗人数 × 该倍数 时预警；人多时自动提高门槛"},
	{Key: "backlog_warn_hours", Name: "预警线：最长等待", Group: "backlog", Type: "number", Unit: "小时", Min: 0.5, Max: 72, Step: 0.5, Precision: 1, Default: 6,
		Description: "存在等待超过该小时数仍未派出的工单时预警"},

	{Key: "backlog_guard_min_orders", Name: "保护线：绝对单量", Group: "backlog", Type: "number", Unit: "单", Min: 1, Max: 500, Step: 1, Precision: 0, Default: 20,
		Description: "待派工单达到该数量即暂停自动派单转人工处置"},
	{Key: "backlog_guard_ratio", Name: "保护线：人均倍数", Group: "backlog", Type: "number", Unit: "倍", Min: 1, Max: 100, Step: 1, Precision: 0, Default: 10,
		Description: "待派 ≥ 在岗人数 × 该倍数 时暂停自动派单"},
	{Key: "backlog_guard_hours", Name: "保护模式：自动恢复等待时长", Group: "backlog", Type: "number", Unit: "小时", Min: 0.5, Max: 72, Step: 0.5, Precision: 1, Default: 12,
		Description: "保护模式（暂停自动派单）期间，若仍有工单等待超过该小时数，则自动恢复自动派单并把积压派出去"},

	{Key: "auto_dispatch_paused", Name: "当前是否暂停自动派单", Group: "runtime", Type: "bool", Min: 0, Max: 1, Step: 1, Default: 0, RuntimeOnly: true,
		Description: "由双阈值保护自动维护：触发保护线置 1，积压回落自动置 0；也可用「恢复自动派单」手动复位"},
}

var customRuleKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)

func findDispatchRuleMeta(key string) *dispatchRuleMeta {
	for i := range dispatchRuleRegistry {
		if dispatchRuleRegistry[i].Key == key {
			return &dispatchRuleRegistry[i]
		}
	}
	return nil
}

func dispatchRuleGroupLabel(key string) string {
	for _, g := range dispatchRuleGroupMeta {
		if g.Key == key {
			return g.Label
		}
	}
	return key
}

// validateRuleValue 校验规则取值（注册表外的自定义规则只做基础校验）。
func validateRuleValue(meta *dispatchRuleMeta, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return errors.New("规则值必须是有效数字")
	}
	if meta == nil {
		if math.Abs(value) > 1e9 {
			return errors.New("规则值过大")
		}
		return nil
	}
	if meta.Type == "bool" {
		if value != 0 && value != 1 {
			return errors.New("开关类规则只能是 0（关）或 1（开）")
		}
		return nil
	}
	if value < meta.Min || value > meta.Max {
		return errors.New("规则值超出允许范围（" + trimFloat(meta.Min) + " ~ " + trimFloat(meta.Max) + "）")
	}
	return nil
}

func trimFloat(v float64) string {
	if v == math.Trunc(v) {
		return strconvFormatInt(int64(v))
	}
	return strconvFormatFloat(v)
}
