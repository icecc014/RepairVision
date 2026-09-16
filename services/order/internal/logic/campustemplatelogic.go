package logic

import (
	"context"
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

// V6.1 区域概览「我的画布」模板库：管理员可把当前画布另存为命名快照，
// 随时对比/加载，并用「使用此模板」把它应用到当前生效地图（其他端可见）。

const campusTemplateNameMax = 100

func campusTemplateToItem(row store.CampusTemplate) types.CampusTemplateItem {
	return types.CampusTemplateItem{
		Id: row.ID, Name: row.Name, Cols: row.Cols, Rows: row.Rows,
		Source: row.Source, UpdatedAt: row.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func campusTemplateToDetail(row store.CampusTemplate) types.CampusTemplateDetail {
	layoutJson := ""
	if row.LayoutJson.Valid {
		layoutJson = row.LayoutJson.String
	}
	return types.CampusTemplateDetail{
		Id: row.ID, Name: row.Name, Cols: row.Cols, Rows: row.Rows,
		Source: row.Source, LayoutJson: layoutJson,
		UpdatedAt: row.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func normalizeTemplateName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errs.BadRequest("画布名称不能为空")
	}
	if len([]rune(name)) > campusTemplateNameMax {
		return "", errs.BadRequest("画布名称过长（最多 100 字）")
	}
	return name, nil
}

// ---------- 列表 ----------

type CampusTemplateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusTemplateListLogic {
	return &CampusTemplateListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusTemplateListLogic) List() (*types.CampusTemplateListResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	rows, err := store.ListCampusTemplates(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	resp := &types.CampusTemplateListResponse{List: make([]types.CampusTemplateItem, 0, len(rows))}
	for _, row := range rows {
		resp.List = append(resp.List, campusTemplateToItem(row))
	}
	return resp, nil
}

// ---------- 读取单个 ----------

type CampusTemplateGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusTemplateGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusTemplateGetLogic {
	return &CampusTemplateGetLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusTemplateGetLogic) Get(req *types.CampusTemplateIdRequest) (*types.CampusTemplateDetailResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	row, err := store.FindCampusTemplate(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("画布不存在或已删除")
		}
		return nil, errs.Internal(err)
	}
	return &types.CampusTemplateDetailResponse{Template: campusTemplateToDetail(*row)}, nil
}

// ---------- 保存（新建或覆盖同名） ----------

type CampusTemplateSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusTemplateSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusTemplateSaveLogic {
	return &CampusTemplateSaveLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusTemplateSaveLogic) Save(req *types.CampusTemplateSaveRequest) (*types.CampusTemplateDetailResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	name, err := normalizeTemplateName(req.Name)
	if err != nil {
		return nil, err
	}
	cols, rows := clampCampusSize(req.Cols, req.Rows)
	layoutJson := strings.TrimSpace(req.LayoutJson)
	if layoutJson == "" {
		layoutJson = "{}"
	}
	existing, err := store.FindCampusTemplateByName(l.ctx, l.svcCtx.DB, name)
	switch {
	case err == nil:
		if !req.Overwrite {
			return nil, errs.Conflict("已存在同名画布：" + name + "（如需覆盖请在弹窗中确认）")
		}
		if uErr := store.UpdateCampusTemplateContent(l.ctx, l.svcCtx.DB, existing.ID, cols, rows, layoutJson); uErr != nil {
			return nil, errs.Internal(uErr)
		}
		updated, fErr := store.FindCampusTemplate(l.ctx, l.svcCtx.DB, existing.ID)
		if fErr != nil {
			return nil, errs.Internal(fErr)
		}
		return &types.CampusTemplateDetailResponse{Template: campusTemplateToDetail(*updated)}, nil
	case !errors.Is(err, sqlx.ErrNotFound):
		return nil, errs.Internal(err)
	}
	id, err := store.InsertCampusTemplate(l.ctx, l.svcCtx.DB, name, cols, rows, layoutJson, "manual")
	if err != nil {
		return nil, errs.Internal(err)
	}
	row, err := store.FindCampusTemplate(l.ctx, l.svcCtx.DB, id)
	if err != nil {
		return nil, errs.Internal(err)
	}
	return &types.CampusTemplateDetailResponse{Template: campusTemplateToDetail(*row)}, nil
}

// ---------- 重命名 ----------

type CampusTemplateRenameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusTemplateRenameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusTemplateRenameLogic {
	return &CampusTemplateRenameLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusTemplateRenameLogic) Rename(req *types.CampusTemplateRenameRequest) (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	name, err := normalizeTemplateName(req.Name)
	if err != nil {
		return nil, err
	}
	if _, err := store.FindCampusTemplate(l.ctx, l.svcCtx.DB, req.Id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("画布不存在或已删除")
		}
		return nil, errs.Internal(err)
	}
	if other, err := store.FindCampusTemplateByName(l.ctx, l.svcCtx.DB, name); err == nil && other.ID != req.Id {
		return nil, errs.Conflict("已存在同名画布：" + name)
	}
	if err := store.RenameCampusTemplate(l.ctx, l.svcCtx.DB, req.Id, name); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}

// ---------- 删除 ----------

type CampusTemplateDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusTemplateDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusTemplateDeleteLogic {
	return &CampusTemplateDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CampusTemplateDeleteLogic) Delete(req *types.CampusTemplateIdRequest) (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if _, err := store.FindCampusTemplate(l.ctx, l.svcCtx.DB, req.Id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("画布不存在或已删除")
		}
		return nil, errs.Internal(err)
	}
	if err := store.DeleteCampusTemplate(l.ctx, l.svcCtx.DB, req.Id); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}

// ---------- 使用此模板（应用到当前生效地图） ----------

type CampusTemplateApplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusTemplateApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusTemplateApplyLogic {
	return &CampusTemplateApplyLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Apply 把模板内容复制进当前生效地图（宿管端/工人端立即可见）。
// 应用前自动把当前地图另存为"应用前备份"，最多保留 5 条，避免误覆盖。
func (l *CampusTemplateApplyLogic) Apply(req *types.CampusTemplateIdRequest) (*types.CampusLayoutResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	row, err := store.FindCampusTemplate(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("画布不存在或已删除")
		}
		return nil, errs.Internal(err)
	}
	layoutJson := ""
	if row.LayoutJson.Valid {
		layoutJson = row.LayoutJson.String
	}
	if err := l.backupCurrent(); err != nil {
		// 备份失败不阻塞应用，但必须留痕
		logx.WithContext(l.ctx).Errorf("auto backup current campus layout failed: %v", err)
	}
	return NewCampusLayoutLogic(l.ctx, l.svcCtx).SaveCampusLayout(&types.CampusLayoutSaveRequest{
		Name:       row.Name,
		Cols:       row.Cols,
		Rows:       row.Rows,
		LayoutJson: layoutJson,
	})
}

// backupCurrent 把当前生效地图存成自动备份模板。
func (l *CampusTemplateApplyLogic) backupCurrent() error {
	current, err := store.FindDefaultCampusLayout(l.ctx, l.svcCtx.DB)
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
	name := "应用前备份 " + nowStamp()
	if existing, err := store.FindCampusTemplateByName(l.ctx, l.svcCtx.DB, name); err == nil {
		if err := store.UpdateCampusTemplateContent(l.ctx, l.svcCtx.DB, existing.ID, current.Cols, current.Rows, raw); err != nil {
			return err
		}
	} else if _, err := store.InsertCampusTemplate(l.ctx, l.svcCtx.DB, name, current.Cols, current.Rows, raw, "auto-backup"); err != nil {
		return err
	}
	return store.PruneCampusTemplatesBySource(l.ctx, l.svcCtx.DB, "auto-backup", 5)
}

// nowStamp 备份命名的秒级时间戳。
func nowStamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// clampCampusSize 画布尺寸兜底（与 CampusLayoutLogic 的上限保持一致）。
func clampCampusSize(cols, rows int64) (int64, int64) {
	if cols <= 0 {
		cols = campusDefaultCols
	}
	if cols > campusMaxCols {
		cols = campusMaxCols
	}
	if rows <= 0 {
		rows = campusDefaultRows
	}
	if rows > campusMaxRows {
		rows = campusMaxRows
	}
	return cols, rows
}

