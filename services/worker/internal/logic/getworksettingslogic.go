package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type GetWorkSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWorkSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWorkSettingsLogic {
	return &GetWorkSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetWorkSettings 返回排班与工作时段配置（不存在时回落到内置默认值）。
func (l *GetWorkSettingsLogic) GetWorkSettings(in *worker.GetWorkSettingsRequest) (*worker.WorkSettingsResponse, error) {
	s, err := store.GetWorkSettings(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	return &worker.WorkSettingsResponse{Settings: settingsToPb(s)}, nil
}