package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/internal/store"
	"map/internal/svc"
	"map/map"
)

type ListBuildingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBuildingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBuildingsLogic {
	return &ListBuildingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListBuildingsLogic) ListBuildings(in *_map.ListBuildingsRequest) (*_map.ListBuildingsResponse, error) {
	buildings, err := store.ListBuildings(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	resp := &_map.ListBuildingsResponse{}
	for _, b := range buildings {
		resp.Buildings = append(resp.Buildings, &_map.Building{
			Id:            b.ID,
			Code:          b.Code,
			Name:          b.Name,
			PosX:          b.PosX,
			PosY:          b.PosY,
			Width:         b.Width,
			Height:        b.Height,
			Floors:        b.Floors,
			FloorHeight:   b.FloorHeight,
			RoomsPerFloor: b.RoomsPerFloor,
			LayoutJson:    b.LayoutJson,
		})
	}
	return resp, nil
}
