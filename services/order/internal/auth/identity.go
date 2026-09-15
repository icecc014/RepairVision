package auth

import (
	"context"
	"encoding/json"
)

const (
	ctxKeyUID        = "uid"
	ctxKeyRole       = "role"
	ctxKeyBuildingID = "bid"
	ctxKeyUsername   = "username"
	ctxKeyName       = "name"
)

type Identity struct {
	UID        int64
	Role       int64
	BuildingID int64
	Username   string
	Name       string
}

func IdentityFromContext(ctx context.Context) (*Identity, bool) {
	uid, ok := toInt64(ctx.Value(ctxKeyUID))
	if !ok {
		return nil, false
	}
	role, ok := toInt64(ctx.Value(ctxKeyRole))
	if !ok {
		return nil, false
	}
	buildingID, _ := toInt64(ctx.Value(ctxKeyBuildingID))
	username, _ := ctx.Value(ctxKeyUsername).(string)
	name, _ := ctx.Value(ctxKeyName).(string)
	return &Identity{
		UID:        uid,
		Role:       role,
		BuildingID: buildingID,
		Username:   username,
		Name:       name,
	}, true
}

// WithSystemIdentity 注入"系统身份"，供定时任务等内部调用复用与管理员相同的校验路径。
func WithSystemIdentity(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, ctxKeyUID, int64(0))
	ctx = context.WithValue(ctx, ctxKeyRole, int64(1))
	ctx = context.WithValue(ctx, ctxKeyUsername, "system")
	ctx = context.WithValue(ctx, ctxKeyName, "自动派单")
	return ctx
}
func toInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return 0, false
		}
		return i, true
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	default:
		return 0, false
	}
}
