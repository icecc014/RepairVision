package logic

import (
	"testing"

	"map/mapclient"
	"worker/workerclient"
)

func TestWorkerCanTake(t *testing.T) {
	off := &workerclient.WorkerInfo{Id: 1, MaxConcurrent: 3, TodayShift: "OFF"}
	if workerCanTake(off, map[int64]int64{}) {
		t.Fatal("OFF worker should not take")
	}
	full := &workerclient.WorkerInfo{Id: 2, MaxConcurrent: 2, TodayShift: "DAY"}
	if workerCanTake(full, map[int64]int64{2: 2}) {
		t.Fatal("full worker should not take")
	}
	ok := &workerclient.WorkerInfo{Id: 3, MaxConcurrent: 2, TodayShift: "DAY"}
	if !workerCanTake(ok, map[int64]int64{3: 1}) {
		t.Fatal("worker with free slot should take")
	}
}

func TestScoreWorkerForOrder(t *testing.T) {
	buildings := []*mapclient.Building{
		{Id: 1, PosX: 0, PosY: 0},
		{Id: 2, PosX: 100, PosY: 0},
	}
	expert := &workerclient.WorkerInfo{
		Id:             1,
		BaseBuildingId: 1,
		Skills:         []*workerclient.SkillInfo{{Name: "电维修", Proficiency: 3}},
	}
	score := scoreWorkerForOrder(expert, buildings, buildings[0], "电维修",
		map[int64]int64{}, 1, 0, 0)
	if score.skillScore != 1 || score.totalScore != 1 {
		t.Fatalf("expected full score, got %+v", score)
	}
}

func TestPickBestOrderSkipsOffAndOverload(t *testing.T) {
	buildings := []*mapclient.Building{
		{Id: 1, PosX: 0, PosY: 0},
		{Id: 2, PosX: 50, PosY: 0},
	}
	off := &workerclient.WorkerInfo{Id: 1, MaxConcurrent: 3, TodayShift: "OFF", BaseBuildingId: 1}
	best := pickBestOrder([]*workerclient.WorkerInfo{off}, buildings, buildings[0], "电维修",
		map[int64]int64{}, map[int64]int64{}, 1, 0, 0)
	if best != nil {
		t.Fatalf("expected no pick, got %+v", best)
	}
}
