package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"order/internal/logic"
	"order/internal/svc"
	"order/internal/types"
)

func WorkerBatchCompleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchCompleteRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewWorkerBatchCompleteLogic(r.Context(), svcCtx)
		resp, err := l.WorkerBatchComplete(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
