package logic

import (
	"context"
	"math"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/mapclient"
	"order/internal/store"
	"order/internal/svc"
	"worker/workerclient"
)

const (
	defaultMaxConcurrent = int64(3)
	shiftOff             = "OFF"
)

// softExtraSlots 批量派单的软上限冗余：并发上限之上再允许 2 个槽位。
const softExtraSlots = int64(2)

// 负载偏离惩罚系数：总得分中最多扣 0.5。
const loadPenaltyWeight = 0.5

// candidateScore 记录单个候选工人的三维得分与总分。
type candidateScore struct {
	workerID      int64
	skillScore    float64
	distanceScore float64
	loadScore     float64
	totalScore    float64
}

// scoreInput 汇总一次评分的全部输入（工单上下文 + 权重 + 路网/负载信息）。
type scoreInput struct {
	buildings   []*mapclient.Building
	current     *mapclient.Building
	faultName   string
	minuteLoads map[int64]int64
	roadNet     *RoadNetwork
	avgLoad     float64
	wSkill      float64
	wDistance   float64
	wLoad       float64
}

// dispatchWeights 读取派单规则权重，缺省 skill/distance/load = 0.4/0.3/0.3。
func dispatchWeights(rules []store.DispatchRule) (skill, distance, load float64) {
	skill, distance, load = 0.4, 0.3, 0.3
	for _, r := range rules {
		switch r.RuleKey {
		case "skill_weight":
			skill = r.RuleValue
		case "distance_weight":
			distance = r.RuleValue
		case "load_weight":
			load = r.RuleValue
		}
	}
	return
}

func skillScore(skills []*workerclient.SkillInfo, faultName string) float64 {
	for _, s := range skills {
		if s.Name == faultName {
			if s.Proficiency >= 3 {
				return 1
			}
			if s.Proficiency <= 1 {
				return 1.0 / 3.0
			}
			return 2.0 / 3.0
		}
	}
	return 0
}

func euclidDistance(a, b *mapclient.Building) float64 {
	dx := a.PosX - b.PosX
	dy := a.PosY - b.PosY
	return math.Sqrt(dx*dx + dy*dy)
}

func maxConcurrentOf(w *workerclient.WorkerInfo) int64 {
	if w.MaxConcurrent > 0 {
		return w.MaxConcurrent
	}
	return defaultMaxConcurrent
}

// workerCanTake 仅保留"当班"约束（当天休息 OFF 不派单）。
// V5.4 起取消硬性并发上限：负载通过 loadPenalty 施加软惩罚，不再因为"已满"而拒派。
func workerCanTake(w *workerclient.WorkerInfo, _ map[int64]int64) bool {
	return w.TodayShift != shiftOff
}

// workerOverloaded 负载参考：在途单数达到/超过最大并发，仅用于展示与告警。
func workerOverloaded(w *workerclient.WorkerInfo, countLoads map[int64]int64) bool {
	return countLoads[w.Id] >= maxConcurrentOf(w)
}

// avgLoadOf 在岗候选人均负载（分钟）：在途工时总和 / 候选人数。
func avgLoadOf(minuteLoads map[int64]int64, workers []*workerclient.WorkerInfo) float64 {
	if len(workers) == 0 {
		return 0
	}
	sum := 0.0
	for _, w := range workers {
		sum += float64(minuteLoads[w.Id])
	}
	return sum / float64(len(workers))
}

// loadPenalty 动态负载均衡惩罚：仅当工人工时负载高于人均 1.2 倍时生效。
func loadPenalty(load, avgLoad float64) float64 {
	if avgLoad <= 0 || load <= avgLoad*1.2 {
		return 0
	}
	return (load - avgLoad) / avgLoad
}

// currentRoadNet 取进程内缓存的校园路网（区域概览道路图元构建，布局变更自动重建）。
func currentRoadNet(ctx context.Context, svcCtx *svc.ServiceContext) *RoadNetwork {
	return campusRoadNet.network(ctx, svcCtx)
}

// 进程内路网缓存：key 为区域概览 JSON，变化即重建。
var campusRoadNet roadNetCache

// distanceScoreOf 距离因子：优先路网最短路（同楼栋满分），无路网数据时回退楼栋坐标欧氏距离。
func distanceScoreOf(in scoreInput, w *workerclient.WorkerInfo) float64 {
	buildingPos := make(map[int64]*mapclient.Building, len(in.buildings))
	for _, b := range in.buildings {
		buildingPos[b.Id] = b
	}
	base := w.BaseBuildingId
	if base <= 0 {
		base = in.current.Id
	}
	if base == in.current.Id {
		return 1 // 同楼栋加成
	}
	if in.roadNet != nil {
		if d, ok := in.roadNet.DistanceBetween(base, in.current.Id); ok {
			if maxRoad := in.roadNet.MaxDistance(); maxRoad > 0 {
				return clamp01(1 - d/maxRoad)
			}
			return 1
		}
	}
	baseB, ok := buildingPos[base]
	if !ok {
		return 0
	}
	maxDist := 0.0
	for _, b := range in.buildings {
		if d := euclidDistance(baseB, b); d > maxDist {
			maxDist = d
		}
	}
	if maxDist <= 0 {
		return 1
	}
	return clamp01(1 - euclidDistance(baseB, in.current)/maxDist)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// scoreWorkerForOrder 计算单个工人对某工单的综合得分：
// 技能 40% + 距离 30% + 负载 30%，再扣除动态负载偏离惩罚（最多 0.5）。
func scoreWorkerForOrder(w *workerclient.WorkerInfo, in scoreInput) candidateScore {
	score := candidateScore{workerID: w.Id}
	score.skillScore = skillScore(w.Skills, in.faultName)
	score.distanceScore = distanceScoreOf(in, w)

	load := in.minuteLoads[w.Id]
	maxLoad := maxConcurrentOf(w) * 30
	if maxLoad <= 0 {
		maxLoad = defaultMaxConcurrent * 30
	}
	if load > maxLoad {
		load = maxLoad
	}
	score.loadScore = 1 - float64(load)/float64(maxLoad)

	total := in.wSkill*score.skillScore + in.wDistance*score.distanceScore + in.wLoad*score.loadScore
	if sum := in.wSkill + in.wDistance + in.wLoad; sum > 0 {
		total /= sum
	}
	total -= loadPenaltyWeight * loadPenalty(float64(in.minuteLoads[w.Id]), in.avgLoad)
	score.totalScore = math.Round(clamp01(total)*10000) / 10000
	return score
}

// pickBestOrder 在候选工人中选得分最高者；同分取 id 较小者。
func pickBestOrder(
	workers []*workerclient.WorkerInfo,
	in scoreInput,
	requiredJobType int64,
) *candidateScore {
	var best *candidateScore
	for _, w := range workers {
		if !workerCanTake(w, nil) {
			continue
		}
		// V5.2 工种匹配：电类只派电工/通用，水类只派水工/通用
		if !jobTypeAllowed(w.JobType, requiredJobType) {
			continue
		}
		score := scoreWorkerForOrder(w, in)
		if best == nil || score.totalScore > best.totalScore ||
			(score.totalScore == best.totalScore && score.workerID < best.workerID) {
			cp := score
			best = &cp
		}
	}
	return best
}

// faultTypeNameMap 读取维修类型字典：code -> 显示名（工人技能按显示名匹配）。
func faultTypeNameMap(ctx context.Context, svcCtx *svc.ServiceContext) (map[string]string, error) {
	list, err := store.ListFaultTypes(ctx, svcCtx.DB)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(list))
	for _, ft := range list {
		result[ft.Code] = ft.Name
	}
	return result, nil
}

// loadMinutesByWorkers 返回每个工人在途单的预计工时总量（分钟），无字段时按 30 分钟兜底。
func loadMinutesByWorkers(ctx context.Context, conn sqlx.Session, workerIDs []int64) (map[int64]int64, error) {
	return store.CountWorkloadByWorkers(ctx, conn, workerIDs)
}

// jobTypeAllowed 工种匹配：required 1=电工 2=水工 0=不限；工人 0 表示通用可接两类。
func jobTypeAllowed(workerJobType, required int64) bool {
	if required <= 0 || workerJobType <= 0 {
		return true
	}
	return workerJobType == required
}