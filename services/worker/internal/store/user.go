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

func FindUserByUsername(ctx context.Context, conn sqlx.Session, username string) (*User, error) {
	return findUserByUsername(ctx, conn, username, true)
}

func FindUserByUsernameAll(ctx context.Context, conn sqlx.Session, username string) (*User, error) {
	return findUserByUsername(ctx, conn, username, false)
}

func findUserByUsername(ctx context.Context, conn sqlx.Session, username string, onlyEnabled bool) (*User, error) {
	var u User
	query := "select " + userColumns + " from users where username = ?"
	args := []any{username}
	if onlyEnabled {
		query += " and status = 1"
	}
	if err := conn.QueryRowCtx(ctx, &u, query, args...); err != nil {
		return nil, err
	}
	return &u, nil
}

func FindUserByID(ctx context.Context, conn sqlx.Session, id int64) (*User, error) {
	var u User
	if err := conn.QueryRowCtx(ctx, &u,
		"select "+userColumns+" from users where id = ?", id); err != nil {
		return nil, err
	}
	return &u, nil
}

func FindUsersByIDs(ctx context.Context, conn sqlx.Session, ids []int64) ([]User, error) {
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

func ListWorkersByBuilding(ctx context.Context, conn sqlx.Session, buildingID int64) ([]User, error) {
	var users []User
	if err := conn.QueryRowsCtx(ctx, &users,
		"select "+userColumns+" from users where role = 2 and status = 1 and id in (select worker_id from worker_buildings where building_id = ?) order by id",
		buildingID); err != nil {
		return nil, err
	}
	return users, nil
}

func ListUsers(ctx context.Context, conn sqlx.Session, role, status int64, keyword string) ([]User, error) {
	query := "select " + userColumns + " from users where 1 = 1"
	var args []any
	if role > 0 {
		query += " and role = ?"
		args = append(args, role)
	}
	if status > 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		query += " and (username like ? or name like ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	query += " order by id"
	var users []User
	if err := conn.QueryRowsCtx(ctx, &users, query, args...); err != nil {
		return nil, err
	}
	return users, nil
}

func InsertUser(ctx context.Context, conn sqlx.Session, u *User) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into users(username, password, role, name, phone, building_id, status) values(?,?,?,?,?,?,?)",
		u.Username, u.Password, u.Role, u.Name, nullableString(u.Phone), nullableInt64(u.BuildingID), u.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateUserProfile(ctx context.Context, conn sqlx.Session, id int64, name, phone string, role int64, status int64, buildingID sql.NullInt64) error {
	_, err := conn.ExecCtx(ctx,
		"update users set name = ?, phone = ?, role = ?, status = ?, building_id = ? where id = ?",
		name, phone, role, status, nullableInt64(buildingID), id)
	return err
}

func UpdateUserPassword(ctx context.Context, conn sqlx.Session, id int64, password string) error {
	_, err := conn.ExecCtx(ctx,
		"update users set password = ? where id = ?", password, id)
	return err
}

func DisableUser(ctx context.Context, conn sqlx.Session, id int64) error {
	_, err := conn.ExecCtx(ctx,
		"update users set status = 0 where id = ?", id)
	return err
}

func InsertWorkerBuilding(ctx context.Context, conn sqlx.Session, workerID, buildingID int64) error {
	_, err := conn.ExecCtx(ctx,
		"insert ignore into worker_buildings(worker_id, building_id) values(?,?)", workerID, buildingID)
	return err
}

func DeleteWorkerBuildings(ctx context.Context, conn sqlx.Session, workerID int64) error {
	_, err := conn.ExecCtx(ctx,
		"delete from worker_buildings where worker_id = ?", workerID)
	return err
}

type workerBuildingRow struct {
	BuildingID int64 `db:"building_id"`
}

func ListWorkerBuildingIDs(ctx context.Context, conn sqlx.Session, workerID int64) ([]int64, error) {
	var rows []workerBuildingRow
	if err := conn.QueryRowsCtx(ctx, &rows,
		"select building_id from worker_buildings where worker_id = ? order by building_id", workerID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.BuildingID)
	}
	return ids, nil
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
