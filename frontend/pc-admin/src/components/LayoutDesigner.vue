<template>
  <el-dialog
    append-to-body
    :model-value="modelValue"
    :title="`楼层布局设计 · ${building?.name || ''}`"
    width="1320px"
    top="3vh"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
  >
    <div class="designer">
      <aside class="tools">
        <div class="tool-title">绘制工具栏</div>
        <button
          v-for="t in CELL_TYPES"
          :key="t.code"
          class="tool-btn"
          :class="{ active: brush === t.code }"
          @click="selectBrush(t.code)"
        >
          <i class="chip" :style="{ background: t.color }" />
          <span>{{ t.label }}</span>
          <em>{{ t.hint }}</em>
        </button>

        <div class="tool-title">画布尺寸</div>
        <div class="size-row">
          <el-input-number v-model="cols" :min="4" :max="16" size="small" @change="applySize" />
          <span class="times">×</span>
          <el-input-number v-model="rows" :min="4" :max="20" size="small" @change="applySize" />
        </div>

        <div class="tool-title">快捷操作</div>
        <el-button size="small" @click="loadTemplate">一键生成标准层</el-button>
        <el-button size="small" @click="clearAll">清空画布</el-button>
        <el-button size="small" type="danger" plain @click="clearCustom">清除自定义布局</el-button>

        <div class="tool-title">快捷键</div>
        <ul class="key-list">
          <li><b>双击</b> 区块进入编辑模式</li>
          <li><b>拖动把手</b> 按格改尺寸</li>
          <li><b>编辑模式拖动</b> 整体移动</li>
          <li><b>Ctrl+Z / Ctrl+Shift+Z</b> 撤销 / 重做</li>
          <li><b>Delete</b> 删除选中区块</li>
          <li><b>Esc</b> 退出编辑模式</li>
        </ul>
      </aside>

      <section class="canvas-area">
        <div class="hint-line" :class="{ warn: !brush || !!editingId, editing: !!editingId }">
          {{ editingId ? '编辑模式：只能调整当前选中的区块（拖把手改尺寸 / 拖本体移动 / 右侧「左扩·右扩」按钮）；按 Esc 或点击空白处退出' : (brush ? '已选择「' + brushLabel + '」：点击网格放置，按住拖动可连续绘制；双击区块可改尺寸' : '请先在左侧选择绘制工具（房间 / 过道 / 楼梯 / 公共区 / 自定义区域 / 擦除），再点击网格开始绘制') }}
        </div>

        <div class="canvas-scroll">
          <div
            ref="canvasRef"
            class="canvas"
            :class="{ 'no-brush': !brush }"
            :style="{ '--cols': cols, '--rows': rows }"
          >
            <div class="slot-layer">
              <button
                v-for="slot in slots"
                :key="slot.key"
                class="slot"
                :class="{ painted: !!slot.blockId }"
                @pointerdown.prevent="onSlotDown(slot)"
                @pointerenter="onSlotEnter(slot)"
              />
            </div>

            <div class="block-layer">
              <div
                v-for="b in grid.blocks"
                :key="b.id"
                class="block"
                :class="[`k-${b.kind}`, { selected: selectedId === b.id, editing: editingId === b.id }]"
                :style="blockStyle(b)"
                @pointerdown.stop="onBlockDown(b, $event)"
                @dblclick.stop="enterEdit(b)"
              >
                <span class="block-label">{{ labelOf(b) }}</span>
                <template v-if="editingId === b.id">
                  <i
                    v-for="h in HANDLES"
                    :key="h"
                    class="handle"
                    :class="`h-${h}`"
                    @pointerdown.stop="onHandleDown(b, h, $event)"
                  />
                </template>
              </div>
            </div>
          </div>
        </div>

        <div class="summary">
          <span>房间 <b>{{ roomCount }}</b> 间</span>
          <span>过道 <b>{{ kindCount('corridor') }}</b> 块</span>
          <span>楼梯 <b>{{ kindCount('stair') }}</b> 块</span>
          <span>公共区 <b>{{ kindCount('public') }}</b> 块</span>
          <span>自定义 <b>{{ kindCount('custom') }}</b> 块</span>
          <span>楼层 <b>{{ building?.floors || 0 }}</b> 层</span>
        </div>
        <p class="tips">
          保存后该楼栋的 2D 平面图与 3D 楼宇都会按这份布局逐层渲染；房间号按「楼层 × 100 + 顺序」自动生成，
          自定义区域不参与编号、也不显示故障红点。
        </p>
      </section>

      <aside class="props">
        <div class="tool-title">区块属性</div>
        <template v-if="selectedBlock">
          <div class="prop-row">
            <span class="prop-label">类型</span>
            <el-select :model-value="selectedBlock.kind" size="small" style="width: 100%" @change="(v: BlockKind) => setKind(v)">
              <el-option label="房间" value="room" />
              <el-option label="过道" value="corridor" />
              <el-option label="楼梯" value="stair" />
              <el-option label="公共区" value="public" />
              <el-option label="自定义区域" value="custom" />
            </el-select>
          </div>
          <div v-if="selectedBlock.kind === 'custom'" class="prop-row">
            <span class="prop-label">名称</span>
            <el-input
              :model-value="selectedBlock.label"
              size="small"
              maxlength="12"
              placeholder="如：洗衣房"
              @change="(v: string) => setLabel(v)"
            />
          </div>
          <div class="prop-row">
            <span class="prop-label">起始格</span>
            <div class="prop-pair">
              <el-input-number :model-value="selectedBlock.row" :min="0" :max="rows - 1" size="small" @change="(v: number) => setProp('row', v)" />
              <el-input-number :model-value="selectedBlock.col" :min="0" :max="cols - 1" size="small" @change="(v: number) => setProp('col', v)" />
            </div>
          </div>
          <div class="prop-row">
            <span class="prop-label">跨格数</span>
            <div class="prop-pair">
              <el-input-number :model-value="selectedBlock.rowSpan" :min="1" :max="rows" size="small" @change="(v: number) => setProp('rowSpan', v)" />
              <el-input-number :model-value="selectedBlock.colSpan" :min="1" :max="cols" size="small" @change="(v: number) => setProp('colSpan', v)" />
            </div>
          </div>
          <div class="prop-row">
            <span class="prop-label">左右边缘（外扩 / 内收）</span>
            <div class="prop-pair">
              <el-button size="small" @click="expandEdge('w', true)">← 左扩</el-button>
              <el-button size="small" @click="expandEdge('w', false)">→ 左收</el-button>
            </div>
            <div class="prop-pair">
              <el-button size="small" @click="expandEdge('e', true)">右扩 →</el-button>
              <el-button size="small" @click="expandEdge('e', false)">← 右收</el-button>
            </div>
          </div>
          <div class="prop-row">
            <span class="prop-label">上下边缘（外扩 / 内收）</span>
            <div class="prop-pair">
              <el-button size="small" @click="expandEdge('n', true)">↑ 上扩</el-button>
              <el-button size="small" @click="expandEdge('n', false)">↓ 上收</el-button>
            </div>
            <div class="prop-pair">
              <el-button size="small" @click="expandEdge('s', true)">下扩 ↓</el-button>
              <el-button size="small" @click="expandEdge('s', false)">↑ 下收</el-button>
            </div>
          </div>
          <el-button size="small" type="danger" plain style="width: 100%" @click="removeSelected">删除该区块</el-button>
          <p class="prop-tip">起始格 / 跨格数 越界或与其它区块重叠时会自动回退到最近合法值；「左扩 / 上扩」会把区块起点一起往外移，因此 1 格宽的房间也能往左、往上变大。</p>
        </template>
        <p v-else class="prop-empty">单击区块可查看属性；双击区块进入编辑模式（显示 8 个把手）。</p>
      </aside>
    </div>

    <template #footer>
      <div class="footer-bar">
        <span class="history-tip">撤销栈 {{ undoStack.length }} 步</span>
        <el-button size="small" :disabled="undoStack.length === 0" @click="undo">撤销</el-button>
        <el-button size="small" :disabled="redoStack.length === 0" @click="redo">重做</el-button>
        <div style="flex: 1" />
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存布局</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminBuilding } from '../api'
import { apiUpdateBuilding } from '../api'
import {
  CELL_TYPES,
  KIND_TEXT,
  buildPlanFromLayout,
  countKind,
  countRooms,
  defaultLayout,
  emptyLayout,
  inBounds,
  kindOfCode,
  makeBlock,
  parseLayout,
  serializeLayout,
  stepMove,
  stepResize,
  blocksOverlap,
  isAreaFree,
  type BlockKind,
  type CellCode,
  type LayoutBlock,
  type LayoutGrid,
  type ResizeHandle,
} from '../utils/layoutGrid'

const props = defineProps<{ modelValue: boolean; building: AdminBuilding | null }>()
const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'saved'): void
}>()

type Handle = ResizeHandle
const HANDLES: Handle[] = ['nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w']

const grid = ref<LayoutGrid>(defaultLayout())
const cols = ref(6)
const rows = ref(10)
const brush = ref<CellCode | null>(null)
const painting = ref(false)
const saving = ref(false)
const selectedId = ref<string | null>(null)
const editingId = ref<string | null>(null)
const canvasRef = ref<HTMLElement | null>(null)
const undoStack = ref<LayoutBlock[][]>([])
const redoStack = ref<LayoutBlock[][]>([])

const selectedBlock = computed(() => grid.value.blocks.find((b) => b.id === selectedId.value) || null)
const roomCount = computed(() => countRooms(grid.value))
const brushLabel = computed(() => CELL_TYPES.find((t) => t.code === brush.value)?.label || '')
const slots = computed(() => {
  const list: { key: string; row: number; col: number; blockId: string | null }[] = []
  for (let r = 0; r < grid.value.rows; r++) {
    for (let c = 0; c < grid.value.cols; c++) {
      const hit = grid.value.blocks.find(
        (b) => r >= b.row && r < b.row + b.rowSpan && c >= b.col && c < b.col + b.colSpan,
      )
      list.push({ key: `${r}-${c}`, row: r, col: c, blockId: hit?.id ?? null })
    }
  }
  return list
})

function cloneBlocks(): LayoutBlock[] {
  return grid.value.blocks.map((b) => ({ ...b }))
}

function pushHistory() {
  undoStack.value.push(cloneBlocks())
  if (undoStack.value.length > 50) undoStack.value.shift()
  redoStack.value = []
}

function undo() {
  const last = undoStack.value.pop()
  if (!last) return
  redoStack.value.push(cloneBlocks())
  grid.value.blocks = last
  syncSelection()
}

function redo() {
  const next = redoStack.value.pop()
  if (!next) return
  undoStack.value.push(cloneBlocks())
  grid.value.blocks = next
  syncSelection()
}

function syncSelection() {
  if (selectedId.value && !grid.value.blocks.some((b) => b.id === selectedId.value)) selectedId.value = null
  if (editingId.value && !grid.value.blocks.some((b) => b.id === editingId.value)) editingId.value = null
}

function kindCount(kind: BlockKind) {
  return countKind(grid.value, kind)
}

function labelOf(b: LayoutBlock) {
  if (b.kind === 'room') {
    const plan = buildPlanFromLayout(grid.value, 1)
    return plan.rooms.find((r) => r.index === roomIndex(b))?.no || ''
  }
  return b.label || KIND_TEXT[b.kind]
}

function roomIndex(b: LayoutBlock) {
  const rooms = grid.value.blocks
    .filter((x) => x.kind === 'room')
    .slice()
    .sort((m, n) => (m.row - n.row) || (m.col - n.col))
  return rooms.findIndex((x) => x.id === b.id) + 1
}

function blockStyle(b: LayoutBlock) {
  return {
    left: `${(b.col / grid.value.cols) * 100}%`,
    top: `${(b.row / grid.value.rows) * 100}%`,
    width: `${(b.colSpan / grid.value.cols) * 100}%`,
    height: `${(b.rowSpan / grid.value.rows) * 100}%`,
  }
}

// ---------- 绘制 ----------

function selectBrush(code: CellCode) {
  brush.value = brush.value === code ? null : code
  painting.value = false
  // 选择工具即退出编辑模式，避免"编辑中又画出新房间"
  editingId.value = null
}

function placeAt(row: number, col: number) {
  if (editingId.value) return // 编辑模式只调整当前区块，不放置新内容
  const kind = brush.value ? kindOfCode(brush.value) : null
  const hit = grid.value.blocks.find((b) => row >= b.row && row < b.row + b.rowSpan && col >= b.col && col < b.col + b.colSpan)
  if (!kind) {
    // 擦除
    if (hit) {
      pushHistory()
      grid.value.blocks = grid.value.blocks.filter((b) => b.id !== hit.id)
      syncSelection()
    }
    return
  }
  if (hit) return
  pushHistory()
  const block = makeBlock(kind, row, col)
  grid.value.blocks.push(block)
  selectedId.value = block.id
}

function onSlotDown(slot: { row: number; col: number }) {
  // 编辑模式：只允许调整当前选中区块，点击空白处退出编辑
  if (editingId.value) {
    editingId.value = null
    selectedId.value = null
    return
  }
  if (!brush.value) {
    ElMessage.warning('请先在左侧选择绘制工具，再点击网格放置')
    return
  }
  painting.value = true
  if (brush.value === '5') {
    createCustomAt(slot.row, slot.col)
    painting.value = false
    return
  }
  placeAt(slot.row, slot.col)
}

function onSlotEnter(slot: { row: number; col: number }) {
  if (editingId.value || !painting.value) return
  placeAt(slot.row, slot.col)
}

async function createCustomAt(row: number, col: number) {
  const hit = grid.value.blocks.find((b) => row >= b.row && row < b.row + b.rowSpan && col >= b.col && col < b.col + b.colSpan)
  if (hit) return
  let label = ''
  try {
    const res = await ElMessageBox.prompt('请输入该区域的名称（如：洗衣房 / 自习室 / 值班室）', '自定义区域', {
      inputValue: '',
      inputValidator: (v: string) => (v && v.trim() ? true : '名称不能为空'),
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
    label = String(res.value || '').trim().slice(0, 12)
  } catch {
    return
  }
  pushHistory()
  const block = makeBlock('custom', row, col, 1, 1, label)
  grid.value.blocks.push(block)
  selectedId.value = block.id
}

// ---------- 选中 / 编辑模式 ----------

function onBlockDown(b: LayoutBlock, ev: PointerEvent) {
  // 编辑模式：只能操作当前区块（点其它区块=切换目标，点本体=整体拖动）
  if (editingId.value) {
    if (editingId.value !== b.id) {
      editingId.value = b.id
      selectedId.value = b.id
      return
    }
    startMove(b, ev)
    return
  }
  if (!brush.value) {
    ElMessage.warning('请先在左侧选择绘制工具，再点击网格放置')
    return
  }
  selectedId.value = b.id
  if (brush.value === '0') {
    pushHistory()
    grid.value.blocks = grid.value.blocks.filter((x) => x.id !== b.id)
    syncSelection()
    return
  }
}

function enterEdit(b: LayoutBlock) {
  selectedId.value = b.id
  editingId.value = editingId.value === b.id ? null : b.id
}

function removeSelected() {
  if (!selectedId.value) return
  pushHistory()
  grid.value.blocks = grid.value.blocks.filter((b) => b.id !== selectedId.value)
  syncSelection()
}

function setKind(kind: BlockKind) {
  const b = selectedBlock.value
  if (!b || b.kind === kind) return
  pushHistory()
  b.kind = kind
  if (kind === 'custom' && !b.label) b.label = '自定义区域'
  if (kind !== 'custom') delete b.label
}

function setLabel(value: string) {
  const b = selectedBlock.value
  if (!b) return
  const next = String(value || '').trim().slice(0, 12)
  if ((b.label || '') === next) return
  pushHistory()
  b.label = next
}

// 先入栈再改值，保证撤销能回到修改前的状态
function setProp(field: 'row' | 'col' | 'rowSpan' | 'colSpan', value: number) {
  const b = selectedBlock.value
  if (!b) return
  const before = { ...b }
  pushHistory()
  b[field] = Number(value) || 0
  applyPropChange()
  if (b.row === before.row && b.col === before.col && b.rowSpan === before.rowSpan && b.colSpan === before.colSpan) {
    undoStack.value.pop() // 无实际变化，丢弃这次历史
  }
}

// 边缘按格扩/收：左扩、上扩会同时移动起点，解决"1 格宽无法往左/往上扩大"的问题
function expandEdge(dir: 'w' | 'e' | 'n' | 's', grow: boolean) {
  const b = selectedBlock.value
  if (!b) return
  const dx = dir === 'w' ? (grow ? -1 : 1) : dir === 'e' ? (grow ? 1 : -1) : 0
  const dy = dir === 'n' ? (grow ? -1 : 1) : dir === 's' ? (grow ? 1 : -1) : 0
  pushHistory()
  const next = stepResize(grid.value, b, dir, dx, dy)
  const changed = next.row !== b.row || next.col !== b.col || next.rowSpan !== b.rowSpan || next.colSpan !== b.colSpan
  if (!changed) {
    undoStack.value.pop()
    ElMessage.warning(grow ? '该方向已到边界或被其它区块挡住' : '该方向已缩到最小 1 格')
    return
  }
  b.row = next.row
  b.col = next.col
  b.rowSpan = next.rowSpan
  b.colSpan = next.colSpan
}

// ---------- 尺寸 / 位置编辑（逐格试探 + 碰撞检测）----------

const drag = reactive({
  mode: '' as '' | 'resize' | 'move',
  id: '',
  handle: 'se' as Handle,
  base: null as LayoutBlock | null,
  startX: 0,
  startY: 0,
  cellW: 1,
  cellH: 1,
})

function beginDrag(mode: 'resize' | 'move', b: LayoutBlock, handle: Handle, ev: PointerEvent) {
  const rect = canvasRef.value?.getBoundingClientRect()
  if (!rect) return
  pushHistory()
  drag.mode = mode
  drag.id = b.id
  drag.handle = handle
  drag.base = { ...b }
  drag.startX = ev.clientX
  drag.startY = ev.clientY
  drag.cellW = rect.width / grid.value.cols
  drag.cellH = rect.height / grid.value.rows
  selectedId.value = b.id
  editingId.value = b.id
}

function onHandleDown(b: LayoutBlock, handle: Handle, ev: PointerEvent) {
  beginDrag('resize', b, handle, ev)
}

function startMove(b: LayoutBlock, ev: PointerEvent) {
  beginDrag('move', b, 'se', ev)
}

function onPointerMove(ev: PointerEvent) {
  if (!drag.mode || !drag.base) return
  const dx = Math.round((ev.clientX - drag.startX) / drag.cellW)
  const dy = Math.round((ev.clientY - drag.startY) / drag.cellH)
  const target = grid.value.blocks.find((b) => b.id === drag.id)
  if (!target) return
  const next = drag.mode === 'resize'
    ? stepResize(grid.value, drag.base, drag.handle, dx, dy)
    : stepMove(grid.value, drag.base, dx, dy)
  target.row = next.row
  target.col = next.col
  target.rowSpan = next.rowSpan
  target.colSpan = next.colSpan
}

function endDrag() {
  if (!drag.mode) return
  const target = grid.value.blocks.find((b) => b.id === drag.id)
  const base = drag.base
  // 没有实际变化时回滚这次历史（避免空撤销）
  if (target && base && target.row === base.row && target.col === base.col &&
    target.rowSpan === base.rowSpan && target.colSpan === base.colSpan) {
    undoStack.value.pop()
  }
  drag.mode = ''
  drag.base = null
  painting.value = false
}

function applyPropChange() {
  const b = selectedBlock.value
  if (!b) return
  const cand: LayoutBlock = {
    ...b,
    row: Math.max(0, Math.min(rows.value - 1, Number(b.row) || 0)),
    col: Math.max(0, Math.min(cols.value - 1, Number(b.col) || 0)),
    rowSpan: Math.max(1, Math.min(rows.value, Number(b.rowSpan) || 1)),
    colSpan: Math.max(1, Math.min(cols.value, Number(b.colSpan) || 1)),
  }
  // 越界 → 贴边；重叠 → 逐步回退到最近合法尺寸
  if (cand.row + cand.rowSpan > rows.value) cand.row = Math.max(0, rows.value - cand.rowSpan)
  if (cand.col + cand.colSpan > cols.value) cand.col = Math.max(0, cols.value - cand.colSpan)
  const conflicts = !isAreaFree(grid.value, cand)
  if (conflicts) {
    const fallback = stepResize(grid.value, { ...b }, 'se', cand.colSpan - b.colSpan, cand.rowSpan - b.rowSpan)
    b.rowSpan = fallback.rowSpan
    b.colSpan = fallback.colSpan
    ElMessage.warning('与其它区块重叠，已回退到最近合法的尺寸')
    return
  }
  b.row = cand.row
  b.col = cand.col
  b.rowSpan = cand.rowSpan
  b.colSpan = cand.colSpan
}

// ---------- 画布尺寸 / 模板 ----------

function applySize() {
  const c = Math.max(4, Math.min(16, Number(cols.value) || 6))
  const r = Math.max(4, Math.min(20, Number(rows.value) || 10))
  const kept: LayoutBlock[] = []
  for (const b of grid.value.blocks) {
    const cand: LayoutBlock = { ...b }
    cand.colSpan = Math.min(cand.colSpan, c)
    cand.rowSpan = Math.min(cand.rowSpan, r)
    cand.col = Math.min(cand.col, c - cand.colSpan)
    cand.row = Math.min(cand.row, r - cand.rowSpan)
    if (inBounds({ cols: c, rows: r } as LayoutGrid, cand) &&
      !kept.some((o) => blocksOverlap(cand, o))) {
      kept.push(cand)
    }
  }
  grid.value = { version: grid.value.version, cols: c, rows: r, blocks: kept }
  cols.value = c
  rows.value = r
  syncSelection()
}

function loadTemplate() {
  pushHistory()
  grid.value = defaultLayout(cols.value, rows.value)
  selectedId.value = null
  editingId.value = null
}

function clearAll() {
  pushHistory()
  grid.value = emptyLayout(cols.value, rows.value)
  selectedId.value = null
  editingId.value = null
}

async function clearCustom() {
  if (!props.building) return
  try {
    await ElMessageBox.confirm('清除后该楼栋将回到内置标准层模板渲染，是否继续？', '清除自定义布局')
  } catch {
    return
  }
  saving.value = true
  try {
    await apiUpdateBuilding(props.building.id, { ...props.building, layoutJson: '' })
    ElMessage.success('已清除自定义布局')
    emit('saved')
    emit('update:modelValue', false)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

// ---------- 键盘快捷键 ----------

function onKeyDown(ev: KeyboardEvent) {
  if (!props.modelValue) return
  const key = ev.key.toLowerCase()
  if ((ev.ctrlKey || ev.metaKey) && key === 'z') {
    ev.preventDefault()
    if (ev.shiftKey) redo()
    else undo()
    return
  }
  if (key === 'delete' || key === 'backspace') {
    if (selectedId.value) {
      ev.preventDefault()
      removeSelected()
    }
    return
  }
  if (key === 'escape') editingId.value = null
}

// ---------- 保存 ----------

async function save() {
  if (!props.building) return
  if (roomCount.value === 0) {
    ElMessage.warning('请至少绘制一间房间')
    return
  }
  saving.value = true
  try {
    await apiUpdateBuilding(props.building.id, {
      code: props.building.code,
      name: props.building.name,
      posX: props.building.posX,
      posY: props.building.posY,
      width: props.building.width,
      height: props.building.height,
      floors: props.building.floors,
      floorHeight: props.building.floorHeight,
      roomsPerFloor: roomCount.value,
      layoutJson: serializeLayout(grid.value),
    })
    ElMessage.success(`布局已保存：每层 ${roomCount.value} 间房`)
    emit('saved')
    emit('update:modelValue', false)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

// ---------- 打开时初始化 ----------

watch(
  () => [props.modelValue, props.building?.id] as const,
  async () => {
    if (!props.modelValue) return
    const parsed = parseLayout(props.building?.layoutJson)
    const base = parsed || defaultLayout()
    grid.value = { version: base.version, cols: base.cols, rows: base.rows, blocks: base.blocks.map((b) => ({ ...b })) }
    cols.value = base.cols
    rows.value = base.rows
    brush.value = null
    painting.value = false
    selectedId.value = null
    editingId.value = null
    undoStack.value = []
    redoStack.value = []
    await nextTick()
  },
  { immediate: true },
)

onMounted(() => {
  window.addEventListener('pointerup', endDrag)
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('pointerup', endDrag)
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('keydown', onKeyDown)
})
</script>

<style scoped>
.designer {
  display: grid;
  grid-template-columns: 215px minmax(0, 1fr) 215px;
  gap: 16px;
}

.tools,
.props {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px;
  background: var(--rv-grad-2);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 16px;
}

.tool-title {
  margin: 6px 0 2px;
  color: #5a6a85;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.tool-btn {
  display: grid;
  grid-template-columns: 14px 1fr;
  gap: 8px;
  align-items: center;
  padding: 8px 10px;
  color: #33415c;
  font-size: 13px;
  text-align: left;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.22s ease;
}

.tool-btn em {
  grid-column: 2;
  color: #94a3b8;
  font-size: 11px;
  font-style: normal;
}

.tool-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 18px rgba(46, 68, 112, 0.12);
}

.tool-btn.active {
  color: #2462d9;
  background: var(--rv-grad-1);
  box-shadow: 0 8px 18px rgba(52, 120, 246, 0.18);
}

.chip {
  width: 14px;
  height: 14px;
  border: 1px solid #9fb2d4;
  border-radius: 4px;
}

.size-row,
.prop-pair {
  display: flex;
  align-items: center;
  gap: 6px;
}

.times {
  color: #94a3b8;
}

.key-list {
  margin: 0;
  padding-left: 16px;
  color: #7c8aa3;
  font-size: 11px;
  line-height: 1.8;
}

.key-list b {
  color: #33415c;
}

.canvas-area {
  min-width: 0;
}

.hint-line {
  margin-bottom: 10px;
  color: #5a6a85;
  font-size: 12px;
}

.hint-line.warn {
  color: #b96b1c;
  font-weight: 600;
}
.hint-line.editing {
  padding: 6px 10px;
  color: #2462d9;
  background: rgba(52, 120, 246, 0.1);
  border-radius: 10px;
}

.canvas-scroll {
  max-height: 70vh;
  overflow: auto;
}

.canvas {
  position: relative;
  width: 100%;
  min-height: 320px;
  background: linear-gradient(135deg, #eef3fc, #e6ecfa);
  border: 1px solid #c8d5ea;
  border-radius: 14px;
  touch-action: none;
}

.slot-layer {
  display: grid;
  grid-template-columns: repeat(var(--cols), minmax(0, 1fr));
  grid-template-rows: repeat(var(--rows), minmax(0, 1fr));
  gap: 3px;
  padding: 8px;
  height: calc(var(--rows) * 46px);
}

.slot {
  background: rgba(255, 255, 255, 0.35);
  border: 1px dashed rgba(150, 170, 205, 0.5);
  border-radius: 6px;
  cursor: crosshair;
  transition: background 0.15s ease;
}

.canvas.no-brush .slot {
  cursor: default;
}

.slot:hover {
  background: rgba(255, 255, 255, 0.75);
}

.slot.painted {
  background: transparent;
  border-color: transparent;
}

.block-layer {
  position: absolute;
  inset: 8px;
  pointer-events: none;
}

.block {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  color: #3d5687;
  font-size: 12px;
  font-weight: 700;
  border-radius: 7px;
  pointer-events: auto;
  cursor: pointer;
  transition: box-shadow 0.18s ease, transform 0.18s ease;
}

.block-label {
  pointer-events: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 0 4px;
}

.k-room {
  background: linear-gradient(135deg, #dcebff, #c7dcf7);
  border: 1px solid #5f7bb5;
}

.k-corridor {
  background: linear-gradient(180deg, #f3f7fd, #e9f0fa);
  border: 1px dashed #a9b8d4;
}

.k-stair {
  background: linear-gradient(135deg, #e6ecf7, #d3dcec);
  border: 1px solid #7d8db3;
}

.k-public {
  background: linear-gradient(135deg, #eaf2fd, #dde8f8);
  border: 1px dashed #9aa9c6;
}

.k-custom {
  color: #6b3fa0;
  background: linear-gradient(135deg, #f2e8ff, #e4d6fb);
  border: 1px solid #a985d8;
}

.block:hover {
  box-shadow: 0 6px 14px rgba(46, 68, 112, 0.18);
}

.block.selected {
  outline: 2px solid #3478f6;
  outline-offset: 1px;
}

.block.editing {
  outline: 2px dashed #2462d9;
  outline-offset: 2px;
  z-index: 3;
}

.handle {
  position: absolute;
  width: 9px;
  height: 9px;
  background: #fff;
  border: 2px solid #3478f6;
  border-radius: 2px;
}

.h-n { top: -5px; left: 50%; margin-left: -5px; cursor: ns-resize; }
.h-s { bottom: -5px; left: 50%; margin-left: -5px; cursor: ns-resize; }
.h-e { right: -5px; top: 50%; margin-top: -5px; cursor: ew-resize; }
.h-w { left: -5px; top: 50%; margin-top: -5px; cursor: ew-resize; }
.h-nw { top: -5px; left: -5px; cursor: nwse-resize; }
.h-ne { top: -5px; right: -5px; cursor: nesw-resize; }
.h-sw { bottom: -5px; left: -5px; cursor: nesw-resize; }
.h-se { bottom: -5px; right: -5px; cursor: nwse-resize; }

.summary {
  display: flex;
  gap: 16px;
  margin-top: 12px;
  color: #5a6a85;
  font-size: 13px;
  flex-wrap: wrap;
}

.summary b {
  color: #2462d9;
}

.tips {
  margin: 8px 0 0;
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.6;
}

.prop-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 8px;
}

.prop-label {
  color: #5a6a85;
  font-size: 12px;
}

.prop-tip,
.prop-empty {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 11px;
  line-height: 1.6;
}

.footer-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.history-tip {
  color: #94a3b8;
  font-size: 12px;
}
</style>
