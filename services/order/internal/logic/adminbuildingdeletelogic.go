package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminBuildingDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminBuildingDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminBuildingDeleteLogic {
	return &AdminBuildingDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminBuildingDeleteLogic) AdminBuildingDelete(req *types.AdminBuildingIdRequest) (resp *types.EmptyResponse, err error) {
	orders, err := store.ListAllOrders(l.ctx, l.svcCtx.DB, 0, req.Id)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if len(orders) > 0 {
		return nil, errs.Conflict("该楼栋已有工单记录，不能删除")
	}
	if _, err := l.svcCtx.MapRpc.DeleteBuilding(l.ctx, &mapclient.BuildingIdRequest{Id: req.Id}); err != nil {
		return nil, rpcBizError(err)
	}
	return &types.EmptyResponse{}, nil
}
