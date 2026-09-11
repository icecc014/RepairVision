package types

// WorkerMapDataRequest 工人端地图/楼层数据请求：days 为报修时间窗（1/3/7/30，默认 3 天）。
type WorkerMapDataRequest struct {
	Days int64 `form:"days,optional"`
}
