package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

// V6.2 区域概览「备份与恢复」：
//   - 每次"使用此模板 / 恢复某一版"前自动备份当前地图；
//   - 也可以手动「立即备份」；
//   - 备份保留份数可配置（sys_settings.campus_backup_keep）。

const (
	campusBackupSource      = "auto-backup"
	campusBackupKeepKey     = "campus_backup_keep"
	campusBackupKeepDefault = 5
	campusBackupKeepMax     = 50
)

// campusBackupKeep 读取"自动备份保留份数"配置。
func campusBackupKeep(ctx context.Context, svcCtx *svc.ServiceContext) int {
	raw, err := store.GetSetting(ctx, svcCtx.DB, campusBackupKeepKey)
	if err != nil {
		return campusBackupKeepDefault
	}
	keep, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || keep < 1 {
		return campusBackupKeepDefault
	}
	if keep > campusBackupKeepMax {
		keep = campusBackupKeepMax
	}
	return keep
}

// summarizeCampusLayout 统计图元数与建筑数，供备份列表展示。
func summarizeCampusLayout(raw string) (blocks int, buildings int) {
	var payload struct {
		Blocks []struct {
			Kind string `json:"kind"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return 0, 0
	}
	for _, b := range payload.Blocks {
		if b.Kind == "building" {
			buildings++
		}
	}
	return len(payload.Blocks), buildings
}

func campusBackupToItem(row store.CampusTemplate) types.CampusBackupItem {
	raw := ""
	if row.LayoutJson.Valid {
		raw = row.LayoutJson.String
	}
	blocks, buildings := summarizeCampusLayout(raw)
	return types.CampusBackupItem{
		Id: row.ID, Name: row.Name, Cols: row.Cols, Rows: row.Rows,
		Blocks: blocks, Buildings: buildings, Bytes: len(raw),
		CreatedAt: row.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// backupCurrentCampus 把当前生效地图存成备份（apply / restore / 手动备份共用）。
func backupCurrentCampus(ctx context.Context, svcCtx *svc.ServiceContext, name string) error {
	current, err := store.FindDefaultCampusLayout(ctx, svcCtx.DB)
	if err != nil {
		return err
	}
	raw := ""
	if current.LayoutJson.Valid {
		raw = strings.TrimSpace(current.LayoutJson.String)
	}
	if len(raw) < 20 {
		return nil // 当前地图为空，无需备份
	}
	if name == "" {
		name = "备份 " + nowStamp()
	}
	if existing, err := store.FindCampusTemplateByName(ctx, svcCtx.DB, name); err == nil {
		if err := store.UpdateCampusTemplateContent(ctx, svcCtx.DB, existing.ID, current.Cols, current.Rows, raw); err != nil {
			return err
		}
	} else if _, err := store.InsertCampusTemplate(ctx, svcCtx.DB, name, current.Cols, current.Rows, raw, campusBackupSource); err != nil {
		return err
	}
	return store.PruneCampusTemplatesBySource(ctx, svcCtx.DB, campusBackupSource, campusBackupKeep(ctx, svcCtx))
}

// ---------- 备份列表 ----------

type CampusBackupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusBackupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusBackupListLogic {
	return &CampusBackupListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusBackupListLogic) List() (*types.CampusBackupListResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	rows, err := store.ListCampusTemplatesBySource(l.ctx, l.svcCtx.DB, campusBackupSource)
	if err != nil {
		return nil, errs.Internal(err)
	}
	resp := &types.CampusBackupListResponse{
		List: make([]types.CampusBackupItem, 0, len(rows)),
		Keep: campusBackupKeep(l.ctx, l.svcCtx),
	}
	for _, row := range rows {
		resp.List = append(resp.List, campusBackupToItem(row))
	}
	return resp, nil
}

// ---------- 立即备份 ----------

type CampusBackupCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusBackupCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusBackupCreateLogic {
	return &CampusBackupCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusBackupCreateLogic) Create() (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if err := backupCurrentCampus(l.ctx, l.svcCtx, "手动备份 "+nowStamp()); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}

// ---------- 恢复某一版 ----------

type CampusBackupRestoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusBackupRestoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusBackupRestoreLogic {
	return &CampusBackupRestoreLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Restore 把某一版备份恢复为当前生效地图（恢复前同样先备份当前地图，可再次回退）。
func (l *CampusBackupRestoreLogic) Restore(req *types.CampusBackupIdRequest) (*types.CampusLayoutResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	row, err := store.FindCampusTemplate(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("备份不存在或已删除")
		}
		return nil, errs.Internal(err)
	}
	layoutJson := ""
	if row.LayoutJson.Valid {
		layoutJson = row.LayoutJson.String
	}
	if strings.TrimSpace(layoutJson) == "" {
		return nil, errs.BadRequest("该备份内容为空，无法恢复")
	}
	if err := backupCurrentCampus(l.ctx, l.svcCtx, ""); err != nil {
		logx.WithContext(l.ctx).Errorf("backup before restore failed: %v", err)
	}
	name := "恢复自 " + row.Name
	if []rune(name) != nil && len([]rune(name)) > campusTemplateNameMax {
		name = string([]rune(name)[:campusTemplateNameMax])
	}
	return NewCampusLayoutLogic(l.ctx, l.svcCtx).SaveCampusLayout(&types.CampusLayoutSaveRequest{
		Name: name, Cols: row.Cols, Rows: row.Rows, LayoutJson: layoutJson,
	})
}

// ---------- 保留份数配置 ----------

type CampusBackupKeepLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusBackupKeepLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusBackupKeepLogic {
	return &CampusBackupKeepLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusBackupKeepLogic) Update(req *types.CampusBackupKeepRequest) (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	keep := req.Keep
	if keep < 1 || keep > campusBackupKeepMax {
		return nil, errs.BadRequest("保留份数需在 1 ~ 50 之间")
	}
	if err := store.SetSetting(l.ctx, l.svcCtx.DB, campusBackupKeepKey, strconv.Itoa(keep),
		"区域概览自动备份保留份数（1~50）"); err != nil {
		return nil, errs.Internal(err)
	}
	if err := store.PruneCampusTemplatesBySource(l.ctx, l.svcCtx.DB, campusBackupSource, keep); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
