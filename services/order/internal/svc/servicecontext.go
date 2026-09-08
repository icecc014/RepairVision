package svc

import (
	"map/mapclient"
	"order/internal/config"
	"worker/workerclient"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	DB        sqlx.SqlConn
	WorkerRpc workerclient.Worker
	MapRpc    mapclient.Map
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		DB:        sqlx.NewMysql(c.DB.DataSource),
		WorkerRpc: workerclient.NewWorker(zrpc.MustNewClient(c.WorkerRpc)),
		MapRpc:    mapclient.NewMap(zrpc.MustNewClient(c.MapRpc)),
	}
}
