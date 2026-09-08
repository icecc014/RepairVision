package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type WorkerMapDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerMapDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerMapDataLogic {
	return &WorkerMapDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkerMapDataLogic) WorkerMapData() (resp *types.WorkerMapDataResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	idsResp, err := l.svcCtx.WorkerRpc.ListManagedBuildings(l.ctx, &workerclient.IdRequest{Id: identity.UID})
	if err != nil {
		return nil, errs.Upstream()
	}
	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	allowed := make(map[int64]struct{})
	for _, id := range idsResp.Ids {
		allowed[id] = struct{}{}
	}
	items := make([]types.WorkerMapBuilding, 0)
	for _, b := range buildingResp.Buildings {
		if _, ok := allowed[b.Id]; ok {
			items = append(items, types.WorkerMapBuilding{
				Id: b.Id, Code: b.Code, Name: b.Name, PosX: b.PosX, PosY: b.PosY,
				Width: b.Width, Height: b.Height, Floors: b.Floors,
				FloorHeight: b.FloorHeight, RoomsPerFloor: b.RoomsPerFloor,
			})
		}
	}
	orders, err := store.ListOrdersByBuildingIDs(l.ctx, l.svcCtx.DB, idsResp.Ids, true)
	if err != nil {
		return nil, errs.Internal(err)
	}
	orderItems, err := buildOrderItems(l.ctx, l.svcCtx, orders)
	if err != nil {
		return nil, err
	}
	return &types.WorkerMapDataResponse{Buildings: items, Orders: orderItems}, nil
}
