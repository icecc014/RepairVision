package logic

import (
	"order/internal/types"
	"worker/workerclient"
)

func leaveStatusText(status int64) string {
	switch status {
	case 1:
		return "待审批"
	case 2:
		return "已通过"
	case 3:
		return "已驳回"
	case 4:
		return "已撤销"
	default:
		return "未知"
	}
}

func leavePbToType(in *workerclient.LeaveItem) types.LeaveItem {
	return types.LeaveItem{
		Id:         in.Id,
		WorkerId:   in.WorkerId,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Reason:     in.Reason,
		Status:     in.Status,
		StatusText: leaveStatusText(in.Status),
		ReviewNote: in.ReviewNote,
		CreatedAt:  in.CreatedAt,
	}
}
