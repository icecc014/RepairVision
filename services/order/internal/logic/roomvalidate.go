package logic

import (
	"fmt"
	"strconv"

	"map/mapclient"
	"order/internal/errs"
)

// validateRoomForBuilding 校验房间号是否真实存在于该楼栋指定楼层。
// 房间号规则沿用当前系统约定：楼层号 * 100 + 房间序号，
// 如 4 楼 16 间房的合法区间为 401-416。
func validateRoomForBuilding(room string, floor int64, b *mapclient.Building) error {
	if floor > b.Floors {
		return errs.BadRequest(fmt.Sprintf("%s 当前只有 %d 层，无法报修 %d 层", b.Name, b.Floors, floor))
	}
	num, err := strconv.Atoi(room)
	if err != nil || num <= 0 {
		return errs.BadRequest("房间号格式不正确")
	}
	base := int(floor) * 100
	perFloor := int(b.RoomsPerFloor)
	if perFloor <= 0 {
		perFloor = 16
	}
	if num < base+1 || num > base+perFloor {
		return errs.BadRequest(fmt.Sprintf(
			"%s %d 层房间号范围为 %d01-%d%02d，当前房间 %s 不在范围内",
			b.Name, floor, floor, floor, perFloor, room))
	}
	return nil
}
