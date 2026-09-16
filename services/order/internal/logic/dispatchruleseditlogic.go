package logic

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/auth"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

// V6.3 派单规则的可视化编辑：批量保存 / 恢复默认 / 删除自定义规则。

type AdminDispatchRulesSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchRulesSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchRulesSaveLogic {
	return &AdminDispatchRulesSaveLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AdminDispatchRulesSave 批量保存规则（值 + 启用状态），保存后立即生效：
// 引擎每次派单都直接读 dispatch_rule_config，无需重启。
func (l *AdminDispatchRulesSaveLogic) AdminDispatchRulesSave(req *types.DispatchRulesSaveRequest) (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	if len(req.Items) == 0 {
		return nil, errs.BadRequest("没有需要保存的规则")
	}
	for _, item := range req.Items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return nil, errs.BadRequest("规则标识不能为空")
		}
		meta := findDispatchRuleMeta(key)
		if meta != nil && meta.RuntimeOnly {
			return nil, errs.BadRequest("「" + meta.Name + "」是运行态规则，由系统自动维护，不能手工修改")
		}
		if meta == nil && !customRuleKeyPattern.MatchString(key) {
			return nil, errs.BadRequest("规则标识格式不合法：" + key + "（应为小写字母开头、可含数字与下划线，如 load_weight）")
		}
		if err := validateRuleValue(meta, item.Value); err != nil {
			name := key
			if meta != nil {
				name = meta.Name
			}
			return nil, errs.BadRequest("「" + name + "」" + err.Error())
		}
		enabled := item.Enabled
		if enabled != 0 {
			enabled = 1
		}
		remark := strings.TrimSpace(item.Remark)
		if remark == "" && meta != nil {
			remark = meta.Description
		}
		if remark == "" {
			if existing, err := store.FindDispatchRuleByKey(l.ctx, l.svcCtx.DB, key); err == nil && existing.Remark.Valid {
				remark = existing.Remark.String
			}
		}
		if err := store.UpsertDispatchRule(l.ctx, l.svcCtx.DB, key, item.Value, enabled, remark); err != nil {
			return nil, errs.Internal(err)
		}
	}
	// 阈值可能被改动：立刻重新评估一次保护线状态，保证看板/暂停态与配置同步
	if _, err := evaluateDispatchGuard(l.ctx, l.svcCtx); err != nil {
		logx.WithContext(l.ctx).Errorf("evaluate guard after rule save failed: %v", err)
	}
	return &types.EmptyResponse{}, nil
}

type AdminDispatchRuleResetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchRuleResetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchRuleResetLogic {
	return &AdminDispatchRuleResetLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AdminDispatchRuleReset 把某条注册规则恢复为默认值。
func (l *AdminDispatchRuleResetLogic) AdminDispatchRuleReset(req *types.DispatchRuleKeyRequest) (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	meta := findDispatchRuleMeta(strings.TrimSpace(req.Key))
	if meta == nil {
		return nil, errs.NotFound("该规则不是系统注册规则，无法恢复默认值")
	}
	if meta.RuntimeOnly {
		return nil, errs.BadRequest("运行态规则由系统维护，不能手工重置")
	}
	if err := store.UpsertDispatchRule(l.ctx, l.svcCtx.DB, meta.Key, meta.Default, 1, meta.Description); err != nil {
		return nil, errs.Internal(err)
	}
	if _, err := evaluateDispatchGuard(l.ctx, l.svcCtx); err != nil {
		logx.WithContext(l.ctx).Errorf("evaluate guard after reset failed: %v", err)
	}
	return &types.EmptyResponse{}, nil
}

type AdminDispatchRuleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDispatchRuleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDispatchRuleDeleteLogic {
	return &AdminDispatchRuleDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AdminDispatchRuleDelete 删除自定义规则（系统注册规则只能停用/恢复默认，不能删）。
func (l *AdminDispatchRuleDeleteLogic) AdminDispatchRuleDelete(req *types.DispatchRuleKeyRequest) (*types.EmptyResponse, error) {
	if _, ok := auth.IdentityFromContext(l.ctx); !ok {
		return nil, errs.Unauthorized("登录状态无效")
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return nil, errs.BadRequest("规则标识不能为空")
	}
	if meta := findDispatchRuleMeta(key); meta != nil {
		return nil, errs.BadRequest("「" + meta.Name + "」是系统注册规则，请改为停用或恢复默认值")
	}
	if err := store.DeleteDispatchRuleByKey(l.ctx, l.svcCtx.DB, key); err != nil {
		return nil, errs.Internal(err)
	}
	return &types.EmptyResponse{}, nil
}

// buildDispatchRulesResponse 组装"分组 + 元数据 + 当前值"的规则清单。
func buildDispatchRulesResponse(ctx context.Context, svcCtx *svc.ServiceContext) (*types.DispatchRulesResponse, error) {
	rows, err := store.ListDispatchRules(ctx, svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	byKey := make(map[string]store.DispatchRule, len(rows))
	for _, r := range rows {
		byKey[r.RuleKey] = r
	}
	resp := &types.DispatchRulesResponse{Groups: []types.DispatchRuleGroupItem{}}
	for _, g := range dispatchRuleGroupMeta {
		group := types.DispatchRuleGroupItem{Key: g.Key, Label: g.Label, Items: []types.DispatchRuleEntry{}}
		for _, meta := range dispatchRuleRegistry {
			if meta.Group != g.Key {
				continue
			}
			entry := types.DispatchRuleEntry{
				Key: meta.Key, Name: meta.Name, Group: meta.Group, Type: meta.Type, Unit: meta.Unit,
				Min: meta.Min, Max: meta.Max, Step: meta.Step, Precision: meta.Precision,
				Value: meta.Default, DefaultValue: meta.Default, Enabled: 1,
				Description: meta.Description, RuntimeOnly: meta.RuntimeOnly, Registered: true,
			}
			if row, ok := byKey[meta.Key]; ok {
				entry.Saved = true
				entry.Value = row.RuleValue
				entry.Enabled = row.Enabled
				entry.UpdatedAt = row.UpdatedAt.Format("2006-01-02 15:04:05")
				if row.Remark.Valid && strings.TrimSpace(row.Remark.String) != "" {
					entry.Remark = row.Remark.String
				}
			}
			group.Items = append(group.Items, entry)
		}
		if len(group.Items) > 0 {
			resp.Groups = append(resp.Groups, group)
		}
	}
	registered := make(map[string]bool, len(dispatchRuleRegistry))
	for _, m := range dispatchRuleRegistry {
		registered[m.Key] = true
	}
	custom := types.DispatchRuleGroupItem{Key: "custom", Label: dispatchRuleGroupLabel("custom"), Items: []types.DispatchRuleEntry{}}
	for _, row := range rows {
		if registered[row.RuleKey] {
			continue
		}
		custom.Items = append(custom.Items, ruleRowToEntry(row))
	}
	if len(custom.Items) > 0 {
		resp.Groups = append(resp.Groups, custom)
	}
	if pending, _, err := store.PendingBacklog(ctx, svcCtx.DB); err == nil {
		resp.PendingCount = pending
	}
	if onDuty, err := countOnDutyWorkers(ctx, svcCtx); err == nil {
		resp.OnDutyCount = onDuty
	}
	if pausedValue, err := store.GetRuleValue(ctx, svcCtx.DB, ruleKeyAutoPaused, 0); err == nil {
		resp.Paused = pausedValue >= 0.5
	}
	resp.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	return resp, nil
}
