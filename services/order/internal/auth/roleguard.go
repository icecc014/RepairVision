package auth

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"order/internal/errs"
)

func RoleGuard(roles ...int64) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok {
				httpx.ErrorCtx(r.Context(), w, errs.Unauthorized("登录状态无效，请重新登录"))
				return
			}
			for _, role := range roles {
				if identity.Role == role {
					next(w, r)
					return
				}
			}
			logx.WithContext(r.Context()).Errorf("forbidden access: uid=%d role=%d path=%s", identity.UID, identity.Role, r.URL.Path)
			httpx.ErrorCtx(r.Context(), w, errs.Forbidden("当前账号无权访问该资源"))
		}
	}
}
