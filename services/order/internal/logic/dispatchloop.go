package logic

import (
	"context"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/svc"
	"order/internal/types"
)

// 兜底扫描周期与事件防抖间隔。
const (
	dispatchLoopInterval = 60 * time.Second
	dispatchDebounce     = 1500 * time.Millisecond
)

var (
	dispatchDebounceMu    sync.Mutex
	dispatchDebounceTimer *time.Timer
)

// StartDispatchLoop 自动派单兜底扫描：每 60 秒检查一次待派队列，
// 非保护模式且有积压时执行一次批量指派（处理漏派）。
func StartDispatchLoop(ctx context.Context, svcCtx *svc.ServiceContext) {
	logx.Infof("dispatch loop started, interval=%s", dispatchLoopInterval)
	ticker := time.NewTicker(dispatchLoopInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runDispatchRecheck(ctx, svcCtx, "ticker")
		}
	}
}

// TriggerDispatchRecheck 事件驱动的重算触发（新工单/完工/上下班/请假/调班等），
// 带 1.5 秒防抖，避免短时间内重复计算。
func TriggerDispatchRecheck(svcCtx *svc.ServiceContext) {
	dispatchDebounceMu.Lock()
	defer dispatchDebounceMu.Unlock()
	if dispatchDebounceTimer != nil {
		dispatchDebounceTimer.Stop()
	}
	dispatchDebounceTimer = time.AfterFunc(dispatchDebounce, func() {
		runDispatchRecheck(context.Background(), svcCtx, "event")
	})
}

// runDispatchRecheck 执行一次兜底指派；返回实际派出的工单数。
func runDispatchRecheck(ctx context.Context, svcCtx *svc.ServiceContext, source string) int {
	if autoDispatchPaused(ctx, svcCtx) {
		logx.WithContext(ctx).Infof("dispatch recheck(%s) skipped: auto dispatch paused by guard", source)
		return 0
	}
	if _, waitMinutes, err := pendingCountForGuard(ctx, svcCtx); err == nil && waitMinutes > 0 {
		// 触发一次保护状态评估，积压越线时自动切换到人工处置模式
		if _, err := evaluateDispatchGuard(ctx, svcCtx); err != nil {
			logx.WithContext(ctx).Errorf("evaluate guard failed: %v", err)
		}
	}
	systemCtx := auth.WithSystemIdentity(ctx)
	resp, err := NewAdminBatchDispatchLogic(systemCtx, svcCtx).AdminBatchDispatch(&types.AdminBatchDispatchRequest{})
	if err != nil {
		logx.WithContext(ctx).Errorf("dispatch recheck(%s) failed: %v", source, err)
		return 0
	}
	if len(resp.Dispatched) > 0 || resp.Remained > 0 {
		logx.WithContext(ctx).Infof("dispatch recheck(%s): dispatched=%d remained=%d",
			source, len(resp.Dispatched), resp.Remained)
	}
	return len(resp.Dispatched)
}

// pendingCountForGuard 读取当前积压指标（数量 + 最长等待分钟）。
func pendingCountForGuard(ctx context.Context, svcCtx *svc.ServiceContext) (int64, int64, error) {
	return pendingBacklogOf(ctx, svcCtx)
}