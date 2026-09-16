package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"strings"

	"map/mapclient"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUsersLogic {
	return &AdminUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUsersLogic) AdminUsers(req *types.AdminUserListRequest) (resp *types.AdminUserListResponse, err error) {
	out, err := l.svcCtx.WorkerRpc.ListUsers(l.ctx, &workerclient.ListUsersRequest{
		Role:    req.Role,
		Status:  req.Status,
		Keyword: req.Keyword,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	buildingNames := make(map[int64]string)
	for _, b := range buildingResp.Buildings {
		// 楼栋名去重：编码已包含在名称前缀时只显示名称（如 code=1 / name=1号宿舍楼）
		if b.Code != "" && !strings.HasPrefix(b.Name, b.Code) {
			buildingNames[b.Id] = b.Code + " " + b.Name
		} else {
			buildingNames[b.Id] = b.Name
		}
	}
	items := make([]types.AdminUserItem, 0, len(out.Users))
	for _, u := range out.Users {
		item := types.AdminUserItem{
			Id:            u.Id,
			Username:      u.Username,
			Name:          u.Name,
			Phone:         u.Phone,
			Role:          u.Role,
			RoleText:      roleText(u.Role),
			JobType:       u.JobType,
			BuildingId:    u.BuildingId,
			Status:        u.Status,
			StatusText:    userStatusText(u.Status),
			BuildingIds:   u.BuildingIds,
			MaxConcurrent: u.MaxConcurrent,
		}
		for _, id := range u.BuildingIds {
			if name := buildingNames[id]; name != "" {
				item.Buildings = append(item.Buildings, name)
			}
		}
		items = append(items, item)
	}
	return &types.AdminUserListResponse{List: items}, nil
}
