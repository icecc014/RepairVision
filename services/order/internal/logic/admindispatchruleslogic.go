package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/store"
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

func (l *AdminDispatchRulesLogic) AdminDispatchRules() (resp *types.DispatchRuleListResponse, err error) {
	list, err := store.ListDispatchRules(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, err
	}
	out := &types.DispatchRuleListResponse{List: []types.DispatchRuleItem{}}
	for _, r := range list {
		remark := ""
		if r.Remark.Valid {
			remark = r.Remark.String
		}
		out.List = append(out.List, types.DispatchRuleItem{
			Id:        r.ID,
			RuleKey:   r.RuleKey,
			RuleValue: fmt.Sprintf("%.4f", r.RuleValue),
			Enabled:   r.Enabled,
			Remark:    remark,
			UpdatedAt: r.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, nil
}
