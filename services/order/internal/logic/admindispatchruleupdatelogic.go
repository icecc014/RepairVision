package logic

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type AdminDispatchRuleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchRuleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchRuleUpdateLogic {
	return &AdminDispatchRuleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDispatchRuleUpdateLogic) AdminDispatchRuleUpdate(req *types.DispatchRuleUpdateRequest) (resp *types.EmptyResponse, err error) {
	if _, err := store.FindDispatchRuleByID(l.ctx, l.svcCtx.DB, req.Id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errs.NotFound("规则不存在")
		}
		return nil, errs.Internal(err)
	}
	valueText := strings.TrimSpace(req.RuleValue)
	if valueText == "" {
		return nil, errs.BadRequest("规则值不能为空")
	}
	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil || value < 0 {
		return nil, errs.BadRequest("规则值必须是不小于0的数字")
	}
	if err := store.UpdateDispatchRule(l.ctx, l.svcCtx.DB, req.Id, value, req.Enabled, strings.TrimSpace(req.Remark)); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
