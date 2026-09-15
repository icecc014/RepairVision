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
	// V5.4：取消硬性并发上限，满载工人仍参与派单（由动态负载惩罚降权）。
	full := &workerclient.WorkerInfo{Id: 2, MaxConcurrent: 2, TodayShift: "DAY"}
	if !workerCanTake(full, map[int64]int64{2: 2}) {
		t.Fatal("V5.4 取消硬上限：满载工人仍应参与派单")
	}
	if !workerOverloaded(full, map[int64]int64{2: 2}) {
		t.Fatal("满载工人应在展示层被标记为负载偏高")
	}
	ok := &workerclient.WorkerInfo{Id: 3, MaxConcurrent: 2, TodayShift: "DAY"}
	if !workerCanTake(ok, map[int64]int64{3: 1}) {
		t.Fatal("worker with free slot should take")
	}
	if workerOverloaded(ok, map[int64]int64{3: 1}) {
		t.Fatal("有空闲容量的工人不应标记为负载偏高")
	}
}

func TestLoadPenalty(t *testing.T) {
	if p := loadPenalty(80, 100); p != 0 {
		t.Fatalf("低于人均负载不惩罚，实际 %v", p)
	}
	if p := loadPenalty(120, 100); p != 0 {
		t.Fatalf("恰好 1.2 倍人均负载不惩罚，实际 %v", p)
	}
	if p := loadPenalty(150, 100); p <= 0 {
		t.Fatalf("高于 1.2 倍人均负载应惩罚，实际 %v", p)
	}
	if p := loadPenalty(50, 0); p != 0 {
		t.Fatalf("人均负载为 0 时不惩罚，实际 %v", p)
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
	score := scoreWorkerForOrder(expert, scoreInput{
		buildings:   buildings,
		current:     buildings[0],
		faultName:   "电维修",
		minuteLoads: map[int64]int64{},
		wSkill:      1,
	})
	if score.skillScore != 1 || score.totalScore != 1 {
		t.Fatalf("expected full score, got %+v", score)
	}
}

func TestPickBestOrderPrefersFreeWorker(t *testing.T) {
	buildings := []*mapclient.Building{
		{Id: 1, PosX: 0, PosY: 0},
		{Id: 2, PosX: 100, PosY: 0},
	}
	busy := &workerclient.WorkerInfo{
		Id: 1, BaseBuildingId: 1, TodayShift: "DAY",
		Skills: []*workerclient.SkillInfo{{Name: "电维修", Proficiency: 3}},
	}
	free := &workerclient.WorkerInfo{
		Id: 2, BaseBuildingId: 1, TodayShift: "DAY",
		Skills: []*workerclient.SkillInfo{{Name: "电维修", Proficiency: 3}},
	}
	in := scoreInput{
		buildings:   buildings,
		current:     buildings[0],
		faultName:   "电维修",
		minuteLoads: map[int64]int64{1: 180, 2: 30},
		avgLoad:     105,
		wSkill:      0.4,
		wDistance:   0.3,
		wLoad:       0.3,
	}
	best := pickBestOrder([]*workerclient.WorkerInfo{busy, free}, in, 0)
	if best == nil || best.workerID != 2 {
		t.Fatalf("应优先选择负载低的工人 2，实际 %+v", best)
	}
	// 高负载工人的惩罚生效：同一输入的 totalScore 应低于空闲工人
	busyScore := scoreWorkerForOrder(busy, in)
	freeScore := scoreWorkerForOrder(free, in)
	if busyScore.totalScore >= freeScore.totalScore {
		t.Fatalf("高负载工人得分应更低：busy=%v free=%v", busyScore.totalScore, freeScore.totalScore)
	}
}

func TestPickBestOrderSkipsOffWorker(t *testing.T) {
	buildings := []*mapclient.Building{
		{Id: 1, PosX: 0, PosY: 0},
		{Id: 2, PosX: 50, PosY: 0},
	}
	off := &workerclient.WorkerInfo{Id: 1, MaxConcurrent: 3, TodayShift: "OFF", BaseBuildingId: 1}
	best := pickBestOrder([]*workerclient.WorkerInfo{off}, scoreInput{
		buildings: buildings, current: buildings[0], faultName: "电维修",
		minuteLoads: map[int64]int64{}, wSkill: 1,
	}, 0)
	if best != nil {
		t.Fatalf("expected no pick, got %+v", best)
	}
}

// 路网可用时距离因子按路网最短路归一化；同楼栋保持满分。
func TestDistanceScoreWithRoadNetwork(t *testing.T) {
	net := buildRoadNetwork(testLayoutJSON(t), 10)
	if net == nil {
		t.Fatal("测试布局应能构建路网")
	}
	buildings := []*mapclient.Building{{Id: 1, PosX: 0, PosY: 0}, {Id: 2, PosX: 100, PosY: 0}}
	same := &workerclient.WorkerInfo{Id: 1, BaseBuildingId: 1, TodayShift: "DAY"}
	other := &workerclient.WorkerInfo{Id: 2, BaseBuildingId: 1, TodayShift: "DAY"}
	inSame := scoreInput{buildings: buildings, current: buildings[0], roadNet: net, minuteLoads: map[int64]int64{}, wDistance: 1}
	if got := scoreWorkerForOrder(same, inSame).distanceScore; got != 1 {
		t.Fatalf("同楼栋距离得分应为 1，实际 %v", got)
	}
	inOther := scoreInput{buildings: buildings, current: buildings[1], roadNet: net, minuteLoads: map[int64]int64{}, wDistance: 1}
	got := scoreWorkerForOrder(other, inOther).distanceScore
	if got < 0 || got >= 1 {
		t.Fatalf("跨楼栋路网距离得分应在 [0,1)，实际 %v", got)
	}
}
// V6：类别 → 所需工种映射
func TestRequiredJobTypeOf(t *testing.T) {
	cases := []struct {
		category string
		want     int64
	}{
		{"electric", 1}, {"water", 2}, {"masonry", 3}, {"wood", 4}, {"other", 0}, {"", 0}, {"UNKNOWN", 0},
	}
	for _, c := range cases {
		if got := requiredJobTypeOf(c.category); got != c.want {
			t.Fatalf("类别 %q 期望工种 %d，实际 %d", c.category, c.want, got)
		}
	}
}

// V6：4 类工种 × 4 类需求的匹配矩阵（通用工人可接任意类别）
func TestJobTypeMatrix(t *testing.T) {
	expect := map[int64]map[int64]bool{
		1: {1: true, 2: false, 3: false, 4: false},  // 电工
		2: {1: false, 2: true, 3: false, 4: false},  // 水工
		3: {1: false, 2: false, 3: true, 4: false},  // 泥瓦工
		4: {1: false, 2: false, 3: false, 4: true},  // 木工
		0: {1: true, 2: true, 3: true, 4: true},     // 通用
	}
	for workerJobType, row := range expect {
		for required := int64(1); required <= 4; required++ {
			want := row[required]
			if got := jobTypeAllowed(workerJobType, required); got != want {
				t.Fatalf("工人工种 %d × 需求 %d：期望 %v，实际 %v", workerJobType, required, want, got)
			}
		}
	}
}