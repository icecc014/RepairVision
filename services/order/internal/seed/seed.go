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
	return nil
}
