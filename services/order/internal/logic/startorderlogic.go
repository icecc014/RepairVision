package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
	"worker/workerclient"
)

type StartOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartOrderLogic {
	return &StartOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// StartOrder 工人开工。派单与工单生成不受工作时段限制，但开工动作受时段约束：
// 非工作时段默认提示并拦截，工人确认后带 force=true 可强制开工（是否允许由配置决定）。
func (l *StartOrderLogic) StartOrder(req *types.StartOrderRequest) (resp *types.StartOrderResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}

	warning := ""
	if status, serr := l.svcCtx.WorkerRpc.GetDutyStatus(l.ctx, &workerclient.DutyStatusRequest{WorkerId: identity.UID}); serr == nil && status != nil && !status.OnDuty {
		period := status.Morning + " / " + status.Afternoon
		if !req.Force {
			return nil, errs.Conflict("当前不在工作时段（" + period + "）：" + status.Reason + "；确认后仍可强制开工")
		}
		if s, err := l.svcCtx.WorkerRpc.GetWorkSettings(l.ctx, &workerclient.GetWorkSettingsRequest{}); err == nil && s.GetSettings().GetAllowForceStart() == 0 {
			return nil, errs.Conflict("系统已禁止非工作时段开工：" + status.Reason)
		}
		warning = "已在非工作时段强制开工（" + status.Reason + "）"
	}

	affected, err := store.StartOrder(l.ctx, l.svcCtx.DB, req.Id, identity.UID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if !affected {
		return nil, errs.Conflict("工单不存在或当前状态不可开工")
	}
	if err := store.AcceptDispatchRecord(l.ctx, l.svcCtx.DB, req.Id, identity.UID); err != nil {
		logx.WithContext(l.ctx).Errorf("mark dispatch accepted failed order=%d worker=%d: %v", req.Id, identity.UID, err)
	}
	if order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id); err == nil {
		l.svcCtx.WS.PublishOrder(ws.OrderEvent{
			Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
			BuildingId: order.BuildingID, WorkerId: identity.UID, Status: order.Status,
		})
		recipients := append([]int64{order.ReporterID}, adminUserIDs(l.ctx, l.svcCtx)...)
		notifyUsers(l.ctx, l.svcCtx, recipients, "start",
			"工单已开工 "+order.OrderNo, order.Room+"室 工人已开工", order.ID)
	}
	return &types.StartOrderResponse{Warning: warning}, nil
}