package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
)

type PublicBuildingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicBuildingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicBuildingsLogic {
	return &PublicBuildingsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Buildings 公开楼栋列表（id / code / name / 层数 / 每层房间数，供公共报修下拉框与房间范围提示）。
func (l *PublicBuildingsLogic) Buildings() (*types.PublicBuildingListResponse, error) {
	resp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	out := &types.PublicBuildingListResponse{List: make([]types.PublicBuildingItem, 0, len(resp.Buildings))}
	for _, b := range resp.Buildings {
		out.List = append(out.List, types.PublicBuildingItem{
			Id: b.Id, Code: b.Code, Name: b.Name,
			Floors: b.Floors, RoomsPerFloor: b.RoomsPerFloor,
		})
	}
	return out, nil
}
