// 宿舍标准层布局模板（依据 AGENTS/宿舍楼层设计.md，以 2 楼为例推广到整栋楼）
// 每层 16 间：北区 01-04、中区 05-12、南区 13-16；两条核心筒（封闭楼梯+公共区域），
// 中间为贯通南北的走廊。坐标使用归一化平面：宽 0-100，深 0-176，2D 与 3D 共用。

export const PLAN_WIDTH = 100
export const PLAN_DEPTH = 176

export interface PlanRoom {
  index: number
  no: string
  zone: 'north' | 'middle' | 'south'
  x: number
  z: number
  w: number
  d: number
}

export interface PlanCore {
  index: number
  z: number
  d: number
}

export interface FloorPlan {
  floor: number
  rooms: PlanRoom[]
  corridor: { x: number; z: number; w: number; d: number }
  cores: PlanCore[]
}

const CORRIDOR_X = 44
const CORRIDOR_W = 12
const COL_W = 44
const NORTH_ROWS = [0, 16]
const NORTH_ROW_H = 14
const CORE_A_Z = 32
const CORE_H = 14
const MIDDLE_ROWS = [48, 68, 88, 108]
const MIDDLE_ROW_H = 18
const CORE_B_Z = 128
const SOUTH_ROWS = [144, 160]
const SOUTH_ROW_H = 14

// 只有 16 间/层的楼栋使用该标准层模板；其它房间数回退到通用网格。
export function supportsCorridorLayout(roomsPerFloor: number) {
  return roomsPerFloor >= 16
}

export function buildFloorPlan(floor: number, roomsPerFloor: number): FloorPlan {
  const rooms: PlanRoom[] = []
  let index = 1
  const push = (zone: PlanRoom['zone'], z: number, d: number) => {
    rooms.push({ index, no: `${floor}${String(index).padStart(2, '0')}`, zone, x: 0, z, w: COL_W, d })
    rooms.push({ index: index + 1, no: `${floor}${String(index + 1).padStart(2, '0')}`, zone, x: 56, z, w: COL_W, d })
    index += 2
  }
  for (const z of NORTH_ROWS) push('north', z, NORTH_ROW_H)
  for (const z of MIDDLE_ROWS) push('middle', z, MIDDLE_ROW_H)
  for (const z of SOUTH_ROWS) push('south', z, SOUTH_ROW_H)
  if (supportsCorridorLayout(roomsPerFloor)) {
    // 标准 16 间模型（上面的 rows 恰好 2+4+2 行 × 2 列）
  }
  return {
    floor,
    rooms: rooms.slice(0, Math.max(1, Math.min(roomsPerFloor, rooms.length))),
    corridor: { x: CORRIDOR_X, z: 0, w: CORRIDOR_W, d: PLAN_DEPTH },
    cores: [
      { index: 1, z: CORE_A_Z, d: CORE_H },
      { index: 2, z: CORE_B_Z, d: CORE_H },
    ],
  }
}

// 通用网格回退：为非 16 间/层楼栋生成行列布局。
export function buildGridRooms(floor: number, roomsPerFloor: number, cols = 4): PlanRoom[] {
  const count = Math.max(roomsPerFloor, 1)
  const realCols = Math.min(cols, count)
  const rows = Math.ceil(count / realCols)
  const gap = 2
  const cellW = (PLAN_WIDTH - gap * (realCols - 1)) / realCols
  const cellD = (PLAN_DEPTH - gap * (rows - 1)) / rows
  const rooms: PlanRoom[] = []
  for (let i = 0; i < count; i++) {
    const col = i % realCols
    const row = Math.floor(i / realCols)
    rooms.push({
      index: i + 1,
      no: `${floor}${String(i + 1).padStart(2, '0')}`,
      zone: 'middle',
      x: col * (cellW + gap),
      z: row * (cellD + gap),
      w: cellW,
      d: cellD,
    })
  }
  return rooms
}

// 工单匹配：兼容完整房间号（401）、楼层内两位（01）与纯序号三种存法。
export function matchRoomOrders<T extends { floor?: number; room?: string }>(
  orders: T[],
  floorNo: number,
  roomNo: string,
): T[] {
  const suffix = roomNo.slice(String(floorNo).length)
  return orders.filter((o) => {
    if (o.floor && o.floor !== floorNo) return false
    const room = (o.room || '').trim()
    if (!room) return false
    return room === roomNo || room === suffix || room.endsWith(suffix)
  })
}
