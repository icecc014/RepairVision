package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type GetDutyStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDutyStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDutyStatusLogic {
	return &GetDutyStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDutyStatus 判断工人在给定时刻是否在岗：
// 在岗 = 今日排班为白班 ∧ 当前处于工作时段 ∧ 当日无已批准请假 ∧ 账号启用。
func (l *GetDutyStatusLogic) GetDutyStatus(in *worker.DutyStatusRequest) (*worker.DutyStatusResponse, error) {
	if in.WorkerId <= 0 {
		return nil, errDate("工人不能为空")
	}
	at := time.Now()
	if raw := strings.TrimSpace(in.At); raw != "" {
		for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02 15:04", "2006-01-02"} {
			if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
				at = t
				break
			}
		}
	}
	settings, err := store.GetWorkSettings(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	resp := &worker.DutyStatusResponse{
		Morning:   settings.MorningStart + "-" + settings.MorningEnd,
		Afternoon: settings.AfternoonStart + "-" + settings.AfternoonEnd,
		Enabled:   true,
	}

	dateStr := at.Format("2006-01-02")
	shift, err := store.FindWorkerShift(l.ctx, l.svcCtx.DB, in.WorkerId, dateStr)
	if err != nil {
		return nil, err
	}
	resp.ShiftType = shift

	user, err := store.FindUserByID(l.ctx, l.svcCtx.DB, in.WorkerId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errDate("工人不存在")
		}
		return nil, err
	}
	resp.Enabled = user.Status == 1

	leaves, err := store.ListApprovedLeavesBetween(l.ctx, l.svcCtx.DB, dateStr, dateStr)
	if err != nil {
		return nil, err
	}
	for _, day := range leaves[in.WorkerId] {
		if day == dateStr {
			resp.OnLeave = true
			break
		}
	}

	resp.InWorkPeriod = inWorkPeriod(at, settings)
	resp.OnDuty = resp.Enabled && !resp.OnLeave && shiftIsWork(shift) && resp.InWorkPeriod
	switch {
	case !resp.Enabled:
		resp.Reason = "账号已停用"
	case resp.OnLeave:
		resp.Reason = "当日已批准请假"
	case !shiftIsWork(shift):
		if strings.TrimSpace(shift) == "" {
			resp.Reason = "今日未排班"
		} else {
			resp.Reason = "今日排班为休息"
		}
	case !resp.InWorkPeriod:
		resp.Reason = "当前不在工作时段（" + workPeriodText(settings) + "）"
	default:
		resp.Reason = "在岗"
	}
	return resp, nil
}