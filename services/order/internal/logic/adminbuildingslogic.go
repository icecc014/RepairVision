package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
)

type AdminBuildingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminBuildingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminBuildingsLogic {
	return &AdminBuildingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminBuildingsLogic) AdminBuildings() (resp *types.AdminBuildingListResponse, err error) {
	out, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	items := make([]types.AdminBuildingItem, 0, len(out.Buildings))
	for _, b := range out.Buildings {
		items = append(items, adminBuildingToItem(*b))
	}
	return &types.AdminBuildingListResponse{List: items}, nil
}
