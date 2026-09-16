package logic

import (
	"context"
	"errors"
	"math"
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

const (
	batchLimit     = 100
	batchDummyCost = int64(10_000_000)
)

var errBatchSkip = errors.New("batch skip")

type AdminBatchDispatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminBatchDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminBatchDispatchLogic {
	return &AdminBatchDispatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminBatchDispatch 对待派队列做容量约束的最小代价批量指派。
// orderIds 为空时使用楼栋筛选；两者都为空时派全部待派单（最多 batchLimit 条）。
func (l *AdminBatchDispatchLogic) AdminBatchDispatch(req *types.AdminBatchDispatchRequest) (resp *types.AdminBatchDispatchResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}

	orders, err := l.loadPendingOrders(req)
	if err != nil {
		return nil, errs.Internal(err)
	}
	resp = &types.AdminBatchDispatchResponse{Dispatched: []types.AdminBatchDispatchItem{}}
	if len(orders) == 0 {
		return resp, nil
	}

	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	buildingByID := make(map[int64]*mapclient.Building, len(buildingResp.Buildings))
	for _, b := range buildingResp.Buildings {
		buildingByID[b.Id] = b
	}
	faultNames, err := faultTypeNameMap(l.ctx, l.svcCtx)
	if err != nil {
		return nil, errs.Internal(err)
	}
	// V6.1 工种过滤：批量派单同样要按故障类型类别限制候选人
	faultCategories, err := faultTypeCategoryMap(l.ctx, l.svcCtx)
	if err != nil {
		return nil, errs.Internal(err)
	}
	rules, err := store.ListDispatchRules(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	wSkill, wDistance, wLoad := dispatchWeights(rules)

	workerByBuilding := make(map[int64][]*workerclient.WorkerInfo)
	buildingIDs := uniqueOrderBuildingIDs(orders)
	for _, bid := range buildingIDs {
		wr, err := l.svcCtx.WorkerRpc.ListWorkersByBuilding(l.ctx,
			&workerclient.BuildingWorkersRequest{
				BuildingId: bid,
				WorkDate:   time.Now().Format("2006-01-02"),
			})
		if err != nil {
			return nil, errs.Upstream()
		}
		workerByBuilding[bid] = wr.Workers
	}

	allWorkerIDs := make([]int64, 0)
	for _, ws := range workerByBuilding {
		for _, w := range ws {
			allWorkerIDs = append(allWorkerIDs, w.Id)
		}
	}
	countLoads, err := store.CountInProgressByWorkers(l.ctx, l.svcCtx.DB, allWorkerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}
	minuteLoads, err := store.CountWorkloadByWorkers(l.ctx, l.svcCtx.DB, allWorkerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}

	// 为每个“当班且有剩余容量”的工人建立容量槽位。
	type slot struct {
		worker *workerclient.WorkerInfo
	}
	// V5.4：取消硬性并发上限，每个在岗工人只占一个指派槽位；
	// 负载改由评分中的动态惩罚体现（高于人均 1.2 倍才扣分），不再因"已满"而拒派。
	slots := make([]slot, 0)
	seenWorker := map[int64]bool{}
	allWorkers := make([]*workerclient.WorkerInfo, 0)
	for _, ws := range workerByBuilding {
		allWorkers = append(allWorkers, ws...)
	}
	leaveSet := onLeaveWorkerIDs(l.ctx, l.svcCtx, allWorkers)
	for _, ws := range workerByBuilding {
		for _, w := range ws {
			if seenWorker[w.Id] || !workerCanTake(w, countLoads) || leaveSet[w.Id] {
				continue
			}
			seenWorker[w.Id] = true
			// 软上限：并发上限 + softExtraSlots 个槽位。超过后不再拒派，
			// 而是由评分中的动态负载惩罚降低该工人的优先级。
			for i := int64(0); i < maxConcurrentOf(w)+softExtraSlots; i++ {
				slots = append(slots, slot{worker: w})
			}
		}
	}
	slotWorkers := make([]*workerclient.WorkerInfo, 0, len(slots))
	for _, s := range slots {
		slotWorkers = append(slotWorkers, s.worker)
	}
	roadNet := currentRoadNet(l.ctx, l.svcCtx)
	avgLoad := avgLoadOf(minuteLoads, slotWorkers)

	type planOrder struct {
		order    store.Order
		building *mapclient.Building
		// V6.1：该工单需要的工种（由故障类型类别推导，0 = 不限）
		requiredJobType int64
	}
	plan := make([]planOrder, 0, len(orders))
	remained := int64(0)
	for _, o := range orders {
		b, ok := buildingByID[o.BuildingID]
		if !ok || buildingByID[o.BuildingID] == nil {
			remained++
			continue
		}
		// V6.1 工种匹配：电/水/泥瓦/木各归其位，通用工人仍可接任意类别
		requiredJobType := orderRequiredJobType(faultCategories, o.FaultType)
		hasEligible := false
		for _, w := range workerByBuilding[o.BuildingID] {
			if workerCanTake(w, countLoads) && jobTypeAllowed(w.JobType, requiredJobType) {
				hasEligible = true
				break
			}
		}
		if !hasEligible {
			// 该楼栋没有工种匹配的可用工人：留在待派队列，交管理员处置
			logx.WithContext(l.ctx).Infof("batch dispatch skip order %s: no worker matches job type %d",
				o.OrderNo, requiredJobType)
			remained++
			continue
		}
		plan = append(plan, planOrder{order: o, building: b, requiredJobType: requiredJobType})
	}
	if len(plan) == 0 || len(slots) == 0 {
		resp.Remained = int64(len(plan)) + remained
		return resp, nil
	}

	nOrders := len(plan)
	nSlots := len(slots)
	n := nOrders
	if nSlots > n {
		n = nSlots
	}
	cost := make([][]int64, n)
	for i := range cost {
		cost[i] = make([]int64, n)
		for j := range cost[i] {
			if j >= nSlots {
				cost[i][j] = batchDummyCost
			} else {
				cost[i][j] = hungarianCostINF
			}
		}
	}
	for i, po := range plan {
		workers := workerByBuilding[po.order.BuildingID]
		for c := 0; c < nSlots; c++ {
			if !containsWorker(workers, slots[c].worker.Id) {
				continue
			}
			// V6.1 工种过滤：工种不匹配的槽位保持 INF 成本，永远不会被指派
			if !jobTypeAllowed(slots[c].worker.JobType, po.requiredJobType) {
				continue
			}
			score := scoreWorkerForOrder(slots[c].worker, scoreInput{
				buildings:   buildingResp.Buildings,
				current:     po.building,
				faultName:   faultNames[po.order.FaultType],
				minuteLoads: minuteLoads,
				roadNet:     roadNet,
				avgLoad:     avgLoad,
				wSkill:      wSkill,
				wDistance:   wDistance,
				wLoad:       wLoad,
			})
			cost[i][c] = int64(math.Round((1 - score.totalScore) * 10000))
		}
	}
	for i := nOrders; i < n; i++ {
		for j := 0; j < n; j++ {
			cost[i][j] = 0
		}
	}

	assignments := hungarian(cost)
	for i, po := range plan {
		col := assignments[i]
		if col < 0 || col >= nSlots {
			remained++
			continue
		}
		if cost[i][col] >= batchDummyCost {
			remained++
			continue
		}
		worker := slots[col].worker
		if !jobTypeAllowed(worker.JobType, po.requiredJobType) {
			remained++
			continue
		}
		score := scoreWorkerForOrder(worker, scoreInput{
			buildings:   buildingResp.Buildings,
			current:     po.building,
			faultName:   faultNames[po.order.FaultType],
			minuteLoads: minuteLoads,
			roadNet:     roadNet,
			avgLoad:     avgLoad,
			wSkill:      wSkill,
			wDistance:   wDistance,
			wLoad:       wLoad,
		})
		if err := l.commitAssignment(po.order, worker.Id, score); err != nil {
			if errors.Is(err, errBatchSkip) {
				remained++
				continue
			}
			return nil, errs.Internal(err)
		}
		resp.Dispatched = append(resp.Dispatched, types.AdminBatchDispatchItem{
			OrderId:  po.order.ID,
			WorkerId: worker.Id,
		})
		notifyUsers(l.ctx, l.svcCtx, []int64{worker.Id, po.order.ReporterID}, "dispatch",
			"工单已派单 "+po.order.OrderNo, po.order.Room+"室 已派单", po.order.ID)
		l.svcCtx.WS.PublishOrder(ws.OrderEvent{
			Type: "order_changed", OrderId: po.order.ID, OrderNo: po.order.OrderNo,
			BuildingId: po.order.BuildingID, WorkerId: worker.Id, Status: store.StatusDispatched,
		})
	}
	resp.Remained = remained
	return resp, nil
}

func (l *AdminBatchDispatchLogic) loadPendingOrders(req *types.AdminBatchDispatchRequest) ([]store.Order, error) {
	if len(req.OrderIds) > 0 {
		return store.ListPendingOrdersByIDs(l.ctx, l.svcCtx.DB, req.OrderIds, store.StatusPending)
	}
	if req.BuildingId > 0 {
		return store.ListPendingOrdersByBuilding(l.ctx, l.svcCtx.DB, req.BuildingId, batchLimit)
	}
	return store.ListPendingOrders(l.ctx, l.svcCtx.DB, batchLimit)
}

func (l *AdminBatchDispatchLogic) commitAssignment(o store.Order, workerID int64, score candidateScore) error {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		ok, err := store.TryAssignOrder(txCtx, session, o.ID, workerID)
		if err != nil {
			return err
		}
		if !ok {
			return errBatchSkip
		}
		return store.InsertDispatchRecord(txCtx, session, o.ID, workerID,
			score.totalScore, score.skillScore, score.distanceScore, score.loadScore)
	})
	return err
}

func uniqueOrderBuildingIDs(orders []store.Order) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0)
	for _, o := range orders {
		if _, ok := seen[o.BuildingID]; ok {
			continue
		}
		seen[o.BuildingID] = struct{}{}
		result = append(result, o.BuildingID)
	}
	return result
}

func containsWorker(workers []*workerclient.WorkerInfo, id int64) bool {
	for _, w := range workers {
		if w.Id == id {
			return true
		}
	}
	return false
}
