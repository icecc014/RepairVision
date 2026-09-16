package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/svc"
	"order/internal/types"
)

type AdminDispatchRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchRulesLogic {
	return &AdminDispatchRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminDispatchRules 返回分组后的派单规则（含元数据、当前值、运行态），供可视化编辑。
func (l *AdminDispatchRulesLogic) AdminDispatchRules() (resp *types.DispatchRulesResponse, err error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	return buildDispatchRulesResponse(l.ctx, l.svcCtx)
}
