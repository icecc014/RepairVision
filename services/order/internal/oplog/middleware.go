package oplog

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

type Pusher interface {
	Push(ctx context.Context, v string) error
}

func Middleware(conn sqlx.SqlConn, pusher Pusher, useKafka bool) rest.Middleware {
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
			if err := appendFileLog(log); err != nil {
				logx.WithContext(r.Context()).Errorf("write file operation log failed: %v", err)
			}
			insertCtx := context.Background()
			if useKafka && pusher != nil {
				body, _ := json.Marshal(log)
				if pushErr := pusher.Push(insertCtx, string(body)); pushErr == nil {
					return
				} else {
					logx.WithContext(r.Context()).Errorf("push operation log to kafka failed: %v", pushErr)
				}
			}
			if err := store.InsertOperationLog(insertCtx, conn, log); err != nil {
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

var (
	fileLogOnce sync.Once
	fileLogDir  string
	fileLogErr  error
)

// appendFileLog 把技术日志以 JSON Lines 写入服务端本地文件（按天切分）。
// 目录默认 /app/logs（容器内），可通过 OPLOG_DIR 覆盖；失败时仅记录，不影响业务。
func appendFileLog(log interface{}) error {
	fileLogOnce.Do(func() {
		fileLogDir = os.Getenv("OPLOG_DIR")
		if fileLogDir == "" {
			fileLogDir = "/app/logs"
		}
		if err := os.MkdirAll(fileLogDir, 0o755); err != nil {
			fileLogErr = err
		}
	})
	if fileLogErr != nil {
		return fileLogErr
	}
	body, err := json.Marshal(log)
	if err != nil {
		return err
	}
	name := "operation-" + time.Now().Format("20060102") + ".log"
	file, err := os.OpenFile(filepath.Join(fileLogDir, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(body, '\n')); err != nil {
		return err
	}
	return nil
}
