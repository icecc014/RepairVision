package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"

	"order/internal/config"
	"order/internal/errs"
	"order/internal/handler"
	"order/internal/oplog"
	"order/internal/seed"
	"order/internal/svc"
	"order/internal/ws"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/order-api.yaml", "the config file")

type envelope struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(c)
	if err := seed.Ensure(context.Background(), ctx.DB); err != nil {
		logx.Must(err)
	}

	httpx.SetOkHandler(func(_ context.Context, v any) any {
		return envelope{Code: 0, Msg: "ok", Data: v}
	})
	httpx.SetErrorHandlerCtx(func(_ context.Context, err error) (int, any) {
		var be *errs.Error
		if errors.As(err, &be) {
			return be.HTTPStatus, envelope{Code: be.HTTPStatus, Msg: be.Message}
		}
		if errors.Is(err, sqlx.ErrNotFound) {
			return http.StatusNotFound, envelope{Code: http.StatusNotFound, Msg: "数据不存在"}
		}
		logx.Error(err)
		return http.StatusInternalServerError, envelope{Code: http.StatusInternalServerError, Msg: "系统繁忙"}
	})

	server := rest.MustNewServer(c.RestConf)
	server.Use(oplog.Middleware(ctx.DB))
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws/orders",
		Handler: ws.Handler(ctx.WS, c.Auth.AccessSecret),
	})
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
