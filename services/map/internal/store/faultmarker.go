package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func UpsertFaultMarker(ctx context.Context, conn sqlx.Session, orderID, buildingID, floor int64, roomNumber string) error {
	_, err := conn.ExecCtx(ctx,
		`insert into fault_markers(order_id, building_id, floor, room_number)
		 values(?,?,?,?) as new
		 on duplicate key update building_id = new.building_id, floor = new.floor, room_number = new.room_number`,
		orderID, buildingID, floor, roomNumber)
	return err
}

func RemoveFaultMarker(ctx context.Context, conn sqlx.Session, orderID int64) error {
	_, err := conn.ExecCtx(ctx,
		"delete from fault_markers where order_id = ?", orderID)
	return err
}
