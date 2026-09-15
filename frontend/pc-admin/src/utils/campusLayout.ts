// 区域概览（校园/建筑群总平面图）网格模型（V5.1）
// 复用楼层布局的整数格几何算法（边界判定、碰撞检测、逐格试探尺寸编辑），
// 但图元类型与楼层不同：建筑 / 道路 / 广场绿地 / 校门 / 自定义。

import {
  blocksOverlap,
  inBounds,
  isAreaFree,
  newBlockId,
  stepMove,
  stepResize,
  type LayoutBlock,
  type LayoutGrid,
  type ResizeHandle,
} from './layoutGrid'

export const CAMPUS_VERSION = 1

export type CampusKind = 'building' | 'road' | 'green' | 'gate' | 'custom'

export interface CampusBlock {
  id: string
  kind: CampusKind
  row: number
  col: number
  rowSpan: number
  colSpan: number
  label?: string
  buildingId?: number
  customBuilding?: boolean
}

export interface CampusGrid {
  version: number
  cols: number
  rows: number
  blocks: CampusBlock[]
}

export interface CampusKindMeta {
  kind: CampusKind
  label: string
  color: string
  hint: string
}

export const CAMPUS_KINDS: CampusKindMeta[] = [
  { kind: 'building', label: '建筑', color: '#dcebff', hint: '关联楼栋或自定义命名' },
  { kind: 'road', label: '道路', color: '#e9e2d6', hint: '连接建筑的路网' },
  { kind: 'green', label: '广场/绿地', color: '#ddf0e3', hint: '广场、绿地、运动场' },
  { kind: 'gate', label: '校门/出入口', color: '#f7e3ef', hint: '校园出入口' },
  { kind: 'custom', label: '自定义', color: '#f2e8ff', hint: '其他设施，可命名' },
]

export const CAMPUS_KIND_TEXT: Record<CampusKind, string> = {
  building: '建筑',
  road: '道路',
  green: '广场/绿地',
  gate: '校门/出入口',
  custom: '自定义',
}

export function makeCampusBlock(kind: CampusKind, row: number, col: number, rowSpan = 1, colSpan = 1, label = ''): CampusBlock {
  const block: CampusBlock = { id: newBlockId(), kind, row, col, rowSpan, colSpan }
  if (label) block.label = label
  return block
}

export function emptyCampus(cols = 40, rows = 30): CampusGrid {
  return { version: CAMPUS_VERSION, cols, rows, blocks: [] }
}

// 示例布局：两排建筑 + 中间主路 + 两条支路 + 校门 + 广场（便于管理员快速起步）
// 示例布局（V6 规模）：16 栋宿舍楼（A1~A8 / B1~B8）+ 主干道 2 条 + 纵向支路 6 条 + 广场 + 校门
export function defaultCampus(cols = 40, rows = 30): CampusGrid {
  const grid = emptyCampus(cols, rows)
  const push = (kind: CampusKind, row: number, col: number, rowSpan = 1, colSpan = 1, label = '') => {
    const block = makeCampusBlock(kind, row, col, rowSpan, colSpan, label)
    if (campusInBounds(grid, block) && campusAreaFree(grid, block)) grid.blocks.push(block)
  }
  push('road', 12, 0, 2, cols)
  push('road', 26, 0, 2, cols)
  for (let k = 0; k < 6; k++) {
    push('road', 0, 3 + k * 7, rows, 1)
  }
  for (let n = 0; n < 8; n++) {
    const col = 4 + n * 4
    push('building', 3, col, 5, 3, `A${n + 1} 宿舍楼`)
    push('building', 17, col, 5, 3, `B${n + 1} 宿舍楼`)
  }
  push('green', 3, 36, 5, 3, '中心广场')
  push('gate', 0, 6, 1, 3, '北门')
  return grid
}
export function parseCampus(json?: string | null): CampusGrid | null {
  if (!json) return null
  try {
    const raw = JSON.parse(json) as Partial<CampusGrid>
    const cols = Math.round(Number(raw.cols))
    const rows = Math.round(Number(raw.rows))
    if (!Number.isFinite(cols) || !Number.isFinite(rows) || cols < 4 || rows < 4 || cols > 80 || rows > 60) return null
    if (!Array.isArray(raw.blocks)) return { version: CAMPUS_VERSION, cols, rows, blocks: [] }
    const blocks: CampusBlock[] = []
    for (const item of raw.blocks as CampusBlock[]) {
      if (!item || !CAMPUS_KIND_TEXT[item.kind as CampusKind]) continue
      const colSpan = Math.max(1, Math.min(cols, Math.round(Number(item.colSpan) || 1)))
      const rowSpan = Math.max(1, Math.min(rows, Math.round(Number(item.rowSpan) || 1)))
      const block: CampusBlock = {
        id: typeof item.id === 'string' && item.id ? item.id : newBlockId(),
        kind: item.kind,
        col: Math.max(0, Math.min(cols - colSpan, Math.round(Number(item.col) || 0))),
        row: Math.max(0, Math.min(rows - rowSpan, Math.round(Number(item.row) || 0))),
        rowSpan,
        colSpan,
      }
      if (typeof item.label === 'string' && item.label.trim()) block.label = item.label.trim()
      if (typeof item.buildingId === 'number' && item.buildingId > 0) block.buildingId = item.buildingId
      if (item.customBuilding) block.customBuilding = true
      if (!campusAreaFree({ ...raw, cols, rows, blocks } as CampusGrid, block)) continue
      blocks.push(block)
    }
    return { version: CAMPUS_VERSION, cols, rows, blocks }
  } catch {
    return null
  }
}

export function serializeCampus(grid: CampusGrid): string {
  return JSON.stringify({ version: CAMPUS_VERSION, cols: grid.cols, rows: grid.rows, blocks: grid.blocks })
}

export function campusCount(grid: CampusGrid, kind: CampusKind): number {
  return grid.blocks.filter((b) => b.kind === kind).length
}

export function buildingBlockOf(grid: CampusGrid, block: CampusBlock): CampusBlock | undefined {
  if (block.kind !== 'building') return undefined
  return grid.blocks.find((b) => b.kind === 'building' && b.id === block.id)
}

// ---------- 复用楼层布局的几何算法（类型转换为通用形状）----------

export function campusInBounds(grid: CampusGrid, b: CampusBlock): boolean {
  return inBounds(grid as unknown as LayoutGrid, b as unknown as LayoutBlock)
}

export function campusOverlap(a: CampusBlock, b: CampusBlock): boolean {
  return blocksOverlap(a as unknown as LayoutBlock, b as unknown as LayoutBlock)
}

export function campusAreaFree(grid: CampusGrid, area: CampusBlock): boolean {
  return isAreaFree(grid as unknown as LayoutGrid, area as unknown as LayoutBlock)
}

export function campusStepResize(grid: CampusGrid, base: CampusBlock, handle: ResizeHandle, dx: number, dy: number): CampusBlock {
  return stepResize(grid as unknown as LayoutGrid, base as unknown as LayoutBlock, handle, dx, dy) as unknown as CampusBlock
}

export function campusStepMove(grid: CampusGrid, base: CampusBlock, dx: number, dy: number): CampusBlock {
  return stepMove(grid as unknown as LayoutGrid, base as unknown as LayoutBlock, dx, dy) as unknown as CampusBlock
}
