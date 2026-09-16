package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
)

type AdminDispatchGuardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchGuardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchGuardLogic {
	return &AdminDispatchGuardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminDispatchGuard 查询双阈值保护状态（预警线/保护线、当前模式、积压指标）。
func (l *AdminDispatchGuardLogic) AdminDispatchGuard() (resp *types.DispatchGuardResponse, err error) {
	state, err := evaluateDispatchGuard(l.ctx, l.svcCtx)
	if err != nil {
		return nil, errs.Internal(err)
	}
	return dispatchGuardToResponse(state), nil
}

type AdminDispatchGuardResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchGuardResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchGuardResumeLogic {
	return &AdminDispatchGuardResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminDispatchGuardResume 管理员处理完积压后恢复自动派单。
func (l *AdminDispatchGuardResumeLogic) AdminDispatchGuardResume() (resp *types.DispatchGuardResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if err := resumeAutoDispatch(l.ctx, l.svcCtx); err != nil {
		return nil, errs.Internal(err)
	}
	logx.WithContext(l.ctx).Infof("auto dispatch resumed by admin")
	// V6.3：恢复后立即触发一次派单重算，积压单当场派出去（不用等 60 秒兜底扫描）
	TriggerDispatchRecheck(l.svcCtx)
	state, err := evaluateDispatchGuard(l.ctx, l.svcCtx)
	if err != nil {
		return nil, errs.Internal(err)
	}
	return dispatchGuardToResponse(state), nil
}

func dispatchGuardToResponse(s *DispatchGuard) *types.DispatchGuardResponse {
	if s == nil {
		return &types.DispatchGuardResponse{Mode: "auto", Level: guardLevelOK}
	}
	return &types.DispatchGuardResponse{
		PendingCount:       s.PendingCount,
		OnDutyCount:        s.OnDutyCount,
		LongestWaitMinutes: s.LongestWaitMinutes,
		WarnRatio:          s.WarnRatio,
		WarnMinOrders:      s.WarnMinOrders,
		GuardRatio:         s.GuardRatio,
		GuardMinOrders:     s.GuardMinOrders,
		WarnHours:          s.WarnHours,
		GuardHours:         s.GuardHours,
		Level:              s.Level,
		Paused:             s.Paused,
		Mode:               s.Mode,
		Message:            s.Message,
		WaitText:           waitText(s.LongestWaitMinutes),
	}
}