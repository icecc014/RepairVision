package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const userColumns = "id, username, password, role, name, phone, building_id, status"

type User struct {
	ID         int64          `db:"id"`
	Username   string         `db:"username"`
	Password   string         `db:"password"`
	Role       int64          `db:"role"`
	Name       string         `db:"name"`
	Phone      sql.NullString `db:"phone"`
	BuildingID sql.NullInt64  `db:"building_id"`
	Status     int64          `db:"status"`
}

func FindUserByUsername(ctx context.Context, conn sqlx.SqlConn, username string) (*User, error) {
	var u User
	if err := conn.QueryRowCtx(ctx, &u,
		"select "+userColumns+" from users where username = ? and status = 1", username); err != nil {
		return nil, err
	}
	return &u, nil
}

func FindUsersByIDs(ctx context.Context, conn sqlx.SqlConn, ids []int64) ([]User, error) {
	if len(ids) == 0 {
		return []User{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	var users []User
	if err := conn.QueryRowsCtx(ctx, &users,
		fmt.Sprintf("select "+userColumns+" from users where id in (%s)", strings.Join(placeholders, ",")), args...); err != nil {
		return nil, err
	}
	return users, nil
}

func ListWorkersByBuilding(ctx context.Context, conn sqlx.SqlConn, buildingID int64) ([]User, error) {
	var users []User
	if err := conn.QueryRowsCtx(ctx, &users,
		"select u."+userColumns+" from users u join worker_buildings wb on wb.worker_id = u.id where u.role = 2 and u.status = 1 and wb.building_id = ? order by u.id",
		buildingID); err != nil {
		return nil, err
	}
	return users, nil
}

func InsertUser(ctx context.Context, conn sqlx.SqlConn, u *User) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into users(username, password, role, name, phone, building_id, status) values(?,?,?,?,?,?,?)",
		u.Username, u.Password, u.Role, u.Name, nullableString(u.Phone), nullableInt64(u.BuildingID), u.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func InsertWorkerBuilding(ctx context.Context, conn sqlx.SqlConn, workerID, buildingID int64) error {
	_, err := conn.ExecCtx(ctx,
		"insert ignore into worker_buildings(worker_id, building_id) values(?,?)", workerID, buildingID)
	return err
}

func nullableString(s sql.NullString) any {
	if !s.Valid || s.String == "" {
		return nil
	}
	return s.String
}

func nullableInt64(n sql.NullInt64) any {
	if !n.Valid || n.Int64 == 0 {
		return nil
	}
	return n.Int64
}
