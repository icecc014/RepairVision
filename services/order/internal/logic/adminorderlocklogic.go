package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminOrderLockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOrderLockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOrderLockLogic {
	return &AdminOrderLockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminOrderLock 锁定/解锁待派工单：锁定后该单不参与自动派单与批量派单（仍可手动改派）。
func (l *AdminOrderLockLogic) AdminOrderLock(req *types.AdminOrderLockRequest) (resp *types.EmptyResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.BadRequest("工单不存在")
		}
		return nil, errs.Internal(err)
	}
	affected, err := store.SetOrderLock(l.ctx, l.svcCtx.DB, order.ID, req.Locked)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("只有待派单的工单可以锁定 / 解锁")
	}
	if req.Locked != 0 {
		logx.WithContext(l.ctx).Infof("order %d locked, excluded from auto dispatch", order.ID)
	} else {
		// 解锁后立即触发一次重算，让该单重新进入自动派单队列
		TriggerDispatchRecheck(l.svcCtx)
	}
	return &types.EmptyResponse{}, nil
}

type AdminOrderPriorityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOrderPriorityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOrderPriorityLogic {
	return &AdminOrderPriorityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminOrderPriority 调整工单优先级（1 最低 ~ 5 最高），待派队列按优先级倒序取单。
func (l *AdminOrderPriorityLogic) AdminOrderPriority(req *types.AdminOrderPriorityRequest) (resp *types.EmptyResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if req.Priority < 1 || req.Priority > 5 {
		return nil, errs.BadRequest("优先级需在 1~5 之间")
	}
	affected, err := store.UpdateOrderPriority(l.ctx, l.svcCtx.DB, req.Id, req.Priority)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("只有待派单或已派单（未开工）的工单可以调整优先级")
	}
	return &types.EmptyResponse{}, nil
}