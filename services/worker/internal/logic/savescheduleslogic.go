package logic

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type SaveSchedulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveSchedulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSchedulesLogic {
	return &SaveSchedulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SaveSchedulesLogic) SaveSchedules(in *worker.SaveSchedulesRequest) (*worker.ScheduleListResponse, error) {
	if len(in.Items) == 0 {
		return &worker.ScheduleListResponse{}, nil
	}
	items := make([]*worker.ScheduleItem, 0, len(in.Items))
	for _, item := range in.Items {
		if item.WorkerId <= 0 {
			return nil, errDate("工人不能为空")
		}
		if _, err := time.Parse("2006-01-02", item.WorkDate); err != nil {
			return nil, errDate("日期格式应为 YYYY-MM-DD")
		}
		shift := strings.ToUpper(strings.TrimSpace(item.ShiftType))
		if !validShiftType(shift) {
			return nil, errDate("班次类型必须为 DAY/MORNING/AFTERNOON/OFF")
		}
		cp := &worker.ScheduleItem{
			WorkerId:  item.WorkerId,
			WorkDate:  item.WorkDate,
			ShiftType: shift,
			Note:      strings.TrimSpace(item.Note),
		}
		items = append(items, cp)
	}
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		for _, item := range items {
			if err := store.UpsertSchedule(txCtx, session, item.WorkerId, item.WorkDate, item.ShiftType, item.Note); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &worker.ScheduleListResponse{Items: items}, nil
}

func validShiftType(shift string) bool {
	switch shift {
	case "DAY", "MORNING", "AFTERNOON", "OFF":
		return true
	}
	return false
}
