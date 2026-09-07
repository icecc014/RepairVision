package logic

import (
	"context"

	"map/mapclient"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping(req *types.PingRequest) (resp *types.PingResponse, err error) {
	workerResp, err := l.svcCtx.WorkerRpc.Ping(l.ctx, &workerclient.Request{Ping: req.Name})
	if err != nil {
		return nil, err
	}
	mapResp, err := l.svcCtx.MapRpc.Ping(l.ctx, &mapclient.Request{Ping: req.Name})
	if err != nil {
		return nil, err
	}
	return &types.PingResponse{
		Worker: workerResp.Pong,
		Map:    mapResp.Pong,
	}, nil
}