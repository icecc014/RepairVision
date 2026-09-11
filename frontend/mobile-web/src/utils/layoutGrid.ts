// 楼层布局网格 v2（V4-S5）：管理员在 PC 端绘制的标准层布局，2D 平面与 3D 楼宇共用同一份数据。
//
// v1（旧）：{ version:1, cols, rows, cells:["112211", ...] }  —— 每格一个字符，只能 1×1。
// v2（现）：{ version:2, cols, rows, blocks:[{ id, kind, row, col, rowSpan, colSpan, label }] }
//          —— 区块可跨格、可带自定义文字，仍是整数格坐标。
//
// 兼容：parseLayout 读到 v1 会自动把每个非 0 单元格升级为 1×1 区块；serializeLayout 一律输出 v2。
// 房间号仍按「楼层 × 100 + 顺序号」生成（顺序 = 行优先、其次列），只有 kind=room 参与编号。

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

export const LAYOUT_VERSION = 2

export type CellCode = '0' | '1' | '2' | '3' | '4' | '5'
export type BlockKind = 'room' | 'corridor' | 'stair' | 'public' | 'custom'

export interface CellTypeMeta {
  code: CellCode
  kind: BlockKind | 'erase'
  label: string
  color: string
  hint: string
}

// 工具栏图例（顺序即工具栏顺序）
export const CELL_TYPES: CellTypeMeta[] = [
  { code: '1', kind: 'room', label: '房间', color: '#dcebff', hint: '宿舍房间，自动编号' },
  { code: '2', kind: 'corridor', label: '过道', color: '#eef3fc', hint: '贯通走廊' },
  { code: '3', kind: 'stair', label: '楼梯', color: '#d9e2f2', hint: '封闭防火楼梯' },
  { code: '4', kind: 'public', label: '公共区', color: '#e7f1fe', hint: '洗衣房 / 活动室等' },
  { code: '5', kind: 'custom', label: '自定义区域', color: '#f2e8ff', hint: '自填名称，如"洗衣房"' },
  { code: '0', kind: 'erase', label: '擦除', color: 'transparent', hint: '点击区块即可删除' },
]

export const KIND_TEXT: Record<BlockKind, string> = {
  room: '房间',
  corridor: '过道',
  stair: '楼梯',
  public: '公共区',
  custom: '自定义区域',
}

const KIND_BY_CODE: Record<string, BlockKind> = {
  '1': 'room',
  '2': 'corridor',
  '3': 'stair',
  '4': 'public',
  '5': 'custom',
}

const CODE_BY_KIND: Record<BlockKind, CellCode> = {
  room: '1',
  corridor: '2',
  stair: '3',
  public: '4',
  custom: '5',
}

export interface LayoutBlock {
  id: string
  kind: BlockKind
  row: number
  col: number
  rowSpan: number
  colSpan: number
  label?: string
}

export interface LayoutGrid {
  version: number
  cols: number
  rows: number
  blocks: LayoutBlock[]
}

export interface PlanBlock {
  type: 'corridor' | 'stair' | 'public' | 'custom'
  x: number
  z: number
  w: number
  d: number
  label?: string
}

export type ResolvedFloorPlan = FloorPlan & {
  custom: boolean
  blocks: PlanBlock[]
}

let blockSeq = 0

export function newBlockId(): string {
  blockSeq += 1
  return `b${Date.now().toString(36)}${blockSeq.toString(36)}`
}

export function codeOf(kind: BlockKind): CellCode {
  return CODE_BY_KIND[kind] || '1'
}

export function kindOfCode(code: string): BlockKind | null {
  return KIND_BY_CODE[code] || null
}

// ---------- 基础创建 ----------

export function emptyLayout(cols = 6, rows = 10): LayoutGrid {
  return { version: LAYOUT_VERSION, cols, rows, blocks: [] }
}

export function makeBlock(kind: BlockKind, row: number, col: number, rowSpan = 1, colSpan = 1, label = ''): LayoutBlock {
  const block: LayoutBlock = { id: newBlockId(), kind, row, col, rowSpan, colSpan }
  if (kind === 'custom') block.label = label || '自定义区域'
  return block
}

// 默认标准层：中间贯通过道 + 两侧共 16 间房 + 两端楼梯与公共区（放不下的部分自动跳过）。
export function defaultLayout(cols = 6, rows = 10): LayoutGrid {
  const grid = emptyLayout(cols, rows)
  const tryPush = (kind: BlockKind, row: number, col: number, rowSpan = 1, colSpan = 1, label = '') => {
    const block = makeBlock(kind, row, col, rowSpan, colSpan, label)
    if (!inBounds(grid, block) || !isAreaFree(grid, block)) return
    grid.blocks.push(block)
  }
  // 走廊：中间两列（不足时取中间一列）
  const corridorCol = cols >= 6 ? 2 : Math.max(0, Math.floor(cols / 2) - 1)
  const corridorSpan = cols >= 6 ? 2 : 1
  tryPush('corridor', 0, corridorCol, rows, corridorSpan)
  // 两端楼梯与公共区
  tryPush('stair', 0, 0, 1, corridorCol || 1)
  tryPush('stair', rows - 1, 0, 1, corridorCol || 1)
  const rightCol = corridorCol + corridorSpan
  const rightSpan = Math.max(1, cols - rightCol)
  tryPush('public', 0, rightCol, 1, rightSpan)
  tryPush('public', rows - 1, rightCol, 1, rightSpan)
  // 两侧房间：中间 4 行
  const startRow = Math.max(1, Math.floor(rows / 2) - 2)
  for (let r = startRow; r < startRow + 4 && r < rows - 1; r++) {
    for (let c = 0; c < corridorCol; c++) tryPush('room', r, c)
    for (let c = rightCol; c < cols; c++) tryPush('room', r, c)
  }
  return grid
}

// ---------- 解析 / 序列化 ----------

function clampInt(value: unknown, min: number, max: number, fallback: number): number {
  const n = Math.round(Number(value))
  if (!Number.isFinite(n)) return fallback
  return Math.min(max, Math.max(min, n))
}

export function parseLayout(json?: string | null): LayoutGrid | null {
  if (!json) return null
  try {
    const raw = JSON.parse(json) as {
      version?: number
      cols?: unknown
      rows?: unknown
      cells?: unknown
      blocks?: unknown
    }
    const cols = clampInt(raw.cols, 2, 24, 0)
    const rows = clampInt(raw.rows, 2, 24, 0)
    if (!cols || !rows) return null
    const blocks: LayoutBlock[] = []
    if (Array.isArray(raw.blocks)) {
      for (const item of raw.blocks as Partial<LayoutBlock>[]) {
        const kind = (item.kind as BlockKind) || KIND_BY_CODE[String(item.kind)]
        if (!kind) continue
        const colSpan = clampInt(item.colSpan, 1, cols, 1)
        const rowSpan = clampInt(item.rowSpan, 1, rows, 1)
        const col = clampInt(item.col, 0, Math.max(0, cols - colSpan), 0)
        const row = clampInt(item.row, 0, Math.max(0, rows - rowSpan), 0)
        const block: LayoutBlock = {
          id: typeof item.id === 'string' && item.id ? item.id : newBlockId(),
          kind,
          row,
          col,
          rowSpan,
          colSpan,
        }
        if (typeof item.label === 'string' && item.label.trim()) block.label = item.label.trim()
        if (!isAreaFree({ cols, rows, blocks } as LayoutGrid, block)) continue
        blocks.push(block)
      }
    } else if (Array.isArray(raw.cells)) {
      // v1 升级：每个非 0 格 → 1×1 区块
      for (let r = 0; r < rows; r++) {
        const line = String((raw.cells as unknown[])[r] ?? '').padEnd(cols, '0')
        for (let c = 0; c < cols; c++) {
          const kind = KIND_BY_CODE[line[c]] || null
          if (!kind) continue
          blocks.push(makeBlock(kind, r, c))
        }
      }
    }
    return { version: LAYOUT_VERSION, cols, rows, blocks }
  } catch {
    return null
  }
}

export function serializeLayout(grid: LayoutGrid): string {
  return JSON.stringify({
    version: LAYOUT_VERSION,
    cols: grid.cols,
    rows: grid.rows,
    blocks: grid.blocks,
  })
}

// v1 兼容视图：把区块压回字符网格（用于调试展示）。
export function toCellRows(grid: LayoutGrid): string[] {
  const rows: string[] = []
  for (let r = 0; r < grid.rows; r++) {
    const line = new Array<string>(grid.cols).fill('0')
    for (const b of grid.blocks) {
      for (let dr = 0; dr < b.rowSpan; dr++) {
        for (let dc = 0; dc < b.colSpan; dc++) {
          const rr = b.row + dr
          const cc = b.col + dc
          if (rr < grid.rows && cc < grid.cols) line[cc] = codeOf(b.kind)
        }
      }
    }
    rows.push(line.join(''))
  }
  return rows
}

// ---------- 几何判定 ----------

export function inBounds(grid: LayoutGrid, b: Pick<LayoutBlock, 'row' | 'col' | 'rowSpan' | 'colSpan'>): boolean {
  return b.row >= 0 && b.col >= 0 && b.rowSpan >= 1 && b.colSpan >= 1 &&
    b.row + b.rowSpan <= grid.rows && b.col + b.colSpan <= grid.cols
}

export function blocksOverlap(
  a: Pick<LayoutBlock, 'row' | 'col' | 'rowSpan' | 'colSpan'>,
  b: Pick<LayoutBlock, 'row' | 'col' | 'rowSpan' | 'colSpan'>,
): boolean {
  return a.row < b.row + b.rowSpan && b.row < a.row + a.rowSpan &&
    a.col < b.col + b.colSpan && b.col < a.col + a.colSpan
}

export function isAreaFree(grid: LayoutGrid, area: LayoutBlock): boolean {
  return grid.blocks.every((b) => b.id === area.id || !blocksOverlap(b, area))
}

export function blockAt(grid: LayoutGrid, row: number, col: number): LayoutBlock | undefined {
  return grid.blocks.find((b) => row >= b.row && row < b.row + b.rowSpan && col >= b.col && col < b.col + b.colSpan)
}

export function usedCellCount(grid: LayoutGrid): number {
  let total = 0
  for (const b of grid.blocks) total += b.rowSpan * b.colSpan
  return total
}

// ---------- 统计与渲染 ----------

export function countRooms(grid: LayoutGrid): number {
  return grid.blocks.filter((b) => b.kind === 'room').length
}

export function countKind(grid: LayoutGrid, kind: BlockKind): number {
  return grid.blocks.filter((b) => b.kind === kind).length
}

// 房间编号：按行优先、其次列排序后依次编号，只有 room 参与。
export function orderedRooms(grid: LayoutGrid): LayoutBlock[] {
  return grid.blocks
    .filter((b) => b.kind === 'room')
    .slice()
    .sort((a, b) => (a.row - b.row) || (a.col - b.col))
}

export function roomNumberOf(grid: LayoutGrid, block: LayoutBlock, floor: number): string {
  const index = orderedRooms(grid).findIndex((b) => b.id === block.id) + 1
  if (index <= 0) return ''
  return `${floor}${String(index).padStart(2, '0')}`
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
  const sorted = grid.blocks.slice().sort((a, b) => (a.row - b.row) || (a.col - b.col))
  let index = 0
  for (const b of sorted) {
    const x = b.col * cw
    const z = b.row * ch
    const w = b.colSpan * cw
    const d = b.rowSpan * ch
    if (b.kind === 'room') {
      index += 1
      rooms.push({
        index,
        no: `${floor}${String(index).padStart(2, '0')}`,
        zone: b.row < rows / 3 ? 'north' : b.row > (rows * 2) / 3 ? 'south' : 'middle',
        x: x + inset,
        z: z + inset,
        w: Math.max(cw * 0.2, w - inset * 2),
        d: Math.max(ch * 0.2, d - inset * 2),
      })
      continue
    }
    if (b.kind === 'corridor') {
      blocks.push({ type: 'corridor', x, z, w, d })
      corridor = corridor
        ? {
            x: Math.min(corridor.x, x),
            z: Math.min(corridor.z, z),
            w: Math.max(corridor.x + corridor.w, x + w) - Math.min(corridor.x, x),
            d: Math.max(corridor.z + corridor.d, z + d) - Math.min(corridor.z, z),
          }
        : { x, z, w, d }
      continue
    }
    if (b.kind === 'stair') {
      blocks.push({ type: 'stair', x, z, w, d, label: b.label })
      cores.push({ index: cores.length + 1, z, d })
      continue
    }
    blocks.push({ type: b.kind === 'public' ? 'public' : 'custom', x, z, w, d, label: b.label })
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

// ---------- 编辑操作：逐格试探 + 碰撞检测 ----------

export type ResizeHandle = 'n' | 's' | 'e' | 'w' | 'nw' | 'ne' | 'sw' | 'se'

// 按格增减尺寸：每一步都检查边界与重叠，遇到冲突立即停在最后一个合法位置。
// handle 含 w/n 时边缘移动会同步调整起点（col/row）与跨格数，实现"向左拖宽一格 / 向右拖回退一格"。
export function stepResize(grid: LayoutGrid, base: LayoutBlock, handle: ResizeHandle, dx: number, dy: number): LayoutBlock {
  let cur: LayoutBlock = { ...base }
  const maxSteps = Math.max(Math.abs(dx), Math.abs(dy))
  for (let s = 1; s <= maxSteps; s++) {
    const cand: LayoutBlock = { ...cur }
    if (Math.abs(dx) >= s) {
      const dir = dx > 0 ? 1 : -1
      if (handle.includes('w')) {
        cand.col += dir
        cand.colSpan -= dir
      } else if (handle.includes('e')) {
        cand.colSpan += dir
      }
    }
    if (Math.abs(dy) >= s) {
      const dir = dy > 0 ? 1 : -1
      if (handle.includes('n')) {
        cand.row += dir
        cand.rowSpan -= dir
      } else if (handle.includes('s')) {
        cand.rowSpan += dir
      }
    }
    if (cand.colSpan < 1 || cand.rowSpan < 1) break
    if (!inBounds(grid, cand)) break
    if (!isAreaFree(grid, cand)) break
    cur = cand
  }
  return cur
}

// 整体平移：同样逐格试探，撞到边界或其它区块即停。
export function stepMove(grid: LayoutGrid, base: LayoutBlock, dx: number, dy: number): LayoutBlock {
  let cur: LayoutBlock = { ...base }
  const maxSteps = Math.max(Math.abs(dx), Math.abs(dy))
  for (let s = 1; s <= maxSteps; s++) {
    const cand: LayoutBlock = { ...cur }
    if (Math.abs(dx) >= s) cand.col += dx > 0 ? 1 : -1
    if (Math.abs(dy) >= s) cand.row += dy > 0 ? 1 : -1
    if (!inBounds(grid, cand)) break
    if (!isAreaFree(grid, cand)) break
    cur = cand
  }
  return cur
}
