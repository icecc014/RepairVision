package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func Sign(secret string, expireSeconds int64, in Identity) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		ctxKeyUID:        in.UID,
		ctxKeyRole:       in.Role,
		ctxKeyBuildingID: in.BuildingID,
		ctxKeyUsername:   in.Username,
		ctxKeyName:       in.Name,
		"iat":            now.Unix(),
		"exp":            now.Add(time.Duration(expireSeconds) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
