package logic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminWorkSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminWorkSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminWorkSettingsLogic {
	return &AdminWorkSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminWorkSettings 查询排班与工作时段配置（双休模式、固定休息日、上午/下午时段）。
func (l *AdminWorkSettingsLogic) AdminWorkSettings() (resp *types.WorkSettingsItem, err error) {
	r, err := l.svcCtx.WorkerRpc.GetWorkSettings(l.ctx, &workerclient.GetWorkSettingsRequest{})
	if err != nil {
		return nil, rpcBizError(err)
	}
	return workSettingsToItem(r.GetSettings()), nil
}

// workSettingsToItem 把 RPC 结果转成 REST 结构；缺省时给出与后端一致的默认值。
func workSettingsToItem(s *workerclient.WorkSettings) *types.WorkSettingsItem {
	item := &types.WorkSettingsItem{
		RestDaysPerWeek:   2,
		RestMode:          "staggered",
		FixedRestWeekdays: "6,7",
		MorningStart:      "08:00",
		MorningEnd:        "12:00",
		AfternoonStart:    "14:00",
		AfternoonEnd:      "18:00",
		AllowForceStart:   1,
	}
	if s == nil {
		return item
	}
	item.RestDaysPerWeek = s.RestDaysPerWeek
	item.RestMode = s.RestMode
	item.FixedRestWeekdays = s.FixedRestWeekdays
	item.MorningStart = s.MorningStart
	item.MorningEnd = s.MorningEnd
	item.AfternoonStart = s.AfternoonStart
	item.AfternoonEnd = s.AfternoonEnd
	item.AllowForceStart = s.AllowForceStart
	return item
}

// pick 返回第一个非空字符串，用于"未传字段沿用旧值"的部分更新。
func pick(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}