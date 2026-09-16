package logic

import (
	"context"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

// AutoAssignBuildingItem 单个工人的分配结果。
type autoAssignItem struct {
	worker *workerclient.User
	ids    []int64
}

// autoAssignWorkerBuildings 按工种把 16 栋楼均匀轮转分配给在岗工人：
// 每种工种（电/水/泥瓦/木）各自把全部楼栋轮转分给该工种的工人，
// 从而保证"每栋楼每类工种都至少有 1 名候选人"，且各人楼栋数尽量均衡。
func autoAssignWorkerBuildings(ctx context.Context, svcCtx *svc.ServiceContext) (*types.AutoAssignResponse, error) {
	usersResp, err := svcCtx.WorkerRpc.ListUsers(ctx, &workerclient.ListUsersRequest{Role: 2, Status: 1})
	if err != nil {
		return nil, errs.Upstream()
	}
	buildingResp, err := svcCtx.MapRpc.ListBuildings(ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	buildingIDs := make([]int64, 0, len(buildingResp.Buildings))
	for _, b := range buildingResp.Buildings {
		buildingIDs = append(buildingIDs, b.Id)
	}
	sort.Slice(buildingIDs, func(i, j int) bool { return buildingIDs[i] < buildingIDs[j] })

	byType := map[int64][]*workerclient.User{}
	for _, u := range usersResp.Users {
		byType[u.JobType] = append(byType[u.JobType], u)
	}
	assigned := map[int64][]int64{}
	coverage := map[int64]int64{}
	for _, jobType := range []int64{1, 2, 3, 4} {
		group := byType[jobType]
		if len(group) == 0 {
			continue
		}
		sort.Slice(group, func(i, j int) bool { return group[i].Id < group[j].Id })
		for i, bid := range buildingIDs {
			w := group[i%len(group)]
			assigned[w.Id] = append(assigned[w.Id], bid)
			coverage[jobType]++
		}
	}

	out := &types.AutoAssignResponse{
		Buildings: int64(len(buildingIDs)),
		Workers:   int64(len(usersResp.Users)),
		List:      []types.AutoAssignWorkerItem{},
	}
	for _, u := range usersResp.Users {
		ids := assigned[u.Id]
		out.List = append(out.List, types.AutoAssignWorkerItem{
			WorkerId:      u.Id,
			Name:          u.Name,
			JobTypeText:   jobTypeText(u.JobType),
			BuildingCount: int64(len(ids)),
			BuildingIds:   ids,
		})
	}
	return out, nil
}

// persistAutoAssign 计算并写回每个工人的管辖楼栋（增量：只更新有变化的工人）。
func persistAutoAssign(ctx context.Context, svcCtx *svc.ServiceContext) (*types.AutoAssignResponse, error) {
	out, err := autoAssignWorkerBuildings(ctx, svcCtx)
	if err != nil {
		return nil, err
	}
	usersResp, err := svcCtx.WorkerRpc.ListUsers(ctx, &workerclient.ListUsersRequest{Role: 2, Status: 1})
	if err != nil {
		return nil, errs.Upstream()
	}
	byID := map[int64]*workerclient.User{}
	for _, u := range usersResp.Users {
		byID[u.Id] = u
	}
	for _, item := range out.List {
		u := byID[item.WorkerId]
		if u == nil || len(item.BuildingIds) == 0 {
			continue
		}
		if sameIDs(u.BuildingIds, item.BuildingIds) {
			continue
		}
		if _, err := svcCtx.WorkerRpc.UpdateUser(ctx, &workerclient.UpdateUserRequest{
			Id:            u.Id,
			Name:          u.Name,
			Phone:         u.Phone,
			Role:          2,
			Status:        1,
			BuildingIds:   item.BuildingIds,
			MaxConcurrent: u.MaxConcurrent,
			JobType:       u.JobType,
		}); err != nil {
			logx.WithContext(ctx).Errorf("auto assign buildings failed worker=%d: %v", u.Id, err)
			return nil, errs.Upstream()
		}
	}
	return out, nil
}

func sameIDs(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]int64(nil), a...)
	y := append([]int64(nil), b...)
	sort.Slice(x, func(i, j int) bool { return x[i] < x[j] })
	sort.Slice(y, func(i, j int) bool { return y[i] < y[j] })
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

type AdminAutoAssignBuildingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminAutoAssignBuildingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAutoAssignBuildingsLogic {
	return &AdminAutoAssignBuildingsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AdminAutoAssignBuildings 管理员一键按工种自动分配全部工人的管辖楼栋。
func (l *AdminAutoAssignBuildingsLogic) AdminAutoAssignBuildings() (*types.AutoAssignResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	return persistAutoAssign(l.ctx, l.svcCtx)
}

// triggerAutoAssign 人员变动后自动重算管辖楼栋（失败只记日志，不影响主流程）。
func triggerAutoAssign(ctx context.Context, svcCtx *svc.ServiceContext) {
	if _, err := persistAutoAssign(ctx, svcCtx); err != nil {
		logx.WithContext(ctx).Errorf("auto assign buildings after user change failed: %v", err)
	}
}
