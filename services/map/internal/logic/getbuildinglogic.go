package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/internal/store"
	"map/internal/svc"
	"map/map"
)

type GetBuildingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBuildingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBuildingLogic {
	return &GetBuildingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBuildingLogic) GetBuilding(in *_map.BuildingIdRequest) (*_map.BuildingResponse, error) {
	b, err := store.FindBuildingByID(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("楼栋不存在")
		}
		return nil, err
	}
	return &_map.BuildingResponse{Building: buildingToPb(*b)}, nil
}
