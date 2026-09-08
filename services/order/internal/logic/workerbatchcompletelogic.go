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
)

type WorkerBatchCompleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerBatchCompleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerBatchCompleteLogic {
	return &WorkerBatchCompleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkerBatchCompleteLogic) WorkerBatchComplete(req *types.BatchCompleteRequest) (resp *types.BatchCompleteResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if req.BuildingId <= 0 || req.FaultType == "" {
		return nil, errs.BadRequest("楼栋和维修类型不能为空")
	}
	ids, err := store.BatchCompleteOrders(l.ctx, l.svcCtx.DB, identity.UID, req.BuildingId, req.FaultType)
	if err != nil {
		return nil, errs.Internal(err)
	}
	for _, id := range ids {
		l.svcCtx.WS.PublishOrder(ws.OrderEvent{
			Type: "order_changed", OrderId: id, BuildingId: req.BuildingId,
			WorkerId: identity.UID, Status: store.StatusCompleted,
		})
	}
	return &types.BatchCompleteResponse{Count: int64(len(ids))}, nil
}
