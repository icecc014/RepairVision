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
	slots := make([]slot, 0)
	for _, ws := range workerByBuilding {
		for _, w := range ws {
			if !workerCanTake(w, countLoads) {
				continue
			}
			avail := maxConcurrentOf(w) - countLoads[w.Id]
			for i := int64(0); i < avail; i++ {
				slots = append(slots, slot{worker: w})
			}
		}
	}

	type planOrder struct {
		order    store.Order
		building *mapclient.Building
	}
	plan := make([]planOrder, 0, len(orders))
	remained := int64(0)
	for _, o := range orders {
		b, ok := buildingByID[o.BuildingID]
		if !ok || buildingByID[o.BuildingID] == nil {
			remained++
			continue
		}
		hasEligible := false
		for _, w := range workerByBuilding[o.BuildingID] {
			if workerCanTake(w, countLoads) {
				hasEligible = true
				break
			}
		}
		if !hasEligible {
			remained++
			continue
		}
		plan = append(plan, planOrder{order: o, building: b})
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
			score := scoreWorkerForOrder(slots[c].worker, buildingResp.Buildings, po.building,
				faultNames[po.order.FaultType], minuteLoads, wSkill, wDistance, wLoad)
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
		score := scoreWorkerForOrder(worker, buildingResp.Buildings, po.building,
			faultNames[po.order.FaultType], minuteLoads, wSkill, wDistance, wLoad)
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
