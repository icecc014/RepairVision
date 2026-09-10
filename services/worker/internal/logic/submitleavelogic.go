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

type SubmitLeaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitLeaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitLeaveLogic {
	return &SubmitLeaveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitLeaveLogic) SubmitLeave(in *worker.SubmitLeaveRequest) (*worker.LeaveItem, error) {
	if in.WorkerId <= 0 {
		return nil, errors.New("工人不能为空")
	}
	start, err := time.Parse("2006-01-02", in.StartDate)
	if err != nil {
		return nil, errors.New("开始日期格式应为 YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", in.EndDate)
	if err != nil {
		return nil, errors.New("结束日期格式应为 YYYY-MM-DD")
	}
	if end.Before(start) {
		return nil, errors.New("结束日期不能早于开始日期")
	}
	if end.Sub(start) > 30*24*time.Hour {
		return nil, errors.New("单次请假不能超过30天")
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return nil, errors.New("请假原因不能为空")
	}
	if len([]rune(reason)) > 200 {
		return nil, errors.New("请假原因不能超过200字")
	}
	overlap, err := store.HasOverlappingLeave(l.ctx, l.svcCtx.DB, in.WorkerId, in.StartDate, in.EndDate)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, errors.New("该时间段已有待审批或已通过的请假申请")
	}
	id, err := store.InsertLeave(l.ctx, l.svcCtx.DB, in.WorkerId, in.StartDate, in.EndDate, reason)
	if err != nil {
		return nil, err
	}
	row, err := store.FindLeaveByID(l.ctx, l.svcCtx.DB, id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("请假申请创建失败")
		}
		return nil, err
	}
	return leaveToPb(*row), nil
}
