package logic

import (
	"context"
	"strings"

	"order/internal/store"
	"order/internal/svc"
	"worker/workerclient"
)

// V6.1 工种匹配辅助。
// 背景：建单路径（createorderlogic）已有 jobTypeAllowed 过滤，但兜底扫描 / 批量派单
// （adminbatchdispatchlogic）此前只判断"是否同一楼栋"，导致"电维修派给水工"的错派。
// 这里提供批量派单共用的类别查询与工种判定。

// faultTypeCategoryMap 读取维修类型字典：code -> 类别（electric/water/masonry/wood/other）。
func faultTypeCategoryMap(ctx context.Context, svcCtx *svc.ServiceContext) (map[string]string, error) {
	list, err := store.ListFaultTypes(ctx, svcCtx.DB)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(list))
	for _, ft := range list {
		result[ft.Code] = normalizeFaultCategory(ft.Category)
	}
	return result, nil
}

// orderRequiredJobType 由故障类型代码推导所需工种；字典里查不到时返回 0（不限）。
func orderRequiredJobType(categories map[string]string, faultType string) int64 {
	category, ok := categories[strings.TrimSpace(faultType)]
	if !ok {
		return 0
	}
	return requiredJobTypeOf(category)
}

// hasJobTypeMatch 判断候选人中是否存在工种匹配（或通用）的工人。
func hasJobTypeMatch(workers []*workerclient.WorkerInfo, required int64) bool {
	for _, w := range workers {
		if jobTypeAllowed(w.JobType, required) {
			return true
		}
	}
	return false
}
