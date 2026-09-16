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
	if err != nil {
		return nil, errs.BadRequest("规则值必须是数字")
	}
	// V6.3：注册规则按元数据校验范围；自定义规则需符合命名规范，并明确标注引擎暂不支持
	meta := findDispatchRuleMeta(key)
	if meta != nil && meta.RuntimeOnly {
		return nil, errs.BadRequest("「" + meta.Name + "」是运行态规则，由系统自动维护，不能手工新增")
	}
	if meta == nil && !customRuleKeyPattern.MatchString(key) {
		return nil, errs.BadRequest("规则标识格式不合法：应为小写字母开头、可含数字与下划线（如 load_weight）")
	}
	if err := validateRuleValue(meta, value); err != nil {
		name := key
		if meta != nil {
			name = meta.Name
		}
		return nil, errs.BadRequest("「" + name + "」" + err.Error())
	}
	enabled := req.Enabled
	if enabled == 0 {
		enabled = 1
	}
	remark := strings.TrimSpace(req.Remark)
	if remark == "" && meta != nil {
		remark = meta.Description
	}
	if _, err := store.FindDispatchRuleByKey(l.ctx, l.svcCtx.DB, key); err == nil {
		return nil, errs.Conflict("该规则已存在，请直接在列表中编辑")
	}
	if _, err := store.CreateDispatchRule(l.ctx, l.svcCtx.DB, key, value, enabled, remark); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errs.Conflict("该规则标识已存在")
		}
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}
