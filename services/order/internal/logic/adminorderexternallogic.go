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
	"order/internal/ws"
)

type AdminOrderExternalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOrderExternalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOrderExternalLogic {
	return &AdminOrderExternalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminOrderExternal 管理员把待处置工单标记为"外援处理"：不参与自动派单，仅记录状态。
func (l *AdminOrderExternalLogic) AdminOrderExternal(req *types.OrderIdRequest) (resp *types.EmptyResponse, err error) {
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
	if order.Status != store.StatusPending {
		return nil, errs.Conflict("只有待派单的工单可以标记外援")
	}
	affected, err := store.MarkOrderExternal(l.ctx, l.svcCtx.DB, order.ID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("工单状态已变化，标记外援失败")
	}

	recipients := append([]int64{order.ReporterID}, adminUserIDs(l.ctx, l.svcCtx)...)
	notifyUsers(l.ctx, l.svcCtx, recipients, "dispatch",
		"工单已转外援 "+order.OrderNo, order.Room+"室 已标记为外援处理", order.ID)
	l.svcCtx.WS.PublishOrder(ws.OrderEvent{
		Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
		BuildingId: order.BuildingID, Status: store.StatusPending,
	})
	return &types.EmptyResponse{}, nil
}