package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminRepairRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRepairRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRepairRecordsLogic {
	return &AdminRepairRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminRepairRecords 返回业务口径的报修记录（楼栋/房间/故障/报修宿管/处理工人/状态/时间）。
func (l *AdminRepairRecordsLogic) AdminRepairRecords(req *types.RepairRecordListRequest) (resp *types.RepairRecordListResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	filter := store.RepairRecordFilter{
		BuildingID: req.BuildingId,
		Status:     req.Status,
		FaultType:  req.FaultType,
		Keyword:    req.Keyword,
		Days:       req.Days,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	}
	if filter.Days <= 0 && filter.StartDate == "" && filter.EndDate == "" {
		filter.Days = 3
	}

	size := req.Size
	if size <= 0 {
		size = 20
	}
	if size > 1000 {
		size = 1000
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}

	total, err := store.CountRepairRecords(l.ctx, l.svcCtx.DB, filter, true)
	if err != nil {
		return nil, errs.Internal(err)
	}
	orders, err := store.ListRepairRecords(l.ctx, l.svcCtx.DB, filter, page, size)
	if err != nil {
		return nil, errs.Internal(err)
	}
	items, err := buildOrderItems(l.ctx, l.svcCtx, orders)
	if err != nil {
		return nil, err
	}

	byStatus, err := store.SummarizeRepairRecords(l.ctx, l.svcCtx.DB, filter)
	if err != nil {
		return nil, errs.Internal(err)
	}
	summary := types.RepairRecordSummary{
		Pending:  byStatus[store.StatusPending] + byStatus[store.StatusDispatched],
		Working:  byStatus[store.StatusWorking],
		Done:     byStatus[store.StatusCompleted],
		Canceled: byStatus[store.StatusCanceled],
	}
	summary.Total = summary.Pending + summary.Working + summary.Done + summary.Canceled

	list := make([]types.RepairRecordItem, 0, len(orders))
	for i := range orders {
		o := orders[i]
		record := types.RepairRecordItem{
			Id:            o.ID,
			OrderNo:       o.OrderNo,
			Title:         o.Title,
			Description:   o.Description,
			BuildingId:    o.BuildingID,
			Room:          o.Room,
			Floor:         o.Floor,
			FaultType:     o.FaultType,
			Status:        o.Status,
			StatusText:    statusText(o.Status),
			ReporterId:    o.ReporterID,
			Source:        o.Source,
			CreatedAt:     o.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if i < len(items) {
			record.BuildingName = items[i].BuildingName
			record.FaultTypeName = items[i].FaultTypeName
			record.ReporterName = items[i].ReporterName
			record.WorkerId = items[i].WorkerId
			record.WorkerName = items[i].WorkerName
			record.WorkerPhone = items[i].WorkerPhone
		}
		if record.FaultTypeName == "" {
			record.FaultTypeName = record.FaultType
		}
		if o.DispatchedAt.Valid {
			record.DispatchedAt = o.DispatchedAt.Time.Format("2006-01-02 15:04:05")
		}
		if o.StartedAt.Valid {
			record.StartedAt = o.StartedAt.Time.Format("2006-01-02 15:04:05")
		}
		if o.CompletedAt.Valid {
			record.CompletedAt = o.CompletedAt.Time.Format("2006-01-02 15:04:05")
		}
		list = append(list, record)
	}
	return &types.RepairRecordListResponse{Total: total, Summary: summary, List: list}, nil
}

