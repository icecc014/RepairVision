package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminDutyOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDutyOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDutyOverviewLogic {
	return &AdminDutyOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminDutyOverview 汇总全部启用工人当前的在岗情况，供排班页展示"当前在岗 N 人"。
func (l *AdminDutyOverviewLogic) AdminDutyOverview() (resp *types.DutyOverviewResponse, err error) {
	out := &types.DutyOverviewResponse{List: []types.DutyStatusItem{}}
	users, err := l.svcCtx.WorkerRpc.ListUsers(l.ctx, &workerclient.ListUsersRequest{Role: 2, Status: 1})
	if err != nil {
		return nil, rpcBizError(err)
	}
	for _, u := range users.GetUsers() {
		status, err := l.svcCtx.WorkerRpc.GetDutyStatus(l.ctx, &workerclient.DutyStatusRequest{WorkerId: u.Id})
		if err != nil {
			continue // 单个工人查询失败不影响整体看板
		}
		item := dutyStatusToItem(u.Id, u.Name, status)
		if out.Morning == "" {
			out.Morning = item.Morning
			out.Afternoon = item.Afternoon
		}
		if item.OnDuty {
			out.OnDutyCount++
		}
		out.Total++
		out.List = append(out.List, *item)
	}
	return out, nil
}