package seed

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
	"worker/internal/store"
)

const defaultPassword = "admin123"

type accountSpec struct {
	Username   string
	Name       string
	Role       int64
	BuildingID int64
	Buildings  []int64
}

var accounts = []accountSpec{
	{Username: "admin", Name: "超级管理员", Role: 1},
	{Username: "dorm1", Name: "1号宿舍管理员", Role: 3, BuildingID: 1},
	{Username: "dorm2", Name: "2号宿舍管理员", Role: 3, BuildingID: 2},
	{Username: "worker1", Name: "李工", Role: 2, Buildings: []int64{1, 2}},
	{Username: "worker2", Name: "王工", Role: 2, Buildings: []int64{2, 3}},
}

func Ensure(ctx context.Context, conn sqlx.SqlConn) error {
	var lastErr error
	for i := 0; i < 30; i++ {
		if err := ensureOnce(ctx, conn); err == nil {
			return nil
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return lastErr
}

func ensureOnce(ctx context.Context, conn sqlx.SqlConn) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	for _, spec := range accounts {
		existing, err := store.FindUserByUsername(ctx, conn, spec.Username)
		if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
			return err
		}
		var userID int64
		if existing == nil {
			user := &store.User{
				Username: spec.Username,
				Password: string(hash),
				Role:     spec.Role,
				Name:     spec.Name,
				Status:   1,
			}
			if spec.BuildingID > 0 {
				user.BuildingID = sql.NullInt64{Int64: spec.BuildingID, Valid: true}
			}
			userID, err = store.InsertUser(ctx, conn, user)
			if err != nil {
				return err
			}
		} else {
			userID = existing.ID
		}
		for _, b := range spec.Buildings {
			if err := store.InsertWorkerBuilding(ctx, conn, userID, b); err != nil {
				return err
			}
		}
	}
	return nil
}
