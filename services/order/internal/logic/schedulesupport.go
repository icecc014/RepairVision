package logic

import (
	"order/internal/types"
	"worker/workerclient"
)

func schedulePbToType(in *workerclient.ScheduleItem) types.ScheduleItem {
	return types.ScheduleItem{
		WorkerId:  in.WorkerId,
		WorkDate:  in.WorkDate,
		ShiftType: in.ShiftType,
		Note:      in.Note,
	}
}

func scheduleTypeToPb(in types.ScheduleItem) *workerclient.ScheduleItem {
	return &workerclient.ScheduleItem{
		WorkerId:  in.WorkerId,
		WorkDate:  in.WorkDate,
		ShiftType: in.ShiftType,
		Note:      in.Note,
	}
}

func schedulePbListToTypes(items []*workerclient.ScheduleItem) []types.ScheduleItem {
	result := make([]types.ScheduleItem, 0, len(items))
	for _, item := range items {
		result = append(result, schedulePbToType(item))
	}
	return result
}
