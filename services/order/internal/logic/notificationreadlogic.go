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

type NotificationReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationReadLogic {
	return &NotificationReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotificationReadLogic) NotificationRead(req *types.NotificationIdRequest) (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if _, err := store.MarkNotificationRead(l.ctx, l.svcCtx.DB, identity.UID, req.Id); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}

type NotificationReadAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNotificationReadAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationReadAllLogic {
	return &NotificationReadAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NotificationReadAllLogic) NotificationReadAll() (resp *types.EmptyResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if err := store.MarkAllNotificationsRead(l.ctx, l.svcCtx.DB, identity.UID); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
