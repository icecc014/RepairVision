package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type ListSchedulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSchedulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSchedulesLogic {
	return &ListSchedulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSchedulesLogic) ListSchedules(in *worker.ScheduleListRequest) (*worker.ScheduleListResponse, error) {
	startDate, endDate, err := normalizeRange(in.StartDate, in.EndDate)
	if err != nil {
		return nil, err
	}
	rows, err := store.ListSchedules(l.ctx, l.svcCtx.DB, in.WorkerId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	resp := &worker.ScheduleListResponse{}
	for _, row := range rows {
		resp.Items = append(resp.Items, scheduleToPb(row))
	}
	return resp, nil
}

func normalizeRange(startDate, endDate string) (string, string, error) {
	const layout = "2006-01-02"
	if startDate == "" {
		startDate = time.Now().Format(layout)
	}
	if endDate == "" {
		endDate = time.Now().AddDate(0, 0, 13).Format(layout)
	}
	if _, err := time.Parse(layout, startDate); err != nil {
		return "", "", errDate("开始日期格式应为 YYYY-MM-DD")
	}
	if _, err := time.Parse(layout, endDate); err != nil {
		return "", "", errDate("结束日期格式应为 YYYY-MM-DD")
	}
	return startDate, endDate, nil
}

type errDate string

func (e errDate) Error() string {
	return string(e)
}
