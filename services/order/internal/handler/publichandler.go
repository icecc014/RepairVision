package handler

import (
	"net"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
	"order/internal/errs"
	"order/internal/logic"
	"order/internal/svc"
	"order/internal/types"
)

// V7 公共报修接口（免 JWT，任何人可访问）。

func PublicCaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewPublicCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.Captcha()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func PublicBuildingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewPublicBuildingsLogic(r.Context(), svcCtx)
		resp, err := l.Buildings()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func PublicReportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublicReportRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		req.ClientIp = publicClientIp(r)
		l := logic.NewPublicReportLogic(r.Context(), svcCtx)
		resp, err := l.PublicReport(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func PublicRoomOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublicRoomOrdersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errs.BadRequest(err.Error()))
			return
		}
		l := logic.NewPublicRoomOrdersLogic(r.Context(), svcCtx)
		resp, err := l.RoomOrders(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// publicClientIp 取真实客户端 IP（Nginx 已透传 X-Real-IP）。
func publicClientIp(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		parts := strings.Split(v, ",")
		return strings.TrimSpace(parts[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
