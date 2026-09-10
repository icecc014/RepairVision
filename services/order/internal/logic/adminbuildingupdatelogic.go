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
		LayoutJson:    req.LayoutJson,
	}}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
