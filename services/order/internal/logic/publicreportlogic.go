package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"order/internal/ws"
	"worker/workerclient"
)

type PublicReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicReportLogic {
	return &PublicReportLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// publicFaultTypeOf 取故障类型字典项（公开三类之一）。
func publicFaultTypeOf(ctx context.Context, svcCtx *svc.ServiceContext, code string) (*store.FaultType, error) {
	list, err := store.ListFaultTypes(ctx, svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	for i := range list {
		if list[i].Code == code {
			return &list[i], nil
		}
	}
	return nil, errs.BadRequest("该故障类型暂未配置，请联系管理员")
}

// workerNameOf 取工人姓名（用于提交后的即时反馈）。
func workerNameOf(ctx context.Context, svcCtx *svc.ServiceContext, workerID int64) string {
	users, err := svcCtx.WorkerRpc.ListUsers(ctx, &workerclient.ListUsersRequest{Role: 2})
	if err != nil {
		return ""
	}
	for _, u := range users.Users {
		if u.Id == workerID {
			return u.Name
		}
	}
	return ""
}

// PublicReport 公共报修提交（免登录）：
// 验证码 → IP 限流 → 敏感词 → 楼栋/房间校验 → 3 天内同类合并 → 建单 → 自动派单。
func (l *PublicReportLogic) PublicReport(req *types.PublicReportRequest) (*types.PublicReportResponse, error) {
	description := strings.TrimSpace(req.Description)
	contact := strings.TrimSpace(req.Contact)
	faultType := strings.TrimSpace(req.FaultType)
	room := strings.TrimSpace(req.Room)

	if req.BuildingId <= 0 {
		return nil, errs.BadRequest("请选择报修楼栋")
	}
	if room == "" {
		return nil, errs.BadRequest("请填写房间号")
	}
	if !publicFaultAllowed(faultType) {
		return nil, errs.BadRequest("请选择故障类型")
	}
	if n := len([]rune(description)); n < publicDescMin || n > publicDescMax {
		return nil, errs.BadRequest(fmt.Sprintf("故障描述需 %d~%d 字", publicDescMin, publicDescMax))
	}
	if req.ReporterType < 1 || req.ReporterType > 3 {
		return nil, errs.BadRequest("请选择报修人身份")
	}
	if len([]rune(contact)) > publicContactMax {
		return nil, errs.BadRequest("联系方式过长")
	}
	if !verifyPublicCaptcha(l.ctx, l.svcCtx, req.CaptchaId, req.CaptchaCode) {
		return nil, errs.BadRequest("验证码错误或已过期，请重新获取")
	}

	if ip := strings.TrimSpace(req.ClientIp); ip != "" {
		if count := l.svcCtx.PubCache.Incr(l.ctx, publicIPRatePrefix+ip, time.Minute); count > publicIPPerMinute {
			return nil, errs.New(429, "提交过于频繁，请稍后再试")
		}
	}
	if word := hitPublicSensitiveWord(l.ctx, l.svcCtx, description); word != "" {
		return nil, errs.BadRequest("描述包含不允许的词汇（" + word + "），请修改后重试")
	}

	building, floor, err := publicBuildingOf(l.ctx, l.svcCtx, req.BuildingId, room)
	if err != nil {
		return nil, err
	}
	fault, err := publicFaultTypeOf(l.ctx, l.svcCtx, faultType)
	if err != nil {
		return nil, err
	}

	// 3 天内同楼栋 + 同房间 + 同类型的未完成工单 → 合并（不新建工单）
	if main, mErr := store.FindRecentRoomOrder(l.ctx, l.svcCtx.DB, building.Id, floor, room, faultType, time.Now().Add(-publicMergeWindow)); mErr == nil {
		if main.Status != store.StatusCompleted && main.Status != store.StatusCanceled {
			extra := fmt.Sprintf("；[%s重复报修 %s] %s", publicReporterText(req.ReporterType), time.Now().Format("01-02 15:04"), description)
			if aErr := store.AppendOrderDescription(l.ctx, l.svcCtx.DB, main.ID, extra); aErr != nil {
				l.Logger.Errorf("append merged description failed: %v", aErr)
			}
			_ = store.InsertDuplicateRecordIfAbsent(l.ctx, l.svcCtx.DB, main.ID, 0, 1)
			l.svcCtx.WS.PublishOrder(ws.OrderEvent{
				Type: "order_changed", OrderId: main.ID, OrderNo: main.OrderNo,
				BuildingId: main.BuildingID, Status: main.Status,
			})
			return &types.PublicReportResponse{
				OrderNo:     main.OrderNo,
				Merged:      true,
				MainOrderNo: main.OrderNo,
				Status:      main.Status,
				StatusText:  statusText(main.Status),
				Message:     "该房间相同类型报修已在处理中，已并入工单 " + main.OrderNo,
			}, nil
		}
	} else if !errors.Is(mErr, sqlx.ErrNotFound) {
		l.Logger.Errorf("public merge lookup failed: %v", mErr)
	}

	// 同房间 30 分钟限流（仅针对会新建工单的提交；同类重复已在上一步合并）
	roomKey := fmt.Sprintf("%s%d:%s", publicRoomRatePrefix, building.Id, room)
	if count := l.svcCtx.PubCache.Incr(l.ctx, roomKey, publicRoomCooldown); count > 1 {
		return nil, errs.New(429, "该房间 30 分钟内已提交过报修，请稍后再试")
	}

	manualReview := int64(0)
	if normalizeFaultCategory(fault.Category) == "other" || fault.AutoDispatch == 0 {
		manualReview = 1
	}
	var contactValue sql.NullString
	if contact != "" {
		contactValue = sql.NullString{String: contact, Valid: true}
	}
	now := time.Now()
	orderNo := fmt.Sprintf("REP%s%09d", now.Format("20060102"), now.UnixNano()%1000000000)
	order := &store.Order{
		OrderNo:         orderNo,
		Title:           fault.Name + "（" + room + "室）",
		Description:     description,
		BuildingID:      building.Id,
		Room:            room,
		Floor:           floor,
		FaultType:       faultType,
		Priority:        1,
		ExpectMinutes:   30,
		Status:          store.StatusPending,
		IsMerged:        0,
		ManualReview:    manualReview,
		ReporterID:      0,
		Source:          "public",
		ReporterType:    req.ReporterType,
		ReporterContact: contactValue,
	}
	orderID, err := store.InsertOrder(l.ctx, l.svcCtx.DB, order)
	if err != nil {
		return nil, errs.Internal(err)
	}

	// 自动派单：复用批量派单引擎（含工种过滤、路网距离、负载与保护线）
	if manualReview == 0 && !autoDispatchPaused(l.ctx, l.svcCtx) {
		systemCtx := auth.WithSystemIdentity(l.ctx)
		if _, dErr := NewAdminBatchDispatchLogic(systemCtx, l.svcCtx).AdminBatchDispatch(
			&types.AdminBatchDispatchRequest{OrderIds: []int64{orderID}}); dErr != nil {
			l.Logger.Errorf("public report auto dispatch failed: %v", dErr)
		}
	}

	resp := &types.PublicReportResponse{
		OrderNo:    orderNo,
		Status:     store.StatusPending,
		StatusText: statusText(store.StatusPending),
		Message:    "报修已提交，等待处理",
	}
	if manualReview == 1 {
		resp.Message = "报修已提交，将由管理员协调处理"
	}
	if fresh, fErr := store.FindOrder(l.ctx, l.svcCtx.DB, orderID); fErr == nil {
		resp.Status = fresh.Status
		resp.StatusText = statusText(fresh.Status)
		if fresh.WorkerID.Valid && fresh.WorkerID.Int64 > 0 {
			resp.Dispatched = true
			resp.WorkerName = workerNameOf(l.ctx, l.svcCtx, fresh.WorkerID.Int64)
		}
	}
	l.svcCtx.WS.PublishOrder(ws.OrderEvent{
		Type: "order_changed", OrderId: orderID, OrderNo: orderNo,
		BuildingId: building.Id, Status: resp.Status,
	})
	return resp, nil
}
