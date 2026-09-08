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

	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/fault-types",
			Handler: FaultTypesHandler(serverCtx),
		},
	}, jwtOpt)

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

	server.AddRoutes(rest.WithMiddleware(auth.RoleGuard(1), []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/orders",
			Handler: AdminOrdersHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/fault-types",
			Handler: AdminFaultTypesHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/fault-types",
			Handler: AdminFaultTypeCreateHandler(serverCtx),
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/fault-types/:id",
			Handler: AdminFaultTypeUpdateHandler(serverCtx),
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/fault-types/:id",
			Handler: AdminFaultTypeDeleteHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/dispatch-rules",
			Handler: AdminDispatchRulesHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/dispatch-rules",
			Handler: AdminDispatchRuleCreateHandler(serverCtx),
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/dispatch-rules/:id",
			Handler: AdminDispatchRuleUpdateHandler(serverCtx),
		},
	}...), jwtOpt)
}
