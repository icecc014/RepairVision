package oplog

import (
	"context"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
	"order/internal/auth"
	"order/internal/store"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func Middleware(conn sqlx.SqlConn) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/ping") || strings.HasPrefix(r.URL.Path, "/ws/") {
				next(w, r)
				return
			}
			start := time.Now()
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
			}
			recorder := &statusRecorder{ResponseWriter: w}
			next(recorder, r)
			if recorder.status == 0 {
				recorder.status = http.StatusOK
			}
			log := &store.OperationLog{
				OpID:         uuid.NewString(),
				Method:       r.Method,
				Path:         r.URL.Path,
				RequestBody:  redact(r.URL.Path, string(bodyBytes)),
				ResponseCode: int64(recorder.status),
				IP:           clientIP(r),
				CostMs:       time.Since(start).Milliseconds(),
				Module:       moduleOf(r.URL.Path),
				Action:       actionOf(r.Method, r.URL.Path),
			}
			if identity, ok := auth.IdentityFromContext(r.Context()); ok {
				log.UserID.Valid = true
				log.UserID.Int64 = identity.UID
				log.Username.Valid = true
				log.Username.String = identity.Username
				log.Role.Valid = true
				log.Role.Int64 = identity.Role
			}
			if err := store.InsertOperationLog(context.Background(), conn, log); err != nil {
				logx.WithContext(r.Context()).Errorf("write operation log failed: %v", err)
			}
		}
	}
}

var passwordPattern = regexp.MustCompile(`"password"\s*:\s*"[^"]*"`)

func redact(path, body string) string {
	if strings.Contains(path, "login") || strings.Contains(path, "reset-password") {
		return passwordPattern.ReplaceAllString(body, `"password":"***"`)
	}
	return body
}

func moduleOf(path string) string {
	switch {
	case strings.Contains(path, "/admin/users"), strings.Contains(path, "/api/login"):
		return "user"
	case strings.Contains(path, "/admin/buildings"):
		return "building"
	case strings.Contains(path, "/admin/fault-types"), strings.Contains(path, "/fault-types"):
		return "fault"
	case strings.Contains(path, "/dispatch-rules"):
		return "dispatch"
	case strings.Contains(path, "/dorm/orders"), strings.Contains(path, "/worker/orders"), strings.Contains(path, "/admin/orders"):
		return "order"
	default:
		return "order"
	}
}

func actionOf(method, path string) string {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "/login") {
		return "login"
	}
	if strings.Contains(lower, "/reset-password") {
		return "reset_password"
	}
	if strings.Contains(lower, "/cancel") {
		return "cancel"
	}
	if strings.Contains(lower, "/start") {
		return "start"
	}
	if strings.Contains(lower, "/complete") {
		return "complete"
	}
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "query"
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
