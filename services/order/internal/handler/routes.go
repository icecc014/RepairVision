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
		{Method: http.MethodGet, Path: "/ping/:name", Handler: PingHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/login", Handler: LoginHandler(serverCtx)},
	})
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/api/fault-types", Handler: FaultTypesHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/notifications", Handler: NotificationListHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/notifications/:id/read", Handler: NotificationReadHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/notifications/read-all", Handler: NotificationReadAllHandler(serverCtx)},
	}, jwtOpt)

	server.AddRoutes(rest.WithMiddleware(auth.RoleGuard(3), []rest.Route{
		{Method: http.MethodGet, Path: "/api/dorm/orders", Handler: DormOrdersHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/dorm/orders/:id", Handler: DormOrderDetailHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/dorm/orders", Handler: CreateOrderHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/dorm/orders/:id/cancel", Handler: CancelOrderHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/dorm/orders/:id/feedback", Handler: DormFeedbackHandler(serverCtx)},
	}...), jwtOpt)

	server.AddRoutes(rest.WithMiddleware(auth.RoleGuard(2), []rest.Route{
		{Method: http.MethodGet, Path: "/api/worker/orders", Handler: WorkerOrdersHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/worker/map-data", Handler: WorkerMapDataHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/worker/schedules", Handler: WorkerSchedulesHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/worker/leaves", Handler: WorkerLeaveSubmitHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/worker/leaves", Handler: WorkerLeaveListHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/worker/leaves/:id/cancel", Handler: WorkerLeaveCancelHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/worker/batch-complete", Handler: WorkerBatchCompleteHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/worker/orders/:id/start", Handler: StartOrderHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/worker/orders/:id/complete", Handler: CompleteOrderHandler(serverCtx)},
	}...), jwtOpt)

	server.AddRoutes(rest.WithMiddleware(auth.RoleGuard(1), []rest.Route{
		{Method: http.MethodGet, Path: "/api/admin/orders", Handler: AdminOrdersHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/schedules", Handler: AdminSchedulesHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/worker-board", Handler: AdminWorkerBoardHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/schedules", Handler: AdminScheduleSaveHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/schedules/generate", Handler: AdminScheduleGenerateHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/leaves", Handler: AdminLeaveListHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/leaves/:id/review", Handler: AdminLeaveReviewHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/stats", Handler: AdminStatsHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/feedback-stats", Handler: AdminFeedbackStatsHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/sla-overview", Handler: AdminSlaOverviewHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/orders/:id/reassign", Handler: AdminOrderReassignHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/orders/batch-dispatch", Handler: AdminBatchDispatchHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/users", Handler: AdminUsersHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/users", Handler: AdminUserCreateHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/api/admin/users/:id", Handler: AdminUserUpdateHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/users/:id/reset-password", Handler: AdminUserResetPasswordHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/admin/users/:id", Handler: AdminUserDeleteHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/buildings", Handler: AdminBuildingsHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/buildings", Handler: AdminBuildingCreateHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/api/admin/buildings/:id", Handler: AdminBuildingUpdateHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/admin/buildings/:id", Handler: AdminBuildingDeleteHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/fault-types", Handler: AdminFaultTypesHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/fault-types", Handler: AdminFaultTypeCreateHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/api/admin/fault-types/:id", Handler: AdminFaultTypeUpdateHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/admin/fault-types/:id", Handler: AdminFaultTypeDeleteHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/dispatch-rules", Handler: AdminDispatchRulesHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/api/admin/dispatch-rules", Handler: AdminDispatchRuleCreateHandler(serverCtx)},
		{Method: http.MethodPut, Path: "/api/admin/dispatch-rules/:id", Handler: AdminDispatchRuleUpdateHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/admin/logs", Handler: AdminOperationLogsHandler(serverCtx)},
	}...), jwtOpt)
}
