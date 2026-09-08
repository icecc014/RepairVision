package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type FaultTypesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFaultTypesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FaultTypesLogic {
	return &FaultTypesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FaultTypesLogic) FaultTypes() (resp *types.FaultTypeListResponse, err error) {
	list, err := store.ListFaultTypes(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	out := &types.FaultTypeListResponse{List: []types.FaultType{}}
	for _, ft := range list {
		out.List = append(out.List, types.FaultType{Id: ft.ID, Code: ft.Code, Name: ft.Name})
	}
	return out, nil
}
