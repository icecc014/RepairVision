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

func (l *WorkerMapDataLogic) WorkerMapData(req *types.WorkerMapDataRequest) (resp *types.WorkerMapDataResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	// V5.4+：工人端只展示"实际派给他的工单所在的楼栋"（而不是管辖楼栋），
	// 管辖楼栋仍保留在账号资料中；这样工人一眼能看到今天真正要去哪几栋。
	days := req.Days
	if days <= 0 {
		days = 3
	}
	orders, err := store.ListActiveOrdersByWorkerWithinDays(l.ctx, l.svcCtx.DB, identity.UID, days)
	if err != nil {
		return nil, errs.Internal(err)
	}
	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	allowed := make(map[int64]struct{})
	for _, o := range orders {
		allowed[o.BuildingID] = struct{}{}
	}
	items := make([]types.WorkerMapBuilding, 0)
	for _, b := range buildingResp.Buildings {
		if _, ok := allowed[b.Id]; ok {
			items = append(items, types.WorkerMapBuilding{
				Id: b.Id, Code: b.Code, Name: b.Name, PosX: b.PosX, PosY: b.PosY,
				Width: b.Width, Height: b.Height, Floors: b.Floors,
				FloorHeight: b.FloorHeight, RoomsPerFloor: b.RoomsPerFloor, LayoutJson: b.LayoutJson,
			})
		}
	}
	// 工单列表同样是"该工人自己的活动工单"（与楼栋保持一致）
	orderItems, err := buildOrderItems(l.ctx, l.svcCtx, orders)
	if err != nil {
		return nil, err
	}
	return &types.WorkerMapDataResponse{Buildings: items, Orders: orderItems}, nil
}
