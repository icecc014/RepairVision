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
	if req.Room == "" || req.FaultType == "" || req.Floor <= 0 {
		return nil, errs.BadRequest("房间、楼层和维修类型不能为空")
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
	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	found := false
	for _, b := range buildingResp.Buildings {
		if b.Id == buildingID {
			found = true
			break
		}
	}
	if !found {
		return nil, errs.BadRequest("楼栋尚未初始化")
	}

	workerResp, err := l.svcCtx.WorkerRpc.ListWorkersByBuilding(l.ctx,
		&workerclient.BuildingWorkersRequest{BuildingId: buildingID})
	if err != nil {
		return nil, errs.Upstream()
	}
	if len(workerResp.Workers) == 0 {
		return nil, errs.BadRequest("该楼栋暂无可用维修工人")
	}
	workerIDs := make([]int64, 0, len(workerResp.Workers))
	for _, w := range workerResp.Workers {
		workerIDs = append(workerIDs, w.Id)
	}
	loads, err := store.CountInProgressByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}
	targetWorker := workerResp.Workers[0].Id
	minLoad := int64(-1)
	for _, w := range workerResp.Workers {
		cnt := loads[w.Id]
		if minLoad < 0 || cnt < minLoad {
			minLoad = cnt
			targetWorker = w.Id
		}
	}

	now := time.Now()
	orderNo := fmt.Sprintf("REP%s%09d", now.Format("20060102"), now.UnixNano()%1000000000)
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = "无补充说明"
	}

	var orderID int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		order := &store.Order{
			OrderNo:     orderNo,
			Title:       faultType.Name + "（" + req.Room + "室）",
			Description: description,
			BuildingID:  buildingID,
			Room:        req.Room,
			Floor:       req.Floor,
			FaultType:   req.FaultType,
			Status:      store.StatusPending,
			IsMerged:    0,
			ReporterID:  identity.UID,
			Source:      "dormitory",
		}
		id, err := store.InsertOrder(txCtx, session, order)
		if err != nil {
			return err
		}
		orderID = id
		if err := store.AssignOrder(txCtx, session, orderID, targetWorker); err != nil {
			return err
		}
		return store.InsertDispatchRecord(txCtx, session, orderID, targetWorker)
	})
	if err != nil {
		return nil, errs.Internal(err)
	}
	return &types.CreateOrderResponse{OrderId: orderID, OrderNo: orderNo}, nil
}
