package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"order/internal/errs"
	"order/internal/logic"
	"order/internal/svc"
	"order/internal/types"
)

func AdminWorkSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminWorkSettingsLogic(r.Context(), svcCtx)
		resp, err := l.AdminWorkSettings()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func AdminWorkSettingsUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WorkSettingsUpdateRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewAdminWorkSettingsUpdateLogic(r.Context(), svcCtx)
		resp, err := l.AdminWorkSettingsUpdate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}