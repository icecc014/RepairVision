package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminOperationLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminOperationLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminOperationLogsLogic {
	return &AdminOperationLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminOperationLogsLogic) AdminOperationLogs(req *types.OperationLogListRequest) (resp *types.OperationLogListResponse, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	logs, total, err := store.ListOperationLogs(l.ctx, l.svcCtx.DB, page, size, req.Module, req.Action, req.Keyword)
	if err != nil {
		return nil, err
	}
	out := &types.OperationLogListResponse{Total: total, List: []types.OperationLogItem{}}
	for _, lv := range logs {
		item := types.OperationLogItem{
			Id:           lv.ID,
			Module:       lv.Module,
			Action:       lv.Action,
			Method:       lv.Method,
			Path:         lv.Path,
			RequestBody:  lv.RequestBody,
			ResponseCode: lv.ResponseCode,
			Ip:           lv.IP,
			CostMs:       lv.CostMs,
			CreatedAt:    lv.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if lv.UserID.Valid {
			item.UserId = lv.UserID.Int64
		}
		if lv.Username.Valid {
			item.Username = lv.Username.String
		}
		if lv.Role.Valid {
			item.Role = lv.Role.Int64
		}
		out.List = append(out.List, item)
	}
	return out, nil
}
