package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type GenerateWeeklyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateWeeklyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateWeeklyLogic {
	return &GenerateWeeklyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GenerateWeekly 生成某周一整周的班次：每名工人每周固定休 1 天，其余为 DAY 班。
// 轮休日按工人顺序 + 周序号错开，避免多人同休。
func (l *GenerateWeeklyLogic) GenerateWeekly(in *worker.GenerateScheduleRequest) (*worker.ScheduleListResponse, error) {
	weekStart, err := time.ParseInLocation("2006-01-02", in.WeekStart, time.Local)
	if err != nil || weekStart.Weekday() != time.Monday {
		return nil, errDate("请提供周一的日期，格式 YYYY-MM-DD")
	}
	workerIDs, err := store.ListWorkerIDs(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	if len(workerIDs) == 0 {
		return &worker.ScheduleListResponse{}, nil
	}
	endDate := weekStart.AddDate(0, 0, 6)
	startStr := weekStart.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")
	weekSeed := int(weekStart.Unix()/86400) % 7
	if weekSeed < 0 {
		weekSeed += 7
	}

	items := make([]*worker.ScheduleItem, 0, len(workerIDs)*7)
	for wi, workerID := range workerIDs {
		offDay := (wi + weekSeed) % 7
		for day := 0; day < 7; day++ {
			shift := "DAY"
			note := ""
			if day == offDay {
				shift = "OFF"
				note = "周模板轮休"
			}
			items = append(items, &worker.ScheduleItem{
				WorkerId:  workerID,
				WorkDate:  weekStart.AddDate(0, 0, day).Format("2006-01-02"),
				ShiftType: shift,
				Note:      note,
			})
		}
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if err := store.ClearScheduleRange(txCtx, session, startStr, endStr); err != nil {
			return err
		}
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
