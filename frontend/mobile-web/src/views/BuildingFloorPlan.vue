<template>
  <div class="plan-wrap">
    <div class="plan-top">
      <div class="floor-tabs">
        <button
          v-for="f in building.floors"
          :key="f"
          class="floor-tab"
          :class="{
            active: activeFloor === f,
            'has-fault': getFloorStatus(f) === 'fault',
            'has-done': getFloorStatus(f) === 'done'
          }"
          @click="switchFloor(f)"
        >
          {{ f }}F
          <i v-if="getFloorStatus(f) === 'fault'" class="tab-dot fault"></i>
          <i v-else-if="getFloorStatus(f) === 'done'" class="tab-dot done"></i>
        </button>
      </div>
      <div class="zoom-bar">
        <button class="zoom-btn" @click="zoomAt(0.8)">＋</button>
        <button class="zoom-btn" @click="zoomAt(1.25)">－</button>
        <button class="zoom-btn wide" @click="resetView">复位</button>
      </div>
    </div>

    <div class="legend-line">
      <span><i class="dot-lg red"></i> 待处理故障</span>
      <span><i class="dot-lg green"></i> 已修好房间</span>
      <span><i class="dot-lg blue"></i> 普通房间</span>
      <span>拖动平移 · 滚轮/双指缩放</span>
    </div>

    <div class="plan-stage">
      <div
        ref="viewportRef"
        class="plan-viewport"
        @wheel.prevent="onWheel"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
        @pointerleave="onPointerUp"
      >
        <svg class="plan-svg" :viewBox="viewBoxStr" preserveAspectRatio="xMidYMid meet" @click="onSvgClick">
          <defs>
            <linearGradient id="gPlanBg" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#f8fafc" />
              <stop offset="100%" stop-color="#f1f5f9" />
            </linearGradient>
            <linearGradient id="gRoom" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#e0f2fe" />
              <stop offset="100%" stop-color="#bae6fd" />
            </linearGradient>
            <linearGradient id="gRoomFault" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#ffe4e6" />
              <stop offset="100%" stop-color="#fecdd3" />
            </linearGradient>
            <linearGradient id="gRoomDone" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#dcfce7" />
              <stop offset="100%" stop-color="#bbf7d0" />
            </linearGradient>
            <linearGradient id="gCorridor" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#ffffff" />
              <stop offset="100%" stop-color="#f8fafc" />
            </linearGradient>
            <linearGradient id="gStair" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#eef2ff" />
              <stop offset="100%" stop-color="#e0e7ff" />
            </linearGradient>
            <linearGradient id="gPublic" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#f1f5f9" />
              <stop offset="100%" stop-color="#e2e8f0" />
            </linearGradient>
          </defs>
          <rect x="1" y="1" :width="PLAN_WIDTH - 2" :height="PLAN_DEPTH - 2" fill="url(#gPlanBg)" stroke="#cbd5e1" stroke-width="1.2" rx="2" />
          <template v-if="!plan.custom">
            <rect :x="plan.corridor.x" :y="plan.corridor.z" :width="plan.corridor.w" :height="plan.corridor.d" fill="url(#gCorridor)" stroke="#cbd5e1" stroke-width="0.6" stroke-dasharray="3 2" />
            <text :x="plan.corridor.x + plan.corridor.w / 2" :y="PLAN_DEPTH / 2" text-anchor="middle" font-size="3.6" fill="#94a3b8" transform="rotate(90, 50, 88)">贯通走廊</text>

            <g v-for="core in plan.cores" :key="core.index">
              <rect x="0" :y="core.z" width="32" :height="core.d" fill="url(#gStair)" stroke="#a5b4fc" stroke-width="0.7" />
              <text x="16" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.2" fill="#4338ca">封闭防火楼梯</text>
              <rect x="32" :y="core.z" width="68" :height="core.d" fill="url(#gPublic)" stroke="#cbd5e1" stroke-width="0.7" stroke-dasharray="2 1.6" />
              <text x="66" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.4" fill="#64748b">公共区域</text>
            </g>
          </template>
          <g v-else>
            <template v-for="(b, bi) in plan.blocks" :key="'blk' + bi">
              <rect :x="b.x" :y="b.z" :width="b.w" :height="b.d" :fill="blockFill(b.type)" :stroke="b.type === 'corridor' ? '#cbd5e1' : '#94a3b8'" stroke-width="0.6" :stroke-dasharray="b.type === 'corridor' ? '3 2' : '0'" />
              <text v-if="b.type !== 'corridor'" :x="b.x + b.w / 2" :y="b.z + b.d / 2 + 1.1" text-anchor="middle" font-size="2.6" fill="#64748b">{{ b.label || (b.type === 'stair' ? '楼梯' : '公共区') }}</text>
            </template>
          </g>

          <g v-for="room in plan.rooms" :key="room.no">
            <rect
              :x="room.x"
              :y="room.z"
              :width="room.w"
              :height="room.d"
              :fill="getRoomFill(room.no)"
              :stroke="getRoomStroke(room.no)"
              :stroke-width="selectedRoomNo === room.no ? 1.8 : 0.8"
              rx="1.5"
              @pointerdown.stop
              @click.stop="select(room)"
              style="cursor: pointer"
            />
            <text
              :x="room.x + room.w / 2"
              :y="room.z + room.d / 2 + 1.4"
              text-anchor="middle"
              font-size="4.2"
              font-weight="bold"
              :fill="getRoomTextColor(room.no)"
              pointer-events="none"
            >
              {{ room.no }}
            </text>
            <!-- 待修故障警示红点 -->
            <circle
              v-if="getRoomAnalysis(room.no).status === 'fault'"
              :cx="room.x + room.w - 4.5"
              :cy="room.z + 4.5"
              r="3.2"
              class="fault-dot"
              fill="#ef4444"
              stroke="#ffffff"
              stroke-width="0.8"
              @pointerdown.stop
              @click.stop="select(room)"
              style="cursor: pointer"
            />
            <!-- 已修好绿色圆点 -->
            <circle
              v-else-if="getRoomAnalysis(room.no).status === 'done'"
              :cx="room.x + room.w - 4.5"
              :cy="room.z + 4.5"
              r="2.6"
              fill="#10b981"
              stroke="#ffffff"
              stroke-width="0.8"
              @pointerdown.stop
              @click.stop="select(room)"
              style="cursor: pointer"
            />
          </g>
        </svg>
      </div>

      <aside v-if="activeFault" class="fault-card" @pointerdown.stop @wheel.stop>
        <div class="fault-head">
          <div class="fault-title-group">
            <span class="fault-room">{{ activeFault.no }}</span>
            <span class="fault-floor">{{ activeFault.floor }} 层</span>
            <span v-if="activeFault.status === 'fault'" class="card-status-badge fault">待维修 ({{ activeFault.unfinished.length }})</span>
            <span v-else-if="activeFault.status === 'done'" class="card-status-badge done">已完工 ({{ activeFault.completed.length }})</span>
            <span v-else class="card-status-badge normal">状态正常</span>
          </div>
          <button class="fault-close" @click="closeFaultCard">✕</button>
        </div>

        <div v-if="activeFault.orders.length === 0" class="fault-empty">
          当前房间无报修工单记录
        </div>
        <div v-for="o in activeFault.orders" :key="o.id" class="fault-item">
          <div class="fault-row">
            <span class="fault-type">{{ o.faultTypeName || '报修' }}</span>
            <span class="fault-status" :class="'fs' + o.status">{{ o.statusText }}</span>
          </div>
          <div class="fault-title">{{ o.title }}</div>
          <div class="fault-label">故障描述</div>
          <div class="fault-text">{{ o.description || '无补充说明' }}</div>
          <div class="fault-label">处理建议</div>
          <div class="fault-text">{{ explain(o).advice }}</div>
          <div class="fault-meta">工单: {{ o.orderNo }} · {{ o.createdAt }}</div>
          <div class="fault-actions">
            <button v-if="o.status === 1 || o.status === 2" class="fault-btn" :disabled="actingId === o.id" @click="act(o, 'start')">
              {{ actingId === o.id ? '处理中…' : '开始维修' }}
            </button>
            <button v-if="o.status === 3" class="fault-btn done" :disabled="actingId === o.id" @click="act(o, 'complete')">
              {{ actingId === o.id ? '处理中…' : '确认完工' }}
            </button>
            <span v-if="o.status === 4" class="badge-done-text">✓ 已完工</span>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { OrderItem, WorkerMapBuilding } from '../api'
import { apiCompleteOrder } from '../api'
import { startOrderFlow } from '../utils/startOrderFlow'
import { PLAN_DEPTH, PLAN_WIDTH, matchRoomOrders, type PlanRoom } from '../utils/floorLayout'
import { resolveFloorPlan } from '../utils/layoutGrid'

function blockFill(type: 'corridor' | 'stair' | 'public') {
  if (type === 'stair') return '#eef2ff'
  if (type === 'public') return '#f8fafc'
  return '#ffffff'
}

const props = defineProps<{ building: WorkerMapBuilding; orders: OrderItem[] }>()
const emit = defineEmits<{
  (e: 'selectRoom', room: { num: string; floor: number; orders: OrderItem[] }): void
  (e: 'refresh'): void
}>()

const activeFloor = ref(1)
const viewportRef = ref<HTMLDivElement | null>(null)
const selectedRoomNo = ref<string | null>(null)

interface ActiveFaultState {
  no: string
  floor: number
  status: 'fault' | 'done' | 'normal'
  orders: OrderItem[]
  unfinished: OrderItem[]
  completed: OrderItem[]
}
const activeFault = ref<ActiveFaultState | null>(null)
const actingId = ref<number | null>(null)

const MIN_W = 24
const MAX_W = PLAN_WIDTH * 1.5
const view = reactive({ x: 0, y: 0, w: PLAN_WIDTH, h: PLAN_DEPTH })
const viewBoxStr = computed(() => `${view.x} ${view.y} ${view.w} ${view.h}`)

const plan = computed(() =>
  resolveFloorPlan(activeFloor.value, props.building.roomsPerFloor, props.building.layoutJson),
)

function getRoomAnalysis(roomNo: string) {
  const bOrders = (props.orders || []).filter((o) => o.buildingId === props.building.id)
  const matched = matchRoomOrders(bOrders, activeFloor.value, roomNo)
  const unfinished = matched.filter((o) => o.status === 1 || o.status === 2 || o.status === 3)
  const completed = matched.filter((o) => o.status === 4)
  let status: 'fault' | 'done' | 'normal' = 'normal'
  if (unfinished.length > 0) {
    status = 'fault'
  } else if (completed.length > 0) {
    status = 'done'
  }
  return { unfinished, completed, status, all: matched }
}

function getRoomFill(roomNo: string) {
  const status = getRoomAnalysis(roomNo).status
  if (status === 'fault') return 'url(#gRoomFault)'
  if (status === 'done') return 'url(#gRoomDone)'
  return 'url(#gRoom)'
}

function getRoomStroke(roomNo: string) {
  if (selectedRoomNo.value === roomNo) return '#2563eb'
  const status = getRoomAnalysis(roomNo).status
  if (status === 'fault') return '#f43f5e'
  if (status === 'done') return '#10b981'
  return '#93c5fd'
}

function getRoomTextColor(roomNo: string) {
  const status = getRoomAnalysis(roomNo).status
  if (status === 'fault') return '#be123c'
  if (status === 'done') return '#047857'
  return '#1e40af'
}

function getFloorStatus(f: number): 'fault' | 'done' | 'normal' {
  const bOrders = (props.orders || []).filter(
    (o) => o.buildingId === props.building.id && (o.floor || 1) === f,
  )
  if (bOrders.some((o) => o.status === 1 || o.status === 2 || o.status === 3)) {
    return 'fault'
  }
  if (bOrders.some((o) => o.status === 4)) {
    return 'done'
  }
  return 'normal'
}

function switchFloor(f: number) {
  activeFloor.value = f
  selectedRoomNo.value = null
  activeFault.value = null
  resetView()
}

function select(room: PlanRoom) {
  selectedRoomNo.value = room.no
  const analysis = getRoomAnalysis(room.no)
  activeFault.value = {
    no: room.no,
    floor: activeFloor.value,
    status: analysis.status,
    orders: analysis.all,
    unfinished: analysis.unfinished,
    completed: analysis.completed,
  }
  emit('selectRoom', { num: room.no, floor: activeFloor.value, orders: analysis.all })
}

function closeFaultCard() {
  activeFault.value = null
  selectedRoomNo.value = null
}

function explain(o: OrderItem): { cause: string; advice: string } {
  switch (o.faultType) {
    case 'electric':
      return {
        cause: '线路接触不良、开关插座损坏或负荷跳闸。',
        advice: '先断开该房间总闸再检修，避免带电操作。',
      }
    case 'water':
      return {
        cause: '管道接头渗漏、角阀老化或下水堵塞。',
        advice: '先关闭角阀/进水阀并清理积水，避免渗漏扩大。',
      }
    default:
      return {
        cause: '设施损坏或需现场排查的具体故障。',
        advice: '请按报修描述携带对应工具上门排查。',
      }
  }
}

async function act(o: OrderItem, action: 'start' | 'complete') {
  try {
    await showConfirmDialog({
      title: action === 'start' ? '确认开工' : '确认完工',
      message: `${action === 'start' ? '开始维修' : '完成'} ${o.title}（${o.orderNo}）？`,
    })
  } catch {
    return
  }
  actingId.value = o.id
  try {
    if (action === 'start') {
      if (!(await startOrderFlow(o.id))) return
      showToast('已开工')
    } else {
      await apiCompleteOrder(o.id)
      showToast('已完工')
    }
    emit('refresh')
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    actingId.value = null
  }
}

function onSvgClick(e: MouseEvent) {
  if (moved) return
  const p = svgPoint(e)
  for (const room of plan.value.rooms) {
    if (p.x >= room.x && p.x <= room.x + room.w && p.y >= room.z && p.y <= room.z + room.d) {
      select(room)
      return
    }
  }
  closeFaultCard()
}

function clampView() {
  view.w = Math.min(Math.max(view.w, MIN_W), MAX_W)
  view.h = (view.w * PLAN_DEPTH) / PLAN_WIDTH
  const padX = view.w * 0.35
  const padY = view.h * 0.35
  view.x = Math.min(Math.max(view.x, -padX), PLAN_WIDTH - view.w + padX)
  view.y = Math.min(Math.max(view.y, -padY), PLAN_DEPTH - view.h + padY)
}

function zoomAt(factor: number, cx = PLAN_WIDTH / 2, cy = PLAN_DEPTH / 2) {
  const oldW = view.w
  const nextW = Math.min(Math.max(oldW * factor, MIN_W), MAX_W)
  const ratio = nextW / oldW
  view.x = cx - (cx - view.x) * ratio
  view.y = cy - (cy - view.y) * ratio
  view.w = nextW
  view.h = (nextW * PLAN_DEPTH) / PLAN_WIDTH
  clampView()
}

function resetView() {
  view.x = 0
  view.y = 0
  view.w = PLAN_WIDTH
  view.h = PLAN_DEPTH
}

function svgPoint(e: PointerEvent | WheelEvent) {
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  const boxAspect = rect.width / Math.max(rect.height, 1)
  const viewAspect = view.w / Math.max(view.h, 1)
  let scale = 1
  let offsetX = 0
  let offsetY = 0
  if (viewAspect > boxAspect) {
    scale = rect.width / view.w
    offsetY = (rect.height - view.h * scale) / 2
  } else {
    scale = rect.height / view.h
    offsetX = (rect.width - view.w * scale) / 2
  }
  return {
    x: view.x + (e.clientX - rect.left - offsetX) / scale,
    y: view.y + (e.clientY - rect.top - offsetY) / scale,
  }
}

function onWheel(e: WheelEvent) {
  const p = svgPoint(e)
  zoomAt(e.deltaY > 0 ? 1.12 : 0.89, p.x, p.y)
}

const pointers = new Map<number, { x: number; y: number }>()
let dragStart: { x: number; y: number; vx: number; vy: number } | null = null
let moved = false
let pinchStartDist = 0
let pinchStartW = PLAN_WIDTH

function onPointerDown(e: PointerEvent) {
  if (pointers.size > 2) pointers.clear()
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  moved = false
  if (pointers.size === 1) {
    dragStart = { x: e.clientX, y: e.clientY, vx: view.x, vy: view.y }
  } else if (pointers.size === 2) {
    const [a, b] = [...pointers.values()]
    pinchStartDist = Math.hypot(a.x - b.x, a.y - b.y) || 1
    pinchStartW = view.w
    dragStart = null
  }
}

function onPointerMove(e: PointerEvent) {
  if (!pointers.has(e.pointerId)) return
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  if (pointers.size === 1 && dragStart) {
    if (Math.abs(e.clientX - dragStart.x) + Math.abs(e.clientY - dragStart.y) > 6) moved = true
    const dx = ((e.clientX - dragStart.x) / rect.width) * view.w
    const dy = ((e.clientY - dragStart.y) / rect.height) * view.h
    view.x = dragStart.vx - dx
    view.y = dragStart.vy - dy
    clampView()
    return
  }
  if (pointers.size === 2) {
    const [a, b] = [...pointers.values()]
    const dist = Math.hypot(a.x - b.x, a.y - b.y) || 1
    const nextW = Math.min(Math.max((pinchStartW * pinchStartDist) / dist, MIN_W), MAX_W)
    const centre = svgPoint(e)
    zoomAt(nextW / view.w, centre.x, centre.y)
  }
}

function onPointerUp(e: PointerEvent) {
  pointers.delete(e.pointerId)
  if (pointers.size === 0) dragStart = null
}
</script>

<style scoped>
.plan-wrap { padding: 4px 10px 10px; }
.plan-top { display: flex; align-items: center; gap: 8px; }
.floor-tabs { display: flex; gap: 6px; flex: 1; overflow-x: auto; }
.floor-tab {
  position: relative;
  flex: 0 0 auto;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.floor-tab.active { color: #fff; background: #2563eb; border-color: #2563eb; }
.floor-tab.has-fault { border-color: #fecdd3; color: #e11d48; }
.floor-tab.has-fault.active { background: #e11d48; border-color: #e11d48; color: #fff; }
.floor-tab.has-done { border-color: #a7f3d0; color: #059669; }
.floor-tab.has-done.active { background: #059669; border-color: #059669; color: #fff; }
.tab-dot { width: 6px; height: 6px; border-radius: 50%; }
.tab-dot.fault { background: #e11d48; }
.tab-dot.done { background: #10b981; }

.zoom-bar { display: flex; gap: 6px; }
.zoom-btn { width: 32px; height: 30px; font-size: 15px; font-weight: 700; color: #1d4ed8; background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 8px; cursor: pointer; }
.zoom-btn.wide { width: auto; padding: 0 10px; font-size: 12px; }

.legend-line { display: flex; gap: 12px; margin: 6px 0; color: #64748b; font-size: 11px; flex-wrap: wrap; align-items: center; }
.dot-lg { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 4px; }
.dot-lg.red { background: #ef4444; }
.dot-lg.green { background: #10b981; }
.dot-lg.blue { background: #93c5fd; border: 1px solid #3b82f6; }

.plan-stage { position: relative; }
.plan-viewport { height: calc(82vh - 200px); min-height: 260px; overflow: hidden; touch-action: none; cursor: grab; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; }
.plan-svg { width: 100%; height: 100%; display: block; }
.fault-dot {
  transform-box: fill-box;
  transform-origin: center;
  animation: rv-dot-pulse 2s ease-out infinite;
}
@keyframes rv-dot-pulse {
  0% { opacity: 1; transform: scale(1); }
  60% { opacity: 0.75; transform: scale(1.4); }
  100% { opacity: 1; transform: scale(1); }
}

.fault-card {
  position: absolute; top: 6px; right: 6px; width: min(80%, 320px);
  max-height: calc(100% - 12px); overflow: auto;
  background: rgba(255, 255, 255, 0.96); backdrop-filter: blur(12px);
  border: 1px solid #e2e8f0; border-radius: 12px; box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  padding: 12px; z-index: 20;
}
.fault-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; border-bottom: 1px solid #f1f5f9; padding-bottom: 6px; }
.fault-title-group { display: flex; align-items: center; gap: 6px; }
.fault-room { font-size: 16px; font-weight: 800; color: #0f172a; }
.fault-floor { font-size: 12px; color: #64748b; }
.card-status-badge { font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 4px; }
.card-status-badge.fault { background: #fee2e2; color: #dc2626; }
.card-status-badge.done { background: #dcfce7; color: #15803d; }
.card-status-badge.normal { background: #f1f5f9; color: #64748b; }
.fault-close { width: 22px; height: 22px; border-radius: 50%; border: none; background: #f1f5f9; color: #64748b; cursor: pointer; font-size: 11px; }

.fault-empty { font-size: 12px; color: #94a3b8; padding: 12px 0; text-align: center; }
.fault-item { padding: 8px 0; border-bottom: 1px dashed #f1f5f9; }
.fault-item:last-child { border-bottom: none; }
.fault-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.fault-type { font-size: 11px; font-weight: 700; color: #2563eb; background: #eff6ff; padding: 2px 6px; border-radius: 4px; }
.fault-status { font-size: 11px; font-weight: 600; }
.fault-status.fs1, .fault-status.fs2 { color: #d97706; }
.fault-status.fs3 { color: #2563eb; }
.fault-status.fs4 { color: #16a34a; }
.fault-title { font-size: 13px; font-weight: 700; color: #1e293b; margin-bottom: 4px; }
.fault-label { font-size: 10px; color: #94a3b8; margin-top: 4px; }
.fault-text { font-size: 12px; color: #334155; line-height: 1.4; }
.fault-meta { font-size: 10px; color: #94a3b8; margin-top: 6px; }
.fault-actions { display: flex; align-items: center; gap: 8px; margin-top: 8px; }
.fault-btn { padding: 4px 12px; font-size: 11px; font-weight: 600; color: #fff; background: #2563eb; border: none; border-radius: 6px; cursor: pointer; }
.fault-btn.done { background: #10b981; }
.badge-done-text { font-size: 11px; font-weight: 700; color: #059669; }
</style>
