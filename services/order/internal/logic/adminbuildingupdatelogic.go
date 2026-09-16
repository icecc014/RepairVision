package logic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
)

// layoutClearLayoutSentinel：建筑更新接口的布局哨兵值。
// 传空字符串表示"本次不动布局"，传该哨兵表示"显式清空自定义布局（回退内置标准层）"。
const layoutClearLayoutSentinel = "__DEFAULT__"

type AdminBuildingUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminBuildingUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminBuildingUpdateLogic {
	return &AdminBuildingUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminBuildingUpdateLogic) AdminBuildingUpdate(req *types.AdminBuildingUpdateRequest) (resp *types.EmptyResponse, err error) {
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, errs.BadRequest("楼栋编码和名称不能为空")
	}
	if req.Floors <= 0 || req.RoomsPerFloor <= 0 {
		return nil, errs.BadRequest("楼层数和每层房间数必须大于0")
	}
	// V6.1 布局保护：建筑信息管理的编辑（改名 / 改坐标 / 改层数）不得清空布局设计，
	// 只有布局设计器显式传哨兵时才清空。此前表单不带 layoutJson → 后端写空串 → 布局丢失。
	layoutJson := strings.TrimSpace(req.LayoutJson)
	switch layoutJson {
	case layoutClearLayoutSentinel:
		layoutJson = ""
	case "":
		current, err := l.svcCtx.MapRpc.GetBuilding(l.ctx, &mapclient.BuildingIdRequest{Id: req.Id})
		if err != nil {
			return nil, rpcBizError(err)
		}
		if current.Building != nil {
			layoutJson = current.Building.LayoutJson
		}
	}
	if _, err := l.svcCtx.MapRpc.UpdateBuilding(l.ctx, &mapclient.SaveBuildingRequest{Building: &mapclient.Building{
		Id:            req.Id,
		Code:          strings.TrimSpace(req.Code),
		Name:          strings.TrimSpace(req.Name),
		PosX:          req.PosX,
		PosY:          req.PosY,
		Width:         req.Width,
		Height:        req.Height,
		Floors:        req.Floors,
		FloorHeight:   req.FloorHeight,
		RoomsPerFloor: req.RoomsPerFloor,
		LayoutJson:    layoutJson,
	}}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
