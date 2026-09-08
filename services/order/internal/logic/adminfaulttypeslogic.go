package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminFaultTypesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminFaultTypesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFaultTypesLogic {
	return &AdminFaultTypesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminFaultTypesLogic) AdminFaultTypes() (resp *types.AdminFaultTypeListResponse, err error) {
	list, err := store.ListAllFaultTypes(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	out := &types.AdminFaultTypeListResponse{List: []types.AdminFaultTypeItem{}}
	for _, ft := range list {
		out.List = append(out.List, types.AdminFaultTypeItem{
			Id: ft.ID, Code: ft.Code, Name: ft.Name, Sort: ft.Sort, Status: ft.Status,
		})
	}
	return out, nil
}
