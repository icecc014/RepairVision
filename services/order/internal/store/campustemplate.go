package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CampusTemplate 区域概览模板（"我的画布"）：管理员命名的地图快照。
type CampusTemplate struct {
	ID         int64          `db:"id"`
	Name       string         `db:"name"`
	Cols       int64          `db:"grid_cols"`
	Rows       int64          `db:"grid_rows"`
	LayoutJson sql.NullString `db:"layout_json"`
	Source     string         `db:"source"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

const campusTemplateColumns = `id, name, grid_cols, grid_rows, layout_json, source, updated_at`

// ListCampusTemplates 返回全部模板，最近修改的在前。
func ListCampusTemplates(ctx context.Context, conn sqlx.Session) ([]CampusTemplate, error) {
	var list []CampusTemplate
	if err := conn.QueryRowsCtx(ctx, &list,
		"select "+campusTemplateColumns+" from campus_layout_templates order by updated_at desc, id desc"); err != nil {
		return nil, err
	}
	return list, nil
}

// FindCampusTemplate 按 ID 查询模板；不存在时返回 sqlx.ErrNotFound。
func FindCampusTemplate(ctx context.Context, conn sqlx.Session, id int64) (*CampusTemplate, error) {
	var row CampusTemplate
	if err := conn.QueryRowCtx(ctx, &row,
		"select "+campusTemplateColumns+" from campus_layout_templates where id = ?", id); err != nil {
		return nil, err
	}
	return &row, nil
}

// FindCampusTemplateByName 按名称查询模板；不存在时返回 sqlx.ErrNotFound。
func FindCampusTemplateByName(ctx context.Context, conn sqlx.Session, name string) (*CampusTemplate, error) {
	var row CampusTemplate
	if err := conn.QueryRowCtx(ctx, &row,
		"select "+campusTemplateColumns+" from campus_layout_templates where name = ?", name); err != nil {
		return nil, err
	}
	return &row, nil
}

// InsertCampusTemplate 新建模板，返回自增 ID。
func InsertCampusTemplate(ctx context.Context, conn sqlx.Session, name string, cols, rows int64, layoutJson, source string) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		"insert into campus_layout_templates(name, grid_cols, grid_rows, layout_json, source) values(?, ?, ?, ?, ?)",
		name, cols, rows, layoutJson, source)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateCampusTemplateContent 覆盖模板内容（同名保存 / 自动备份复用）。
func UpdateCampusTemplateContent(ctx context.Context, conn sqlx.Session, id int64, cols, rows int64, layoutJson string) error {
	_, err := conn.ExecCtx(ctx,
		"update campus_layout_templates set grid_cols = ?, grid_rows = ?, layout_json = ? where id = ?",
		cols, rows, layoutJson, id)
	return err
}

// RenameCampusTemplate 重命名模板。
func RenameCampusTemplate(ctx context.Context, conn sqlx.Session, id int64, name string) error {
	_, err := conn.ExecCtx(ctx, "update campus_layout_templates set name = ? where id = ?", name, id)
	return err
}

// DeleteCampusTemplate 删除模板。
func DeleteCampusTemplate(ctx context.Context, conn sqlx.Session, id int64) error {
	_, err := conn.ExecCtx(ctx, "delete from campus_layout_templates where id = ?", id)
	return err
}

// PruneCampusTemplatesBySource 只保留某来源（如自动备份）最新的 keep 条。
func PruneCampusTemplatesBySource(ctx context.Context, conn sqlx.Session, source string, keep int) error {
	_, err := conn.ExecCtx(ctx,
		`delete from campus_layout_templates where source = ? and id not in (
			select id from (
				select id from campus_layout_templates where source = ? order by id desc limit ?
			) keep_rows
		)`, source, source, keep)
	return err
}
