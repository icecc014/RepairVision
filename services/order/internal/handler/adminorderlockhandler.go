package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"order/internal/errs"
	"order/internal/logic"
	"order/internal/svc"
	"order/internal/types"
)

func AdminOrderLockHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminOrderLockRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewAdminOrderLockLogic(r.Context(), svcCtx)
		resp, err := l.AdminOrderLock(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func AdminOrderPriorityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminOrderPriorityRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewAdminOrderPriorityLogic(r.Context(), svcCtx)
		resp, err := l.AdminOrderPriority(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}