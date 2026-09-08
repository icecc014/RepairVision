package seed

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"map/internal/store"
)

var buildings = []store.Building{
	{ID: 1, Code: "A1", Name: "1号宿舍楼", PosX: 100, PosY: 220, Width: 60, Height: 36, Floors: 6, FloorHeight: 3.5, RoomsPerFloor: 20},
	{ID: 2, Code: "A2", Name: "2号宿舍楼", PosX: 300, PosY: 220, Width: 60, Height: 36, Floors: 6, FloorHeight: 3.5, RoomsPerFloor: 20},
	{ID: 3, Code: "B1", Name: "3号宿舍楼", PosX: 200, PosY: 60, Width: 70, Height: 40, Floors: 5, FloorHeight: 3.5, RoomsPerFloor: 18},
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
	for i := range buildings {
		if err := store.InsertBuilding(ctx, conn, &buildings[i]); err != nil {
			return err
		}
		if err := store.UpdateBuildingRoomsPerFloor(ctx, conn, buildings[i].ID, 16); err != nil {
			return err
		}
	}
	return nil
}
