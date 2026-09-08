package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"order/internal/logic"
	"order/internal/svc"
)

func WorkerMapDataHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewWorkerMapDataLogic(r.Context(), svcCtx)
		resp, err := l.WorkerMapData()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
