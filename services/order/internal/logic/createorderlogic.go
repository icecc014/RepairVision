package logic

import (
	"context"
	"errors"
	"fmt"
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
		&workerclient.BuildingWorkersRequest{BuildingId: buildingID, WorkDate: time.Now().Format("2006-01-02")})
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
		countLoads, err := store.CountInProgressByWorkers(l.ctx, l.svcCtx.DB, workerIDs(workerResp.Workers))
		if err != nil {
			return nil, errs.Internal(err)
		}
		minuteLoads, err := store.CountWorkloadByWorkers(l.ctx, l.svcCtx.DB, workerIDs(workerResp.Workers))
		if err != nil {
			return nil, errs.Internal(err)
		}
		best = pickBestOrder(workerResp.Workers, buildingResp.Buildings, currentBuilding, faultType.Name,
			countLoads, minuteLoads, skillW, distW, loadW)
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
