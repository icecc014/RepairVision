package handler

import (
	"net/http"

	"order/internal/auth"
	"order/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	jwtOpt := rest.WithJwt(serverCtx.Config.Auth.AccessSecret)

	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/ping/:name",
			Handler: PingHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/login",
			Handler: LoginHandler(serverCtx),
		},
	})

	// 任意登录角色可用的基础数据
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/fault-types",
			Handler: FaultTypesHandler(serverCtx),
		},
	}, jwtOpt)

	// 宿管（role=3）：本栋工单管理
	server.AddRoutes(rest.WithMiddleware(auth.RoleGuard(3), []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/dorm/orders",
			Handler: DormOrdersHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/dorm/orders/:id",
			Handler: DormOrderDetailHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/dorm/orders",
			Handler: CreateOrderHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/dorm/orders/:id/cancel",
			Handler: CancelOrderHandler(serverCtx),
		},
	}...), jwtOpt)

	// 工人（role=2）：我的工单与开工/完工
	server.AddRoutes(rest.WithMiddleware(auth.RoleGuard(2), []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/worker/orders",
			Handler: WorkerOrdersHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/worker/orders/:id/start",
			Handler: StartOrderHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/worker/orders/:id/complete",
			Handler: CompleteOrderHandler(serverCtx),
		},
	}...), jwtOpt)

	// 管理员（role=1）：只读工单总览
	server.AddRoutes(rest.WithMiddleware(auth.RoleGuard(1), []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/orders",
			Handler: AdminOrdersHandler(serverCtx),
		},
	}...), jwtOpt)
}
