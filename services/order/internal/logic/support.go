package logic

import (
	"context"
	"strings"

	"map/mapclient"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

func statusText(status int64) string {
	switch status {
	case store.StatusPending:
		return "待派单"
	case store.StatusDispatched:
		return "已派单"
	case store.StatusWorking:
		return "维修中"
	case store.StatusCompleted:
		return "已完成"
	case store.StatusCanceled:
		return "已取消"
	default:
		return "未知"
	}
}

func buildOrderItems(ctx context.Context, svcCtx *svc.ServiceContext, orders []store.Order) ([]types.OrderItem, error) {
	items := make([]types.OrderItem, 0, len(orders))
	if len(orders) == 0 {
		return items, nil
	}

	buildingIDs := map[int64]struct{}{}
	userIDs := map[int64]struct{}{}
	for i := range orders {
		buildingIDs[orders[i].BuildingID] = struct{}{}
		if orders[i].ReporterID > 0 {
			userIDs[orders[i].ReporterID] = struct{}{}
		}
		if orders[i].WorkerID.Valid && orders[i].WorkerID.Int64 > 0 {
			userIDs[orders[i].WorkerID.Int64] = struct{}{}
		}
	}

	buildingNames := make(map[int64]string)
	userNames := make(map[int64]string)
	faultNames := make(map[string]string)

	orderIDList := make([]int64, 0, len(orders))
	for i := range orders {
		orderIDList = append(orderIDList, orders[i].ID)
	}
	dispatchMap, err := store.ListDispatchSummaries(ctx, svcCtx.DB, orderIDList)
	if err != nil {
		return nil, err
	}

	bidList := make([]int64, 0, len(buildingIDs))
	for id := range buildingIDs {
		bidList = append(bidList, id)
	}
	if len(bidList) > 0 {
		buildingResp, err := svcCtx.MapRpc.ListBuildings(ctx, &mapclient.ListBuildingsRequest{})
		if err != nil {
			return nil, errs.Upstream()
		}
		for _, b := range buildingResp.Buildings {
			buildingNames[b.Id] = b.Name
		}
	}

	uidList := make([]int64, 0, len(userIDs))
	for id := range userIDs {
		uidList = append(uidList, id)
	}
	if len(uidList) > 0 {
		userResp, err := svcCtx.WorkerRpc.GetUsers(ctx, &workerclient.UserIdsRequest{Ids: uidList})
		if err != nil {
			return nil, errs.Upstream()
		}
		for _, u := range userResp.Users {
			userNames[u.Id] = u.Name
		}
	}

	faultTypes, err := store.ListFaultTypes(ctx, svcCtx.DB)
	if err != nil {
		return nil, err
	}
	for _, ft := range faultTypes {
		faultNames[ft.Code] = ft.Name
	}

	for i := range orders {
		o := orders[i]
		item := types.OrderItem{
			Id:            o.ID,
			OrderNo:       o.OrderNo,
			Title:         o.Title,
			Description:   o.Description,
			BuildingId:    o.BuildingID,
			BuildingName:  buildingNames[o.BuildingID],
			Room:          o.Room,
			Floor:         o.Floor,
			FaultType:     o.FaultType,
			FaultTypeName: faultNames[o.FaultType],
			Status:        o.Status,
			StatusText:    statusText(o.Status),
			ReporterId:    o.ReporterID,
			ReporterName:  userNames[o.ReporterID],
			Source:        o.Source,
			CreatedAt:     o.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     o.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if o.WorkerID.Valid {
			item.WorkerId = o.WorkerID.Int64
			item.WorkerName = userNames[o.WorkerID.Int64]
		}
		if ds, ok := dispatchMap[o.ID]; ok {
			item.DispatchScore = ds.Score
			item.SkillScore = ds.SkillScore
			item.DistanceScore = ds.DistanceScore
			item.LoadScore = ds.LoadScore
		}
		items = append(items, item)
	}
	return items, nil
}

func rpcBizError(err error) *errs.Error {
	msg := err.Error()
	if idx := strings.LastIndex(msg, "desc = "); idx >= 0 {
		msg = msg[idx+len("desc = "):]
	}
	msg = strings.TrimSpace(msg)
	if strings.Contains(msg, "已存在") {
		return errs.Conflict(msg)
	}
	if msg != "" {
		return errs.BadRequest(msg)
	}
	return errs.Upstream()
}

func roleText(role int64) string {
	switch role {
	case 1:
		return "超级管理员"
	case 2:
		return "维修工人"
	case 3:
		return "宿管"
	default:
		return "未知"
	}
}

func userStatusText(status int64) string {
	if status == 1 {
		return "启用"
	}
	return "停用"
}

func adminBuildingToItem(b mapclient.Building) types.AdminBuildingItem {
	return types.AdminBuildingItem{
		Id:            b.Id,
		Code:          b.Code,
		Name:          b.Name,
		PosX:          b.PosX,
		PosY:          b.PosY,
		Width:         b.Width,
		Height:        b.Height,
		Floors:        b.Floors,
		FloorHeight:   b.FloorHeight,
		RoomsPerFloor: b.RoomsPerFloor,
	}
}
