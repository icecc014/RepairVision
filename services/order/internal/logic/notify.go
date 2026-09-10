package logic

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/store"
	"order/internal/svc"
	"worker/workerclient"
)

// adminUserIDs 返回全部启用管理员 ID。
func adminUserIDs(ctx context.Context, svcCtx *svc.ServiceContext) []int64 {
	resp, err := svcCtx.WorkerRpc.ListUsers(ctx, &workerclient.ListUsersRequest{Role: 1, Status: 1})
	if err != nil {
		logx.WithContext(ctx).Errorf("list admins for notify failed: %v", err)
		return nil
	}
	ids := make([]int64, 0, len(resp.Users))
	for _, u := range resp.Users {
		ids = append(ids, u.Id)
	}
	return ids
}

// notifyUsers 给指定用户写入通知（尽力而为，不阻断主流程）。
func notifyUsers(ctx context.Context, svcCtx *svc.ServiceContext, userIDs []int64, ntype, title, content string, orderID int64) {
	seen := make(map[int64]struct{})
	ids := make([]int64, 0, len(userIDs))
	for _, id := range userIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return
	}
	users, err := svcCtx.WorkerRpc.GetUsers(ctx, &workerclient.UserIdsRequest{Ids: ids})
	if err != nil {
		logx.WithContext(ctx).Errorf("get users for notify failed: %v", err)
		return
	}
	roleByID := make(map[int64]int64, len(users.Users))
	for _, u := range users.Users {
		roleByID[u.Id] = u.Role
	}
	orderRef := sql.NullInt64{}
	if orderID > 0 {
		orderRef = sql.NullInt64{Int64: orderID, Valid: true}
	}
	for _, id := range ids {
		n := &store.Notification{
			UserID:  id,
			Role:    roleByID[id],
			Type:    ntype,
			Title:   title,
			Content: content,
			OrderID: orderRef,
		}
		if err := store.InsertNotification(ctx, svcCtx.DB, n); err != nil {
			logx.WithContext(ctx).Errorf("insert notification failed user=%d: %v", id, err)
		}
	}
}
