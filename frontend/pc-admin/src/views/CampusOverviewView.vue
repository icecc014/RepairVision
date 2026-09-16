<template>
  <AdminShell title="区域概览" subtitle="绘制校园 / 建筑群总平面图（建筑、道路、广场、校门），作为派单算法的空间数据基础">
    <section class="panel">
      <div class="toolbar">
        <el-input v-model="name" placeholder="概览图名称" style="width: 200px" />
        <el-button size="small" @click="loadTemplate">生成示例布局</el-button>
        <el-button size="small" @click="clearAll">清空画布</el-button>
        <el-button size="small" @click="reload">放弃修改</el-button>
        <div style="flex: 1" />
        <span class="toolbar-tip">画布 {{ cols }} × {{ rows }} 格</span>
        <el-input-number v-model="cols" :min="10" :max="140" size="small" style="width: 100px" />
        <el-input-number v-model="rows" :min="10" :max="80" size="small" style="width: 100px" />
        <el-button size="small" @click="applySize">应用画布尺寸</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存概览</el-button>
      </div>
    </section>

    <div class="layout">
      <section class="panel tools">
        <div class="panel-title">图元工具</div>
        <button
          v-for="t in CAMPUS_KINDS"
          :key="t.kind"
          class="tool-btn"
          :class="{ active: brush === t.kind }"
          @click="selectBrush(t.kind)"
        >
          <i class="chip" :style="{ background: t.color }" />
          <span>{{ t.label }}</span>
          <em>{{ t.hint }}</em>
        </button>
        <button class="tool-btn" :class="{ active: brush === 'erase' }" @click="selectBrush('erase')">
          <i class="chip" style="background: #fff" />
          <span>擦除</span>
          <em>点击图元即可删除</em>
        </button>

        <div class="panel-title">统计</div>
        <div class="stat-line">建筑 <b>{{ campusCount(grid, 'building') }}</b> 个</div>
        <div class="stat-line">道路 <b>{{ campusCount(grid, 'road') }}</b> 块</div>
        <div class="stat-line">广场绿地 <b>{{ campusCount(grid, 'green') }}</b> 块</div>
        <div class="stat-line">校门 <b>{{ campusCount(grid, 'gate') }}</b> 个</div>
        <div class="stat-line">自定义 <b>{{ campusCount(grid, 'custom') }}</b> 块</div>

        <div class="panel-title">快捷键</div>
        <ul class="key-list">
          <li><b>单击</b> 放置图元</li>
          <li><b>双击</b> 进入编辑模式</li>
          <li><b>拖把手</b> 按格改尺寸</li>
          <li><b>Ctrl+Z</b> 撤销 / <b>Delete</b> 删除</li>
          <li><b>Esc</b> 退出编辑</li>
          <li><b>+ / -</b> 鼠标在画布上时缩放</li>
          <li><b>F</b> 适应窗口 · <b>空格+拖动</b> 平移画布</li>
        </ul>
      </section>

      <section class="panel canvas-area">
        <div v-if="loading" class="hint-line">正在加载服务器上的区域概览…</div>
        <div class="hint-line" :class="{ warn: !brush || !!editingId }">
          {{ editingId ? '编辑模式：只能调整当前图元（拖把手改尺寸 / 拖本体移动）；按 Esc 或点击空白处退出'
            : (brush === 'erase' ? '擦除模式：点击图元即可删除'
              : (brush ? '已选择「' + brushLabel + '」：点击网格放置图元' : '请先在左侧选择图元工具，再点击网格开始绘制')) }}
        </div>
        <div class="canvas-tools">
          <span class="toolbar-tip">画布缩放</span>
          <el-button size="small" @click="zoomOut">－</el-button>
          <span class="zoom-text">{{ Math.round(zoom * 100) }}%</span>
          <el-button size="small" @click="zoomIn">＋</el-button>
          <el-button size="small" @click="zoomReset">重置</el-button>
          <el-button size="small" @click="fitToCanvas">适应窗口</el-button>
          <span class="toolbar-tip">快捷键：鼠标在画布上时按 + / - 缩放，F 适应窗口，空格+拖动平移（或长按空白处 0.3 秒）</span>
        </div>
        <div class="canvas-scroll" ref="scrollRef" @pointerenter="canvasHover = true" @pointerleave="canvasHover = false">
          <div class="canvas-stage" :style="stageStyle">
            <div
              ref="canvasRef"
              class="canvas"
              :style="{ '--cols': grid.cols, '--rows': grid.rows, width: baseW + 'px', transform: `scale(${zoom})`, transformOrigin: 'top left' }"
            >
            <div class="slot-layer">
              <button
                v-for="slot in slots"
                :key="slot.key"
                class="slot"
                :class="{ painted: !!slot.blockId }"
                @pointerdown.prevent="onSlotDown(slot, $event)"
              />
            </div>
            <div class="block-layer">
              <div
                v-for="b in grid.blocks"
                :key="b.id"
                class="block"
                draggable="false"
                :class="[`k-${b.kind}`, { selected: selectedId === b.id, editing: editingId === b.id }]"
                :style="blockStyle(b)"
                @pointerdown.stop.prevent="onBlockDown(b, $event)"
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
        </div>
      </section>

      <div class="side-col">
      <section class="panel props">
          <div class="panel-title">图元属性</div>
          <template v-if="selectedBlock">
            <div class="prop-row">
              <span class="prop-label">类型</span>
              <el-select :model-value="selectedBlock.kind" size="small" style="width: 100%" @change="setKind">
                <el-option v-for="t in CAMPUS_KINDS" :key="t.kind" :label="t.label" :value="t.kind" />
              </el-select>
            </div>
            <template v-if="selectedBlock.kind === 'building'">
              <div class="prop-row">
                <span class="prop-label">名称来源</span>
                <el-radio-group :model-value="selectedBlock.customBuilding ? 'custom' : 'building'" size="small" @change="setBuildingSource">
                  <el-radio-button value="building">已有楼栋</el-radio-button>
                  <el-radio-button value="custom">自定义</el-radio-button>
                </el-radio-group>
              </div>
              <div v-if="!selectedBlock.customBuilding" class="prop-row">
                <span class="prop-label">关联楼栋</span>
                <el-select :model-value="selectedBlock.buildingId" size="small" style="width: 100%" @change="setBuilding">
                <el-option v-if="buildings.length === 0" label="楼栋列表加载失败，请刷新页面重试" :value="0" disabled />
                  <el-option v-for="b in buildings" :key="b.id" :label="`${b.name}（${b.code}）`" :value="b.id" />
                </el-select>
              </div>
            </template>
            <div v-if="needLabel" class="prop-row">
              <span class="prop-label">名称</span>
              <el-input
                :model-value="selectedBlock.label"
                size="small"
                maxlength="16"
                placeholder="如：第二食堂 / 图书馆"
                @input="setLabelLive"
                @change="setLabel"
              />
            </div>
            <div class="prop-row">
              <span class="prop-label">操作</span>
            <div class="prop-pair">
              <el-button size="small" @click="editingId = editingId === selectedBlock.id ? null : selectedBlock.id">
                {{ editingId === selectedBlock.id ? '完成编辑' : '进入编辑' }}
              </el-button>
              <el-button size="small" type="danger" plain @click="removeSelected">删除图元</el-button>
            </div>
          </div>
          <div class="prop-row">
            <span class="prop-label">起始格（行 / 列）</span>
              <div class="prop-pair">
                <el-input-number :model-value="selectedBlock.row" :min="0" :max="grid.rows - 1" size="small" @change="(v: number) => setProp('row', v)" />
                <el-input-number :model-value="selectedBlock.col" :min="0" :max="grid.cols - 1" size="small" @change="(v: number) => setProp('col', v)" />
              </div>
            </div>
            <div class="prop-row">
              <span class="prop-label">跨格数（行 / 列）</span>
              <div class="prop-pair">
                <el-input-number :model-value="selectedBlock.rowSpan" :min="1" :max="grid.rows" size="small" @change="(v: number) => setProp('rowSpan', v)" />
                <el-input-number :model-value="selectedBlock.colSpan" :min="1" :max="grid.cols" size="small" @change="(v: number) => setProp('colSpan', v)" />
              </div>
            </div>
            <div class="prop-row">
              <span class="prop-label">左右边缘</span>
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
              <span class="prop-label">上下边缘</span>
              <div class="prop-pair">
                <el-button size="small" @click="expandEdge('n', true)">↑ 上扩</el-button>
                <el-button size="small" @click="expandEdge('n', false)">↓ 上收</el-button>
              </div>
              <div class="prop-pair">
                <el-button size="small" @click="expandEdge('s', true)">下扩 ↓</el-button>
                <el-button size="small" @click="expandEdge('s', false)">↑ 下收</el-button>
              </div>
            </div>
            <el-button size="small" type="danger" plain style="width: 100%" @click="removeSelected">删除该图元</el-button>
          </template>
          <p v-else class="prop-empty">单击图元查看属性；双击进入编辑模式（显示 8 个把手）。<br />建筑图元可关联"建筑信息管理"中的楼栋，或选择自定义后手动命名。</p>
        </section>
<section class="panel distances">
          <div class="panel-title">路网与距离</div>
          <p class="dist-tip">
            比例尺：1 格 = {{ distances.gridMeters }} 米 · 道路格 {{ distances.roadCells }} 个 ·
            最远建筑间距 {{ distances.maxMeters }} 米（派单距离即按路网最短路计算）
          </p>
          <el-table :data="distances.buildings" size="small" border max-height="300">
            <el-table-column prop="name" label="建筑" min-width="130" />
            <el-table-column label="接入路网" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.connected ? 'success' : 'info'" size="small">{{ row.connected ? '已接入' : '未接入' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="entryCount" label="入口格" width="90" align="center" />
            <el-table-column label="最远通勤" width="110" align="center">
              <template #default="{ row }">{{ row.connected ? row.maxMeters + ' 米' : '回退欧氏' }}</template>
            </el-table-column>
          </el-table>
          <p v-if="distances.pairs.length" class="dist-tip" style="margin-top: 10px">建筑间路网距离（米）</p>
          <el-table v-if="distances.pairs.length" :data="distances.pairs" size="small" border max-height="300">
            <el-table-column label="起点" min-width="120">
              <template #default="{ row }">{{ nameOfBuilding(row.fromId) }}</template>
            </el-table-column>
            <el-table-column label="终点" min-width="120">
              <template #default="{ row }">{{ nameOfBuilding(row.toId) }}</template>
            </el-table-column>
            <el-table-column label="路网距离" width="110" align="center">
              <template #default="{ row }">{{ row.meters }} 米</template>
            </el-table-column>
          </el-table>
        </section>
      </div>
    </div>
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminBuilding, CampusDistance } from '../api'
import {
  apiAdminBuildings,
  apiAdminCampusDistances,
  apiAdminCampusLayout,
  apiSaveCampusLayout,
} from '../api'
import AdminShell from '../components/AdminShell.vue'
import {
  CAMPUS_KINDS,
  CAMPUS_KIND_TEXT,
  campusAreaFree,
  campusCount,
  campusInBounds,
  campusStepMove,
  campusStepResize,
  emptyCampus,
  defaultCampus,
  makeCampusBlock,
  parseCampus,
  serializeCampus,
  type CampusBlock,
  type CampusGrid,
  type CampusKind,
} from '../utils/campusLayout'
import type { ResizeHandle } from '../utils/layoutGrid'

type Brush = CampusKind | 'erase'
const HANDLES: ResizeHandle[] = ['nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w']

const grid = ref<CampusGrid>(emptyCampus())
const cols = ref(100)
const rows = ref(40)
const name = ref('校园总览')
const brush = ref<Brush | null>(null)
const selectedId = ref<string | null>(null)
const editingId = ref<string | null>(null)
const canvasRef = ref<HTMLElement | null>(null)
const buildings = ref<AdminBuilding[]>([])
const saving = ref(false)
const loading = ref(false)
const undoStack = ref<CampusBlock[][]>([])
// ---------- 画布平移与缩放 ----------
const scrollRef = ref<HTMLElement | null>(null)
const zoom = ref(1)
const baseW = ref(900)
const baseH = ref(420)
const stageStyle = computed(() => ({ width: baseW.value * zoom.value + 'px', height: baseH.value * zoom.value + 'px' }))
const pan = reactive({ active: false, startX: 0, startY: 0, scrollLeft: 0, scrollTop: 0 })
const canvasHover = ref(false)
const spacePanReady = ref(false)
let panTimer: ReturnType<typeof setTimeout> | undefined

function measureCanvas() {
  const sc = scrollRef.value
  if (!sc) return
  baseW.value = Math.max(320, sc.clientWidth - 4)
  const el = canvasRef.value
  if (el) baseH.value = Math.max(320, el.offsetHeight || 420)
}

function zoomIn() { zoom.value = Math.min(4, Math.round((zoom.value + 0.2) * 10) / 10); setTimeout(measureCanvas, 0) }
function zoomOut() { zoom.value = Math.max(0.25, Math.round((zoom.value - 0.2) * 10) / 10); setTimeout(measureCanvas, 0) }
function zoomReset() { zoom.value = 1; setTimeout(measureCanvas, 0) }

// 适应窗口：让整张画布完整落在可视区域内（高度优先，宽度不超过 100%）
function fitToCanvas() {
  const sc = scrollRef.value
  if (!sc) return
  measureCanvas()
  const byH = sc.clientHeight > 0 ? sc.clientHeight / baseH.value : 1
  const byW = sc.clientWidth > 0 ? sc.clientWidth / baseW.value : 1
  const next = Math.max(0.25, Math.min(4, Math.min(byH, byW)))
  zoom.value = Math.round(next * 100) / 100
  setTimeout(() => {
    measureCanvas()
    if (scrollRef.value) { scrollRef.value.scrollLeft = 0; scrollRef.value.scrollTop = 0 }
  }, 0)
}

// 长按空白处 → 平移画布（拖动 scrollLeft/scrollTop）
function startPan(ev: PointerEvent) {
  const sc = scrollRef.value
  if (!sc) return
  pan.startX = ev.clientX
  pan.startY = ev.clientY
  pan.scrollLeft = sc.scrollLeft
  pan.scrollTop = sc.scrollTop
  if (panTimer) clearTimeout(panTimer)
  panTimer = setTimeout(() => {
    pan.active = true
    ElMessage.info('平移模式：拖动查看画布，松开结束')
  }, 300)
}

function onPanMove(ev: PointerEvent) {
  const sc = scrollRef.value
  if (!pan.active || !sc) return
  sc.scrollLeft = pan.scrollLeft - (ev.clientX - pan.startX)
  sc.scrollTop = pan.scrollTop - (ev.clientY - pan.startY)
}

function endPan() {
  if (panTimer) { clearTimeout(panTimer); panTimer = undefined }
  pan.active = false
}
const redoStack = ref<CampusBlock[][]>([])
const distances = ref<CampusDistance>({
  gridMeters: 10,
  maxMeters: 0,
  roadCells: 0,
  buildings: [],
  pairs: [],
})

async function loadDistances() {
  try {
    distances.value = await apiAdminCampusDistances()
  } catch {
    // 路网距离面板失败不阻塞设计器
  }
}

function nameOfBuilding(id: number) {
  return distances.value.buildings.find((b) => b.buildingId === id)?.name || `#${id}`
}

const selectedBlock = computed(() => grid.value.blocks.find((b) => b.id === selectedId.value) || null)
const brushLabel = computed(() => (brush.value === 'erase' ? '擦除' : CAMPUS_KIND_TEXT[brush.value as CampusKind] || ''))
const needLabel = computed(() => {
  const b = selectedBlock.value
  if (!b) return false
  // 建筑图元允许直接命名：有关联楼栋时同步楼栋名，未关联时即"自定义命名"
  return b.kind !== 'road'
})
const slots = computed(() => {
  const out: { key: string; row: number; col: number; blockId: string | null }[] = []
  for (let r = 0; r < grid.value.rows; r++) {
    for (let c = 0; c < grid.value.cols; c++) {
      const hit = grid.value.blocks.find((b) => r >= b.row && r < b.row + b.rowSpan && c >= b.col && c < b.col + b.colSpan)
      out.push({ key: `${r}-${c}`, row: r, col: c, blockId: hit?.id ?? null })
    }
  }
  return out
})

function cloneBlocks(): CampusBlock[] {
  return grid.value.blocks.map((b) => ({ ...b }))
}
function pushHistory() {
  undoStack.value.push(cloneBlocks())
  if (undoStack.value.length > 50) undoStack.value.shift()
  redoStack.value = []
}
function syncSelection() {
  if (selectedId.value && !grid.value.blocks.some((b) => b.id === selectedId.value)) selectedId.value = null
  if (editingId.value && !grid.value.blocks.some((b) => b.id === editingId.value)) editingId.value = null
}
function labelOf(b: CampusBlock) {
  // 关联了"建筑信息管理"里的楼栋时，名称跟随楼栋（改名后设计器/三端同步）
  if (b.kind === 'building' && b.buildingId) {
    const linked = buildings.value.find((x) => x.id === b.buildingId)
    if (linked) return linked.name
  }
  if (b.label) return b.label
  return CAMPUS_KIND_TEXT[b.kind]
}
function blockStyle(b: CampusBlock) {
  return {
    left: `${(b.col / grid.value.cols) * 100}%`,
    top: `${(b.row / grid.value.rows) * 100}%`,
    width: `${(b.colSpan / grid.value.cols) * 100}%`,
    height: `${(b.rowSpan / grid.value.rows) * 100}%`,
  }
}

// ---------- 绘制与编辑 ----------
function selectBrush(kind: Brush) {
  brush.value = brush.value === kind ? null : kind
  editingId.value = null
}
function onSlotDown(slot: { row: number; col: number }, ev?: PointerEvent) {
  // 未选工具（或鼠标中键）时：长按空白处平移画布
  if (ev && (spacePanReady.value || !brush.value || ev.button === 1)) {
    startPan(ev)
    return
  }
  if (editingId.value) {
    editingId.value = null
    selectedId.value = null
    return
  }
  if (!brush.value) {
    ElMessage.warning('请先在左侧选择图元工具，再点击网格放置')
    return
  }
  const hit = grid.value.blocks.find((b) => slot.row >= b.row && slot.row < b.row + b.rowSpan && slot.col >= b.col && slot.col < b.col + b.colSpan)
  if (brush.value === 'erase') {
    if (!hit) return
    pushHistory()
    grid.value.blocks = grid.value.blocks.filter((b) => b.id !== hit.id)
    syncSelection()
    return
  }
  if (hit) return
  pushHistory()
  const kind = brush.value
  const block = makeCampusBlock(kind, slot.row, slot.col, kind === 'building' ? 3 : 1, kind === 'building' ? 4 : 1)
  if (kind === 'building') block.label = '新建筑'
  if (kind !== 'building' && kind !== 'road') block.label = CAMPUS_KIND_TEXT[kind]
  grid.value.blocks.push(block)
  selectedId.value = block.id
}
function onBlockDown(b: CampusBlock, ev: PointerEvent) {
  // 空格 + 拖动 / 鼠标中键：优先平移画布
  if (spacePanReady.value || ev.button === 1) {
    startPan(ev)
    return
  }
  if (editingId.value) {
    if (editingId.value !== b.id) {
      editingId.value = b.id
      selectedId.value = b.id
      return
    }
    beginDrag('move', b, 'se', ev)
    return
  }
  if (brush.value === 'erase') {
    pushHistory()
    grid.value.blocks = grid.value.blocks.filter((x) => x.id !== b.id)
    syncSelection()
    return
  }
  selectedId.value = b.id
  startLongPress(b, ev)
}
function enterEdit(b: CampusBlock) {
  selectedId.value = b.id
  editingId.value = editingId.value === b.id ? null : b.id
}
function removeSelected() {
  if (!selectedId.value) return
  pushHistory()
  grid.value.blocks = grid.value.blocks.filter((b) => b.id !== selectedId.value)
  syncSelection()
}

// ---------- 属性编辑 ----------
function setKind(kind: CampusKind) {
  const b = selectedBlock.value
  if (!b || b.kind === kind) return
  pushHistory()
  b.kind = kind
  if (kind !== 'building') {
    delete b.buildingId
    delete b.customBuilding
  }
  if (!b.label) b.label = CAMPUS_KIND_TEXT[kind]
}
function setBuildingSource(source: string) {
  const b = selectedBlock.value
  if (!b) return
  pushHistory()
  if (source === 'custom') {
    b.customBuilding = true
    delete b.buildingId
    if (!b.label || b.label === '新建筑') b.label = ''
  } else {
    delete b.customBuilding
    b.buildingId = buildings.value[0]?.id
    b.label = buildings.value[0]?.name || b.label
  }
}
function setBuilding(id: number) {
  const b = selectedBlock.value
  if (!b) return
  pushHistory()
  b.buildingId = id
  b.label = buildings.value.find((x) => x.id === id)?.name || b.label
}
function setLabelLive(value: string) {
  const b = selectedBlock.value
  if (!b) return
  const next = String(value || '').slice(0, 16)
  if ((b.label || '') === next) return
  b.label = next
  // 建筑图元手动改名即视为"自定义命名"，避免仍被当作待关联楼栋
  if (b.kind === 'building' && !b.buildingId) b.customBuilding = true
}

function setLabel(value: string) {
  const b = selectedBlock.value
  if (!b) return
  const next = String(value || '').trim().slice(0, 16)
  if ((b.label || '') === next) return
  pushHistory()
  b.label = next
  if (b.kind === 'building' && !b.buildingId) b.customBuilding = true
}
function setProp(field: 'row' | 'col' | 'rowSpan' | 'colSpan', value: number) {
  const b = selectedBlock.value
  if (!b) return
  const before = { ...b }
  pushHistory()
  b[field] = Number(value) || 0
  const cand: CampusBlock = {
    ...b,
    rowSpan: Math.max(1, Math.min(grid.value.rows, b.rowSpan)),
    colSpan: Math.max(1, Math.min(grid.value.cols, b.colSpan)),
  }
  cand.row = Math.max(0, Math.min(grid.value.rows - cand.rowSpan, b.row))
  cand.col = Math.max(0, Math.min(grid.value.cols - cand.colSpan, b.col))
  if (!campusInBounds(grid.value, cand) || !campusAreaFree(grid.value, cand)) {
    const fallback = campusStepResize(grid.value, before, 'se', cand.colSpan - before.colSpan, cand.rowSpan - before.rowSpan)
    b.rowSpan = fallback.rowSpan
    b.colSpan = fallback.colSpan
    ElMessage.warning('与其它图元重叠或越界，已回退到最近合法值')
  } else {
    b.row = cand.row
    b.col = cand.col
    b.rowSpan = cand.rowSpan
    b.colSpan = cand.colSpan
  }
  if (b.row === before.row && b.col === before.col && b.rowSpan === before.rowSpan && b.colSpan === before.colSpan) {
    undoStack.value.pop()
  }
}
function expandEdge(dir: 'w' | 'e' | 'n' | 's', grow: boolean) {
  const b = selectedBlock.value
  if (!b) return
  const dx = dir === 'w' ? (grow ? -1 : 1) : dir === 'e' ? (grow ? 1 : -1) : 0
  const dy = dir === 'n' ? (grow ? -1 : 1) : dir === 's' ? (grow ? 1 : -1) : 0
  pushHistory()
  const next = campusStepResize(grid.value, b, dir, dx, dy)
  const changed = next.row !== b.row || next.col !== b.col || next.rowSpan !== b.rowSpan || next.colSpan !== b.colSpan
  if (!changed) {
    undoStack.value.pop()
    ElMessage.warning(grow ? '该方向已到边界或被其它图元挡住' : '该方向已缩到最小 1 格')
    return
  }
  b.row = next.row
  b.col = next.col
  b.rowSpan = next.rowSpan
  b.colSpan = next.colSpan
}

// ---------- 拖拽 ----------
const drag = reactive({
  mode: '' as '' | 'resize' | 'move',
  id: '',
  handle: 'se' as ResizeHandle,
  base: null as CampusBlock | null,
  startX: 0,
  startY: 0,
  cellW: 1,
  cellH: 1,
})
function beginDrag(mode: 'resize' | 'move', b: CampusBlock, handle: ResizeHandle, ev: PointerEvent) {
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
function onHandleDown(b: CampusBlock, handle: ResizeHandle, ev: PointerEvent) {
  beginDrag('resize', b, handle, ev)
}
function onPointerMove(ev: PointerEvent) {
  if (!drag.mode || !drag.base) return
  const dx = Math.round((ev.clientX - drag.startX) / drag.cellW)
  const dy = Math.round((ev.clientY - drag.startY) / drag.cellH)
  const target = grid.value.blocks.find((b) => b.id === drag.id)
  if (!target) return
  const next = drag.mode === 'resize'
    ? campusStepResize(grid.value, drag.base, drag.handle, dx, dy)
    : campusStepMove(grid.value, drag.base, dx, dy)
  target.row = next.row
  target.col = next.col
  target.rowSpan = next.rowSpan
  target.colSpan = next.colSpan
}
function endDrag() {
  if (!drag.mode) return
  const target = grid.value.blocks.find((b) => b.id === drag.id)
  const base = drag.base
  if (target && base && target.row === base.row && target.col === base.col &&
    target.rowSpan === base.rowSpan && target.colSpan === base.colSpan) {
    undoStack.value.pop()
  }
  drag.mode = ''
  drag.base = null
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
function onKeyUp(ev: KeyboardEvent) {
  if (ev.code === 'Space') spacePanReady.value = false
}
function onKeyDown(ev: KeyboardEvent) {
  const key = ev.key.toLowerCase()
  // 仅在鼠标位于画布区域时响应，且不占用 Ctrl+滚轮（避免与浏览器缩放冲突）
  if (canvasHover.value) {
    if (ev.key === '+' || ev.key === '=' || key === 'add') {
      ev.preventDefault()
      zoomIn()
      return
    }
    if (ev.key === '-' || ev.key === '_' || key === 'subtract') {
      ev.preventDefault()
      zoomOut()
      return
    }
    if (key === 'f') {
      ev.preventDefault()
      fitToCanvas()
      return
    }
    if (ev.code === 'Space') {
      ev.preventDefault()
      spacePanReady.value = true
      return
    }
  }
  if ((ev.ctrlKey || ev.metaKey) && key === 'z') {
    ev.preventDefault()
    if (ev.shiftKey) redo()
    else undo()
    return
  }
  if ((key === 'delete' || key === 'backspace') && selectedId.value) {
    const tag = (ev.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA') return
    ev.preventDefault()
    removeSelected()
    return
  }
  if (key === 'escape') editingId.value = null
}

// ---------- 画布操作与保存 ----------
function applySize() {
  const c = Math.max(10, Math.min(140, Number(cols.value) || 40))
  const r = Math.max(10, Math.min(80, Number(rows.value) || 30))
  const kept: CampusBlock[] = []
  for (const b of grid.value.blocks) {
    const cand: CampusBlock = { ...b }
    cand.colSpan = Math.min(cand.colSpan, c)
    cand.rowSpan = Math.min(cand.rowSpan, r)
    cand.col = Math.min(cand.col, c - cand.colSpan)
    cand.row = Math.min(cand.row, r - cand.rowSpan)
    if (!kept.some((o) => o.row < cand.row + cand.rowSpan && cand.row < o.row + o.rowSpan &&
      o.col < cand.col + cand.colSpan && cand.col < o.col + o.colSpan)) {
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
  grid.value = defaultCampus(cols.value, rows.value)
  selectedId.value = null
  editingId.value = null
}
function clearAll() {
  pushHistory()
  grid.value = emptyCampus(cols.value, rows.value)
  selectedId.value = null
  editingId.value = null
}
async function load() {
  loading.value = true
  try {
    let data = await apiAdminCampusLayout()
    if (!data.layoutJson || data.layoutJson.length < 10) {
      // 自愈：极少数情况下首次返回为空布局（浏览器/网络侧旧响应），自动重取一次
      console.warn('[campus] 首次返回空布局，300ms 后自动重取', data)
      await new Promise((r) => setTimeout(r, 300))
      data = await apiAdminCampusLayout()
      console.info('[campus] 重取后 layoutJson 长度 =', (data.layoutJson || '').length)
    }
    name.value = data.name || '校园总览'
    cols.value = data.cols || 40
    rows.value = data.rows || 30
    let parsed = parseCampus(data.layoutJson)
    if (!parsed && data.layoutJson && data.layoutJson.length > 2) {
      // 兜底：若图元相互重叠导致严格解析失败，改用宽松解析，避免整页空白
      try {
        const raw = JSON.parse(data.layoutJson) as { cols?: number; rows?: number; blocks?: CampusBlock[] }
        if (raw && Array.isArray(raw.blocks)) {
          parsed = {
            version: 1,
            cols: Number(raw.cols) || cols.value,
            rows: Number(raw.rows) || rows.value,
            blocks: raw.blocks.filter((b) => b && b.kind),
          }
          ElMessage.warning('概览中存在重叠图元，已按宽松模式加载（建议修正后重新保存）')
        }
      } catch {
        // ignore
      }
    }
    grid.value = parsed || emptyCampus(cols.value, rows.value)
    selectedId.value = null
    editingId.value = null
    undoStack.value = []
    redoStack.value = []
    console.info("[campus] loaded blocks =", grid.value.blocks.length, "cols/rows =", grid.value.cols, grid.value.rows)
  } catch (err) {
    ElMessage.error("加载区域概览失败：" + (err as Error).message)
  } finally {
    loading.value = false
  }
}
function reload() {
  load()
  ElMessage.info('已重新加载服务器上的概览数据')
}
async function save() {
  if (campusCount(grid.value, 'building') === 0) {
    try {
      await ElMessageBox.confirm('当前还没有建筑图元，仍要保存吗？', '保存区域概览')
    } catch {
      return
    }
  }
  saving.value = true
  try {
    await apiSaveCampusLayout({
      name: name.value || '校园总览',
      cols: grid.value.cols,
      rows: grid.value.rows,
      layoutJson: serializeCampus(grid.value),
    })
    ElMessage.success('区域概览已保存，宿管端与工人端可查看')
    loadDistances()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}
onMounted(async () => {
  window.addEventListener('pointerup', endDrag)
  window.addEventListener('pointermove', onPressMove)
  window.addEventListener('pointerup', finishPressDrag)
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('keydown', onKeyDown)
  window.addEventListener('pointermove', onPanMove)
  window.addEventListener('pointerup', endPan)
  window.addEventListener('resize', measureCanvas)
  window.addEventListener('keyup', onKeyUp)
  try {
    buildings.value = await apiAdminBuildings()
  } catch {
    // 楼栋列表失败不影响绘制
  }
  // 先等组件完全挂载（el-input-number 等会写回 cols/rows），再加载服务器数据，
  // 避免"挂载过程中的控件事件把刚载入的图元覆盖掉"导致首屏空白。
  await nextTick()
  await load()
  measureCanvas()
  loadDistances()
})
onUnmounted(() => {
  window.removeEventListener('pointerup', endDrag)
  window.removeEventListener('pointermove', onPressMove)
  window.removeEventListener('pointerup', finishPressDrag)
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('keydown', onKeyDown)
  window.removeEventListener('pointermove', onPanMove)
  window.removeEventListener('pointerup', endPan)
  window.removeEventListener('resize', measureCanvas)
  window.removeEventListener('keyup', onKeyUp)
})
// ---------- #7 长按拖动（只改位置、不改尺寸；松手时若重叠则回退） ----------
const pressDrag = reactive({
  timer: 0,
  active: false,
  id: '',
  base: null as CampusBlock | null,
  startX: 0,
  startY: 0,
  cellW: 1,
  cellH: 1,
})

function startLongPress(b: CampusBlock, ev: PointerEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  if (!rect.width || !rect.height) return
  pressDrag.cellW = rect.width / grid.value.cols
  pressDrag.cellH = rect.height / grid.value.rows
  pressDrag.startX = ev.clientX
  pressDrag.startY = ev.clientY
  pressDrag.id = b.id
  pressDrag.base = { ...b }
  cancelLongPress()
  pressDrag.timer = window.setTimeout(() => {
    pressDrag.active = true
    pushHistory()
    ElMessage.info('拖动模式：移动到目标位置后松开鼠标')
  }, 300)
}

function cancelLongPress() {
  if (pressDrag.timer) {
    clearTimeout(pressDrag.timer)
    pressDrag.timer = 0
  }
}

function onPressMove(ev: PointerEvent) {
  if (!pressDrag.active || !pressDrag.base) return
  const target = grid.value.blocks.find((x) => x.id === pressDrag.id)
  const base = pressDrag.base
  if (!target || !base) return
  const dx = Math.round((ev.clientX - pressDrag.startX) / pressDrag.cellW)
  const dy = Math.round((ev.clientY - pressDrag.startY) / pressDrag.cellH)
  // 只改位置、不改尺寸；拖动过程允许暂时重叠，松手时统一校验
  target.row = Math.max(0, Math.min(grid.value.rows - target.rowSpan, base.row + dy))
  target.col = Math.max(0, Math.min(grid.value.cols - target.colSpan, base.col + dx))
}

function overlapsOthers(b: CampusBlock) {
  return grid.value.blocks.some((o) => o.id !== b.id &&
    b.row < o.row + o.rowSpan && o.row < b.row + b.rowSpan &&
    b.col < o.col + o.colSpan && o.col < b.col + b.colSpan)
}

function finishPressDrag() {
  cancelLongPress()
  if (!pressDrag.active) return
  pressDrag.active = false
  const target = grid.value.blocks.find((x) => x.id === pressDrag.id)
  const base = pressDrag.base
  pressDrag.base = null
  if (!target || !base) return
  if (target.row === base.row && target.col === base.col) {
    undoStack.value.pop() // 没移动：丢弃这次撤销点
    return
  }
  if (overlapsOthers(target)) {
    target.row = base.row
    target.col = base.col
    undoStack.value.pop()
    ElMessage.warning('该位置与其它图元重叠，已还原到拖动前的位置')
  }
}
</script>

<style scoped>
.distances {
  margin-top: 0;
}
.dist-tip {
  margin: 0 0 10px;
  color: #5a6a85;
  font-size: 12px;
  line-height: 1.6;
}
.panel {
  padding: 16px 18px;
  margin-bottom: 16px;
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 10px 30px rgba(46, 68, 112, 0.08);
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.toolbar-tip {
  color: #8a97ad;
  font-size: 12px;
}
.layout {
  align-items: start;
  display: grid;
  grid-template-columns: 210px minmax(0, 1fr) 330px;
  gap: 16px;
}
.panel-title {
  margin: 6px 0 8px;
  color: #2b3445;
  font-size: 13px;
  font-weight: 800;
}
.tool-btn {
  display: grid;
  grid-template-columns: 14px 1fr;
  gap: 8px;
  align-items: center;
  width: 100%;
  padding: 8px 10px;
  margin-bottom: 6px;
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
.stat-line {
  color: #5a6a85;
  font-size: 12px;
  line-height: 1.9;
}
.stat-line b {
  color: #2462d9;
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
.canvas-scroll {
  max-height: 74vh;
  overflow: auto;
}
.canvas {
  user-select: none;
  -webkit-user-drag: none;
  position: relative;
  width: 100%;
  min-height: 420px;
  background: linear-gradient(135deg, #f2f6fb, #e8eef8);
  border: 1px solid #c8d5ea;
  border-radius: 14px;
  touch-action: none;
}
.slot-layer {
  display: grid;
  grid-template-columns: repeat(var(--cols), minmax(0, 1fr));
  grid-template-rows: repeat(var(--rows), minmax(0, 1fr));
  gap: 1px;
  padding: 8px;
  height: calc(var(--rows) * 20px);
}
.slot {
  background: rgba(255, 255, 255, 0.45);
  border: 1px dashed rgba(150, 170, 205, 0.35);
  border-radius: 3px;
  cursor: crosshair;
}
.slot:hover {
  background: rgba(255, 255, 255, 0.85);
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
  color: #33415c;
  font-size: 11px;
  font-weight: 700;
  border-radius: 6px;
  pointer-events: auto;
  cursor: pointer;
}
.block-label {
  pointer-events: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 0 4px;
}
.k-building { background: linear-gradient(135deg, #dcebff, #c7dcf7); border: 1px solid #5f7bb5; }
.k-road { background: linear-gradient(135deg, #efe8da, #e2d8c4); border: 1px dashed #b3a68c; }
.k-green { background: linear-gradient(135deg, #ddf0e3, #c8e6d2); border: 1px solid #7fb894; }
.k-gate { background: linear-gradient(135deg, #f7e3ef, #eccfdf); border: 1px solid #c084a5; }
.k-custom { background: linear-gradient(135deg, #f2e8ff, #e4d6fb); border: 1px solid #a985d8; }
.block.selected { outline: 2px solid #3478f6; outline-offset: 1px; }
.block.editing { outline: 2px dashed #2462d9; outline-offset: 2px; z-index: 3; }
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
.side-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  max-height: 80vh;
  overflow: auto;
}
.props {
  min-width: 0;
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
.prop-pair {
  display: flex;
  align-items: center;
  gap: 6px;
}
.prop-empty {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 11px;
  line-height: 1.7;
}
.canvas-tools {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.zoom-text {
  min-width: 46px;
  color: #2b3445;
  font-size: 12px;
  font-weight: 700;
  text-align: center;
}
.canvas-stage {
  position: relative;
}
.canvas-stage .canvas {
  position: absolute;
  top: 0;
  left: 0;
}
</style>
