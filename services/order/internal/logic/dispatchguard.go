package logic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/store"
	"order/internal/svc"
	"worker/workerclient"
)

// 双阈值保护配置键（存 order_db.dispatch_rule_config，可在"派单规则"页调整）
const (
	ruleKeyWarnRatio     = "backlog_warn_ratio"     // 预警线：待派 > 在岗 × 该倍数
	ruleKeyGuardRatio    = "backlog_guard_ratio"    // 保护线：待派 > 在岗 × 该倍数
	ruleKeyWarnHours     = "backlog_warn_hours"     // 预警线：存在等待超过该小时数的工单
	ruleKeyGuardHours    = "backlog_guard_hours"    // 保护线：存在等待超过该小时数的工单
	ruleKeyAutoPaused    = "auto_dispatch_paused"   // 1 = 已暂停自动派单（人工处置模式）
	defaultWarnRatio     = 3.0
	defaultGuardRatio    = 5.0
	defaultWarnHours     = 2.0
	defaultGuardHours    = 4.0
)

// 保护级别
const (
	guardLevelOK    = "ok"
	guardLevelWarn  = "warn"
	guardLevelGuard = "guard"
)

// DispatchGuard 双阈值保护状态：预警线只提醒，保护线暂停自动派单转人工处置。
type DispatchGuard struct {
	PendingCount       int64   `json:"pendingCount"`
	OnDutyCount        int64   `json:"onDutyCount"`
	LongestWaitMinutes int64   `json:"longestWaitMinutes"`
	WarnRatio          float64 `json:"warnRatio"`
	GuardRatio         float64 `json:"guardRatio"`
	WarnHours          float64 `json:"warnHours"`
	GuardHours         float64 `json:"guardHours"`
	Level              string  `json:"level"`
	Paused             bool    `json:"paused"`
	Mode               string  `json:"mode"`
	Message            string  `json:"message"`
}

// evaluateDispatchGuard 汇总当前积压指标与保护级别（保护线触发时自动暂停自动派单）。
func evaluateDispatchGuard(ctx context.Context, svcCtx *svc.ServiceContext) (*DispatchGuard, error) {
	pending, waitMinutes, err := store.PendingBacklog(ctx, svcCtx.DB)
	if err != nil {
		return nil, err
	}
	onDuty, err := countOnDutyWorkers(ctx, svcCtx)
	if err != nil {
		return nil, err
	}
	warnRatio, _ := store.GetRuleValue(ctx, svcCtx.DB, ruleKeyWarnRatio, defaultWarnRatio)
	guardRatio, _ := store.GetRuleValue(ctx, svcCtx.DB, ruleKeyGuardRatio, defaultGuardRatio)
	warnHours, _ := store.GetRuleValue(ctx, svcCtx.DB, ruleKeyWarnHours, defaultWarnHours)
	guardHours, _ := store.GetRuleValue(ctx, svcCtx.DB, ruleKeyGuardHours, defaultGuardHours)
	pausedValue, _ := store.GetRuleValue(ctx, svcCtx.DB, ruleKeyAutoPaused, 0)

	// 无人在岗时按 1 人计算：此时任何待派工单都会触碰保护线，符合"无人可用转人工"的语义。
	base := onDuty
	if base <= 0 {
		base = 1
	}
	state := &DispatchGuard{
		PendingCount:       pending,
		OnDutyCount:        onDuty,
		LongestWaitMinutes: waitMinutes,
		WarnRatio:          warnRatio,
		GuardRatio:         guardRatio,
		WarnHours:          warnHours,
		GuardHours:         guardHours,
		Paused:             pausedValue >= 0.5,
		Level:              guardLevelOK,
	}
	guardHit, warnHit := guardThresholds(pending, base, waitMinutes, warnRatio, guardRatio, warnHours, guardHours)
	state.Level, state.Message = guardLevelText(guardHit, warnHit)

	if guardHit && !state.Paused {
		if err := store.SetRuleValue(ctx, svcCtx.DB, ruleKeyAutoPaused, 1); err != nil {
			return nil, err
		}
		state.Paused = true
		logx.WithContext(ctx).Errorf("dispatch guard triggered: pending=%d onDuty=%d waitMinutes=%d, auto dispatch paused",
			pending, onDuty, waitMinutes)
	}
	state.Mode = "auto"
	if state.Paused {
		state.Mode = "manual"
	}
	return state, nil
}

// autoDispatchPaused 自动派单是否处于暂停（保护模式）。
func autoDispatchPaused(ctx context.Context, svcCtx *svc.ServiceContext) bool {
	v, err := store.GetRuleValue(ctx, svcCtx.DB, ruleKeyAutoPaused, 0)
	if err != nil {
		return false // 读取失败不阻塞派单
	}
	return v >= 0.5
}

// resumeAutoDispatch 管理员处理完积压后恢复自动派单。
func resumeAutoDispatch(ctx context.Context, svcCtx *svc.ServiceContext) error {
	return store.SetRuleValue(ctx, svcCtx.DB, ruleKeyAutoPaused, 0)
}

// countOnDutyWorkers 统计当前在岗工人数（白班 ∧ 时段内 ∧ 未请假 ∧ 启用）。
func countOnDutyWorkers(ctx context.Context, svcCtx *svc.ServiceContext) (int64, error) {
	users, err := svcCtx.WorkerRpc.ListUsers(ctx, &workerclient.ListUsersRequest{Role: 2, Status: 1})
	if err != nil {
		return 0, err
	}
	count := int64(0)
	for _, u := range users.GetUsers() {
		status, err := svcCtx.WorkerRpc.GetDutyStatus(ctx, &workerclient.DutyStatusRequest{WorkerId: u.Id})
		if err != nil {
			continue
		}
		if status.GetOnDuty() {
			count++
		}
	}
	return count, nil
}

// waitText 把等待分钟数转成便于展示的中文描述。
func waitText(minutes int64) string {
	if minutes <= 0 {
		return "无积压"
	}
	if minutes < 60 {
		return strconv.FormatInt(minutes, 10) + " 分钟"
	}
	return strconv.FormatFloat(float64(minutes)/60, 'f', 1, 64) + " 小时"
}

// guardThresholds 纯函数：判断是否触及保护线 / 预警线。
// base 为在岗人数（无人时调用方按 1 计），waitMinutes 为最长等待分钟数。
func guardThresholds(pending, base, waitMinutes int64, warnRatio, guardRatio, warnHours, guardHours float64) (guardHit, warnHit bool) {
	guardHit = float64(pending) > float64(base)*guardRatio || float64(waitMinutes) > guardHours*60
	warnHit = float64(pending) > float64(base)*warnRatio || float64(waitMinutes) > warnHours*60
	return
}

// guardLevelText 纯函数：把阈值判定转成级别与提示文案。
func guardLevelText(guardHit, warnHit bool) (level, message string) {
	switch {
	case guardHit:
		return guardLevelGuard, "积压已触发保护线：自动派单暂停，请增援 / 调班 / 外援或手动指派后再恢复自动派单"
	case warnHit:
		return guardLevelWarn, "积压已触发预警线：请关注待派工单与在岗人数"
	default:
		return guardLevelOK, "运行正常"
	}
}