package logic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type PublicRoomOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicRoomOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicRoomOrdersLogic {
	return &PublicRoomOrdersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// RoomOrders 公共查询：按楼栋 + 房间号返回该房间的报修进度（脱敏）。
func (l *PublicRoomOrdersLogic) RoomOrders(req *types.PublicRoomOrdersRequest) (*types.PublicRoomOrdersResponse, error) {
	room := strings.TrimSpace(req.Room)
	building, _, err := publicBuildingOf(l.ctx, l.svcCtx, req.BuildingId, room)
	if err != nil {
		return nil, err
	}
	orders, err := store.ListOrdersByRoom(l.ctx, l.svcCtx.DB, building.Id, room, 20)
	if err != nil {
		return nil, errs.Internal(err)
	}
	faultNames := map[string]string{}
	if faultTypes, err := store.ListFaultTypes(l.ctx, l.svcCtx.DB); err == nil {
		for _, ft := range faultTypes {
			faultNames[ft.Code] = ft.Name
		}
	}
	workerNames := map[int64]string{}
	if users, err := l.svcCtx.WorkerRpc.ListUsers(l.ctx, &workerclient.ListUsersRequest{Role: 2}); err == nil {
		for _, u := range users.Users {
			workerNames[u.Id] = u.Name
		}
	}
	resp := &types.PublicRoomOrdersResponse{
		BuildingName: building.Name,
		Room:         room,
		Orders:       make([]types.PublicRoomOrderItem, 0, len(orders)),
	}
	for _, o := range orders {
		if o.IsMerged == 2 {
			continue // 子单不再展示，避免同一报修出现两条
		}
		item := types.PublicRoomOrderItem{
			FaultTypeName: faultNames[o.FaultType],
			Description:   trimRunes(o.Description, 120),
			StatusText:    statusText(o.Status),
			CreatedAt:     o.CreatedAt.Format("2006-01-02 15:04"),
		}
		if o.CompletedAt.Valid {
			item.CompletedAt = o.CompletedAt.Time.Format("2006-01-02 15:04")
		}
		if o.WorkerID.Valid && o.WorkerID.Int64 > 0 {
			item.WorkerName = workerNames[o.WorkerID.Int64]
		}
		resp.Orders = append(resp.Orders, item)
	}
	return resp, nil
}

func trimRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "…"
}
