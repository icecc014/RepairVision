package types

// CampusLayoutResponse 区域概览（供管理端编辑与三端只读查看）。
type CampusLayoutResponse struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Cols       int64  `json:"cols"`
	Rows       int64  `json:"rows"`
	LayoutJson string `json:"layoutJson"`
	UpdatedAt  string `json:"updatedAt"`
}

// CampusLayoutSaveRequest 保存区域概览。
type CampusLayoutSaveRequest struct {
	Name       string `json:"name,optional"`
	Cols       int64  `json:"cols,optional"`
	Rows       int64  `json:"rows,optional"`
	LayoutJson string `json:"layoutJson"`
}
