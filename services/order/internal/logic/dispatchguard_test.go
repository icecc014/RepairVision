package logic

import "testing"

func TestGuardThresholds(t *testing.T) {
	// 在岗 4 人，预警 3 倍 = 12，保护 5 倍 = 20；等待预警 2 小时、保护 4 小时
	cases := []struct {
		name             string
		pending          int64
		base             int64
		waitMinutes      int64
		wantGuard        bool
		wantWarn         bool
	}{
		{"正常", 5, 4, 30, false, false},
		{"恰好预警线不触发", 12, 4, 120, false, false},
		{"预警线（数量）", 13, 4, 30, false, true},
		{"保护线（数量）", 21, 4, 30, true, true},
		{"预警线（等待 3 小时）", 1, 4, 180, false, true},
		{"保护线（等待 5 小时）", 1, 4, 300, true, true},
		{"无人在岗按 1 人计：6 单即触保护线", 6, 1, 10, true, true},
	}
	for _, c := range cases {
		guardHit, warnHit := guardThresholds(c.pending, c.base, c.waitMinutes, 3, 5, 2, 4)
		if guardHit != c.wantGuard || warnHit != c.wantWarn {
			t.Fatalf("%s: pending=%d base=%d wait=%d -> guard=%v warn=%v，期望 guard=%v warn=%v",
				c.name, c.pending, c.base, c.waitMinutes, guardHit, warnHit, c.wantGuard, c.wantWarn)
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