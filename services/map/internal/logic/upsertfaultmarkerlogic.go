package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/internal/store"
	"map/internal/svc"
	"map/map"
)

type UpsertFaultMarkerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpsertFaultMarkerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertFaultMarkerLogic {
	return &UpsertFaultMarkerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpsertFaultMarkerLogic) UpsertFaultMarker(in *_map.FaultMarkerUpsertRequest) (*_map.Response, error) {
	if err := store.UpsertFaultMarker(l.ctx, l.svcCtx.DB, in.OrderId, in.BuildingId, in.Floor, in.RoomNumber); err != nil {
		return nil, err
	}
	return &_map.Response{Pong: "ok"}, nil
}
