package ws

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/net/websocket"
)

type handlerClaims struct {
	UID        int64  `json:"uid"`
	Role       int64  `json:"role"`
	BuildingID int64  `json:"bid"`
	Name       string `json:"name"`
}

func Handler(hub *Hub, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenText := r.URL.Query().Get("token")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenText, &claims, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "无效的登录凭证")
			return
		}
		identity := handlerClaims{
			UID:        jsonNumber(claims["uid"]),
			Role:       jsonNumber(claims["role"]),
			BuildingID: jsonNumber(claims["bid"]),
		}
		if identity.UID <= 0 {
			writeError(w, http.StatusUnauthorized, "无效的登录凭证")
			return
		}
		keys := make([]string, 0, 3)
		if identity.Role == 1 {
			keys = append(keys, "all")
		} else if identity.Role == 3 {
			keys = append(keys, "building:"+intToString(identity.BuildingID))
		} else if identity.Role == 2 {
			keys = append(keys, "worker:"+intToString(identity.UID))
		} else {
			writeError(w, http.StatusForbidden, "当前角色不支持实时订阅")
			return
		}

		server := websocket.Server{
			Handshake: func(_ *websocket.Config, _ *http.Request) error { return nil },
			Handler: func(conn *websocket.Conn) {
				c := &client{
					conn: conn,
					send: make(chan []byte, 16),
					keys: keys,
				}
				hub.register(c)
				defer hub.unregister(c)
				go hub.pump(c)
				for {
					var msg string
					if err := websocket.Message.Receive(conn, &msg); err != nil {
						return
					}
				}
			},
		}
		server.ServeHTTP(w, r)
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "msg": msg, "data": nil})
}

func jsonNumber(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case json.Number:
		i, _ := t.Int64()
		return i
	default:
		return 0
	}
}
