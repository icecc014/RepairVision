package logic

import (
	"map/internal/store"
	"map/map"
)

func buildingToPb(b store.Building) *_map.Building {
	return &_map.Building{
		Id:            b.ID,
		Code:          b.Code,
		Name:          b.Name,
		PosX:          b.PosX,
		PosY:          b.PosY,
		Width:         b.Width,
		Height:        b.Height,
		Floors:        b.Floors,
		FloorHeight:   b.FloorHeight,
		RoomsPerFloor: b.RoomsPerFloor,
	}
}

func pbToBuilding(in *_map.Building) *store.Building {
	return &store.Building{
		ID:            in.Id,
		Code:          in.Code,
		Name:          in.Name,
		PosX:          in.PosX,
		PosY:          in.PosY,
		Width:         in.Width,
		Height:        in.Height,
		Floors:        in.Floors,
		FloorHeight:   in.FloorHeight,
		RoomsPerFloor: in.RoomsPerFloor,
	}
}
