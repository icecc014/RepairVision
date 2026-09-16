package logic

import "testing"

func TestGuardThresholds(t *testing.T) {
	// V6.3 默认：预警 = max(在岗×5, 10) 单 或 等待 > 6 小时；
	// 保护（暂停）= 待派 ≥ max(在岗×10, 20) 单；等待 > 12 小时则强制恢复。
	cases := []struct {
		name        string
		pending     int64
		base        int64
		waitMinutes int64
		wantGuard   bool
		wantWarn    bool
		wantForce   bool
	}{
		{"正常（3 单、等待 30 分钟）", 3, 1, 30, false, false, false},
		{"在岗 1 人、9 单不触预警", 9, 1, 30, false, false, false},
		{"在岗 1 人、10 单只到预警线", 10, 1, 30, false, true, false},
		{"在岗 1 人、20 单触保护线", 20, 1, 30, true, true, false},
		{"在岗 4 人：19 单不触预警（门槛 20）", 19, 4, 30, false, false, false},
		{"在岗 4 人：20 单触预警", 20, 4, 30, false, true, false},
		{"在岗 4 人：40 单触保护线", 40, 4, 30, true, true, false},
		{"等待 5 小时不触预警（门槛 6 小时）", 1, 4, 300, false, false, false},
		{"等待 7 小时只预警、不暂停", 1, 4, 420, false, true, false},
		{"等待 13 小时：仍不暂停，但触发强制恢复", 1, 4, 780, false, true, true},
	}
	for _, c := range cases {
		guardHit, warnHit, forceResume := guardThresholds(c.pending, c.base, c.waitMinutes, 5, 10, 10, 20, 6, 12)
		if guardHit != c.wantGuard || warnHit != c.wantWarn || forceResume != c.wantForce {
			t.Fatalf("%s: pending=%d base=%d wait=%d -> guard=%v warn=%v force=%v，期望 guard=%v warn=%v force=%v",
				c.name, c.pending, c.base, c.waitMinutes, guardHit, warnHit, forceResume, c.wantGuard, c.wantWarn, c.wantForce)
		}
	}
}

func TestGuardLevelText(t *testing.T) {
	if lv, _ := guardLevelText(false, false); lv != guardLevelOK {
		t.Fatalf("期望 ok，实际 %s", lv)
	}
	if lv, _ := guardLevelText(false, true); lv != guardLevelWarn {
		t.Fatalf("期望 warn，实际 %s", lv)
	}
	if lv, msg := guardLevelText(true, true); lv != guardLevelGuard || msg == "" {
		t.Fatalf("期望 guard 且带文案，实际 %s / %s", lv, msg)
	}
}