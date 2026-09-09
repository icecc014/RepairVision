package logic

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/mapclient"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
	"worker/workerclient"
)

type candidateScore struct {
	workerID      int64
	skillScore    float64
	distanceScore float64
	loadScore     float64
	totalScore    float64
}

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderRequest) (resp *types.CreateOrderResponse, err error) {
	identity, ok := auth.IdentityFromContext(l.ctx)
	if !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	room := strings.TrimSpace(req.Room)
	if room == "" || req.FaultType == "" {
		return nil, errs.BadRequest("房间号和维修类型不能为空")
	}
	floor := req.Floor
	if floor <= 0 {
		if len(room) < 2 {
			return nil, errs.BadRequest("房间号格式不正确，如 401 表示 4 层 01 房")
		}
		first := int(room[0] - '0')
		if first < 1 || first > 9 {
			return nil, errs.BadRequest("房间号第一位必须是楼层数字")
		}
		floor = int64(first)
	} else if !strings.HasPrefix(room, fmt.Sprintf("%d", floor)) {
		return nil, errs.BadRequest(fmt.Sprintf("楼层 %d 的房间号应以 %d 开头，如 %d01", floor, floor, floor))
	}

	faultType, err := store.FindFaultTypeByCode(l.ctx, l.svcCtx.DB, req.FaultType)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.BadRequest("维修类型不存在")
		}
		return nil, errs.Internal(err)
	}

	buildingID := identity.BuildingID
	if buildingID <= 0 {
		return nil, errs.Forbidden("当前账号未绑定楼栋")
	}

	// F16 粗粒度去重：同一楼栋/楼层/房间/类型，2小时内未关闭工单则阻止重复报修
	dup, err := store.FindRecentDuplicateOrder(l.ctx, l.svcCtx.DB, buildingID, floor, req.Room, req.FaultType, time.Now().Add(-2*time.Hour))
	if err == nil && dup != nil {
		return nil, errs.Conflict(fmt.Sprintf("该房间近期已有同类维修工单（%s），请勿重复报修", dup.OrderNo))
	}
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errs.Internal(err)
	}

	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	var currentBuilding *mapclient.Building
	for _, b := range buildingResp.Buildings {
		if b.Id == buildingID {
			currentBuilding = b
			break
		}
	}
	if currentBuilding == nil {
		return nil, errs.BadRequest("楼栋尚未初始化")
	}

	// F17 加权派单候选
	if err := validateRoomForBuilding(req.Room, floor, currentBuilding); err != nil {
		return nil, err
	}
	workerResp, err := l.svcCtx.WorkerRpc.ListWorkersByBuilding(l.ctx,
		&workerclient.BuildingWorkersRequest{BuildingId: buildingID})
	if err != nil {
		return nil, errs.Upstream()
	}

	rules, err := store.ListDispatchRules(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	skillW, distW, loadW := dispatchWeights(rules)
	autoDispatch := true
	for _, r := range rules {
		if r.RuleKey == "auto_dispatch_enabled" && (r.Enabled == 0 || r.RuleValue < 0.5) {
			autoDispatch = false
		}
	}

	now := time.Now()
	orderNo := fmt.Sprintf("REP%s%09d", now.Format("20060102"), now.UnixNano()%1000000000)
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = "无补充说明"
	}

	var best *candidateScore
	if autoDispatch && len(workerResp.Workers) > 0 {
		loads, err := store.CountInProgressByWorkers(l.ctx, l.svcCtx.DB, workerIDs(workerResp.Workers))
		if err != nil {
			return nil, errs.Internal(err)
		}
		best = l.pickBest(workerResp.Workers, buildingResp.Buildings, currentBuilding, faultType.Name, loads, skillW, distW, loadW)
	}

	var orderID int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		order := &store.Order{
			OrderNo:       orderNo,
			Title:         faultType.Name + "（" + req.Room + "室）",
			Description:   description,
			BuildingID:    buildingID,
			Room:          room,
			Floor:         floor,
			FaultType:     req.FaultType,
			Priority:      1,
			ExpectMinutes: 30,
			Status:        store.StatusPending,
			IsMerged:      0,
			ReporterID:    identity.UID,
			Source:        "dormitory",
		}
		id, err := store.InsertOrder(txCtx, session, order)
		if err != nil {
			return err
		}
		orderID = id
		if best != nil {
			if err := store.AssignOrder(txCtx, session, orderID, best.workerID); err != nil {
				return err
			}
			return store.InsertDispatchRecord(txCtx, session, orderID, best.workerID,
				best.totalScore, best.skillScore, best.distanceScore, best.loadScore)
		}
		return nil
	})
	if err != nil {
		return nil, errs.Internal(err)
	}
	if _, markerErr := l.svcCtx.MapRpc.UpsertFaultMarker(l.ctx, &mapclient.FaultMarkerUpsertRequest{
		OrderId: orderID, BuildingId: buildingID, Floor: floor, RoomNumber: room,
	}); markerErr != nil {
		logx.WithContext(l.ctx).Errorf("upsert fault marker failed: %v", markerErr)
	}
	status := store.StatusPending
	workerID := int64(0)
	if best != nil {
		status = store.StatusDispatched
		workerID = best.workerID
	}
	l.svcCtx.WS.PublishOrder(ws.OrderEvent{
		Type: "order_changed", OrderId: orderID, OrderNo: orderNo,
		BuildingId: buildingID, WorkerId: workerID, Status: status,
	})
	return &types.CreateOrderResponse{OrderId: orderID, OrderNo: orderNo}, nil
}

func workerIDs(workers []*workerclient.WorkerInfo) []int64 {
	ids := make([]int64, 0, len(workers))
	for _, w := range workers {
		ids = append(ids, w.Id)
	}
	return ids
}

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

func (l *CreateOrderLogic) pickBest(
	workers []*workerclient.WorkerInfo,
	buildings []*mapclient.Building,
	current *mapclient.Building,
	faultName string,
	loads map[int64]int64,
	wSkill, wDistance, wLoad float64,
) *candidateScore {
	buildingPos := make(map[int64]*mapclient.Building)
	for _, b := range buildings {
		buildingPos[b.Id] = b
	}

	maxLoad := int64(0)
	maxDist := 0.0
	for _, w := range workers {
		cnt := loads[w.Id]
		if cnt > maxLoad {
			maxLoad = cnt
		}
		base := w.BaseBuildingId
		if base <= 0 {
			base = current.Id
		}
		if baseB, ok := buildingPos[base]; ok {
			d := distance(baseB, current)
			if d > maxDist {
				maxDist = d
			}
		}
	}

	var best *candidateScore
	for _, w := range workers {
		score := candidateScore{workerID: w.Id}
		score.skillScore = skillScore(w.Skills, faultName)
		cnt := loads[w.Id]
		if maxLoad <= 0 {
			score.loadScore = 1
		} else {
			score.loadScore = 1 - float64(cnt)/float64(maxLoad)
		}
		base := w.BaseBuildingId
		if base <= 0 {
			base = current.Id
		}
		if baseB, ok := buildingPos[base]; ok {
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
		if best == nil || score.totalScore > best.totalScore ||
			(score.totalScore == best.totalScore && score.workerID < best.workerID) {
			cp := score
			best = &cp
		}
	}
	return best
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
