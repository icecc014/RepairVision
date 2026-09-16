package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/mapclient"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
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
	campusMaxCols     = 140
	campusMaxRows     = 80
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
	// 方案 1：区域概览 → 建筑信息管理 单向同步（把图元中心格换算为建筑 2D 坐标）
	if err := l.syncBuildingPositions(layoutJson); err != nil {
		logx.WithContext(l.ctx).Errorf("sync building positions failed: %v", err)
	}
	// V5.1+：区域概览保存后广播，宿管端 / 工人端可就地刷新（无需手动重进页面）
	l.svcCtx.WS.PublishOrder(ws.OrderEvent{
		Type: "campus_changed", OrderId: 0, OrderNo: "", BuildingId: 0, Status: 0,
	})
	return campusToResponse(updated), nil
}

// syncBuildingPositions 把区域概览中"关联楼栋"图元的中心格换算成建筑 2D 坐标写回建筑信息管理。
// 约定：1 格 = 10 米，pos_x = (col + colSpan/2) × 10，pos_y = (row + rowSpan/2) × 10；
// 这样 V4 的欧氏距离与 V5 的路网距离使用同一套地理基准。
func (l *CampusLayoutLogic) syncBuildingPositions(raw string) error {
	var layout campusLayoutJSON
	if err := json.Unmarshal([]byte(raw), &layout); err != nil {
		return err
	}
	if len(layout.Blocks) == 0 {
		return nil
	}
	resp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return err
	}
	byID := make(map[int64]*mapclient.Building, len(resp.Buildings))
	for _, b := range resp.Buildings {
		byID[b.Id] = b
	}
	for _, blk := range layout.Blocks {
		if blk.Kind != "building" || blk.BuildingID <= 0 {
			continue
		}
		b := byID[blk.BuildingID]
		if b == nil {
			continue
		}
		rowSpan, colSpan := blk.RowSpan, blk.ColSpan
		if rowSpan <= 0 {
			rowSpan = 1
		}
		if colSpan <= 0 {
			colSpan = 1
		}
		// 坐标口径（V6 起）：以区域概览画布左上角为原点，画布区域为第四象限（向下为负），
		// 以"建筑图元左上角"为坐标应用点，1 格 = 10 米：
		//   距左 5 格、距上 5 格 → (50, -50)
		posX := float64(blk.Col) * defaultGridMeters
		posY := -float64(blk.Row) * defaultGridMeters
		if _, err := l.svcCtx.MapRpc.UpdateBuilding(l.ctx, &mapclient.SaveBuildingRequest{
			Building: &mapclient.Building{
				Id: b.Id, Code: b.Code, Name: b.Name, PosX: posX, PosY: posY,
				Width: b.Width, Height: b.Height, Floors: b.Floors,
				FloorHeight: b.FloorHeight, RoomsPerFloor: b.RoomsPerFloor, LayoutJson: b.LayoutJson,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}
