<template>
  <div class="admin-visual">
    <div class="tabs">
      <button class="tab" :class="{ active: mode === 'plan' }" @click="mode = 'plan'">楼层户型</button>
      <button class="tab" :class="{ active: mode === '3d' }" @click="switch3d">3D 立体</button>
    </div>

    <div v-if="mode === 'plan'" class="plan-tab">
      <div class="plan-top">
        <div class="floor-tabs">
          <button
            v-for="f in building.floors"
            :key="f"
            class="floor-chip"
            :class="{ active: floor === f }"
            @click="floor = f"
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
      <div class="plan-head">
        {{ building.name }} · 第 {{ floor }} 层 · 每层 {{ building.roomsPerFloor }} 间 · 拖动平移 / 滚轮缩放
      </div>
      <div
        class="plan-viewport"
        @wheel.prevent="onWheel"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
        @pointerleave="onPointerUp"
      >
        <svg class="plan-svg" :viewBox="viewBoxStr" preserveAspectRatio="xMidYMid meet">
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
              :x="room.x" :y="room.z" :width="room.w" :height="room.d"
              :fill="roomOrders(room).length ? '#fee2e2' : '#dbeafe'"
              stroke="#1e3a8a" stroke-width="0.8"
            />
            <text :x="room.x + room.w / 2" :y="room.z + room.d / 2 + 1.4" text-anchor="middle" font-size="4.6" font-weight="bold" fill="#1e3a8a">
              {{ room.no }}
            </text>
            <circle v-if="roomOrders(room).length" :cx="room.x + room.w - 5" :cy="room.z + 5" r="3.4" fill="#ef4444" />
          </g>
        </svg>
      </div>
      <p class="hint">北区 01-04 · 中区 05-12 · 南区 13-16 · 每栋楼按同样标准层逐层映射</p>
    </div>

    <div v-else class="three-tab">
      <div ref="mountRef" class="three-mount"></div>
      <p v-if="webglError" class="error">{{ errorText || 'WebGL 初始化失败，可切回楼层户型' }}</p>
      <p class="hint">拖拽旋转 · 滚轮缩放 · 每层含走廊与两处核心筒</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref } from 'vue'
import type { AdminBuilding, OrderItem } from '../api'
import {
  PLAN_DEPTH,
  PLAN_WIDTH,
  buildFloorPlan,
  buildGridRooms,
  matchRoomOrders,
  supportsCorridorLayout,
  type PlanRoom,
} from '../utils/floorLayout'

const props = defineProps<{ building: AdminBuilding; orders: OrderItem[] }>()
const mode = ref<'plan' | '3d'>('plan')
const floor = ref(1)
const mountRef = ref<HTMLDivElement | null>(null)
const webglError = ref(false)
const errorText = ref('')

const MIN_W = 24
const MAX_W = PLAN_WIDTH * 1.4
const view = reactive({ x: 0, y: 0, w: PLAN_WIDTH, h: PLAN_DEPTH })
const viewBoxStr = computed(() => `${view.x} ${view.y} ${view.w} ${view.h}`)

const plan = computed(() => {
  if (supportsCorridorLayout(props.building.roomsPerFloor)) {
    return buildFloorPlan(floor.value, props.building.roomsPerFloor)
  }
  return {
    floor: floor.value,
    rooms: buildGridRooms(floor.value, props.building.roomsPerFloor || 8),
    corridor: { x: 0, z: 0, w: 0, d: 0 },
    cores: [] as { index: number; z: number; d: number }[],
  }
})

function roomOrders(room: PlanRoom) {
  return matchRoomOrders(
    props.orders.filter((o) => o.buildingId === props.building.id),
    floor.value,
    room.no,
  )
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
  const p = svgPoint(e)
  zoomAt(e.deltaY > 0 ? 1.12 : 0.89, p.x, p.y)
}

const pointers = new Map<number, { x: number; y: number }>()
let dragStart: { x: number; y: number; vx: number; vy: number } | null = null

function onPointerDown(e: PointerEvent) {
  (e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  dragStart = { x: e.clientX, y: e.clientY, vx: view.x, vy: view.y }
}

function onPointerMove(e: PointerEvent) {
  if (!pointers.has(e.pointerId) || !dragStart) return
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  const dx = ((e.clientX - dragStart.x) / rect.width) * view.w
  const dy = ((e.clientY - dragStart.y) / rect.height) * view.h
  view.x = dragStart.vx - dx
  view.y = dragStart.vy - dy
  clampView()
}

function onPointerUp(e: PointerEvent) {
  pointers.delete(e.pointerId)
  if (pointers.size === 0) dragStart = null
}

async function switch3d() {
  mode.value = '3d'
  webglError.value = false
  errorText.value = ''
  await nextTick()
  if (mountRef.value) init3d()
}

async function init3d() {
  if (!mountRef.value) return
  try {
    const THREE = await import('three')
    const controlsModule = await import('three/examples/jsm/controls/OrbitControls.js')
    const el = mountRef.value
    const width = el.clientWidth || 700
    const height = 430
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x0b1e45)
    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 4000)
    const renderer = new THREE.WebGLRenderer({ antialias: true })
    renderer.setSize(width, height)
    el.innerHTML = ''
    el.appendChild(renderer.domElement)

    const boxW = 140
    const boxD = (boxW * PLAN_DEPTH) / PLAN_WIDTH
    const floorH = 10
    camera.position.set(boxW * 1.5, floorH * 3.2, boxD * 1.4)
    scene.add(new THREE.AmbientLight(0xffffff, 0.92))
    const dir = new THREE.DirectionalLight(0xffffff, 0.85)
    dir.position.set(80, 160, 60)
    scene.add(dir)

    const controls = new controlsModule.OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true
    controls.target.set(0, floorH * 1.6, 0)
    controls.update()

    const ground = new THREE.Mesh(new THREE.PlaneGeometry(boxW * 3, boxD * 2), new THREE.MeshLambertMaterial({ color: '#12264e', side: THREE.DoubleSide }))
    ground.rotation.x = -Math.PI / 2
    ground.position.y = -1.2
    scene.add(ground)

    const totalFloors = Math.max(props.building.floors, 1)
    for (let f = 0; f < totalFloors; f++) {
      const yBase = f * (floorH + 1)
      const slab = new THREE.Mesh(new THREE.BoxGeometry(boxW + 6, 0.9, boxD + 6), new THREE.MeshLambertMaterial({ color: '#1e3a8a' }))
      slab.position.y = yBase + 0.45
      scene.add(slab)
      const p = supportsCorridorLayout(props.building.roomsPerFloor)
        ? buildFloorPlan(f + 1, props.building.roomsPerFloor)
        : { rooms: buildGridRooms(f + 1, props.building.roomsPerFloor || 8), corridor: { x: 0, z: 0, w: 0, d: 0 }, cores: [] as { index: number; z: number; d: number }[] }

      for (const room of p.rooms) {
        const w = (room.w / PLAN_WIDTH) * boxW * 0.92
        const d = (room.d / PLAN_DEPTH) * boxD * 0.92
        const x = ((room.x + room.w / 2 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW
        const z = ((room.z + room.d / 2 - PLAN_DEPTH / 2) / PLAN_DEPTH) * boxD
        const mat = new THREE.MeshLambertMaterial({ color: f % 2 === 0 ? '#60a5fa' : '#34d399', transparent: true, opacity: 0.85 })
        const tile = new THREE.Mesh(new THREE.BoxGeometry(w, 1, d), mat)
        tile.position.set(x, yBase + 1, z)
        scene.add(tile)
      }

      if (p.corridor.w > 0) {
        const cw = (p.corridor.w / PLAN_WIDTH) * boxW
        const cd = (p.corridor.d / PLAN_DEPTH) * boxD
        const corridor = new THREE.Mesh(new THREE.BoxGeometry(cw, 0.7, cd), new THREE.MeshLambertMaterial({ color: '#cbd5e1' }))
        corridor.position.set(0, yBase + 0.8, 0)
        scene.add(corridor)
      }

      for (const core of p.cores) {
        const cw = (32 / PLAN_WIDTH) * boxW
        const cd = (core.d / PLAN_DEPTH) * boxD
        const cz = ((core.z + core.d / 2 - PLAN_DEPTH / 2) / PLAN_DEPTH) * boxD
        const stair = new THREE.Mesh(
          new THREE.BoxGeometry(cw, floorH * 0.9, cd),
          new THREE.MeshLambertMaterial({ color: '#94a3b8', transparent: true, opacity: 0.8 }),
        )
        stair.position.set(((16 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW, yBase + floorH * 0.45, cz)
        scene.add(stair)
      }
    }

    const animate = () => {
      controls.update()
      renderer.render(scene, camera)
      requestAnimationFrame(animate)
    }
    animate()
  } catch (e) {
    console.error(e)
    webglError.value = true
    errorText.value = String((e as Error)?.message || e)
  }
}
</script>

<style scoped>
.admin-visual { width: 100%; }
.tabs { display: flex; gap: 10px; margin-bottom: 10px; }
.tab { padding: 8px 18px; border-radius: 999px; border: 1px solid #c7d2fe; background: #eef2ff; cursor: pointer; }
.tab.active { background: #2563eb; color: #fff; }
.plan-top { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.floor-tabs { display: flex; gap: 6px; flex: 1; flex-wrap: wrap; }
.floor-chip { padding: 5px 12px; font-size: 12px; background: #eef2ff; border: 1px solid #c7d2fe; border-radius: 999px; cursor: pointer; }
.floor-chip.active { color: #fff; background: #2563eb; }
.zoom-bar { display: flex; gap: 6px; }
.zoom-btn { width: 34px; height: 30px; font-size: 15px; font-weight: 700; color: #1d4ed8; background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 8px; cursor: pointer; }
.zoom-btn.wide { width: auto; padding: 0 12px; font-size: 12px; }
.plan-head { margin-bottom: 6px; font-weight: 700; }
.plan-viewport { height: 460px; overflow: hidden; touch-action: none; cursor: grab; background: #fff; border: 1px solid #e2e8f0; border-radius: 10px; }
.plan-svg { width: 100%; height: 100%; display: block; }
.three-mount { width: 100%; height: 430px; }
.error { color: #dc2626; }
.hint { color: #94a3b8; text-align: center; font-size: 12px; }
</style>