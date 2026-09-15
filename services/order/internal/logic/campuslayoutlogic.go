package logic

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type CampusLayoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusLayoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusLayoutLogic {
	return &CampusLayoutLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

const (
	campusDefaultCols = 40
	campusDefaultRows = 30
	campusMaxCols     = 80
	campusMaxRows     = 60
)

// loadDefault 读取默认区域概览，不存在时自动创建一张空白图。
func (l *CampusLayoutLogic) loadDefault() (*store.CampusLayout, error) {
	row, err := store.FindDefaultCampusLayout(l.ctx, l.svcCtx.DB)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, sqlx.ErrNotFound) {
		return nil, err
	}
	if _, err := store.InsertCampusLayout(l.ctx, l.svcCtx.DB, "默认区域概览", campusDefaultCols, campusDefaultRows); err != nil {
		return nil, err
	}
	return store.FindDefaultCampusLayout(l.ctx, l.svcCtx.DB)
}

func campusToResponse(row *store.CampusLayout) *types.CampusLayoutResponse {
	layoutJson := ""
	if row.LayoutJson.Valid {
		layoutJson = row.LayoutJson.String
	}
	return &types.CampusLayoutResponse{
		Id:         row.ID,
		Name:       row.Name,
		Cols:       row.Cols,
		Rows:       row.Rows,
		LayoutJson: layoutJson,
		UpdatedAt:  row.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// CampusLayout 查询区域概览（管理端、宿管端、工人端共用，只读）。
func (l *CampusLayoutLogic) CampusLayout() (*types.CampusLayoutResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	row, err := l.loadDefault()
	if err != nil {
		return nil, errs.Internal(err)
	}
	return campusToResponse(row), nil
}

// SaveCampusLayout 管理员保存区域概览。
func (l *CampusLayoutLogic) SaveCampusLayout(req *types.CampusLayoutSaveRequest) (*types.CampusLayoutResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	row, err := l.loadDefault()
	if err != nil {
		return nil, errs.Internal(err)
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "默认区域概览"
	}
	cols := req.Cols
	if cols <= 0 {
		cols = campusDefaultCols
	}
	if cols > campusMaxCols {
		cols = campusMaxCols
	}
	rows := req.Rows
	if rows <= 0 {
		rows = campusDefaultRows
	}
	if rows > campusMaxRows {
		rows = campusMaxRows
	}
	layoutJson := strings.TrimSpace(req.LayoutJson)
	if layoutJson == "" {
		layoutJson = "{}"
	}
	if err := store.SaveCampusLayout(l.ctx, l.svcCtx.DB, row.ID, name, cols, rows, layoutJson); err != nil {
		return nil, errs.Internal(err)
	}
	updated := &store.CampusLayout{
		ID: row.ID, Name: name, IsDefault: 1, Cols: cols, Rows: rows,
		LayoutJson: sql.NullString{String: layoutJson, Valid: true}, UpdatedAt: time.Now(),
	}
	return campusToResponse(updated), nil
}
