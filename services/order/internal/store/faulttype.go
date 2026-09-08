package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FaultType struct {
	ID     int64  `db:"id"`
	Code   string `db:"code"`
	Name   string `db:"name"`
	Status int64  `db:"status"`
}

func ListFaultTypes(ctx context.Context, conn sqlx.Session) ([]FaultType, error) {
	var list []FaultType
	if err := conn.QueryRowsCtx(ctx, &list,
		"select id, code, name, status from fault_types where status = 1 order by sort, id"); err != nil {
		return nil, err
	}
	return list, nil
}

func FindFaultTypeByCode(ctx context.Context, conn sqlx.Session, code string) (*FaultType, error) {
	var ft FaultType
	if err := conn.QueryRowCtx(ctx, &ft,
		"select id, code, name, status from fault_types where code = ? and status = 1", code); err != nil {
		return nil, err
	}
	return &ft, nil
}

func InsertFaultType(ctx context.Context, conn sqlx.Session, code, name string) error {
	_, err := conn.ExecCtx(ctx,
		"insert ignore into fault_types(code, name, parent_id, sort, status) values(?,?,0,0,1)", code, name)
	return err
}
