<template>
  <div class="campus-wrap">
    <div class="campus-zoom">
      <button class="campus-retry" @click="zoomOut">－ 缩小</button>
      <span class="campus-scale">{{ Math.round(zoom * 100) }}%</span>
      <button class="campus-retry" @click="zoomIn">＋ 放大</button>
      <button class="campus-retry" @click="zoomReset">重置</button>
      <span v-if="focusedLabel" class="campus-focus">已定位：{{ focusedLabel }}</span>
    </div>
    <div v-if="!layout" class="campus-empty">
      <span v-if="loading">正在加载区域概览…</span>
      <span v-else-if="errorMsg">{{ '加载失败：' + errorMsg }}</span>
      <span v-else>是否加载区域示意图？</span>
      <button class="campus-retry" @click="load">{{ loading ? '加载中…' : '加载' }}</button>
    </div>
    <template v-else>
      <div
        ref="viewportRef"
        class="campus-viewport"
        @wheel.prevent="onWheel"
        @touchstart="onTouchStart"
        @touchmove="onTouchMove"
        @touchend="onTouchEnd"
      >
        <div class="campus-stage" :style="stageStyle">
          <svg class="campus-svg" :width="svgW" :height="svgH" :viewBox="`0 0 ${layout.cols} ${layout.rows}`" preserveAspectRatio="none">
            <defs>
              <linearGradient id="campusBg" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#f2f6fb" />
                <stop offset="100%" stop-color="#e8eef8" />
              </linearGradient>
            </defs>
            <rect x="0.2" y="0.2" :width="layout.cols - 0.4" :height="layout.rows - 0.4" fill="url(#campusBg)" stroke="#c7d3e8" stroke-width="0.25" rx="0.8" />
            <g v-for="b in layout.blocks" :key="b.id">
              <rect
                :x="b.col + 0.12"
                :y="b.row + 0.12"
                :width="Math.max(0.2, b.colSpan - 0.24)"
                :height="Math.max(0.2, b.rowSpan - 0.24)"
                :fill="b.color || fillOf(b.kind)"
                :stroke="isFocused(b) || isHighlight(b) ? '#dc2626' : strokeOf(b.kind)"
                :stroke-width="isFocused(b) || isHighlight(b) ? 0.75 : 0.22"
                :class="{ 'campus-pulse': isFocused(b) || isHighlight(b) }"
                rx="0.35"
                @click="tapBlock(b)"
              />
              <g v-if="isFocused(b) || isHighlight(b)">
                <circle
                  :cx="b.col + b.colSpan / 2"
                  :cy="b.row + b.rowSpan / 2"
                  :r="Math.max(0.8, Math.min(b.colSpan, b.rowSpan) * 0.48)"
                  fill="none"
                  stroke="#ef4444"
                  stroke-width="0.35"
                  class="target-radar-ring"
                />
                <circle
                  :cx="b.col + b.colSpan / 2"
                  :cy="b.row + b.rowSpan / 2"
                  r="0.3"
                  fill="#ef4444"
                />
              </g>
              <text
                v-if="labelPlan(b)"
                :x="b.col + b.colSpan / 2"
                :y="labelBaseY(b)"
                text-anchor="middle"
                :font-size="labelPlan(b)!.size"
                fill="#33415c"
                class="campus-label"
              >
                <tspan
                  v-for="(line, li) in labelPlan(b)!.lines"
                  :key="li"
                  :x="b.col + b.colSpan / 2"
                  :dy="li === 0 ? 0 : labelPlan(b)!.size * 1.18"
                >{{ line }}</tspan>
              </text>
            </g>
          </svg>
        </div>
      </div>
      <div v-if="active" ref="cardRef" class="campus-card">
        <div class="card-head">
          <span class="card-title">{{ activeTitle }}</span>
          <button class="card-close" @click="active = null">✕</button>
        </div>
        <div class="card-row"><span class="k">类型</span><span class="v">{{ kindText(active.kind) }}</span></div>
        <div v-if="activeInfo" class="card-row"><span class="k">编码</span><span class="v">{{ activeInfo.code || '—' }}</span></div>
        <div v-if="activeInfo" class="card-row"><span class="k">楼层</span><span class="v">{{ activeInfo.floors }} 层 · 每层 {{ activeInfo.roomsPerFloor }} 间</span></div>
        <div v-if="activeInfo" class="card-row"><span class="k">待处理</span><span class="v">{{ activeCount }} 单</span></div>
        <div v-if="!activeInfo" class="card-tip">自定义设施，未关联建筑信息</div>
      </div>
      <div class="campus-legend">
        <span><i class="lg building" />建筑</span>
        <span><i class="lg road" />道路</span>
        <span><i class="lg green" />广场/绿地</span>
        <span><i class="lg gate" />校门</span>
        <span v-if="buildingCount > 0">共 {{ buildingCount }} 处建筑</span>
      </div>
      <div class="campus-tip">缩放时高亮宿舍自动锁定于正中；双指或按钮缩放；点建筑可查看待处理单数</div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { apiCampusMap } from '../api'

interface CampusBlock {
  id: string
  kind: string
  row: number
  col: number
  rowSpan: number
  colSpan: number
  label?: string
  buildingId?: number
  color?: string
}

const props = defineProps<{
  buildings?: { id: number; name: string; code?: string; floors?: number; roomsPerFloor?: number }[]
  counts?: Record<number, number>
  highlightBuildingId?: number
  autoLoad?: boolean
}>()

const loading = ref(false)
const errorMsg = ref('')
const started = ref(false)
const zoom = ref(1)
const layout = ref<{ cols: number; rows: number; blocks: CampusBlock[] } | null>(null)
const viewportRef = ref<HTMLElement | null>(null)
const unitPx = ref(3)
const svgW = computed(() => (layout.value ? layout.value.cols * unitPx.value * zoom.value : 0))
const svgH = computed(() => (layout.value ? layout.value.rows * unitPx.value * zoom.value : 0))
const active = ref<CampusBlock | null>(null)
const cardRef = ref<HTMLElement | null>(null)
const focusedBuildingId = ref<number | null>(null)

// 动态计算视口 padding，使视口内任意角落的宿舍都能完全置中
const padX = computed(() => {
  const vp = viewportRef.value
  return vp ? Math.max(30, Math.floor(vp.clientWidth / 2)) : 180
})
const padY = computed(() => {
  const vp = viewportRef.value
  return vp ? Math.max(30, Math.floor(vp.clientHeight / 2)) : 150
})
const stageStyle = computed(() => ({
  padding: `${padY.value}px ${padX.value}px`,
  display: 'inline-block',
  minWidth: '100%',
  boxSizing: 'content-box' as const,
}))

const activeInfo = computed(() => {
  const id = active.value?.buildingId
  if (!id) return null
  return props.buildings?.find((b) => b.id === id) || null
})
const activeTitle = computed(() => {
  if (!active.value) return ''
  if (activeInfo.value) return activeInfo.value.name
  return active.value.label || KIND_TEXT[active.value.kind] || '图元'
})
const activeCount = computed(() => {
  const id = active.value?.buildingId
  if (!id) return 0
  return props.counts?.[id] ?? 0
})
const focusedLabel = computed(() => {
  const id = focusedBuildingId.value || props.highlightBuildingId
  if (!id) return ''
  return props.buildings?.find((b) => b.id === id)?.name || ''
})

// 核心聚焦目标查找
function getTargetBlock(): CampusBlock | null {
  if (!layout.value) return null
  if (focusedBuildingId.value) {
    const hit = layout.value.blocks.find((b) => b.buildingId === focusedBuildingId.value)
    if (hit) return hit
  }
  if (props.highlightBuildingId) {
    const hit = layout.value.blocks.find((b) => b.buildingId === props.highlightBuildingId)
    if (hit) return hit
  }
  if (active.value) return active.value
  return null
}

// 保证目标建筑在任何缩放级别下，都绝对精准位于视口正中央
function scrollBlockIntoView(target: CampusBlock, smooth = true) {
  const vp = viewportRef.value
  if (!vp) return
  const px = unitPx.value * zoom.value
  const cx = padX.value + (target.col + target.colSpan / 2) * px
  const cy = padY.value + (target.row + target.rowSpan / 2) * px
  vp.scrollTo({
    left: Math.max(0, cx - vp.clientWidth / 2),
    top: Math.max(0, cy - vp.clientHeight / 2),
    behavior: smooth ? 'smooth' : 'auto',
  })
}

function scrollCenter(smooth = false) {
  const vp = viewportRef.value
  if (!vp || !layout.value) return
  const px = unitPx.value * zoom.value
  const cx = padX.value + (layout.value.cols / 2) * px
  const cy = padY.value + (layout.value.rows / 2) * px
  vp.scrollTo({
    left: Math.max(0, cx - vp.clientWidth / 2),
    top: Math.max(0, cy - vp.clientHeight / 2),
    behavior: smooth ? 'smooth' : 'auto',
  })
}

function applyZoom(nextZoom: number) {
  const target = getTargetBlock()
  zoom.value = Math.max(0.6, Math.min(4, Math.round(nextZoom * 100) / 100))
  nextTick(() => {
    measure()
    nextTick(() => {
      if (target) {
        scrollBlockIntoView(target, false)
      } else {
        scrollCenter(false)
      }
    })
  })
}

function zoomIn() {
  applyZoom(zoom.value + 0.3)
}

function zoomOut() {
  applyZoom(zoom.value - 0.3)
}

function zoomReset() {
  zoom.value = 1
  const target = getTargetBlock()
  nextTick(() => {
    measure()
    nextTick(() => {
      if (target) {
        scrollBlockIntoView(target, false)
      } else {
        scrollCenter(false)
      }
    })
  })
}

function onWheel(e: WheelEvent) {
  const delta = e.deltaY > 0 ? -0.25 : 0.25
  applyZoom(zoom.value + delta)
}

let touchStartDist = 0
let touchStartZoom = 1

function onTouchStart(e: TouchEvent) {
  if (e.touches.length === 2) {
    const t1 = e.touches[0]
    const t2 = e.touches[1]
    touchStartDist = Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY) || 1
    touchStartZoom = zoom.value
  }
}

function onTouchMove(e: TouchEvent) {
  if (e.touches.length === 2) {
    e.preventDefault()
    const t1 = e.touches[0]
    const t2 = e.touches[1]
    const dist = Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY) || 1
    const nextZoom = touchStartZoom * (dist / touchStartDist)
    applyZoom(nextZoom)
  }
}

function onTouchEnd(e: TouchEvent) {
  if (e.touches.length < 2) {
    touchStartDist = 0
  }
}

watch(
  () => props.highlightBuildingId,
  (newId) => {
    if (newId) {
      focusBuilding(newId, false)
    }
  },
)

const KIND_TEXT: Record<string, string> = {
  building: '建筑',
  road: '道路',
  green: '广场',
  gate: '校门',
  custom: '设施',
}

const buildingCount = computed(() => (layout.value ? layout.value.blocks.filter((b) => b.kind === 'building').length : 0))

function fillOf(kind: string) {
  switch (kind) {
    case 'building':
      return '#dcebff'
    case 'road':
      return '#efe8da'
    case 'green':
      return '#ddf0e3'
    case 'gate':
      return '#f7e3ef'
    default:
      return '#f2e8ff'
  }
}

function strokeOf(kind: string) {
  switch (kind) {
    case 'building':
      return '#5f7bb5'
    case 'road':
      return '#c3b598'
    case 'green':
      return '#7fb894'
    case 'gate':
      return '#c084a5'
    default:
      return '#a985c8'
  }
}

function kindText(kind: string) {
  return KIND_TEXT[kind] || kind
}
function isHighlight(b: CampusBlock) {
  return !!props.highlightBuildingId && b.buildingId === props.highlightBuildingId
}
function isFocused(b: CampusBlock) {
  return !!focusedBuildingId.value && b.buildingId === focusedBuildingId.value
}
const MIN_LABEL_SIZE = 0.5

function labelCandidates(text: string): string[] {
  const chars = Array.from(text)
  const list = [text]
  for (const n of [4, 3, 2, 1]) {
    if (chars.length > n) list.push(chars.slice(0, n).join(''))
  }
  return list
}

function fitLabel(text: string, b: CampusBlock) {
  const chars = Array.from(text)
  const w = Math.max(0.8, b.colSpan - 0.32)
  const h = Math.max(0.55, b.rowSpan - 0.32)
  for (let lines = 1; lines <= 3; lines++) {
    const perLine = Math.ceil(chars.length / lines)
    const size = Math.min(w / (perLine * 1.02), h / (lines * 1.2), 2.4)
    if (size >= MIN_LABEL_SIZE) {
      const out: string[] = []
      for (let i = 0; i < chars.length; i += perLine) out.push(chars.slice(i, i + perLine).join(''))
      return { lines: out, size }
    }
  }
  return null
}

function labelPlan(b: CampusBlock) {
  if (b.kind === 'road') return null
  const text = displayLabel(b)
  if (!text) return null
  const chars = Array.from(text)
  const primary = [text]
  for (const n of [4, 3]) if (chars.length > n) primary.push(chars.slice(0, n).join(''))
  let best: { lines: string[]; size: number } | null = null
  for (const candidate of primary) {
    const plan = fitLabel(candidate, b)
    if (!plan) continue
    if (!best || plan.size >= best.size * 1.2) best = plan
  }
  if (best) return best
  for (const n of [2, 1]) {
    if (chars.length <= n) continue
    const plan = fitLabel(chars.slice(0, n).join(''), b)
    if (plan) return plan
  }
  return null
}
function labelBaseY(b: CampusBlock) {
  const plan = labelPlan(b)
  const cy = b.row + b.rowSpan / 2
  if (!plan) return cy
  return cy + plan.size * 0.35 - ((plan.lines.length - 1) * plan.size * 1.18) / 2
}
function tapBlock(b: CampusBlock) {
  active.value = b
  focusedBuildingId.value = b.buildingId ?? null
}
function measure() {
  const vp = viewportRef.value
  if (!vp || !layout.value) return
  const width = vp.clientWidth || 320
  unitPx.value = Math.max(1.5, width / layout.value.cols)
}

function focusBuilding(buildingId: number, openCard = true): boolean {
  if (!layout.value) {
    started.value = true
    void load().then(() => focusBuilding(buildingId, openCard))
    return true
  }
  const target = layout.value?.blocks.find((b) => b.buildingId === buildingId)
  if (!target) return false
  zoom.value = Math.min(4, Math.max(zoom.value, 2.2))
  focusedBuildingId.value = buildingId
  if (openCard) active.value = target
  nextTick(() => {
    measure()
    nextTick(() => scrollBlockIntoView(target, true))
  })
  return true
}

function displayLabel(b: CampusBlock) {
  if (b.buildingId) {
    const hit = props.buildings?.find((x) => x.id === b.buildingId)
    if (hit) return hit.name
  }
  return b.label || KIND_TEXT[b.kind] || ''
}

async function load() {
  loading.value = true
  try {
    let data = await apiCampusMap()
    if (!data.layoutJson || data.layoutJson.length < 10) {
      console.warn('[campus] 首次返回空布局，300ms 后自动重取', data)
      await new Promise((r) => setTimeout(r, 300))
      data = await apiCampusMap()
    }
    const raw = data.layoutJson ? JSON.parse(data.layoutJson) : null
    if (raw && Array.isArray(raw.blocks)) {
      layout.value = {
        cols: Number(raw.cols) || 40,
        rows: Number(raw.rows) || 30,
        blocks: raw.blocks.filter((b: CampusBlock) => b && b.kind),
      }
      errorMsg.value = ''
      await nextTick()
      measure()
      const target = getTargetBlock()
      if (target) {
        scrollBlockIntoView(target, false)
      }
    } else {
      layout.value = null
    }
  } catch (err) {
    layout.value = null
    errorMsg.value = (err as Error).message || '加载失败'
  } finally {
    loading.value = false
  }
}

let resizeTimer = 0
function onResize() {
  if (resizeTimer) window.clearTimeout(resizeTimer)
  resizeTimer = window.setTimeout(() => measure(), 150)
}

function reload() {
  errorMsg.value = ''
  if (props.autoLoad !== false) load()
}

onMounted(() => {
  if (props.autoLoad === true) {
    started.value = true
    load()
  }
  window.addEventListener('rv-campus-refresh', reload)
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('rv-campus-refresh', reload)
  window.removeEventListener('resize', onResize)
})

defineExpose({ load, focusBuilding, zoomIn, zoomOut, zoomReset })
</script>

<style scoped>
.campus-wrap {
  padding: 4px 0;
}
.campus-viewport {
  position: relative;
  overflow: auto;
  max-height: 46vh;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-x pan-y;
}
.campus-stage {
  display: inline-block;
  min-width: 100%;
  box-sizing: content-box;
}
.campus-svg {
  display: block;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05);
}
.campus-empty {
  padding: 18px 10px;
  color: var(--rv-text-light, #94a3b8);
  font-size: 13px;
  text-align: center;
}
.campus-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
  color: var(--rv-text-sub, #5a6a85);
  font-size: 11px;
}
.campus-legend .lg {
  display: inline-block;
  width: 10px;
  height: 10px;
  margin-right: 4px;
  border-radius: 2px;
  vertical-align: -1px;
}
.lg.building { background: #dcebff; border: 1px solid #5f7bb5; }
.lg.road { background: #efe8da; border: 1px solid #c3b598; }
.lg.green { background: #ddf0e3; border: 1px solid #7fb894; }
.lg.gate { background: #f7e3ef; border: 1px solid #c084a5; }
.campus-retry {
  margin-left: 8px;
  padding: 2px 10px;
  border: 1px solid #c8d5ea;
  border-radius: 999px;
  background: #fff;
  color: #2462d9;
  font-size: 12px;
}
.campus-label {
  pointer-events: none;
  font-weight: 600;
  paint-order: stroke;
  stroke: #ffffff;
  stroke-width: 0.14;
  stroke-linejoin: round;
}
.campus-pulse {
  animation: campus-pulse 1.1s ease-in-out infinite;
}
@keyframes campus-pulse {
  0%, 100% { opacity: 1; stroke-width: 0.8; }
  50% { opacity: 0.45; stroke-width: 1.1; }
}
.target-radar-ring {
  animation: radar-expand 1.4s cubic-bezier(0.25, 1, 0.5, 1) infinite;
  transform-origin: center;
  pointer-events: none;
}
@keyframes radar-expand {
  0% {
    r: 0.4;
    opacity: 1;
    stroke-width: 0.5;
  }
  100% {
    r: 2.2;
    opacity: 0;
    stroke-width: 0.1;
  }
}
.campus-card {
  margin-top: 8px;
  background: #ffffff;
  border: 1px solid #dbe4f0;
  border-radius: 10px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.1);
  padding: 8px 10px;
  font-size: 12px;
  color: #334155;
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}
.card-title {
  font-weight: 700;
  color: #1f2a3d;
}
.card-close {
  border: none;
  background: transparent;
  color: #94a3b8;
  font-size: 13px;
  padding: 0 2px;
}
.card-row {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  line-height: 1.7;
}
.card-row .k {
  color: #94a3b8;
}
.card-row .v {
  font-weight: 600;
}
.card-tip {
  margin-top: 4px;
  color: #94a3b8;
  font-size: 11px;
}
.campus-focus {
  color: #dc2626;
  font-size: 11px;
  font-weight: 600;
}
.campus-tip {
  margin-top: 6px;
  font-size: 11px;
  color: #94a3b8;
}
.campus-zoom {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  margin-bottom: 6px;
  font-size: 12px;
  color: #64748b;
}
</style>
