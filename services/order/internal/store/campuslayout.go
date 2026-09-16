package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CampusLayout 区域概览（校园/建筑群总平面图）。
type CampusLayout struct {
	ID         int64          `db:"id"`
	Name       string         `db:"name"`
	IsDefault  int64          `db:"is_default"`
	Cols       int64          `db:"grid_cols"`
	Rows       int64          `db:"grid_rows"`
	LayoutJson sql.NullString `db:"layout_json"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

const campusLayoutColumns = `id, name, is_default, grid_cols, grid_rows, layout_json, updated_at`

// FindDefaultCampusLayout 取默认区域概览；不存在时返回 sqlx.ErrNotFound。
func FindDefaultCampusLayout(ctx context.Context, conn sqlx.Session) (*CampusLayout, error) {
	var row CampusLayout
	if err := conn.QueryRowCtx(ctx, &row,
		"select "+campusLayoutColumns+" from campus_layouts where is_default = 1 order by id limit 1"); err != nil {
		return nil, err
	}
	return &row, nil
}

// InsertCampusLayout 新建区域概览，返回自增 ID。
func InsertCampusLayout(ctx context.Context, conn sqlx.Session, name string, cols, rows int64) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into campus_layouts(name, is_default, grid_cols, grid_rows) values(?, 1, ?, ?)", name, cols, rows)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// SaveCampusLayout 保存区域概览内容。
func SaveCampusLayout(ctx context.Context, conn sqlx.Session, id int64, name string, cols, rows int64, layoutJson string) error {
	_, err := conn.ExecCtx(ctx,
		"update campus_layouts set name = ?, grid_cols = ?, grid_rows = ?, layout_json = ? where id = ?",
		name, cols, rows, layoutJson, id)
	return err
}
