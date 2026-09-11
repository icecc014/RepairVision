package types

// AdminLeaveCreateRequest 管理员代工人登记请假（直接置为已通过）。
type AdminLeaveCreateRequest struct {
	WorkerId  int64  `json:"workerId"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Reason    string `json:"reason"`
}
