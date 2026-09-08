package logic

import (
	"context"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/internal/store"
	"map/internal/svc"
	"map/map"
)

type UpdateBuildingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBuildingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBuildingLogic {
	return &UpdateBuildingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateBuildingLogic) UpdateBuilding(in *_map.SaveBuildingRequest) (*_map.BuildingResponse, error) {
	if in.Building == nil || in.Building.Id <= 0 {
		return nil, errors.New("楼栋ID不能为空")
	}
	b := pbToBuilding(in.Building)
	b.Code = strings.TrimSpace(b.Code)
	b.Name = strings.TrimSpace(b.Name)
	if b.Code == "" || b.Name == "" {
		return nil, errors.New("楼栋编码和名称不能为空")
	}
	if b.Floors <= 0 || b.RoomsPerFloor <= 0 {
		return nil, errors.New("楼层数和每层房间数必须大于0")
	}
	if _, err := store.FindBuildingByID(l.ctx, l.svcCtx.DB, b.ID); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("楼栋不存在")
		}
		return nil, err
	}
	if err := store.UpdateBuilding(l.ctx, l.svcCtx.DB, b); err != nil {
		return nil, err
	}
	return &_map.BuildingResponse{Building: buildingToPb(*b)}, nil
}
