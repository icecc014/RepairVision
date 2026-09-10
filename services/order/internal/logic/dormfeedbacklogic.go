package logic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
)

type DormFeedbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDormFeedbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DormFeedbackLogic {
	return &DormFeedbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DormFeedbackLogic) DormFeedback(req *types.DormFeedbackRequest) (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if req.Rating < 1 || req.Rating > 5 {
		return nil, errs.BadRequest("评价星级需在1-5之间")
	}
	order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.BadRequest("工单不存在")
		}
		return nil, errs.Internal(err)
	}
	if order.BuildingID != identity.BuildingID {
		return nil, errs.Forbidden("只能评价本栋楼工单")
	}
	if order.Status != store.StatusCompleted {
		return nil, errs.Conflict("仅已完工工单可评价")
	}
	if _, err := store.FindFeedbackByOrderID(l.ctx, l.svcCtx.DB, order.ID); err == nil {
		return nil, errs.Conflict("该工单已评价")
	} else if !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errs.Internal(err)
	}
	workerID := sql.NullInt64{}
	if order.WorkerID.Valid && order.WorkerID.Int64 > 0 {
		workerID = order.WorkerID
	}
	comment := strings.TrimSpace(req.Comment)
	if len([]rune(comment)) > 500 {
		return nil, errs.BadRequest("评价内容不能超过500字")
	}
	if err := store.InsertFeedback(l.ctx, l.svcCtx.DB, order.ID, order.BuildingID, workerID, req.Rating, comment); err != nil {
		return nil, errs.Internal(err)
	}
	recipients := adminUserIDs(l.ctx, l.svcCtx)
	workerEventID := int64(0)
	if order.WorkerID.Valid {
		recipients = append(recipients, order.WorkerID.Int64)
		workerEventID = order.WorkerID.Int64
	}
	notifyUsers(l.ctx, l.svcCtx, recipients, "feedback",
		"收到服务评价 "+order.OrderNo, order.Room+"室 收到新的服务评价", order.ID)
	l.svcCtx.WS.PublishOrder(ws.OrderEvent{
		Type: "order_changed", OrderId: order.ID, OrderNo: order.OrderNo,
		BuildingId: order.BuildingID, WorkerId: workerEventID, Status: order.Status,
	})
	return &types.EmptyResponse{}, nil
}
