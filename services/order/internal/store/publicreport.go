package store

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// V7 公共报修的存储层辅助：合并追加、合并审计、按房间查询。

// AppendOrderDescription 把重复报修的补充说明追加到主单描述（合并语义）。
func AppendOrderDescription(ctx context.Context, conn sqlx.Session, orderID int64, extra string) error {
	_, err := conn.ExecCtx(ctx,
		"update orders set description = left(concat(description, ?), 500), updated_at = now() where id = ?",
		extra, orderID)
	return err
}

// InsertDuplicateRecordIfAbsent 记录一次合并（审计用，重复合并同一主单不报错）。
func InsertDuplicateRecordIfAbsent(ctx context.Context, conn sqlx.Session, mainOrderID, subOrderID int64, score float64) error {
	_, err := conn.ExecCtx(ctx,
		"insert ignore into order_duplicate_records(main_order_id, sub_order_id, similarity_score) values(?, ?, ?)",
		mainOrderID, subOrderID, score)
	return err
}

// ListOrdersByRoom 查询某房间的工单（按创建时间倒序）。
func ListOrdersByRoom(ctx context.Context, conn sqlx.Session, buildingID int64, room string, limit int64) ([]Order, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var list []Order
	if err := conn.QueryRowsCtx(ctx, &list,
		orderBase+"where building_id = ? and room = ? order by created_at desc, id desc limit ?",
		buildingID, room, limit); err != nil {
		return nil, err
	}
	return list, nil
}

// FindRecentRoomOrder 找 window 内同房间同类型、且不是子单的最近工单（用于 3 天内合并判定）。
func FindRecentRoomOrder(ctx context.Context, conn sqlx.Session, buildingID, floor int64, room, faultType string, since time.Time) (*Order, error) {
	var o Order
	if err := conn.QueryRowCtx(ctx, &o,
		orderBase+"where building_id = ? and floor = ? and room = ? and fault_type = ? and created_at >= ? and is_merged <> 2 order by id desc limit 1",
		buildingID, floor, room, faultType, since); err != nil {
		return nil, err
	}
	return &o, nil
}

// ListSensitiveWords 读取敏感词表。
func ListSensitiveWords(ctx context.Context, conn sqlx.Session) ([]string, error) {
	var rows []struct {
		Word string `db:"word"`
	}
	if err := conn.QueryRowsCtx(ctx, &rows, "select word from sensitive_words"); err != nil {
		return nil, err
	}
	words := make([]string, 0, len(rows))
	for _, r := range rows {
		words = append(words, r.Word)
	}
	return words, nil
}
