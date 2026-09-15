package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FaultType struct {
	ID           int64  `db:"id"`
	Code         string `db:"code"`
	Name         string `db:"name"`
	Category     string `db:"category"`
	AutoDispatch int64  `db:"auto_dispatch"`
	Sort         int64  `db:"sort"`
	Status       int64  `db:"status"`
}

func ListFaultTypes(ctx context.Context, conn sqlx.Session) ([]FaultType, error) {
	var list []FaultType
	if err := conn.QueryRowsCtx(ctx, &list,
		"select id, code, name, coalesce(category,'other') category, coalesce(auto_dispatch,1) auto_dispatch, sort, status from fault_types where status = 1 order by sort, id"); err != nil {
		return nil, err
	}
	return list, nil
}

func ListAllFaultTypes(ctx context.Context, conn sqlx.Session) ([]FaultType, error) {
	var list []FaultType
	if err := conn.QueryRowsCtx(ctx, &list,
		"select id, code, name, coalesce(category,'other') category, coalesce(auto_dispatch,1) auto_dispatch, sort, status from fault_types order by sort, id"); err != nil {
		return nil, err
	}
	return list, nil
}

func FindFaultTypeByCode(ctx context.Context, conn sqlx.Session, code string) (*FaultType, error) {
	var ft FaultType
	if err := conn.QueryRowCtx(ctx, &ft,
		"select id, code, name, coalesce(category,'other') category, coalesce(auto_dispatch,1) auto_dispatch, sort, status from fault_types where code = ? and status = 1", code); err != nil {
		return nil, err
	}
	return &ft, nil
}

func FindFaultTypeByID(ctx context.Context, conn sqlx.Session, id int64) (*FaultType, error) {
	var ft FaultType
	if err := conn.QueryRowCtx(ctx, &ft,
		"select id, code, name, coalesce(category,'other') category, coalesce(auto_dispatch,1) auto_dispatch, sort, status from fault_types where id = ?", id); err != nil {
		return nil, err
	}
	return &ft, nil
}

func InsertFaultType(ctx context.Context, conn sqlx.Session, code, name string) error {
	_, err := conn.ExecCtx(ctx,
		"insert ignore into fault_types(code, name, parent_id, sort, status) values(?,?,0,0,1)", code, name)
	return err
}

func CreateFaultType(ctx context.Context, conn sqlx.Session, code, name string, sort int64) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into fault_types(code, name, parent_id, sort, status) values(?,?,0,?,1)", code, name, sort)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateFaultType(ctx context.Context, conn sqlx.Session, id, sort, status int64, name string) error {
	_, err := conn.ExecCtx(ctx,
		"update fault_types set name = ?, sort = ?, status = ? where id = ?", name, sort, status, id)
	return err
}

func DisableFaultType(ctx context.Context, conn sqlx.Session, id int64) error {
	_, err := conn.ExecCtx(ctx,
		"update fault_types set status = 0 where id = ?", id)
	return err
}

// CreateFaultTypeFull 新建故障类型（含类别与是否自动派单）。
func CreateFaultTypeFull(ctx context.Context, conn sqlx.Session, code, name, category string, autoDispatch, sort int64) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into fault_types(code, name, category, auto_dispatch, parent_id, sort, status) values(?,?,?,?,0,?,1)",
		code, name, category, autoDispatch, sort)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateFaultTypeFull 更新故障类型（含类别与是否自动派单）。
func UpdateFaultTypeFull(ctx context.Context, conn sqlx.Session, id, sort, status int64, name, category string, autoDispatch int64) error {
	_, err := conn.ExecCtx(ctx,
		"update fault_types set name = ?, category = ?, auto_dispatch = ?, sort = ?, status = ? where id = ?",
		name, category, autoDispatch, sort, status, id)
	return err
}
