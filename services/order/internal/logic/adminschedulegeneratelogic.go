package logic

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminScheduleGenerateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminScheduleGenerateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminScheduleGenerateLogic {
	return &AdminScheduleGenerateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminScheduleGenerate 一键生成周排班：把「每周休息天数」「每栋楼最少在岗人数」透传给 worker 服务，
// 并把告警里的楼栋编号替换成楼栋名称，方便管理员直接阅读。
func (l *AdminScheduleGenerateLogic) AdminScheduleGenerate(req *types.ScheduleGenerateRequest) (resp *types.ScheduleListResponse, err error) {
	out, err := l.svcCtx.WorkerRpc.GenerateWeekly(l.ctx, &workerclient.GenerateScheduleRequest{
		WeekStart:       req.WeekStart,
		RestDaysPerWeek: req.RestDaysPerWeek,
		MinPerBuilding:  req.MinPerBuilding,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	warnings := out.Warnings
	if len(warnings) > 0 {
		warnings = l.replaceBuildingNames(warnings)
	}
	return &types.ScheduleListResponse{
		List:     schedulePbListToTypes(out.Items),
		Warnings: warnings,
	}, nil
}

func (l *AdminScheduleGenerateLogic) replaceBuildingNames(warnings []string) []string {
	resp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return warnings
	}
	names := make(map[int64]string, len(resp.Buildings))
	for _, b := range resp.Buildings {
		names[b.Id] = b.Name
	}
	result := make([]string, 0, len(warnings))
	for _, w := range warnings {
		line := w
		for id, name := range names {
			line = strings.ReplaceAll(line, fmt.Sprintf("楼栋#%d", id), name)
		}
		result = append(result, line)
	}
	return result
}
