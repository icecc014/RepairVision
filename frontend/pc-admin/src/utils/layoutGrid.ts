// 楼层布局网格（V4-S3）：管理员在 PC 端画出的一层布局，2D 平面与 3D 楼宇共用同一份数据。
// 网格单元格编码：0=空白 / 1=房间 / 2=过道 / 3=楼梯 / 4=公共区。
// 房间号按“楼层 × 100 + 顺序号”自动生成（如 4 层第 1 间 = 401），与工单里的房间号一致。

import {
  PLAN_DEPTH,
  PLAN_WIDTH,
  buildFloorPlan,
  buildGridRooms,
  supportsCorridorLayout,
  type FloorPlan,
  type PlanCore,
  type PlanRoom,
} from './floorLayout'

export const LAYOUT_VERSION = 1

export type CellCode = '0' | '1' | '2' | '3' | '4'

export interface CellTypeMeta {
  code: CellCode
  label: string
  color: string
  hint: string
}

// 设计器工具栏使用的图例（顺序即工具栏顺序）。
export const CELL_TYPES: CellTypeMeta[] = [
  { code: '1', label: '房间', color: '#dcebff', hint: '宿舍房间，自动编号' },
  { code: '2', label: '过道', color: '#eef3fc', hint: '贯通走廊' },
  { code: '3', label: '楼梯', color: '#d9e2f2', hint: '封闭防火楼梯' },
  { code: '4', label: '公共区', color: '#e7f1fe', hint: '洗衣房 / 活动室等' },
  { code: '0', label: '空白', color: 'transparent', hint: '擦除为空白' },
]

export interface LayoutGrid {
  version: number
  cols: number
  rows: number
  cells: string[]
}

export interface PlanBlock {
  type: 'corridor' | 'stair' | 'public'
  x: number
  z: number
  w: number
  d: number
}

export type ResolvedFloorPlan = FloorPlan & {
  custom: boolean
  blocks: PlanBlock[]
}

export function emptyLayout(cols = 6, rows = 10): LayoutGrid {
  const cells: string[] = []
  for (let r = 0; r < rows; r++) cells.push('0'.repeat(cols))
  return { version: LAYOUT_VERSION, cols, rows, cells }
}

// 默认标准层：中间贯通过道 + 两侧各 8 间房（共 16 间）+ 两端楼梯与公共区。
export function defaultLayout(cols = 6, rows = 10): LayoutGrid {
  const grid = emptyLayout(cols, rows)
  const put = (r: number, c: number, code: CellCode) => {
    const line = grid.cells[r] || ''
    grid.cells[r] = line.slice(0, c) + code + line.slice(c + 1)
  }
  for (let r = 0; r < rows; r++) {
    put(r, 2, '2')
    put(r, 3, '2')
  }
  for (let r = Math.floor(rows / 2) - 2; r <= Math.floor(rows / 2) + 1; r++) {
    if (r < 0 || r >= rows) continue
    for (const c of [0, 1, 4, 5]) put(r, c, '1')
  }
  put(0, 0, '3')
  put(0, 1, '3')
  put(rows - 1, 0, '3')
  put(rows - 1, 1, '3')
  put(0, 4, '4')
  put(0, 5, '4')
  put(rows - 1, 4, '4')
  put(rows - 1, 5, '4')
  return grid
}

export function parseLayout(json?: string | null): LayoutGrid | null {
  if (!json) return null
  try {
    const raw = JSON.parse(json) as Partial<LayoutGrid>
    const cols = Number(raw.cols)
    const rows = Number(raw.rows)
    if (!Number.isFinite(cols) || !Number.isFinite(rows)) return null
    if (cols < 2 || cols > 24 || rows < 2 || rows > 24) return null
    if (!Array.isArray(raw.cells) || raw.cells.length < rows) return null
    const cells = raw.cells.slice(0, rows).map((line) => String(line).padEnd(cols, '0').slice(0, cols))
    return { version: Number(raw.version) || LAYOUT_VERSION, cols, rows, cells }
  } catch {
    return null
  }
}

export function serializeLayout(grid: LayoutGrid): string {
  return JSON.stringify({ version: LAYOUT_VERSION, cols: grid.cols, rows: grid.rows, cells: grid.cells })
}

export function countRooms(grid: LayoutGrid): number {
  let total = 0
  for (const line of grid.cells) {
    for (const ch of line) if (ch === '1') total++
  }
  return total
}

// 把网格映射到 2D/3D 共用的归一化平面（宽 0-100，深 0-176）。
export function buildPlanFromLayout(grid: LayoutGrid, floor: number): ResolvedFloorPlan {
  const cols = Math.max(1, grid.cols)
  const rows = Math.max(1, grid.rows)
  const cw = PLAN_WIDTH / cols
  const ch = PLAN_DEPTH / rows
  const inset = Math.min(cw, ch) * 0.06
  const rooms: PlanRoom[] = []
  const blocks: PlanBlock[] = []
  const cores: PlanCore[] = []
  let corridor: { x: number; z: number; w: number; d: number } | null = null
  let index = 0
  for (let r = 0; r < rows; r++) {
    const line = (grid.cells[r] || '').padEnd(cols, '0')
    for (let c = 0; c < cols; c++) {
      const code = line[c]
      const x = c * cw
      const z = r * ch
      if (code === '1') {
        index++
        rooms.push({
          index,
          no: `${floor}${String(index).padStart(2, '0')}`,
          zone: r < rows / 3 ? 'north' : r > (rows * 2) / 3 ? 'south' : 'middle',
          x: x + inset,
          z: z + inset,
          w: cw - inset * 2,
          d: ch - inset * 2,
        })
      } else if (code === '2') {
        blocks.push({ type: 'corridor', x, z, w: cw, d: ch })
        corridor = corridor
          ? {
              x: Math.min(corridor.x, x),
              z: Math.min(corridor.z, z),
              w: Math.max(corridor.x + corridor.w, x + cw) - Math.min(corridor.x, x),
              d: Math.max(corridor.z + corridor.d, z + ch) - Math.min(corridor.z, z),
            }
          : { x, z, w: cw, d: ch }
      } else if (code === '3') {
        blocks.push({ type: 'stair', x, z, w: cw, d: ch })
        cores.push({ index: cores.length + 1, z, d: ch })
      } else if (code === '4') {
        blocks.push({ type: 'public', x, z, w: cw, d: ch })
      }
    }
  }
  return {
    floor,
    rooms,
    corridor: corridor || { x: 0, z: 0, w: 0, d: 0 },
    cores,
    custom: true,
    blocks,
  }
}

// 优先使用管理员绘制的布局；没有配置时回退到内置标准层模板。
export function resolveFloorPlan(floor: number, roomsPerFloor: number, layoutJson?: string | null): ResolvedFloorPlan {
  const grid = parseLayout(layoutJson)
  if (grid) {
    const plan = buildPlanFromLayout(grid, floor)
    if (plan.rooms.length > 0) return plan
  }
  if (supportsCorridorLayout(roomsPerFloor)) {
    return { ...buildFloorPlan(floor, roomsPerFloor), custom: false, blocks: [] }
  }
  return {
    floor,
    rooms: buildGridRooms(floor, Math.max(roomsPerFloor || 1, 1)),
    corridor: { x: 0, z: 0, w: 0, d: 0 },
    cores: [],
    custom: false,
    blocks: [],
  }
}
