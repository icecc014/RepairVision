package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/internal/store"
	"map/internal/svc"
	"map/map"
)

type RemoveFaultMarkerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveFaultMarkerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveFaultMarkerLogic {
	return &RemoveFaultMarkerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveFaultMarkerLogic) RemoveFaultMarker(in *_map.FaultMarkerOrderRequest) (*_map.Response, error) {
	if err := store.RemoveFaultMarker(l.ctx, l.svcCtx.DB, in.OrderId); err != nil {
		return nil, err
	}
	return &_map.Response{Pong: "ok"}, nil
}
