package logic

import (
	"order/internal/errs"
	"order/internal/store"
	"worker/workerclient"
)

// crossBuildingWorker 跨楼栋手动支援：
// 管理员在"本楼栋没有在岗且工种匹配的工人"时，可以指定其他在岗同工种工人支援，
// 这里负责校验（在岗 + 工种匹配）并返回可直接参与评分/指派的信息。
func (l *AdminOrderReassignLogic) crossBuildingWorker(order *store.Order, workerID int64) (*workerclient.WorkerInfo, error) {
	if workerID <= 0 {
		return nil, errs.BadRequest("请选择目标工人")
	}
	users, err := l.svcCtx.WorkerRpc.ListUsers(l.ctx, &workerclient.ListUsersRequest{Role: 2, Status: 1})
	if err != nil {
		return nil, errs.Upstream()
	}
	var target *workerclient.User
	for _, u := range users.Users {
		if u.Id == workerID {
			target = u
			break
		}
	}
	if target == nil {
		return nil, errs.BadRequest("工人不存在或已停用")
	}
	status, err := l.svcCtx.WorkerRpc.GetDutyStatus(l.ctx, &workerclient.DutyStatusRequest{WorkerId: workerID})
	if err != nil {
		return nil, errs.Upstream()
	}
	if !status.GetOnDuty() {
		reason := status.GetReason()
		if reason == "" {
			reason = "今天休息 / 请假"
		}
		return nil, errs.Conflict("该工人当前不在岗（" + reason + "），请选择在岗工人")
	}
	required := int64(0)
	if faultTypes, err := store.ListFaultTypes(l.ctx, l.svcCtx.DB); err == nil {
		for _, ft := range faultTypes {
			if ft.Code == order.FaultType {
				required = requiredJobTypeOf(normalizeFaultCategory(ft.Category))
				break
			}
		}
	}
	if !jobTypeAllowed(target.JobType, required) {
		return nil, errs.Conflict("该工人工种不匹配（该单需要" + jobTypeText(required) + "）")
	}
	maxConcurrent := target.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = defaultMaxConcurrent
	}
	shift := status.GetShiftType()
	if shift == "" {
		shift = "DAY"
	}
	return &workerclient.WorkerInfo{
		Id:            target.Id,
		Username:      target.Username,
		Name:          target.Name,
		JobType:       target.JobType,
		MaxConcurrent: maxConcurrent,
		TodayShift:    shift,
	}, nil
}
