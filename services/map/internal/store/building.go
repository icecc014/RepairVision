package store

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Building struct {
	ID            int64   `db:"id"`
	Code          string  `db:"code"`
	Name          string  `db:"name"`
	PosX          float64 `db:"pos_x"`
	PosY          float64 `db:"pos_y"`
	Width         float64 `db:"width"`
	Height        float64 `db:"height"`
	Floors        int64   `db:"floors"`
	FloorHeight   float64 `db:"floor_height"`
	RoomsPerFloor int64   `db:"rooms_per_floor"`
}

const buildingSelect = `select id, code, name,
	cast(pos_x as double) pos_x,
	cast(pos_y as double) pos_y,
	cast(coalesce(width, 50) as double) width,
	cast(coalesce(height, 30) as double) height,
	coalesce(floors, 1) floors,
	cast(coalesce(floor_height, 3.5) as double) floor_height,
	coalesce(rooms_per_floor, 10) rooms_per_floor
	from buildings`

func ListBuildings(ctx context.Context, conn sqlx.Session) ([]Building, error) {
	var buildings []Building
	if err := conn.QueryRowsCtx(ctx, &buildings, buildingSelect+" order by id"); err != nil {
		return nil, err
	}
	return buildings, nil
}

func FindBuildingByID(ctx context.Context, conn sqlx.Session, id int64) (*Building, error) {
	var b Building
	if err := conn.QueryRowCtx(ctx, &b, buildingSelect+" where id = ?", id); err != nil {
		return nil, err
	}
	return &b, nil
}

func InsertBuilding(ctx context.Context, conn sqlx.Session, b *Building) error {
	_, err := conn.ExecCtx(ctx,
		`insert ignore into buildings(id, code, name, pos_x, pos_y, width, height, floors, floor_height, rooms_per_floor)
		 values(?,?,?,?,?,?,?,?,?,?)`,
		b.ID, b.Code, b.Name, b.PosX, b.PosY, b.Width, b.Height, b.Floors, b.FloorHeight, b.RoomsPerFloor)
	return err
}

func CreateBuilding(ctx context.Context, conn sqlx.Session, b *Building) (int64, error) {
	result, err := conn.ExecCtx(ctx,
		`insert into buildings(code, name, pos_x, pos_y, width, height, floors, floor_height, rooms_per_floor)
		 values(?,?,?,?,?,?,?,?,?)`,
		b.Code, b.Name, b.PosX, b.PosY, b.Width, b.Height, b.Floors, b.FloorHeight, b.RoomsPerFloor)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateBuilding(ctx context.Context, conn sqlx.Session, b *Building) error {
	_, err := conn.ExecCtx(ctx,
		`update buildings set code = ?, name = ?, pos_x = ?, pos_y = ?, width = ?, height = ?,
			floors = ?, floor_height = ?, rooms_per_floor = ? where id = ?`,
		b.Code, b.Name, b.PosX, b.PosY, b.Width, b.Height, b.Floors, b.FloorHeight, b.RoomsPerFloor, b.ID)
	return err
}

func DeleteBuilding(ctx context.Context, conn sqlx.Session, id int64) error {
	_, err := conn.ExecCtx(ctx,
		"delete from buildings where id = ?", id)
	return err
}
