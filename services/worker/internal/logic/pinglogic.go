package logic

import (
	"context"

	"worker/internal/svc"
	"worker/worker"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(in *worker.Request) (*worker.Response, error) {
	// todo: add your logic here and delete this line

	return &worker.Response{Pong: "worker-rpc:" + in.Ping}, nil
}
