package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"order/internal/errs"
	"order/internal/logic"
	"order/internal/svc"
	"order/internal/types"
)

func WorkerLeaveSubmitHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WorkerLeaveSubmitRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewWorkerLeaveSubmitLogic(r.Context(), svcCtx)
		resp, err := l.WorkerLeaveSubmit(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func WorkerLeaveListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WorkerLeaveListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewWorkerLeaveListLogic(r.Context(), svcCtx)
		resp, err := l.WorkerLeaveList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func WorkerLeaveCancelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LeaveIdRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewWorkerLeaveCancelLogic(r.Context(), svcCtx)
		resp, err := l.WorkerLeaveCancel(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func AdminLeaveListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminLeaveListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewAdminLeaveListLogic(r.Context(), svcCtx)
		resp, err := l.AdminLeaveList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func AdminLeaveReviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LeaveReviewRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewAdminLeaveReviewLogic(r.Context(), svcCtx)
		resp, err := l.AdminLeaveReview(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
