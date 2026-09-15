package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminWorkSettingsUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminWorkSettingsUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminWorkSettingsUpdateLogic {
	return &AdminWorkSettingsUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminWorkSettingsUpdate 更新排班与工作时段配置；未传字段沿用当前值。
func (l *AdminWorkSettingsUpdateLogic) AdminWorkSettingsUpdate(req *types.WorkSettingsUpdateRequest) (resp *types.WorkSettingsItem, err error) {
	cur, err := l.svcCtx.WorkerRpc.GetWorkSettings(l.ctx, &workerclient.GetWorkSettingsRequest{})
	if err != nil {
		return nil, rpcBizError(err)
	}
	base := workSettingsToItem(cur.GetSettings())

	restDays := base.RestDaysPerWeek
	if req.RestDaysPerWeek != 0 || base.RestDaysPerWeek == 0 {
		restDays = req.RestDaysPerWeek
	}
	allowForce := base.AllowForceStart
	if req.AllowForceStart != 0 {
		allowForce = 1
	} else if req.RestMode != "" || req.MorningStart != "" {
		allowForce = 0
	}

	r, err := l.svcCtx.WorkerRpc.UpdateWorkSettings(l.ctx, &workerclient.UpdateWorkSettingsRequest{
		RestDaysPerWeek:   restDays,
		RestMode:          pick(req.RestMode, base.RestMode),
		FixedRestWeekdays: pick(req.FixedRestWeekdays, base.FixedRestWeekdays),
		MorningStart:      pick(req.MorningStart, base.MorningStart),
		MorningEnd:        pick(req.MorningEnd, base.MorningEnd),
		AfternoonStart:    pick(req.AfternoonStart, base.AfternoonStart),
		AfternoonEnd:      pick(req.AfternoonEnd, base.AfternoonEnd),
		AllowForceStart:   allowForce,
	})
	if err != nil {
		return nil, rpcBizError(err)
	}
	return workSettingsToItem(r.GetSettings()), nil
}