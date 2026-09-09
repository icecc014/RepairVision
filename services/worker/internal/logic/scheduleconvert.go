package logic

import (
	"worker/internal/store"
	"worker/worker"
)

func scheduleToPb(row store.ScheduleRow) *worker.ScheduleItem {
	note := ""
	if row.Note.Valid {
		note = row.Note.String
	}
	return &worker.ScheduleItem{
		WorkerId:  row.WorkerID,
		WorkDate:  row.WorkDate.Format("2006-01-02"),
		ShiftType: row.ShiftType,
		Note:      note,
	}
}
