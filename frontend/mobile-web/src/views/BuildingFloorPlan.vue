<template>
  <div class="plan-wrap">
    <div class="plan-top">
      <div class="floor-tabs">
        <button
          v-for="f in building.floors"
          :key="f"
          class="floor-tab"
          :class="{ active: activeFloor === f }"
          @click="activeFloor = f"
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
      <span><i class="red"></i> 待处理工单</span>
      <span><i class="blue"></i> 普通房间</span>
      <span>拖动平移 · 滚轮/双指缩放 · 点击房间</span>
    </div>

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
      <svg
        class="plan-svg"
        :viewBox="viewBoxStr"
        preserveAspectRatio="xMidYMid meet"
      >
        <rect x="1" y="1" :width="PLAN_WIDTH - 2" :height="PLAN_DEPTH - 2" fill="#f8fafc" stroke="#1e293b" stroke-width="1.6" rx="1.5" />
        <rect :x="plan.corridor.x" :y="plan.corridor.z" :width="plan.corridor.w" :height="plan.corridor.d" fill="#eef2f7" stroke="#94a3b8" stroke-width="0.6" stroke-dasharray="3 2" />
        <text :x="plan.corridor.x + plan.corridor.w / 2" :y="PLAN_DEPTH / 2" text-anchor="middle" font-size="4" fill="#94a3b8" transform="rotate(90, 50, 88)">过道</text>

        <g v-for="core in plan.cores" :key="core.index">
          <rect x="0" :y="core.z" width="32" :height="core.d" fill="#e2e8f0" stroke="#475569" stroke-width="0.7" />
          <text x="16" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.4" fill="#334155">封闭防火楼梯</text>
          <rect x="32" :y="core.z" width="68" :height="core.d" fill="#e8eef7" stroke="#64748b" stroke-width="0.7" stroke-dasharray="2 1.6" />
          <text x="66" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.8" fill="#64748b">公共区域</text>
        </g>

        <g v-for="room in plan.rooms" :key="room.no">
          <rect
            :x="room.x"
            :y="room.z"
            :width="room.w"
            :height="room.d"
            :fill="roomOrders(room)[0] ? '#fee2e2' : '#dbeafe'"
            stroke="#1e3a8a"
            stroke-width="0.8"
            @click="select(room)"
            style="cursor: pointer"
          />
          <text
            :x="room.x + room.w / 2"
            :y="room.z + room.d / 2 + 1.4"
            text-anchor="middle"
            font-size="4.6"
            font-weight="bold"
            fill="#1e3a8a"
          >
            {{ room.no }}
          </text>
          <circle
            v-if="roomOrders(room).length"
            :cx="room.x + room.w - 5"
            :cy="room.z + 5"
            r="3.4"
            fill="#ef4444"
            @click="select(room)"
            style="cursor: pointer"
          />
        </g>
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'
import {
  PLAN_DEPTH,
  PLAN_WIDTH,
  buildFloorPlan,
  buildGridRooms,
  matchRoomOrders,
  supportsCorridorLayout,
  type PlanRoom,
} from '../utils/floorLayout'

const props = defineProps<{ building: WorkerMapBuilding; orders: OrderItem[] }>()
const emit = defineEmits<{ (e: 'selectRoom', room: { num: string; floor: number; orders: OrderItem[] }): void }>()

const activeFloor = ref(1)
const viewportRef = ref<HTMLDivElement | null>(null)

const MIN_W = 24
const MAX_W = PLAN_WIDTH * 1.4
const view = reactive({ x: 0, y: 0, w: PLAN_WIDTH, h: PLAN_DEPTH })

const viewBoxStr = computed(() => `${view.x} ${view.y} ${view.w} ${view.h}`)

const plan = computed(() => {
  if (supportsCorridorLayout(props.building.roomsPerFloor)) {
    return buildFloorPlan(activeFloor.value, props.building.roomsPerFloor)
  }
  return {
    floor: activeFloor.value,
    rooms: buildGridRooms(activeFloor.value, props.building.roomsPerFloor || 8),
    corridor: { x: 0, z: 0, w: 0, d: 0 },
    cores: [] as { index: number; z: number; d: number }[],
  }
})

function roomOrders(room: PlanRoom) {
  return matchRoomOrders(
    props.orders.filter((o) => o.buildingId === props.building.id),
    activeFloor.value,
    room.no,
  )
}

function select(room: PlanRoom) {
  emit('selectRoom', { num: room.no, floor: activeFloor.value, orders: roomOrders(room) })
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
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  const nx = (e.clientX - rect.left) / rect.width
  const ny = (e.clientY - rect.top) / rect.height
  return { x: view.x + nx * view.w, y: view.y + ny * view.h }
}

function onWheel(e: WheelEvent) {
  if (e.ctrlKey || Math.abs(e.deltaY) > 0) {
    const p = svgPoint(e)
    zoomAt(e.deltaY > 0 ? 1.12 : 0.89, p.x, p.y)
  }
}

const pointers = new Map<number, { x: number; y: number }>()
let dragStart: { x: number; y: number; vx: number; vy: number } | null = null
let pinchStartDist = 0
let pinchStartW = PLAN_WIDTH

function onPointerDown(e: PointerEvent) {
  (e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
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
.plan-wrap { padding: 4px 10px 12px; }
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
.plan-viewport { height: calc(82vh - 190px); min-height: 240px; overflow: hidden; touch-action: none; cursor: grab; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; }
.plan-svg { width: 100%; height: 100%; display: block; }
</style>