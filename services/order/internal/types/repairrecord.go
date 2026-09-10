package types

// RepairRecordItem 管理端“报修记录”行：以业务口径呈现一次宿舍报修的完整信息。
type RepairRecordItem struct {
	Id            int64  `json:"id"`
	OrderNo       string `json:"orderNo"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	BuildingId    int64  `json:"buildingId"`
	BuildingName  string `json:"buildingName"`
	Room          string `json:"room"`
	Floor         int64  `json:"floor"`
	FaultType     string `json:"faultType"`
	FaultTypeName string `json:"faultTypeName"`
	Status        int64  `json:"status"`
	StatusText    string `json:"statusText"`
	ReporterId    int64  `json:"reporterId"`
	ReporterName  string `json:"reporterName,optional"`
	WorkerId      int64  `json:"workerId,optional"`
	WorkerName    string `json:"workerName,optional"`
	WorkerPhone   string `json:"workerPhone,optional"`
	Source        string `json:"source"`
	CreatedAt     string `json:"createdAt"`
	DispatchedAt  string `json:"dispatchedAt,optional"`
	StartedAt     string `json:"startedAt,optional"`
	CompletedAt   string `json:"completedAt,optional"`
}

// RepairRecordSummary 当前筛选条件下的业务汇总（忽略状态筛选）。
type RepairRecordSummary struct {
	Total    int64 `json:"total"`
	Pending  int64 `json:"pending"`
	Working  int64 `json:"working"`
	Done     int64 `json:"done"`
	Canceled int64 `json:"canceled"`
}

type RepairRecordListRequest struct {
	BuildingId int64  `form:"buildingId,optional"`
	Status     int64  `form:"status,optional"`
	FaultType  string `form:"faultType,optional"`
	Keyword    string `form:"keyword,optional"`
	Days       int64  `form:"days,optional"`
	StartDate  string `form:"startDate,optional"`
	EndDate    string `form:"endDate,optional"`
	Page       int64  `form:"page,optional"`
	Size       int64  `form:"size,optional"`
}

type RepairRecordListResponse struct {
	Total   int64               `json:"total"`
	Summary RepairRecordSummary `json:"summary"`
	List    []RepairRecordItem  `json:"list"`
}
