package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"

	"order/internal/config"
	"order/internal/errs"
	"order/internal/handler"
	"order/internal/oplog"
	"order/internal/seed"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/ws"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-queue/kq"
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
	var logPusher oplog.Pusher
	if c.Kafka.Enabled {
		logx.Infof("kafka async op-log enabled, brokers=%v topic=%s", c.Kafka.Brokers, c.Kafka.Topic)
		logPusher = kq.NewPusher(c.Kafka.Brokers, c.Kafka.Topic, kq.WithAllowAutoTopicCreation())
		logReader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  c.Kafka.Brokers,
			GroupID:  c.Kafka.Group,
			Topic:    c.Kafka.Topic,
			MinBytes: 1e3,
			MaxBytes: 10e6,
		})
		defer logReader.Close()
		go func() {
			for {
				msg, err := logReader.FetchMessage(context.Background())
				if err != nil {
					logx.Error(err)
					continue
				}
				var logEntry store.OperationLog
				if err := json.Unmarshal(msg.Value, &logEntry); err != nil {
					logx.Error(err)
					continue
				}
				if err := store.InsertOperationLog(context.Background(), ctx.DB, &logEntry); err != nil {
					logx.Error(err)
				}
				if err := logReader.CommitMessages(context.Background(), msg); err != nil {
					logx.Error(err)
				}
			}
		}()
	}
	server.Use(oplog.Middleware(ctx.DB, logPusher, c.Kafka.Enabled))
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	// WebSocket 需要 Hijacker，go-zero rest 包装不支持，故由独立 HTTP 端口承载
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws/orders", ws.Handler(ctx.WS, c.Auth.AccessSecret))
	wsServer := &http.Server{Addr: ":8890", Handler: wsMux}
	go func() {
		fmt.Printf("Starting ws server at :8890...\n")
		if err := wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Must(err)
		}
	}()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
