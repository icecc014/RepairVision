<template>
  <div class="plan-wrap">
    <div class="plan-top">
      <div class="floor-tabs">
        <button
          v-for="f in building.floors"
          :key="f"
          class="floor-tab"
          :class="{ active: activeFloor === f }"
          @click="switchFloor(f)"
        >
          {{ f }}F
        </button>
      </div>
      <div class="zoom-bar">
        <button class="zoom-btn" @click="zoomAt(0.8)">＋</button>
        <button class="zoom-btn" @click="zoomAt(1.25)">－</button>
        <button class="zoom-btn wide" @click="resetView">复位</button>
      </div>
    </div>

    <div class="legend-line">
      <span><i class="red"></i> 待处理故障（点击看解释）</span>
      <span><i class="blue"></i> 普通房间</span>
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
              <stop offset="0%" stop-color="#eef3fc" />
              <stop offset="100%" stop-color="#e6ecfa" />
            </linearGradient>
            <linearGradient id="gRoom" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#dcebff" />
              <stop offset="100%" stop-color="#c7dcf7" />
            </linearGradient>
            <linearGradient id="gRoomFault" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#fde3ea" />
              <stop offset="100%" stop-color="#f8c9d6" />
            </linearGradient>
            <linearGradient id="gCorridor" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#f3f7fd" />
              <stop offset="100%" stop-color="#e9f0fa" />
            </linearGradient>
            <linearGradient id="gStair" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#e6ecf7" />
              <stop offset="100%" stop-color="#d3dcec" />
            </linearGradient>
            <linearGradient id="gPublic" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#eaf2fd" />
              <stop offset="100%" stop-color="#dde8f8" />
            </linearGradient>
          </defs>
          <rect x="1" y="1" :width="PLAN_WIDTH - 2" :height="PLAN_DEPTH - 2" fill="url(#gPlanBg)" stroke="#7d8db3" stroke-width="1.6" rx="1.5" />
          <template v-if="!plan.custom">
          <rect :x="plan.corridor.x" :y="plan.corridor.z" :width="plan.corridor.w" :height="plan.corridor.d" fill="url(#gCorridor)" stroke="#a9b8d4" stroke-width="0.6" stroke-dasharray="3 2" />
          <text :x="plan.corridor.x + plan.corridor.w / 2" :y="PLAN_DEPTH / 2" text-anchor="middle" font-size="4" fill="#94a3b8" transform="rotate(90, 50, 88)">过道</text>

          <g v-for="core in plan.cores" :key="core.index">
            <rect x="0" :y="core.z" width="32" :height="core.d" fill="url(#gStair)" stroke="#7d8db3" stroke-width="0.7" />
            <text x="16" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.4" fill="#334155">封闭防火楼梯</text>
            <rect x="32" :y="core.z" width="68" :height="core.d" fill="url(#gPublic)" stroke="#9aa9c6" stroke-width="0.7" stroke-dasharray="2 1.6" />
            <text x="66" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.8" fill="#64748b">公共区域</text>
          </g>
          </template>
          <g v-else>
            <template v-for="(b, bi) in plan.blocks" :key="'blk' + bi">
              <rect :x="b.x" :y="b.z" :width="b.w" :height="b.d" :fill="blockFill(b.type)" :stroke="b.type === 'corridor' ? '#a9b8d4' : '#7d8db3'" stroke-width="0.6" :stroke-dasharray="b.type === 'corridor' ? '3 2' : '0'" />
              <text v-if="b.type !== 'corridor'" :x="b.x + b.w / 2" :y="b.z + b.d / 2 + 1.1" text-anchor="middle" font-size="2.6" fill="#64748b">{{ b.type === 'stair' ? '楼梯' : '公共区' }}</text>
            </template>
          </g>

          <g v-for="room in plan.rooms" :key="room.no">
            <rect
              :x="room.x"
              :y="room.z"
              :width="room.w"
              :height="room.d"
              :fill="roomOrders(room)[0] ? 'url(#gRoomFault)' : 'url(#gRoom)'"
              stroke="#5f7bb5"
              stroke-width="0.8"
              @pointerdown.stop
              @click.stop="select(room)"
              style="cursor: pointer"
            />
            <text
              :x="room.x + room.w / 2"
              :y="room.z + room.d / 2 + 1.4"
              text-anchor="middle"
              font-size="4.6"
              font-weight="bold"
              fill="#3d5687"
            >
              {{ room.no }}
            </text>
            <circle
              v-if="roomOrders(room).length"
              :cx="room.x + room.w - 5"
              :cy="room.z + 5"
              r="3.4"
              class="fault-dot" fill="#e0648a"
              @pointerdown.stop
              @click.stop="select(room)"
              style="cursor: pointer"
            />
          </g>
        </svg>
      </div>

      <aside v-if="activeFault" class="fault-card" @pointerdown.stop @wheel.stop>
        <div class="fault-head">
          <div>
            <span class="fault-room">{{ activeFault.no }}</span>
            <span class="fault-floor">{{ activeFault.floor }} 层</span>
          </div>
          <button class="fault-close" @click="activeFault = null">✕</button>
        </div>

        <div v-for="o in activeFault.orders" :key="o.id" class="fault-item">
          <div class="fault-row">
            <span class="fault-type">{{ o.faultTypeName }}</span>
            <span class="fault-status" :class="'fs' + o.status">{{ o.statusText }}</span>
          </div>
          <div class="fault-label">可能原因</div>
          <div class="fault-text">{{ explain(o).cause }}</div>
          <div class="fault-label">故障描述</div>
          <div class="fault-text">{{ o.description || '无补充说明' }}</div>
          <div class="fault-label">处理建议</div>
          <div class="fault-text">{{ explain(o).advice }}</div>
          <div class="fault-meta">报修 {{ o.createdAt }} · {{ o.workerName || '待派单' }}</div>
          <div class="fault-actions">
            <button v-if="o.status === 2" class="fault-btn" :disabled="actingId === o.id" @click="act(o, 'start')">
              {{ actingId === o.id ? '处理中…' : '开工' }}
            </button>
            <button v-if="o.status === 3" class="fault-btn done" :disabled="actingId === o.id" @click="act(o, 'complete')">
              {{ actingId === o.id ? '处理中…' : '完工' }}
            </button>
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
import { apiCompleteOrder, apiStartOrder } from '../api'
import { PLAN_DEPTH, PLAN_WIDTH, matchRoomOrders, type PlanRoom } from '../utils/floorLayout'
import { resolveFloorPlan } from '../utils/layoutGrid'

function blockFill(type: 'corridor' | 'stair' | 'public') {
  if (type === 'stair') return '#dbe3f3'
  if (type === 'public') return '#e7f1fe'
  return '#eef3fc'
}

const props = defineProps<{ building: WorkerMapBuilding; orders: OrderItem[] }>()
const emit = defineEmits<{
  (e: 'selectRoom', room: { num: string; floor: number; orders: OrderItem[] }): void
  (e: 'refresh'): void
}>()

const activeFloor = ref(1)
const viewportRef = ref<HTMLDivElement | null>(null)
const activeFault = ref<{ no: string; floor: number; orders: OrderItem[] } | null>(null)
const actingId = ref<number | null>(null)

const MIN_W = 24
const MAX_W = PLAN_WIDTH * 1.4
const view = reactive({ x: 0, y: 0, w: PLAN_WIDTH, h: PLAN_DEPTH })
const viewBoxStr = computed(() => `${view.x} ${view.y} ${view.w} ${view.h}`)

const plan = computed(() =>
  resolveFloorPlan(activeFloor.value, props.building.roomsPerFloor, props.building.layoutJson),
)

function roomOrders(room: PlanRoom) {
  return matchRoomOrders(
    props.orders.filter((o) => o.buildingId === props.building.id),
    activeFloor.value,
    room.no,
  ).filter((o) => o.status === 1 || o.status === 2 || o.status === 3)
}

function switchFloor(f: number) {
  activeFloor.value = f
  activeFault.value = null
  resetView()
}

function select(room: PlanRoom) {
  const orders = roomOrders(room)
  activeFault.value = orders.length > 0 ? { no: room.no, floor: activeFloor.value, orders } : null
  emit('selectRoom', { num: room.no, floor: activeFloor.value, orders })
}

function explain(o: OrderItem): { cause: string; advice: string } {
  switch (o.faultType) {
    case 'electric':
      return {
        cause: '常见于线路接触不良、开关/插座损坏、灯具故障或负载跳闸。',
        advice: '先断开该房间电源再检修，避免湿手操作；涉及总闸请联系电工。',
      }
    case 'water':
      return {
        cause: '常见于管道接头渗漏、阀门老化、下水堵塞或水压异常。',
        advice: '先关闭角阀/进水阀并清理积水，避免渗漏扩大到楼下。',
      }
    default:
      return {
        cause: '设施损坏或需要现场排查的具体故障。',
        advice: '请按报修描述携带工具上门确认，必要时上报更换配件。',
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
      await apiStartOrder(o.id)
      showToast('已开工')
    } else {
      await apiCompleteOrder(o.id)
      showToast('已完工')
    }
    activeFault.value = null
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
      const orders = roomOrders(room)
      activeFault.value = orders.length > 0 ? { no: room.no, floor: activeFloor.value, orders } : null
      return
    }
  }
  activeFault.value = null
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
    moved = false
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
.floor-tab { flex: 0 0 auto; padding: 6px 13px; font-size: 12px; background: #eef2ff; border: 1px solid #c7d2fe; border-radius: 999px; cursor: pointer; }
.floor-tab.active { color: #fff; background: #2563eb; }
.zoom-bar { display: flex; gap: 6px; }
.zoom-btn { width: 32px; height: 30px; font-size: 15px; font-weight: 700; color: #1d4ed8; background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 8px; cursor: pointer; }
.zoom-btn.wide { width: auto; padding: 0 10px; font-size: 12px; }
.legend-line { display: flex; gap: 12px; margin: 6px 0; color: #64748b; font-size: 11px; flex-wrap: wrap; }
.red { display: inline-block; width: 8px; height: 8px; background: #ef4444; border-radius: 50%; }
.blue { display: inline-block; width: 8px; height: 8px; background: #dbeafe; border: 1px solid #1e3a8a; border-radius: 2px; }
.plan-stage { position: relative; }
.plan-viewport { height: calc(82vh - 200px); min-height: 260px; overflow: hidden; touch-action: none; cursor: grab; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; }
.plan-svg { width: 100%; height: 100%; display: block; }
.fault-dot {
  transform-box: fill-box;
  transform-origin: center;
  animation: rv-dot-pulse 2.2s ease-out infinite;
}
@keyframes rv-dot-pulse {
  0% { opacity: 1; transform: scale(1); }
  60% { opacity: 0.75; transform: scale(1.55); }
  100% { opacity: 1; transform: scale(1); }
}
.fault-card {
  position: absolute; top: 6px; right: 6px; width: min(76%, 320px);
  max-height: calc(100% - 12px); overflow: auto;
  z-index: 6;
  padding: 12px 14px; color: #2b3445;
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 16px;
  box-shadow: 0 16px 36px rgba(46, 68, 112, 0.18);
}
.fault-head { display: flex; align-items: center; justify-content: space-between; }
.fault-room { color: var(--rv-primary-deep); font-size: 16px; font-weight: 800; }
.fault-floor { margin-left: 6px; color: #6b7a94; font-size: 12px; }
.fault-close { width: 24px; height: 24px; color: var(--rv-text-sub); background: rgba(120,145,190,0.14); border: none; border-radius: 50%; cursor: pointer; }
.fault-item { margin-top: 8px; padding-top: 8px; border-top: 1px dashed rgba(120,145,190,0.24); }
.fault-row { display: flex; align-items: center; gap: 8px; }
.fault-type { padding: 2px 8px; color: #2462d9; font-size: 12px; background: var(--rv-grad-1); border-radius: 999px; }
.fault-status { font-size: 12px; font-weight: 700; }
.fs1, .fs2 { color: #b96b1c; }
.fs3 { color: #2462d9; }
.fs4 { color: #17865a; }
.fs5 { color: #5a6a85; }
.fault-label { margin-top: 7px; color: #2462d9; font-size: 11px; }
.fault-text { margin-top: 2px; color: #2b3445; font-size: 12px; line-height: 1.5; }
.fault-meta { margin-top: 6px; color: #6b7a94; font-size: 11px; }
.fault-actions { margin-top: 8px; display: flex; gap: 8px; }
.fault-btn { padding: 5px 14px; color: #fff; font-size: 12px; background: linear-gradient(135deg,#7fb2ff,#3478f6); border: none; border-radius: 999px; cursor: pointer; }
.fault-btn.done { background: linear-gradient(135deg,#7fe6c8,#22b573); }
</style>