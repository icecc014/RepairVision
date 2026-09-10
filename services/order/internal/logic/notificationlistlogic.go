package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type NotificationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationListLogic {
	return &NotificationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotificationListLogic) NotificationList(req *types.NotificationListRequest) (resp *types.NotificationListResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	rows, total, err := store.ListNotifications(l.ctx, l.svcCtx.DB, identity.UID, req.UnreadOnly == 1, req.Page, req.Size)
	if err != nil {
		return nil, errs.Internal(err)
	}
	unread, err := store.CountUnreadNotifications(l.ctx, l.svcCtx.DB, identity.UID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	resp = &types.NotificationListResponse{
		Total:  total,
		Unread: unread,
		List:   make([]types.NotificationItem, 0, len(rows)),
	}
	for _, n := range rows {
		item := types.NotificationItem{
			Id:        n.ID,
			Type:      n.Type,
			Title:     n.Title,
			Content:   n.Content,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if n.OrderID.Valid {
			item.OrderId = n.OrderID.Int64
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}
