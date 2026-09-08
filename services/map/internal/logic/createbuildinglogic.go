package logic

import (
	"context"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"map/internal/store"
	"map/internal/svc"
	"map/map"
)

type CreateBuildingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateBuildingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBuildingLogic {
	return &CreateBuildingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateBuildingLogic) CreateBuilding(in *_map.SaveBuildingRequest) (*_map.BuildingResponse, error) {
	if in.Building == nil {
		return nil, errors.New("楼栋信息不能为空")
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
	id, err := store.CreateBuilding(l.ctx, l.svcCtx.DB, b)
	if err != nil {
		return nil, err
	}
	b.ID = id
	return &_map.BuildingResponse{Building: buildingToPb(*b)}, nil
}
