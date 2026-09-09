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

// candidateScore 记录单个候选工人的三维得分与总分。
type candidateScore struct {
	workerID      int64
	skillScore    float64
	distanceScore float64
	loadScore     float64
	totalScore    float64
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

func distance(a, b *mapclient.Building) float64 {
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

// workerCanTake 容量与当班约束：在途单数未达最大并发、且当天不是 OFF。
func workerCanTake(w *workerclient.WorkerInfo, countLoads map[int64]int64) bool {
	if w.TodayShift == shiftOff {
		return false
	}
	return countLoads[w.Id] < maxConcurrentOf(w)
}

// scoreWorkerForOrder 计算单个工人在某工单上的三维得分（不校验容量，供改派记录用）。
func scoreWorkerForOrder(
	w *workerclient.WorkerInfo,
	buildings []*mapclient.Building,
	current *mapclient.Building,
	faultName string,
	minuteLoads map[int64]int64,
	wSkill, wDistance, wLoad float64,
) candidateScore {
	score := candidateScore{workerID: w.Id}
	score.skillScore = skillScore(w.Skills, faultName)
	load := minuteLoads[w.Id]
	maxLoad := maxConcurrentOf(w) * 30
	if maxLoad <= 0 {
		maxLoad = defaultMaxConcurrent * 30
	}
	if load > maxLoad {
		load = maxLoad
	}
	score.loadScore = 1 - float64(load)/float64(maxLoad)

	buildingPos := make(map[int64]*mapclient.Building, len(buildings))
	for _, b := range buildings {
		buildingPos[b.Id] = b
	}
	base := w.BaseBuildingId
	if base <= 0 {
		base = current.Id
	}
	if baseB, ok := buildingPos[base]; ok {
		maxDist := 0.0
		for _, b := range buildings {
			d := distance(baseB, b)
			if d > maxDist {
				maxDist = d
			}
		}
		d := distance(baseB, current)
		if maxDist > 0 {
			score.distanceScore = 1 - d/maxDist
		} else {
			score.distanceScore = 1
		}
	} else {
		score.distanceScore = 0
	}

	total := wSkill*score.skillScore + wDistance*score.distanceScore + wLoad*score.loadScore
	sum := wSkill + wDistance + wLoad
	if sum > 0 {
		total /= sum
	}
	score.totalScore = math.Round(total*10000) / 10000
	return score
}

// pickBestOrder 在候选工人中选得分最高者；同分取 id 较小者。
func pickBestOrder(
	workers []*workerclient.WorkerInfo,
	buildings []*mapclient.Building,
	current *mapclient.Building,
	faultName string,
	countLoads map[int64]int64,
	minuteLoads map[int64]int64,
	wSkill, wDistance, wLoad float64,
) *candidateScore {
	var best *candidateScore
	for _, w := range workers {
		if !workerCanTake(w, countLoads) {
			continue
		}
		score := scoreWorkerForOrder(w, buildings, current, faultName, minuteLoads, wSkill, wDistance, wLoad)
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
