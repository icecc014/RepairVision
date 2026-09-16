package logic

import "testing"

func TestRoadNetworkMiniCase(t *testing.T) {
	mini := `{"version":1,"cols":12,"rows":6,"blocks":[{"id":"r1","kind":"road","row":2,"col":0,"rowSpan":1,"colSpan":12},{"id":"b1","kind":"building","row":3,"col":1,"rowSpan":2,"colSpan":2,"buildingId":7,"label":"test"}]}`
	net := buildRoadNetwork(mini, 10)
	if net == nil {
		t.Fatal("net is nil: 解析失败或道路收集为 0")
	}
	t.Logf("roadCells=%d buildings=%v", net.RoadCellCount(), net.Buildings())
}