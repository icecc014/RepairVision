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

type workerSkillSpec struct {
	Username    string
	Skill       string
	Proficiency int64
}

var accounts = []accountSpec{
	{Username: "admin", Name: "超级管理员", Role: 1},
	{Username: "dorm1", Name: "1号宿舍管理员", Role: 3, BuildingID: 1},
	{Username: "dorm2", Name: "2号宿舍管理员", Role: 3, BuildingID: 2},
	{Username: "dorm3", Name: "3号宿舍管理员", Role: 3, BuildingID: 3},
	{Username: "dorm4", Name: "4号宿舍管理员", Role: 3, BuildingID: 4},
	{Username: "dorm5", Name: "5号宿舍管理员", Role: 3, BuildingID: 5},
	{Username: "dorm6", Name: "6号宿舍管理员", Role: 3, BuildingID: 6},
	{Username: "dorm7", Name: "7号宿舍管理员", Role: 3, BuildingID: 7},
	{Username: "dorm8", Name: "8号宿舍管理员", Role: 3, BuildingID: 8},
	{Username: "dorm9", Name: "9号宿舍管理员", Role: 3, BuildingID: 9},
	{Username: "dorm10", Name: "10号宿舍管理员", Role: 3, BuildingID: 10},
	{Username: "dorm11", Name: "11号宿舍管理员", Role: 3, BuildingID: 11},
	{Username: "dorm12", Name: "12号宿舍管理员", Role: 3, BuildingID: 12},
	{Username: "dorm13", Name: "13号宿舍管理员", Role: 3, BuildingID: 13},
	{Username: "dorm14", Name: "14号宿舍管理员", Role: 3, BuildingID: 14},
	{Username: "dorm15", Name: "15号宿舍管理员", Role: 3, BuildingID: 15},
	{Username: "dorm16", Name: "16号宿舍管理员", Role: 3, BuildingID: 16},
	{Username: "dorm17", Name: "17号教学楼管理员", Role: 3, BuildingID: 17},
	{Username: "dorm18", Name: "18号教学楼管理员", Role: 3, BuildingID: 18},
	{Username: "dorm19", Name: "19号教学楼管理员", Role: 3, BuildingID: 19},
	{Username: "dorm20", Name: "20号教学楼管理员", Role: 3, BuildingID: 20},
	{Username: "dorm21", Name: "21号教学楼管理员", Role: 3, BuildingID: 21},
	{Username: "dorm22", Name: "22号教学楼管理员", Role: 3, BuildingID: 22},
	{Username: "dorm23", Name: "23号教学楼管理员", Role: 3, BuildingID: 23},
	{Username: "dorm24", Name: "24号教学楼管理员", Role: 3, BuildingID: 24},
	{Username: "dorm25", Name: "25号教学楼管理员", Role: 3, BuildingID: 25},
	{Username: "dorm26", Name: "26号教学楼管理员", Role: 3, BuildingID: 26},
}

var workerSkills = []workerSkillSpec{
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
	userIDs := make(map[string]int64, len(accounts))
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
			if err := store.SetUserBaseBuildingIfEmpty(ctx, conn, userID, spec.BuildingID); err != nil {
				return err
			}
		}
		userIDs[spec.Username] = userID
		for _, b := range spec.Buildings {
			if err := store.InsertWorkerBuilding(ctx, conn, userID, b); err != nil {
				return err
			}
		}
	}
	skillIDs := make(map[string]int64)
	for _, ws := range workerSkills {
		skillID, exists := skillIDs[ws.Skill]
		if !exists {
			skillID, err = store.EnsureSkill(ctx, conn, ws.Skill)
			if err != nil {
				return err
			}
			skillIDs[ws.Skill] = skillID
		}
		workerID, ok := userIDs[ws.Username]
		if !ok {
			continue
		}
		if err := store.EnsureWorkerSkill(ctx, conn, workerID, skillID, ws.Proficiency); err != nil {
			return err
		}
	}
	return nil
}
