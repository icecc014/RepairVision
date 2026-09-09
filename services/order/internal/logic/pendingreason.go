package logic

import (
	"context"
	"time"

	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

// enrichPendingReasons 为待派工单(状态1)推导滞留原因，仅管理员列表使用。
func enrichPendingReasons(ctx context.Context, svcCtx *svc.ServiceContext, orders []store.Order, items []types.OrderItem) error {
	pendingBuildingIDs := make(map[int64]struct{})
	for _, o := range orders {
		if o.Status == store.StatusPending {
			pendingBuildingIDs[o.BuildingID] = struct{}{}
		}
	}
	if len(pendingBuildingIDs) == 0 {
		return nil
	}

	rules, err := store.ListDispatchRules(ctx, svcCtx.DB)
	if err != nil {
		return err
	}
	autoDispatch := true
	for _, r := range rules {
		if r.RuleKey == "auto_dispatch_enabled" && (r.Enabled == 0 || r.RuleValue < 0.5) {
			autoDispatch = false
		}
	}

	today := time.Now().Format("2006-01-02")
	workersByBuilding := make(map[int64][]*workerclient.WorkerInfo)
	allWorkerIDs := make([]int64, 0)
	for bid := range pendingBuildingIDs {
		wr, err := svcCtx.WorkerRpc.ListWorkersByBuilding(ctx,
			&workerclient.BuildingWorkersRequest{BuildingId: bid, WorkDate: today})
		if err != nil {
			return err
		}
		workersByBuilding[bid] = wr.Workers
		for _, w := range wr.Workers {
			allWorkerIDs = append(allWorkerIDs, w.Id)
		}
	}
	countLoads, err := store.CountInProgressByWorkers(ctx, svcCtx.DB, allWorkerIDs)
	if err != nil {
		return err
	}

	reasonByOrder := make(map[int64]string)
	for bid := range pendingBuildingIDs {
		workers := workersByBuilding[bid]
		if len(workers) == 0 {
			for _, o := range orders {
				if o.BuildingID == bid && o.Status == store.StatusPending {
					reasonByOrder[o.ID] = "该楼栋暂无维修工人"
				}
			}
			continue
		}
		anyAvailable := false
		for _, w := range workers {
			if workerCanTake(w, countLoads) {
				anyAvailable = true
				break
			}
		}
		for _, o := range orders {
			if o.BuildingID != bid || o.Status != store.StatusPending {
				continue
			}
			if !autoDispatch {
				reasonByOrder[o.ID] = "自动派单已关闭，等待管理员手动派单"
			} else if !anyAvailable {
				reasonByOrder[o.ID] = "工人休息或已达最大并发"
			} else {
				reasonByOrder[o.ID] = "等待调度系统处理"
			}
		}
	}

	for i := range items {
		if reason := reasonByOrder[items[i].Id]; reason != "" {
			items[i].PendingReason = reason
		}
	}
	return nil
}
