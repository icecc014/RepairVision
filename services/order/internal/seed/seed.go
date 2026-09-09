package seed

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"order/internal/store"
)

type faultSpec struct {
	Code string
	Name string
}

var faults = []faultSpec{
	{Code: "electric", Name: "电维修"},
	{Code: "water", Name: "水维修"},
	{Code: "other", Name: "其他"},
}

type ruleSpec struct {
	Key    string
	Value  float64
	Remark string
}

var rules = []ruleSpec{
	{Key: "skill_weight", Value: 0.4, Remark: "技能匹配权重"},
	{Key: "distance_weight", Value: 0.3, Remark: "距离权重（P2加权派单启用）"},
	{Key: "load_weight", Value: 0.3, Remark: "在途负载权重"},
	{Key: "auto_dispatch_enabled", Value: 1, Remark: "创建工单后是否自动派单"},
	{Key: "pending_timeout_hours", Value: 2, Remark: "待派超时阈值(小时)"},
	{Key: "dispatched_timeout_hours", Value: 4, Remark: "已派未开工超时阈值(小时)"},
}

func Ensure(ctx context.Context, conn sqlx.SqlConn) error {
	var lastErr error
	for i := 0; i < 30; i++ {
		if err := ensureOnce(ctx, conn); err == nil {
			return nil
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return lastErr
}

func ensureOnce(ctx context.Context, conn sqlx.SqlConn) error {
	for _, f := range faults {
		if err := store.InsertFaultType(ctx, conn, f.Code, f.Name); err != nil {
			return err
		}
	}
	for _, r := range rules {
		if err := store.InsertDispatchRuleIfAbsent(ctx, conn, r.Key, r.Value, r.Remark); err != nil {
			return err
		}
	}
	return nil
}
