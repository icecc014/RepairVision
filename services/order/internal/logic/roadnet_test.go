package logic

import (
	"encoding/json"
	"testing"
)

// 构造 12x6 平面图：第 3 行是主干道，两栋建筑分别贴在道路两侧。
func testLayoutJSON(t *testing.T) string {
	t.Helper()
	layout := map[string]any{
		"version": 1,
		"cols":    12,
		"rows":    6,
		"blocks": []map[string]any{
			{"id": "r1", "kind": "road", "row": 3, "col": 0, "rowSpan": 1, "colSpan": 12},
			{"id": "b1", "kind": "building", "row": 1, "col": 1, "rowSpan": 2, "colSpan": 2, "buildingId": 1, "label": "1号宿舍楼"},
			{"id": "b2", "kind": "building", "row": 1, "col": 8, "rowSpan": 2, "colSpan": 2, "buildingId": 2, "label": "2号宿舍楼"},
			{"id": "b3", "kind": "custom", "row": 5, "col": 10, "rowSpan": 1, "colSpan": 1, "label": "食堂"},
		},
	}
	raw, err := json.Marshal(layout)
	if err != nil {
		t.Fatalf("marshal layout: %v", err)
	}
	return string(raw)
}

func TestBuildRoadNetworkDistance(t *testing.T) {
	net := buildRoadNetwork(testLayoutJSON(t), 10)
	if net == nil {
		t.Fatal("期望构建出路网，实际为 nil")
	}
	d, ok := net.DistanceBetween(1, 2)
	if !ok {
		t.Fatal("1 号楼与 2 号楼应可通过道路连通")
	}
	// 1 号楼入口格为 (3,1)、(3,2)（与主干道相邻），2 号楼入口格为 (3,8)、(3,9)；
	// 取最短入口组合 (3,2) -> (3,8) 共 6 格 = 60 米。
	if d != 60 {
		t.Fatalf("期望 60 米，实际 %v 米", d)
	}
	if d2, ok2 := net.DistanceBetween(2, 1); !ok2 || d2 != d {
		t.Fatalf("距离应对称：expect %v, got %v (ok=%v)", d, d2, ok2)
	}
	if d3, ok3 := net.DistanceBetween(1, 1); !ok3 || d3 != 0 {
		t.Fatalf("同楼栋距离应为 0，实际 %v (ok=%v)", d3, ok3)
	}
	if _, ok4 := net.DistanceBetween(1, 99); ok4 {
		t.Fatal("未在布局中的楼栋不应有路网距离")
	}
	if !net.IsIsolated(99) {
		t.Fatal("未在布局中的楼栋应视为孤立")
	}
	if got := len(net.Buildings()); got != 2 {
		t.Fatalf("布局中应有 2 栋建筑，实际 %d", got)
	}
}

func TestBuildRoadNetworkIsolatedBuilding(t *testing.T) {
	layout := map[string]any{
		"version": 1,
		"cols":    8,
		"rows":    6,
		"blocks": []map[string]any{
			{"id": "r1", "kind": "road", "row": 0, "col": 0, "rowSpan": 1, "colSpan": 8},
			{"id": "b1", "kind": "building", "row": 1, "col": 0, "rowSpan": 1, "colSpan": 1, "buildingId": 11},
			{"id": "b2", "kind": "building", "row": 5, "col": 6, "rowSpan": 1, "colSpan": 1, "buildingId": 12},
		},
	}
	raw, _ := json.Marshal(layout)
	net := buildRoadNetwork(string(raw), 10)
	if net == nil {
		t.Fatal("期望构建出路网")
	}
	if d, ok := net.DistanceBetween(11, 12); ok {
		t.Fatalf("孤立建筑不应有路网距离，实际 %v", d)
	}
	if !net.IsIsolated(12) {
		t.Fatal("12 号楼未接道路，应标记为孤立")
	}
}

func TestBuildRoadNetworkDegraded(t *testing.T) {
	if n := buildRoadNetwork("", 10); n != nil {
		t.Fatal("空布局应返回 nil")
	}
	if n := buildRoadNetwork("{", 10); n != nil {
		t.Fatal("非法 JSON 应返回 nil")
	}
	noRoad := `{"version":1,"cols":4,"rows":4,"blocks":[{"id":"b1","kind":"building","row":0,"col":0,"rowSpan":1,"colSpan":1,"buildingId":1}]}`
	if n := buildRoadNetwork(noRoad, 10); n != nil {
		t.Fatal("无道路图元应返回 nil（回退欧氏距离）")
	}
}