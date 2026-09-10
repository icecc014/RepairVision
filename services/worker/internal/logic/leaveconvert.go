package logic

import (
	"worker/internal/store"
	"worker/worker"
)

func leaveToPb(row store.LeaveRequest) *worker.LeaveItem {
	reviewerID := int64(0)
	if row.ReviewerID.Valid {
		reviewerID = row.ReviewerID.Int64
	}
	reviewNote := ""
	if row.ReviewNote.Valid {
		reviewNote = row.ReviewNote.String
	}
	return &worker.LeaveItem{
		Id:         row.ID,
		WorkerId:   row.WorkerID,
		StartDate:  row.StartDate.Format("2006-01-02"),
		EndDate:    row.EndDate.Format("2006-01-02"),
		Reason:     row.Reason,
		Status:     row.Status,
		ReviewerId: reviewerID,
		ReviewNote: reviewNote,
		CreatedAt:  row.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
