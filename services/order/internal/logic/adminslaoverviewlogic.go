package logic

import (
	"context"
	"math"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminSlaOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminSlaOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSlaOverviewLogic {
	return &AdminSlaOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminSlaOverviewLogic) AdminSlaOverview() (resp *types.SlaOverviewResponse, err error) {
	rules, err := store.ListDispatchRules(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	pendingHours := int64(2)
	dispatchedHours := int64(4)
	for _, r := range rules {
		switch r.RuleKey {
		case "pending_timeout_hours":
			if r.RuleValue > 0 {
				pendingHours = int64(math.Round(r.RuleValue))
			}
		case "dispatched_timeout_hours":
			if r.RuleValue > 0 {
				dispatchedHours = int64(math.Round(r.RuleValue))
			}
		}
	}

	pendingOverdue, err := store.CountPendingOverdue(l.ctx, l.svcCtx.DB, pendingHours)
	if err != nil {
		return nil, errs.Internal(err)
	}
	dispatchedOverdue, err := store.CountDispatchedOverdue(l.ctx, l.svcCtx.DB, dispatchedHours)
	if err != nil {
		return nil, errs.Internal(err)
	}
	avgDispatch, err := store.AvgDispatchMinutes(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	avgRepair, err := store.AvgRepairMinutes(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	overdue, err := store.ListSlaOverdueOrders(l.ctx, l.svcCtx.DB, pendingHours, dispatchedHours)
	if err != nil {
		return nil, errs.Internal(err)
	}
	items, err := buildOrderItems(l.ctx, l.svcCtx, overdue)
	if err != nil {
		return nil, err
	}

	return &types.SlaOverviewResponse{
		PendingTimeoutHours:    pendingHours,
		DispatchedTimeoutHours: dispatchedHours,
		PendingOverdue:         pendingOverdue,
		DispatchedOverdue:      dispatchedOverdue,
		AvgDispatchMinutes:     math.Round(avgDispatch*100) / 100,
		AvgRepairMinutes:       math.Round(avgRepair*100) / 100,
		OverdueOrders:          items,
	}, nil
}
