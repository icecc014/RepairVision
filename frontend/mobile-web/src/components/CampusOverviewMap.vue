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
      <div ref="viewportRef" class="campus-viewport" @scroll="onScroll">
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
            :stroke="isFocused(b) ? '#dc2626' : (isHighlight(b) ? '#2563eb' : strokeOf(b.kind))"
            :stroke-width="isFocused(b) ? 0.7 : (isHighlight(b) ? 0.5 : 0.22)"
            :class="{ 'campus-pulse': isFocused(b) }"
            rx="0.35"
            @click="tapBlock(b, $event)"
          />
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
      <div v-if="active" class="campus-card" :style="cardStyle">
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
      </div>
      <div class="campus-legend">
        <span><i class="lg building" />建筑</span>
        <span><i class="lg road" />道路</span>
        <span><i class="lg green" />广场/绿地</span>
        <span><i class="lg gate" />校门</span>
        <span v-if="buildingCount > 0">共 {{ buildingCount }} 处建筑</span>
      </div>
      <div class="campus-tip">用上方按钮缩放、拖动查看其他区域；点建筑可看名称/编码/楼层/待处理单数</div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
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
const activeAt = ref({ x: 0, y: 0 })
const focusedBuildingId = ref<number | null>(null)
const scroll = ref({ left: 0, top: 0 })
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
  const id = focusedBuildingId.value
  if (!id) return ''
  return props.buildings?.find((b) => b.id === id)?.name || ''
})
const cardStyle = computed(() => {
  const vp = viewportRef.value
  const width = 176
  const maxLeft = vp ? Math.max(4, vp.clientWidth - width - 6) : 4
  const anchorX = activeAt.value.x - scroll.value.left + 6
  const anchorY = activeAt.value.y - scroll.value.top + 6
  return { left: Math.min(Math.max(4, anchorX), maxLeft) + 'px', top: Math.max(4, anchorY) + 'px', width: width + 'px' }
})
function zoomIn() { zoom.value = Math.min(4, Math.round((zoom.value + 0.25) * 100) / 100) }
function zoomOut() { zoom.value = Math.max(1, Math.round((zoom.value - 0.25) * 100) / 100) }
function zoomReset() {
  zoom.value = 1
  focusedBuildingId.value = null
  active.value = null
  nextTick(() => {
    const vp = viewportRef.value
    if (vp) {
      vp.scrollLeft = 0
      vp.scrollTop = 0
      scroll.value = { left: 0, top: 0 }
    }
  })
}

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
// 字体自适应：在建筑框内试 1~3 行，能放下就返回字号与分行；
// 放不下就不画字（点建筑看详情卡），保证文字永远不溢出图形。
function labelPlan(b: CampusBlock) {
  if (b.kind === 'road') return null
  const text = displayLabel(b)
  if (!text) return null
  const chars = Array.from(text)
  const w = Math.max(0.8, b.colSpan - 0.32)
  const h = Math.max(0.6, b.rowSpan - 0.32)
  for (let lines = 1; lines <= 3; lines++) {
    const perLine = Math.ceil(chars.length / lines)
    const size = Math.min(w / (perLine * 1.02), h / (lines * 1.2), 2.4)
    if (size >= 0.78) {
      const out: string[] = []
      for (let i = 0; i < chars.length; i += perLine) out.push(chars.slice(i, i + perLine).join(''))
      return { lines: out, size }
    }
  }
  return null
}
function labelBaseY(b: CampusBlock) {
  const plan = labelPlan(b)
  const cy = b.row + b.rowSpan / 2
  if (!plan) return cy
  return cy + plan.size * 0.35 - ((plan.lines.length - 1) * plan.size * 1.18) / 2
}
function onScroll() {
  const vp = viewportRef.value
  if (!vp) return
  scroll.value = { left: vp.scrollLeft, top: vp.scrollTop }
}
function tapBlock(b: CampusBlock, ev: MouseEvent) {
  const vp = viewportRef.value
  if (!vp) return
  const rect = vp.getBoundingClientRect()
  active.value = b
  focusedBuildingId.value = b.buildingId ?? null
  activeAt.value = { x: ev.clientX - rect.left, y: ev.clientY - rect.top }
}
function measure() {
  const vp = viewportRef.value
  if (!vp || !layout.value) return
  const width = vp.clientWidth || 320
  unitPx.value = Math.max(1.5, width / layout.value.cols)
}
// 供外部（下方楼栋工单列表）调用：缩放并把该建筑平滑滚到视口中央
function focusBuilding(buildingId: number, openCard = true): boolean {
  // 还没加载过概览：先自动加载再定位（点楼栋列表时不用先手动点"加载"）
  if (!layout.value) {
    started.value = true
    void load().then(() => focusBuilding(buildingId, openCard))
    return true
  }
  const target = layout.value?.blocks.find((b) => b.buildingId === buildingId)
  if (!target) return false
  zoom.value = Math.min(4, Math.max(zoom.value, 2.5))
  focusedBuildingId.value = buildingId
  if (openCard) active.value = target
  nextTick(() => {
    const vp = viewportRef.value
    if (!vp) return
    measure()
    // 再等一帧：SVG 用新的宽高渲染完成后，容器的可滚动范围才是对的，否则滚动会被夹到 0
    nextTick(() => {
      const px = unitPx.value * zoom.value
      const cx = (target.col + target.colSpan / 2) * px
      const cy = (target.row + target.rowSpan / 2) * px
      vp.scrollTo({ left: Math.max(0, cx - vp.clientWidth / 2), top: Math.max(0, cy - vp.clientHeight / 2), behavior: 'smooth' })
      activeAt.value = { x: Math.min(vp.clientWidth - 16, cx), y: Math.max(8, cy - vp.clientHeight / 2 + 14) }
      window.setTimeout(() => {
        scroll.value = { left: vp.scrollLeft, top: vp.scrollTop }
      }, 420)
    })
  })
  return true
}

// 建筑名称以"建筑信息管理"中的最新名称优先，其次用概览内保存的名称
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
      // 自愈：首次返回空布局时自动重取一次
      console.warn('[campus] 首次返回空布局，300ms 后自动重取', data)
      await new Promise((r) => setTimeout(r, 300))
      data = await apiCampusMap()
      console.info('[campus] 重取后 layoutJson 长度 =', (data.layoutJson || '').length)
    }
    const raw = data.layoutJson ? JSON.parse(data.layoutJson) : null
    if (raw && Array.isArray(raw.blocks)) {  // 有图元就渲染；blocks 为空时也渲染空白画布，避免误报"未绘制"
      layout.value = {
        cols: Number(raw.cols) || 40,
        rows: Number(raw.rows) || 30,
        blocks: raw.blocks.filter((b: CampusBlock) => b && b.kind),
      }
      errorMsg.value = ''
      await nextTick()
      measure()
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
function startLoad() {
  started.value = true
  load()
}

function reload() {
  errorMsg.value = ''
  if (props.autoLoad !== false) load()
}

onMounted(() => {
  // 默认不自动加载：工人端先显示“是否加载区域示意图？”，点击后加载
  if (props.autoLoad === true) {
    started.value = true
    load()
  }
  // 管理端保存区域概览后会通过 WebSocket 广播，就地刷新
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
.campus-svg {
  border-radius: 12px;
  background: #fff;
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
.campus-viewport {
  position: relative;
  overflow: auto;
  max-height: 46vh;
  border-radius: 12px;
  background: #fff;
  -webkit-overflow-scrolling: touch;
}
.campus-svg {
  display: block;
}
.campus-label {
  pointer-events: none;
  font-weight: 600;
}
.campus-pulse {
  animation: campus-pulse 1.1s ease-in-out infinite;
}
@keyframes campus-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}
.campus-card {
  position: absolute;
  z-index: 5;
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid #dbe4f0;
  border-radius: 10px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.16);
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
