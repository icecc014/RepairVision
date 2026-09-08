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
