package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminStatsLogic {
	return &AdminStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminStatsLogic) AdminStats() (resp *types.AdminStatsResponse, err error) {
	statusRows, err := store.CountOrdersByStatus(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	statusMap := map[int64]int64{}
	for _, row := range statusRows {
		statusMap[row.Status] = row.Cnt
	}
	statusItems := make([]types.StatsStatusItem, 0, 5)
	for status := int64(1); status <= 5; status++ {
		statusItems = append(statusItems, types.StatsStatusItem{
			Status: status, StatusText: statusText(status), Count: statusMap[status],
		})
	}

	buildingRows, err := store.CountOrdersByBuilding(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	names := make(map[int64]string)
	for _, b := range buildingResp.Buildings {
		names[b.Id] = b.Name
	}
	buildingItems := make([]types.StatsBuildingItem, 0, len(buildingRows))
	for _, row := range buildingRows {
		buildingItems = append(buildingItems, types.StatsBuildingItem{
			BuildingId: row.BuildingID, BuildingName: names[row.BuildingID], Count: row.Cnt,
		})
	}

	faultRows, err := store.CountOrdersByFaultType(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	faultTypes, err := store.ListFaultTypes(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	faultNames := make(map[string]string)
	for _, f := range faultTypes {
		faultNames[f.Code] = f.Name
	}
	faultItems := make([]types.StatsFaultItem, 0, len(faultRows))
	for _, row := range faultRows {
		faultItems = append(faultItems, types.StatsFaultItem{
			FaultType: row.FaultType, FaultTypeName: faultNames[row.FaultType], Count: row.Cnt,
		})
	}

	days, err := store.CountOrdersLastDays(l.ctx, l.svcCtx.DB, 7)
	if err != nil {
		return nil, errs.Internal(err)
	}
	dayMap := make(map[string]int64)
	for _, d := range days {
		dayMap[d.Day] = d.Cnt
	}
	recent := make([]types.StatsDayItem, 0, 7)
	now := time.Now()
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		recent = append(recent, types.StatsDayItem{Date: day, Count: dayMap[day]})
	}
	return &types.AdminStatsResponse{
		Status: statusItems, Buildings: buildingItems, Faults: faultItems, Recent: recent,
	}, nil
}
