package logic

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminDispatchRuleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchRuleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchRuleCreateLogic {
	return &AdminDispatchRuleCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDispatchRuleCreateLogic) AdminDispatchRuleCreate(req *types.DispatchRuleCreateRequest) (resp *types.EmptyResponse, err error) {
	key := strings.TrimSpace(req.RuleKey)
	valueText := strings.TrimSpace(req.RuleValue)
	if key == "" || valueText == "" {
		return nil, errs.BadRequest("规则标识和值不能为空")
	}
	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil || value < 0 {
		return nil, errs.BadRequest("规则值必须是不小于0的数字")
	}
	enabled := req.Enabled
	if enabled == 0 {
		enabled = 1
	}
	if _, err := store.CreateDispatchRule(l.ctx, l.svcCtx.DB, key, value, enabled, strings.TrimSpace(req.Remark)); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errs.Conflict("该规则标识已存在")
		}
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
