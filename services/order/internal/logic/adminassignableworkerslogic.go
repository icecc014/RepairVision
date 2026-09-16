package logic

import (
	"context"
	"errors"
	"sort"
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

// V6.5 手动派单 / 改派的可选工人：只列出"今天在岗 + 工种匹配"的工人，
// 并区分"本楼栋管辖"与"全校跨区支援"，前端据此禁用休息/工种不符的选项。

type AdminAssignableWorkersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminAssignableWorkersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAssignableWorkersLogic {
	return &AdminAssignableWorkersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminAssignableWorkersLogic) AdminAssignableWorkers(req *types.OrderIdRequest) (*types.AssignableWorkerListResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	order, err := store.FindOrder(l.ctx, l.svcCtx.DB, req.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("工单不存在")
		}
		return nil, errs.Internal(err)
	}

	faultName, category := "", ""
	if faultTypes, err := store.ListFaultTypes(l.ctx, l.svcCtx.DB); err == nil {
		for _, ft := range faultTypes {
			if ft.Code == order.FaultType {
				faultName = ft.Name
				category = normalizeFaultCategory(ft.Category)
				break
			}
		}
	}
	requiredJobType := requiredJobTypeOf(category)

	buildings, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	buildingName := make(map[int64]string, len(buildings.Buildings))
	for _, b := range buildings.Buildings {
		buildingName[b.Id] = b.Name
	}

	// 本楼栋管辖工人（含今日班次）
	inBuilding := map[int64]*workerclient.WorkerInfo{}
	today := time.Now().Format("2006-01-02")
	if wr, err := l.svcCtx.WorkerRpc.ListWorkersByBuilding(l.ctx,
		&workerclient.BuildingWorkersRequest{BuildingId: order.BuildingID, WorkDate: today}); err == nil {
		for _, w := range wr.Workers {
			inBuilding[w.Id] = w
		}
	}

	// 全校启用工人（含跨楼栋支援候选）
	users, err := l.svcCtx.WorkerRpc.ListUsers(l.ctx, &workerclient.ListUsersRequest{Role: 2, Status: 1})
	if err != nil {
		return nil, errs.Upstream()
	}
	workerIDs := make([]int64, 0, len(users.Users))
	for _, u := range users.Users {
		workerIDs = append(workerIDs, u.Id)
	}
	countLoads, err := store.CountInProgressByWorkers(l.ctx, l.svcCtx.DB, workerIDs)
	if err != nil {
		return nil, errs.Internal(err)
	}

	resp := &types.AssignableWorkerListResponse{
		List:            []types.AssignableWorkerItem{},
		RequiredJobType: requiredJobType,
		FaultTypeName:   faultName,
	}
	if requiredJobType > 0 {
		resp.RequiredJobText = jobTypeText(requiredJobType)
	}
	for _, u := range users.Users {
		info := inBuilding[u.Id]
		onDuty, working, enabled, onLeave := false, false, true, false
		if status, err := l.svcCtx.WorkerRpc.GetDutyStatus(l.ctx, &workerclient.DutyStatusRequest{WorkerId: u.Id}); err == nil {
			onDuty = status.GetOnDuty()
			enabled = status.GetEnabled()
			onLeave = status.GetOnLeave()
			// 今天排了白班且未请假 / 未停用，就可以承接工单（非工作时段仍可指派，工人上班后开工）
			working = enabled && !onLeave && status.GetShiftType() != shiftOff
		}
		todayShift := ""
		if info != nil {
			todayShift = info.TodayShift
		}
		maxConcurrent := u.MaxConcurrent
		if info != nil && info.MaxConcurrent > 0 {
			maxConcurrent = info.MaxConcurrent
		}
		if maxConcurrent <= 0 {
			maxConcurrent = defaultMaxConcurrent
		}
		jobMatch := jobTypeAllowed(u.JobType, requiredJobType)
		selectable := info != nil && working && jobMatch
		reason := ""
		switch {
		case !enabled:
			reason = "账号已停用"
		case onLeave:
			reason = "今天已批准请假"
		case !working:
			reason = "今天轮休（OFF）"
		case !jobMatch:
			reason = "工种不匹配（该单需要" + resp.RequiredJobText + "）"
		case info == nil:
			reason = "不负责该楼栋，可在下方跨区支援里选择"
		case !onDuty:
			reason = "当前非工作时段（今日白班），仍可指派"
		}
		item := types.AssignableWorkerItem{
			Id: u.Id, Name: u.Name, Username: u.Username,
			JobType: u.JobType, JobTypeText: jobTypeText(u.JobType),
			OnDuty: onDuty, Working: working, Enabled: enabled, OnLeave: onLeave, TodayShift: todayShift,
			InProgress: countLoads[u.Id], MaxConcurrent: maxConcurrent,
			InBuilding: info != nil, Selectable: selectable, Reason: reason,
		}
		for _, bid := range u.BuildingIds {
			if name, ok := buildingName[bid]; ok {
				item.BuildingNames = append(item.BuildingNames, name)
			}
		}
		// 跨区支援：本楼栋没人可派时，今天排班且工种匹配的其他工人也可以指派
		if selectable {
			item.Selectable = true
		} else if info == nil && working && jobMatch {
			item.Selectable = true
			item.Reason = "跨区支援（不负责该楼栋）"
		}
		resp.List = append(resp.List, item)
	}
	// 本楼栋可派 → 跨区可派 → 本楼栋其他 → 跨区其他；同级按在途数升序
	sort.SliceStable(resp.List, func(i, j int) bool {
		a, b := resp.List[i], resp.List[j]
		rank := func(x types.AssignableWorkerItem) int {
			switch {
			case x.Selectable && x.InBuilding:
				return 0
			case x.Selectable:
				return 1
			case x.InBuilding:
				return 2
			default:
				return 3
			}
		}
		if rank(a) != rank(b) {
			return rank(a) < rank(b)
		}
		return strings.Compare(a.Username, b.Username) < 0
	})
	return resp, nil
}
