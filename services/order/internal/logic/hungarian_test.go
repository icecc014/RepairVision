package logic

import "testing"

func TestHungarianMinimalCost(t *testing.T) {
	cost := [][]int64{
		{1, 2, 3},
		{3, 1, 2},
		{2, 3, 1},
	}
	got := hungarian(cost)
	total := int64(0)
	for i, col := range got {
		if col < 0 || col >= len(cost) {
			t.Fatalf("row %d unassigned", i)
		}
		total += cost[i][col]
	}
	if total != 3 {
		t.Fatalf("expected minimal total cost 3, got %d", total)
	}
}

func TestHungarianCapacitySlots(t *testing.T) {
	// 2 个订单、2 名工人各 1 个槽位：最小总代价应选 (row0,col1)+(row1,col0)=5。
	cost := [][]int64{
		{8, 1},
		{4, 9},
	}
	got := hungarian(cost)
	total := int64(0)
	for i, col := range got {
		total += cost[i][col]
	}
	if total != 5 {
		t.Fatalf("expected 5, got %d", total)
	}
}
