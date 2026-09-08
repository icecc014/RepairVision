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

type AdminBuildingCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminBuildingCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminBuildingCreateLogic {
	return &AdminBuildingCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminBuildingCreateLogic) AdminBuildingCreate(req *types.AdminBuildingCreateRequest) (resp *types.EmptyResponse, err error) {
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, errs.BadRequest("楼栋编码和名称不能为空")
	}
	if req.Floors <= 0 || req.RoomsPerFloor <= 0 {
		return nil, errs.BadRequest("楼层数和每层房间数必须大于0")
	}
	if _, err := l.svcCtx.MapRpc.CreateBuilding(l.ctx, &mapclient.SaveBuildingRequest{Building: &mapclient.Building{
		Code:          strings.TrimSpace(req.Code),
		Name:          strings.TrimSpace(req.Name),
		PosX:          req.PosX,
		PosY:          req.PosY,
		Width:         req.Width,
		Height:        req.Height,
		Floors:        req.Floors,
		FloorHeight:   req.FloorHeight,
		RoomsPerFloor: req.RoomsPerFloor,
	}}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
