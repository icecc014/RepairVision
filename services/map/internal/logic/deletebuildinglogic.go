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

type DeleteBuildingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteBuildingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBuildingLogic {
	return &DeleteBuildingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteBuildingLogic) DeleteBuilding(in *_map.BuildingIdRequest) (*_map.Response, error) {
	if _, err := store.FindBuildingByID(l.ctx, l.svcCtx.DB, in.Id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("楼栋不存在")
		}
		return nil, err
	}
	if err := store.DeleteBuilding(l.ctx, l.svcCtx.DB, in.Id); err != nil {
		return nil, err
	}
	return &_map.Response{Pong: "ok"}, nil
}
