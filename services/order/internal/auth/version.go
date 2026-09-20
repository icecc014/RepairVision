package auth

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// V9.8 方案A2：密码版本号。
// 改密成功后版本 +1，使已签发的旧 token 失效（受缓存生效延迟影响，通常 <1s）。
// 版本号存于 Redis（order-api 的 PubCache）；未注入缓存时所有校验失败开放，避免锁死全站。

// VersionStore 是版本号所需的最小缓存能力（*pubcache.Store 满足）。
type VersionStore interface {
	Get(ctx context.Context, key string) (string, bool)
	Incr(ctx context.Context, key string, ttl time.Duration) int64
}

var pwdVerStore VersionStore

// 版本号键保留一年：既避免永久脏键，也远长于任何 token 有效期。
const pwdVerTTL = 365 * 24 * time.Hour

// UsePwdVersionStore 由服务启动时注入缓存句柄。
func UsePwdVersionStore(s VersionStore) { pwdVerStore = s }

// PwdVersionKey 返回某用户的版本号缓存键。
func PwdVersionKey(uid int64) string { return "pwdver:" + strconv.FormatInt(uid, 10) }

// CurrentPwdVersion 读取当前版本号，缺省 0。
func CurrentPwdVersion(ctx context.Context, s VersionStore, uid int64) int64 {
	if s == nil || uid <= 0 {
		return 0
	}
	if raw, ok := s.Get(ctx, PwdVersionKey(uid)); ok {
		if n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); err == nil {
			return n
		}
	}
	return 0
}

// BumpPwdVersion 在改密成功后递增版本号。
func BumpPwdVersion(ctx context.Context, s VersionStore, uid int64) {
	if s == nil || uid <= 0 {
		return
	}
	s.Incr(ctx, PwdVersionKey(uid), pwdVerTTL)
}

// pwdVersionOK 校验 token 内的版本号是否仍有效。
// 以下情况一律放行（失败开放）：未注入缓存、系统身份、旧 token 无版本号。
func pwdVersionOK(ctx context.Context, id *Identity) bool {
	if pwdVerStore == nil || id == nil || id.UID <= 0 {
		return true
	}
	tokenVer, ok := toInt64(ctx.Value(ctxKeyPwdVer))
	if !ok {
		return true
	}
	return tokenVer >= CurrentPwdVersion(ctx, pwdVerStore, id.UID)
}