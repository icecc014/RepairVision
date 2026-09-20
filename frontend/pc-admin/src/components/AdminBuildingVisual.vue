<template>
  <div class="admin-visual-root">
    <!-- 顶部控制条：模式切换、楼层切片、待修指标 -->
    <div class="visual-header-bar">
      <div class="mode-switch-group">
        <button
          class="mode-btn"
          :class="{ active: mode === 'plan' }"
          @click="mode = 'plan'"
        >
          <span class="btn-icon">📐</span>
          <span>2D 楼层户型</span>
        </button>
        <button
          class="mode-btn"
          :class="{ active: mode === '3d' }"
          @click="switch3d"
        >
          <span class="btn-icon">🏢</span>
          <span>3D 立体沙盘</span>
        </button>
      </div>

      <!-- 楼层过滤器 -->
      <div class="floor-filter-wrap">
        <span class="filter-title">楼层透视：</span>
        <div class="floor-chip-scroller">
          <button
            class="floor-chip-btn"
            :class="{ active: selectedFloor === 0 && mode === '3d' }"
            v-if="mode === '3d'"
            @click="setFloor(0)"
          >
            整栋全景
          </button>
          <button
            v-for="f in building.floors"
            :key="f"
            class="floor-chip-btn"
            :class="{ active: (mode === 'plan' ? currentPlanFloor === f : selectedFloor === f) }"
            @click="setFloor(f)"
          >
            {{ f }}F
            <span v-if="getFloorStatus(f) === 'fault'" class="chip-fault-dot"></span>
            <span v-else-if="getFloorStatus(f) === 'done'" class="chip-done-dot"></span>
          </button>
        </div>
      </div>

      <!-- 概览状态标牌 -->
      <div class="header-badges">
        <div class="stat-badge" :class="{ 'has-faults': totalFaultRoomsCount > 0 }">
          <span class="pulse-dot"></span>
          <span class="badge-text">待修: <strong>{{ totalFaultRoomsCount }}</strong> 间</span>
        </div>
        <div class="stat-badge done-badge" v-if="totalDoneRoomsCount > 0">
          <span class="done-dot"></span>
          <span class="badge-text">已修: <strong>{{ totalDoneRoomsCount }}</strong> 间</span>
        </div>
      </div>
    </div>

    <!-- 主交互工作区：左侧 2D/3D 画布，右侧工单联动面板 -->
    <div class="visual-stage-container">
      <!-- 2D 平面视口 -->
      <div v-show="mode === 'plan'" class="canvas-panel plan-panel">
        <div class="panel-sub-toolbar">
          <span class="sub-hint">💡 滚轮平滑缩放 · 鼠标拖拽平移 · 点击房间查看报修工单</span>
          <div class="zoom-action-group">
            <button class="square-tool-btn" title="放大" @click="zoomAt(0.8)">＋</button>
            <button class="square-tool-btn" title="缩小" @click="zoomAt(1.25)">－</button>
            <button class="text-tool-btn" @click="resetView">复位视口</button>
          </div>
        </div>

        <div
          ref="planBoxRef" class="plan-viewport-box"
          @wheel.prevent="onWheel"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
          @pointerleave="onPointerUp"
        >
          <svg class="plan-svg-canvas" :viewBox="viewBoxStr" preserveAspectRatio="xMidYMid meet" @click="onSvgClick">
            <defs>
              <linearGradient id="v4BgGrad" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#f8fafc" />
                <stop offset="100%" stop-color="#edf2f7" />
              </linearGradient>
              <linearGradient id="v4RoomNormal" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#e0f2fe" />
                <stop offset="100%" stop-color="#bae6fd" />
              </linearGradient>
              <linearGradient id="v4RoomFault" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#ffe4e6" />
                <stop offset="100%" stop-color="#fecdd3" />
              </linearGradient>
              <linearGradient id="v4RoomDone" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#dcfce7" />
                <stop offset="100%" stop-color="#bbf7d0" />
              </linearGradient>
              <linearGradient id="v4Corridor" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#ffffff" />
                <stop offset="100%" stop-color="#f1f5f9" />
              </linearGradient>
              <linearGradient id="v4Stair" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#f1f5f9" />
                <stop offset="100%" stop-color="#e2e8f0" />
              </linearGradient>
              <linearGradient id="v4Public" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#f8fafc" />
                <stop offset="100%" stop-color="#e2e8f0" />
              </linearGradient>
            </defs>

            <!-- 建筑基座外框 -->
            <rect x="1" y="1" :width="PLAN_WIDTH - 2" :height="PLAN_DEPTH - 2" fill="url(#v4BgGrad)" stroke="#cbd5e1" stroke-width="1.2" rx="4" />

            <!-- 走廊与核心筒 -->
            <template v-if="!activeFloorPlan.custom">
              <rect
                :x="activeFloorPlan.corridor.x"
                :y="activeFloorPlan.corridor.z"
                :width="activeFloorPlan.corridor.w"
                :height="activeFloorPlan.corridor.d"
                fill="url(#v4Corridor)"
                stroke="#cbd5e1"
                stroke-width="0.6"
                stroke-dasharray="3 2"
              />
              <text :x="activeFloorPlan.corridor.x + activeFloorPlan.corridor.w / 2" :y="PLAN_DEPTH / 2" text-anchor="middle" font-size="3.6" fill="#94a3b8" transform="rotate(90, 50, 88)">贯通走廊</text>

              <g v-for="core in activeFloorPlan.cores" :key="core.index">
                <rect x="0" :y="core.z" width="32" :height="core.d" fill="url(#v4Stair)" stroke="#94a3b8" stroke-width="0.7" />
                <text x="16" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.2" fill="#475569">封闭防火楼梯</text>
                <rect x="32" :y="core.z" width="68" :height="core.d" fill="url(#v4Public)" stroke="#cbd5e1" stroke-width="0.7" stroke-dasharray="2 1.6" />
                <text x="66" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.4" fill="#64748b">公共区域</text>
              </g>
            </template>

            <!-- 自定义网格布局 -->
            <g v-else>
              <template v-for="(b, bi) in activeFloorPlan.blocks" :key="'blk' + bi">
                <rect
                  :x="b.x" :y="b.z" :width="b.w" :height="b.d"
                  :fill="blockFill(b.type)"
                  :stroke="b.type === 'corridor' ? '#cbd5e1' : '#94a3b8'"
                  stroke-width="0.6"
                  :stroke-dasharray="b.type === 'corridor' ? '3 2' : '0'"
                />
                <text v-if="b.type !== 'corridor'" :x="b.x + b.w / 2" :y="b.z + b.d / 2 + 1.1" text-anchor="middle" font-size="2.6" fill="#64748b">
                  {{ b.label || (b.type === 'stair' ? '楼梯' : '公共区') }}
                </text>
              </template>
            </g>

            <!-- 房间矩形元素 -->
            <g v-for="room in activeFloorPlan.rooms" :key="room.no">
              <rect
                :x="room.x" :y="room.z" :width="room.w" :height="room.d"
                :fill="getRoomStyle(currentPlanFloor, room.no).fill"
                :stroke="activeSelectedRoom?.no === room.no && activeSelectedRoom?.floor === currentPlanFloor ? '#2563eb' : getRoomStyle(currentPlanFloor, room.no).stroke"
                :stroke-width="activeSelectedRoom?.no === room.no && activeSelectedRoom?.floor === currentPlanFloor ? 1.8 : 0.8"
                rx="2"
                class="room-svg-shape"
                @click.stop="onRoomClick(currentPlanFloor, room.no)"
              />
              <text
                :x="room.x + room.w / 2"
                :y="room.z + room.d / 2 + 1.2"
                text-anchor="middle"
                font-size="3.4"
                font-weight="700"
                :fill="getRoomStyle(currentPlanFloor, room.no).text"
                pointer-events="none"
              >
                {{ room.no }}
              </text>
              <!-- 待修故障警示脉冲点 -->
              <circle
                v-if="getRoomAnalysis(currentPlanFloor, room.no).status === 'fault'"
                :cx="room.x + room.w - 3.8"
                :cy="room.z + 3.8"
                r="2.4"
                fill="#f43f5e"
                stroke="#ffffff"
                stroke-width="0.8"
                class="fault-radar-dot"
                pointer-events="none"
              />
              <!-- 已完成绿色指示点 -->
              <circle
                v-else-if="getRoomAnalysis(currentPlanFloor, room.no).status === 'done'"
                :cx="room.x + room.w - 3.8"
                :cy="room.z + 3.8"
                r="2.0"
                fill="#10b981"
                stroke="#ffffff"
                stroke-width="0.8"
                pointer-events="none"
              />
            </g>
          </svg>
        </div>
      </div>

      <!-- 3D 沙盘视口 -->
      <div v-show="mode === '3d'" class="canvas-panel three-panel">
        <div class="panel-sub-toolbar">
          <span class="sub-hint">💡 鼠标左键旋转 · 右键平移 · 滚轮缩放 · 点击三维房间定位工单</span>
          <div class="three-legend-box">
            <span class="legend-tag normal"><i class="color-dot"></i> 正常房间</span>
            <span class="legend-tag done"><i class="color-dot done-dot-legend"></i> 已修好房间</span>
            <span class="legend-tag fault"><i class="color-dot beacon-dot"></i> 待修房间 (悬浮Beacon)</span>
          </div>
          <button class="text-tool-btn" @click="reset3dCamera">重置沙盘视角</button>
        </div>

        <div
          ref="mountRef"
          class="three-canvas-mount"
          @pointerdown="onThreePointerDown"
          @pointerup="onThreePointerUp"
          @pointermove="onThreePointerMove"
        ></div>

        <!-- 悬停三维房间悬浮提示 -->
        <div
          v-if="hoverRoomInfo"
          class="three-tooltip"
          :style="{ left: hoverRoomInfo.x + 'px', top: hoverRoomInfo.y + 'px' }"
        >
          <div class="tt-room">{{ hoverRoomInfo.roomNo }} 室 ({{ hoverRoomInfo.floor }}F)</div>
          <div class="tt-desc" :class="{ alert: hoverRoomInfo.faultCount > 0, done: hoverRoomInfo.faultCount === 0 && hoverRoomInfo.doneCount > 0 }">
            {{ hoverRoomInfo.faultCount > 0 ? ('待修工单: ' + hoverRoomInfo.faultCount + ' 条') : (hoverRoomInfo.doneCount > 0 ? ('已修好: ' + hoverRoomInfo.doneCount + ' 条') : '设施运行正常') }}
          </div>
        </div>

        <div v-if="webglError" class="three-error-mask">
          <p class="error-msg">⚠️ {{ errorText || 'WebGL 初始化失败，可切换回 2D 户型图查看' }}</p>
        </div>
      </div>

      <!-- 右侧房间工单联动侧边栏（毛玻璃拟态） -->
      <aside v-if="activeSelectedRoom" class="room-order-drawer">
        <div class="drawer-header-row">
          <div class="drawer-room-badge">
            <span class="room-code">{{ activeSelectedRoom.no }}</span>
            <span class="room-floor-pill">{{ activeSelectedRoom.floor }} 层标准间</span>
          </div>
          <button class="drawer-close-circle" title="关闭面板" @click="activeSelectedRoom = null">✕</button>
        </div>

        <div class="drawer-scroll-body">
          <div class="orders-summary-bar" :class="{ 'warning-status': selectedRoomOrders.length > 0 }">
            <span v-if="selectedRoomOrders.length > 0">
              🚨 发现 <strong>{{ selectedRoomOrders.length }}</strong> 起待处理报修
            </span>
            <span v-else>
              ✅ 该房间当前设施齐全，无报修记录
            </span>
          </div>

          <div v-if="selectedRoomOrders.length > 0" class="order-list-cards">
            <div v-for="ord in selectedRoomOrders" :key="ord.id" class="order-item-card">
              <div class="card-head">
                <span class="ord-no">{{ ord.orderNo || '#' + ord.id }}</span>
                <span class="ord-status" :class="'status-' + ord.status">{{ statusName(ord.status, ord.statusText) }}</span>
              </div>
              <div class="card-fault">
                <span class="fault-icon">📌</span>
                <span class="fault-text">{{ ord.faultTypeName || ord.faultType || '日常维修' }}</span>
              </div>
              <div class="card-desc">
                {{ ord.description || '无具体报修详情' }}
              </div>
              <div class="card-foot">
                <span class="card-worker" v-if="ord.workerName">
                  👤 师傅：{{ ord.workerName }}
                </span>
                <span class="card-worker" v-else>⏳ 待指派师傅</span>
                <span class="card-priority" v-if="ord.priority && ord.priority > 1">急件 P{{ ord.priority }}</span>
              </div>
            </div>
          </div>

          <div v-else class="empty-room-state">
            <div class="empty-glass-icon">🏢</div>
            <p class="empty-title">当前房间运转正常</p>
            <p class="empty-sub">没有处于派单或维修中的工单</p>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import type { AdminBuilding, OrderItem } from '../api'
import { PLAN_DEPTH, PLAN_WIDTH, matchRoomOrders } from '../utils/floorLayout'
import { resolveFloorPlan } from '../utils/layoutGrid'

const props = defineProps<{
  building: AdminBuilding
  orders: OrderItem[]
}>()

function blockFill(type: 'corridor' | 'stair' | 'public') {
  if (type === 'stair') return '#f1f5f9'
  if (type === 'public') return '#f8fafc'
  return '#ffffff'
}

// 视图模式与楼层状态
const mode = ref<'plan' | '3d'>('plan')
const currentPlanFloor = ref(1)
const selectedFloor = ref(0) // 0 表示整栋全景

// 当前选中查看工单的房间：{ floor, no }
const activeSelectedRoom = ref<{ floor: number; no: string } | null>(null)

// 3D 悬浮提示信息
const hoverRoomInfo = ref<{ x: number; y: number; floor: number; roomNo: string; faultCount: number } | null>(null)

// 2D 平面计算属性
const activeFloorPlan = computed(() => {
  const f = mode.value === 'plan' ? currentPlanFloor.value : (selectedFloor.value === 0 ? 1 : selectedFloor.value)
  return resolveFloorPlan(f, props.building.roomsPerFloor, props.building.layoutJson)
})

// 房间综合工单与状态分析
function getRoomAnalysis(floor: number, roomNo: string) {
  const bOrders = (props.orders || []).filter(o => o.buildingId === props.building.id)
  const matched = matchRoomOrders(bOrders, floor, roomNo)
  const unfinished = matched.filter(o => o.status === 1 || o.status === 2 || o.status === 3)
  const completed = matched.filter(o => o.status === 4)
  let status: 'fault' | 'done' | 'normal' = 'normal'
  if (unfinished.length > 0) {
    status = 'fault'
  } else if (completed.length > 0) {
    status = 'done'
  }
  return {
    unfinished,
    completed,
    status,
    total: matched.length
  }
}

function getRoomStyle(floor: number, roomNo: string) {
  const { status } = getRoomAnalysis(floor, roomNo)
  if (status === 'fault') {
    return { fill: 'url(#v4RoomFault)', stroke: '#f43f5e', text: '#e11d48' }
  } else if (status === 'done') {
    return { fill: 'url(#v4RoomDone)', stroke: '#10b981', text: '#047857' }
  }
  return { fill: 'url(#v4RoomNormal)', stroke: '#7dd3fc', text: '#0369a1' }
}

function getFloorStatus(floor: number): 'fault' | 'done' | 'normal' {
  const bOrders = (props.orders || []).filter(o => o.buildingId === props.building.id && (o.floor || 1) === floor)
  if (bOrders.some(o => o.status === 1 || o.status === 2 || o.status === 3)) {
    return 'fault'
  }
  if (bOrders.some(o => o.status === 4)) {
    return 'done'
  }
  return 'normal'
}

// 楼栋总待修与已修好房间数统计
const totalFaultRoomsCount = computed(() => {
  const buildingOrders = (props.orders || []).filter(o => o.buildingId === props.building.id && (o.status === 1 || o.status === 2 || o.status === 3))
  const roomKeySet = new Set<string>()
  for (const ord of buildingOrders) {
    const f = ord.floor || 1
    const r = ord.room || ''
    if (r) roomKeySet.add(`${f}-${r}`)
  }
  return roomKeySet.size
})

const totalDoneRoomsCount = computed(() => {
  const buildingOrders = (props.orders || []).filter(o => o.buildingId === props.building.id)
  const roomKeySet = new Set<string>()
  for (const ord of buildingOrders) {
    const f = ord.floor || 1
    const r = ord.room || ''
    if (r) {
      const key = `${f}-${r}`
      const analysis = getRoomAnalysis(f, r)
      if (analysis.status === 'done') {
        roomKeySet.add(key)
      }
    }
  }
  return roomKeySet.size
})

// 查询某楼层某房间的活跃报修工单
function hasFloorFaults(floor: number): boolean {
  return getFloorStatus(floor) === 'fault'
}

function getRoomOrders(floor: number, roomNo: string) {
  return getRoomAnalysis(floor, roomNo).unfinished
}

// 当前选中房间关联的全部工单
const selectedRoomOrders = computed(() => {
  if (!activeSelectedRoom.value) return []
  const bOrders = (props.orders || []).filter(o => o.buildingId === props.building.id)
  return matchRoomOrders(bOrders, activeSelectedRoom.value.floor, activeSelectedRoom.value.no)
})

function setFloor(f: number) {
  if (mode.value === 'plan') {
    currentPlanFloor.value = f === 0 ? 1 : f
  } else {
    selectedFloor.value = f
    rebuild3dScene()
  }
}

function onRoomClick(floor: number, roomNo: string) {
  activeSelectedRoom.value = { floor, no: roomNo }
}

function statusName(status: number, text?: string) {
  if (text) return text
  switch (status) {
    case 1: return '待接单'
    case 2: return '已派工'
    case 3: return '维修中'
    case 4: return '已完成'
    case 5: return '已取消'
    default: return '待处理'
  }
}

// ---------- 2D 平面视口全景自适应、拖拽与缩放 ----------
const planBoxRef = ref<HTMLDivElement | null>(null)
const MIN_W = 30
let maxW = 300
const view = reactive({ x: 0, y: 0, w: PLAN_WIDTH, h: PLAN_DEPTH })
const viewBoxStr = computed(() => "" + view.x + " " + view.y + " " + view.w + " " + view.h)

let dragStart: { px: number; py: number; vx: number; vy: number } | null = null
const pointers = new Map<number, { x: number; y: number }>()

function fitToContainer() {
  // 针对宽 100，深 176 的狭长户型，设置舒适的全局全景视野
  // 预留足够留白（左右上下至少 30% 呼吸空间），彻底避免放大过度
  const boxH = PLAN_DEPTH * 1.35 // 约 237.6
  const boxW = Math.max(PLAN_WIDTH * 2.2, boxH * 1.3) // 宽屏舒适留白
  view.w = boxW
  view.h = boxH
  view.x = -(boxW - PLAN_WIDTH) / 2
  view.y = -(boxH - PLAN_DEPTH) / 2
  maxW = Math.max(500, boxW * 2)
}

function clampView() {
  view.w = Math.min(Math.max(view.w, MIN_W), maxW)
  const ratio = (planBoxRef.value && planBoxRef.value.clientWidth && planBoxRef.value.clientHeight)
    ? (planBoxRef.value.clientWidth / planBoxRef.value.clientHeight)
    : (PLAN_WIDTH / PLAN_DEPTH)
  view.h = view.w / ratio
}

function zoomAt(factor: number, cx = view.x + view.w / 2, cy = view.y + view.h / 2) {
  const newW = view.w * factor
  const clampedW = Math.min(Math.max(newW, MIN_W), maxW)
  const ratio = clampedW / view.w
  view.x = cx - (cx - view.x) * ratio
  view.y = cy - (cy - view.y) * ratio
  view.w = clampedW
  clampView()
}

function resetView() {
  fitToContainer()
}

function onWheel(e: WheelEvent) {
  const factor = e.deltaY > 0 ? 1.15 : 0.87
  zoomAt(factor)
}

function onPointerDown(e: PointerEvent) {
  const target = e.target as HTMLElement | SVGElement | null
  if (target && (target.classList?.contains('room-svg-shape') || target.closest?.('.room-svg-group'))) {
    return
  }
  const el = e.currentTarget as HTMLElement
  el.setPointerCapture?.(e.pointerId)
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  if (pointers.size === 1) {
    dragStart = { px: e.clientX, py: e.clientY, vx: view.x, vy: view.y }
  }
}

function onPointerMove(e: PointerEvent) {
  if (!pointers.has(e.pointerId)) return
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  if (pointers.size === 1 && dragStart) {
    const el = (e.currentTarget as HTMLElement).closest('.plan-viewport-box') as HTMLElement | null
    const scale = (el?.clientWidth || 700) / view.w
    const dx = (e.clientX - dragStart.px) / scale
    const dy = (e.clientY - dragStart.py) / scale
    view.x = dragStart.vx - dx
    view.y = dragStart.vy - dy
    clampView()
  }
}

function onPointerUp(e: PointerEvent) {
  pointers.delete(e.pointerId)
  if (pointers.size === 0) dragStart = null
}

function onSvgClick() {
  // 点击空白处保留选中
}

// ---------- 3D 数字沙盘（Three.js 浅色微水泥与 Beacon 粒子渲染） ----------
const mountRef = ref<HTMLDivElement | null>(null)
const webglError = ref(false)
const errorText = ref('')

let threeScene: any = null
let threeCamera: any = null
let threeRenderer: any = null
let threeControls: any = null
let animFrameId: number | null = null
let raycaster: any = null
let mouseVec: any = null
let resizeObserver: any = null
let selectionWireframe: any = null
const roomMeshList: any[] = []
const beaconMeshList: any[] = []
let pointerDownPos = { x: 0, y: 0 }

async function switch3d() {
  mode.value = '3d'
  webglError.value = false
  errorText.value = ''
  await nextTick()
  if (mountRef.value) {
    init3d()
  }
}

async function init3d() {
  if (!mountRef.value) return
  cleanupThree()

  try {
    const THREE = await import('three')
    const controlsModule = await import('three/examples/jsm/controls/OrbitControls.js')

    const el = mountRef.value
    const width = el.clientWidth || 760
    const height = el.clientHeight || 540

    // 1. 场景与高明度柔和天空浅灰底色
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0xf1f5fb)
    threeScene = scene

    // 2. 摄像机与渲染器
    const camera = new THREE.PerspectiveCamera(42, width / height, 1, 3000)
    threeCamera = camera
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.setSize(width, height)
    renderer.shadowMap.enabled = true
    renderer.shadowMap.type = THREE.PCFSoftShadowMap
    el.innerHTML = ''
    el.appendChild(renderer.domElement)
    threeRenderer = renderer

    raycaster = new THREE.Raycaster()
    mouseVec = new THREE.Vector2()

    // 3. 柔和环境光与暖柔平行光
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.95)
    scene.add(ambientLight)

    const hemiLight = new THREE.HemisphereLight(0xffffff, 0xdfe7f6, 0.6)
    hemiLight.position.set(0, 200, 0)
    scene.add(hemiLight)

    const dirLight = new THREE.DirectionalLight(0xffffff, 0.75)
    dirLight.position.set(120, 220, 140)
    dirLight.castShadow = true
    dirLight.shadow.mapSize.width = 1024
    dirLight.shadow.mapSize.height = 1024
    scene.add(dirLight)

    // 4. 轨道控制器
    const controls = new controlsModule.OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true
    controls.dampingFactor = 0.08
    controls.maxPolarAngle = Math.PI / 2.05
    threeControls = controls

    // 5. 构建立体建筑模型
    buildSceneGeometry(THREE)

    // 6. 初始视角调节
    reset3dCamera()

    // 7. 动画渲染循环
    const clock = new THREE.Clock()
    const animate = () => {
      animFrameId = requestAnimationFrame(animate)
      const elapsed = clock.getElapsedTime()

      // Beacon 浮动呼吸动画
      for (const b of beaconMeshList) {
        b.rotation.y = elapsed * 1.8
        b.position.y = b.userData.baseY + Math.sin(elapsed * 3.2) * 0.45
      }

      controls.update()
      renderer.render(scene, camera)
    }
    animate()

    window.addEventListener('resize', handleResize)
    if (typeof window.ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => {
        handleResize()
      })
      resizeObserver.observe(el)
    }
  } catch (err: any) {
    console.error('Three.js init error:', err)
    webglError.value = true
    errorText.value = String(err?.message || err)
  }
}

function handleResize() {
  if (!mountRef.value || !threeRenderer || !threeCamera) return
  const width = mountRef.value.clientWidth
  const height = mountRef.value.clientHeight
  if (width && height) {
    threeCamera.aspect = width / height
    threeCamera.updateProjectionMatrix()
    threeRenderer.setSize(width, height)
  }
}

function buildSceneGeometry(THREE: any) {
  if (!threeScene) return
  roomMeshList.length = 0
  beaconMeshList.length = 0

  const boxW = 120
  const boxD = (boxW * PLAN_DEPTH) / PLAN_WIDTH
  const roomHeight = 15.0
  const floorGap = 5.0
  const floorStep = roomHeight + floorGap

  // 1. 基底柔和晶格底板
  const groundGeo = new THREE.PlaneGeometry(boxW * 2.4, boxD * 2.0)
  const groundMat = new THREE.MeshStandardMaterial({
    color: 0xf1f5f9,
    roughness: 0.8,
    metalness: 0.1,
    side: THREE.DoubleSide
  })
  const ground = new THREE.Mesh(groundGeo, groundMat)
  ground.rotation.x = -Math.PI / 2
  ground.position.y = -0.1
  ground.receiveShadow = true
  threeScene.add(ground)

  const grid = new THREE.GridHelper(Math.max(boxW, boxD) * 2.2, 22, 0xdbeafe, 0xe2e8f0)
  grid.position.y = 0
  threeScene.add(grid)

  const totalFloors = Math.max(props.building.floors, 1)

  // 2. BIM 精致微晶材质定义 (彻底消除多余粗笨灰板，升级为半透微晶体块与建筑工程轮廓线)
  const corridorMat = new THREE.MeshStandardMaterial({
    color: 0xf8fafc,
    roughness: 0.6,
    metalness: 0.05,
    transparent: true,
    opacity: 0.75
  })

  const stairMat = new THREE.MeshStandardMaterial({
    color: 0xe0e7ff,
    roughness: 0.5,
    metalness: 0.1,
    transparent: true,
    opacity: 0.7
  })

  const stairEdgeMat = new THREE.LineBasicMaterial({
    color: 0x818cf8,
    linewidth: 1,
    transparent: true,
    opacity: 0.6
  })

  const normalRoomMat = new THREE.MeshStandardMaterial({
    color: 0x93c5fd,
    roughness: 0.2,
    metalness: 0.15,
    transparent: true,
    opacity: 0.55
  })

  const normalEdgeMat = new THREE.LineBasicMaterial({
    color: 0x3b82f6,
    linewidth: 1,
    transparent: true,
    opacity: 0.7
  })

  const faultRoomMat = new THREE.MeshStandardMaterial({
    color: 0xf43f5e,
    emissive: 0xe11d48,
    emissiveIntensity: 0.55,
    roughness: 0.15,
    metalness: 0.1,
    transparent: true,
    opacity: 0.82
  })

  const faultEdgeMat = new THREE.LineBasicMaterial({
    color: 0xffe4e6,
    linewidth: 1.5
  })

  const doneRoomMat = new THREE.MeshStandardMaterial({
    color: 0x10b981,
    emissive: 0x059669,
    emissiveIntensity: 0.35,
    roughness: 0.25,
    metalness: 0.1,
    transparent: true,
    opacity: 0.68
  })

  const doneEdgeMat = new THREE.LineBasicMaterial({
    color: 0x6ee7b7,
    linewidth: 1.2
  })

  for (let f = 1; f <= totalFloors; f++) {
    const isIsolated = selectedFloor.value > 0
    if (isIsolated && selectedFloor.value !== f) {
      continue
    }

    const yBase = (f - 1) * floorStep
    const p = resolveFloorPlan(f, props.building.roomsPerFloor, props.building.layoutJson)

    // A. 轻量化通廊地面
    if (p.corridor.w > 0) {
      const cw = (p.corridor.w / PLAN_WIDTH) * boxW
      const cd = (p.corridor.d / PLAN_DEPTH) * boxD
      const corridor = new THREE.Mesh(new THREE.BoxGeometry(cw, 0.15, cd), corridorMat)
      corridor.position.set(0, yBase + 0.08, 0)
      corridor.receiveShadow = true
      threeScene.add(corridor)
    }

    // B. 楼梯间/核心筒 (带 BIM 细线边框)
    for (const core of p.cores) {
      const cw = (32 / PLAN_WIDTH) * boxW
      const cd = (core.d / PLAN_DEPTH) * boxD
      const cz = ((core.z + core.d / 2 - PLAN_DEPTH / 2) / PLAN_DEPTH) * boxD
      const stairGeo = new THREE.BoxGeometry(cw, roomHeight * 0.9, cd)
      const stair = new THREE.Mesh(stairGeo, stairMat)
      stair.position.set(((16 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW, yBase + roomHeight * 0.45, cz)
      stair.castShadow = true
      stair.receiveShadow = true
      const stairEdges = new THREE.LineSegments(new THREE.EdgesGeometry(stairGeo), stairEdgeMat)
      stair.add(stairEdges)
      threeScene.add(stair)
    }

    // C. 房间微晶体块 (带精密 BIM 轮廓线与待修发光标牌)
    for (const room of p.rooms) {
      const w = (room.w / PLAN_WIDTH) * boxW * 0.94
      const d = (room.d / PLAN_DEPTH) * boxD * 0.94
      const x = ((room.x + room.w / 2 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW
      const z = ((room.z + room.d / 2 - PLAN_DEPTH / 2) / PLAN_DEPTH) * boxD

      const roomAnalysis = getRoomAnalysis(f, room.no)
      const hasFault = roomAnalysis.status === 'fault'
      const isDone = roomAnalysis.status === 'done'

      const roomGeo = new THREE.BoxGeometry(w, roomHeight, d)
      const roomMesh = new THREE.Mesh(
        roomGeo,
        hasFault ? faultRoomMat : (isDone ? doneRoomMat : normalRoomMat)
      )
      roomMesh.position.set(x, yBase + roomHeight / 2, z)
      roomMesh.castShadow = true
      roomMesh.receiveShadow = true

      // BIM 精密轮廓描边
      const edgeWire = new THREE.LineSegments(
        new THREE.EdgesGeometry(roomGeo),
        hasFault ? faultEdgeMat : (isDone ? doneEdgeMat : normalEdgeMat)
      )
      roomMesh.add(edgeWire)

      roomMesh.userData = {
        floor: f,
        roomNo: room.no,
        hasFault,
        isDone,
        faultCount: roomAnalysis.unfinished.length,
        doneCount: roomAnalysis.completed.length
      }

      threeScene.add(roomMesh)
      roomMeshList.push(roomMesh)

      // 3D 浮空房间号标牌 (Sprite 标牌)
      const spriteCanvas = document.createElement('canvas')
      spriteCanvas.width = 128
      spriteCanvas.height = 64
      const sCtx = spriteCanvas.getContext('2d')
      // 只有在逐层查看(isIsolated)或者待修房间才显示房间号标牌，全景不杂乱
      if (sCtx) {
        let bgColor = 'rgba(15, 23, 42, 0.8)'
        let strokeColor = '#93c5fd'
        if (hasFault) {
          bgColor = 'rgba(244, 63, 94, 0.92)'
          strokeColor = '#ffe4e6'
        } else if (isDone) {
          bgColor = 'rgba(16, 185, 129, 0.92)'
          strokeColor = '#dcfce7'
        }
        sCtx.fillStyle = bgColor
        sCtx.beginPath()
        sCtx.roundRect(14, 8, 100, 48, 10)
        sCtx.fill()
        sCtx.lineWidth = 2.5
        sCtx.strokeStyle = strokeColor
        sCtx.stroke()
        sCtx.font = 'bold 26px -apple-system, sans-serif'
        sCtx.fillStyle = '#ffffff'
        sCtx.textAlign = 'center'
        sCtx.textBaseline = 'middle'
        sCtx.fillText(room.no, 64, 32)
        const spriteTex = new THREE.CanvasTexture(spriteCanvas)
        const spriteMat = new THREE.SpriteMaterial({
          map: spriteTex,
          transparent: true,
          depthTest: true,
          depthWrite: false
        })
        const roomSprite = new THREE.Sprite(spriteMat)
        roomSprite.scale.set(9.5, 4.8, 1)
        roomSprite.position.set(x, yBase + roomHeight + 2.4, z)
        roomSprite.userData = {
          floor: f,
          roomNo: room.no,
          hasFault,
          isDone,
          faultCount: roomAnalysis.unfinished.length,
          doneCount: roomAnalysis.completed.length
        }
        threeScene.add(roomSprite)
        roomMeshList.push(roomSprite)
      }

      // 待修房间：悬浮旋转水晶八面体 Beacon 光标
      if (hasFault) {
        const beaconGeo = new THREE.OctahedronGeometry(1.5, 0)
        const beaconMat = new THREE.MeshStandardMaterial({
          color: 0xff1744,
          emissive: 0xff1744,
          emissiveIntensity: 1.0,
          roughness: 0.1
        })
        const beacon = new THREE.Mesh(beaconGeo, beaconMat)
        const baseY = yBase + roomHeight + 2.2
        beacon.position.set(x, baseY, z)
        beacon.userData = {
          baseY,
          floor: f,
          roomNo: room.no,
          hasFault,
          isDone,
          faultCount: roomAnalysis.unfinished.length,
          doneCount: roomAnalysis.completed.length
        }
        threeScene.add(beacon)
        beaconMeshList.push(beacon)
        roomMeshList.push(beacon)
      }
    }
  }
  update3dSelection(THREE)
}

function update3dSelection(THREE_INST?: any) {
  if (!threeScene) return
  if (selectionWireframe) {
    threeScene.remove(selectionWireframe)
    if (selectionWireframe.geometry) selectionWireframe.geometry.dispose?.()
    if (selectionWireframe.material) selectionWireframe.material.dispose?.()
    selectionWireframe = null
  }
  if (!activeSelectedRoom.value) return

  const targetMesh = roomMeshList.find(m =>
    m.isMesh &&
    m.geometry?.type === 'BoxGeometry' &&
    m.userData?.roomNo === activeSelectedRoom.value?.no &&
    m.userData?.floor === activeSelectedRoom.value?.floor
  )
  if (!targetMesh) return

  const createBox = (THREE: any) => {
    const wireGeo = new THREE.EdgesGeometry(targetMesh.geometry)
    const wireMat = new THREE.LineBasicMaterial({
      color: 0x2563eb,
      linewidth: 3,
      depthTest: false,
      transparent: true,
      opacity: 0.95
    })
    selectionWireframe = new THREE.LineSegments(wireGeo, wireMat)
    selectionWireframe.position.copy(targetMesh.position)
    selectionWireframe.scale.set(1.04, 1.04, 1.04)
    selectionWireframe.renderOrder = 999
    threeScene.add(selectionWireframe)
  }

  if (THREE_INST) {
    createBox(THREE_INST)
  } else {
    import('three').then((THREE) => createBox(THREE))
  }
}

watch(activeSelectedRoom, () => {
  if (mode.value === '3d') {
    update3dSelection()
  }
})

// 监听建筑切换，彻底重置视图与状态，杜绝跨楼残留
watch(() => props.building.id, () => {
  currentPlanFloor.value = 1
  selectedFloor.value = 0
  activeSelectedRoom.value = null
  hoverRoomInfo.value = null
  resetView()
  if (mode.value === '3d') {
    rebuild3dScene()
    reset3dCamera()
  }
})

function rebuild3dScene() {
  if (!threeScene) return
  import('three').then((THREE) => {
    const toRemove = threeScene.children.filter((child: any) =>
      child.type !== 'DirectionalLight' && child.type !== 'AmbientLight' && child.type !== 'HemisphereLight'
    )
    toRemove.forEach((obj) => {
      threeScene.remove(obj)
      if (obj.geometry) obj.geometry.dispose?.()
      if (obj.material) {
        if (Array.isArray(obj.material)) {
          obj.material.forEach((m: any) => { m.map?.dispose?.(); m.dispose?.() })
        } else {
          obj.material.map?.dispose?.()
          obj.material.dispose?.()
        }
      }
    })

    buildSceneGeometry(THREE)
  })
}

function reset3dCamera() {
  if (!threeCamera || !threeControls) return
  const boxW = 120
  const boxD = (boxW * PLAN_DEPTH) / PLAN_WIDTH
  const totalFloors = Math.max(props.building.floors, 1)

  const roomHeight = 15.0
  const floorGap = 5.0
  const floorStep = roomHeight + floorGap

  if (selectedFloor.value === 0) {
    threeCamera.position.set(boxW * 1.5, floorStep * totalFloors + 60, boxD * 1.5)
    threeControls.target.set(0, (floorStep * totalFloors) / 2, 0)
  } else {
    const targetY = (selectedFloor.value - 1) * floorStep + roomHeight / 2
    threeCamera.position.set(boxW * 1.3, targetY + 45, boxD * 1.3)
    threeControls.target.set(0, targetY, 0)
  }
  threeControls.update()
}

function onThreePointerDown(e: PointerEvent) {
  pointerDownPos = { x: e.clientX, y: e.clientY }
}

function onThreePointerUp(e: PointerEvent) {
  const dist = Math.hypot(e.clientX - pointerDownPos.x, e.clientY - pointerDownPos.y)
  if (dist > 12) return

  if (!mountRef.value || !threeCamera || !raycaster || !threeRenderer) return
  const dom = threeRenderer.domElement || mountRef.value
  const rect = dom.getBoundingClientRect()
  mouseVec.x = ((e.clientX - rect.left) / rect.width) * 2 - 1
  mouseVec.y = -((e.clientY - rect.top) / rect.height) * 2 + 1

  raycaster.setFromCamera(mouseVec, threeCamera)
  const intersects = raycaster.intersectObjects(roomMeshList, false)

  if (intersects.length > 0) {
    let obj: any = intersects[0].object
    let data = obj.userData
    while (obj && (!data || !data.roomNo) && obj.parent) {
      obj = obj.parent
      data = obj.userData
    }
    if (data && data.roomNo) {
      activeSelectedRoom.value = {
        floor: data.floor,
        no: data.roomNo
      }
    }
  }
}

function onThreePointerMove(e: PointerEvent) {
  if (!mountRef.value || !threeCamera || !raycaster || !threeRenderer) return
  const dom = threeRenderer.domElement || mountRef.value
  const rect = dom.getBoundingClientRect()
  mouseVec.x = ((e.clientX - rect.left) / rect.width) * 2 - 1
  mouseVec.y = -((e.clientY - rect.top) / rect.height) * 2 + 1

  raycaster.setFromCamera(mouseVec, threeCamera)
  const intersects = raycaster.intersectObjects(roomMeshList, false)

  if (intersects.length > 0) {
    mountRef.value.style.cursor = 'pointer'
    let obj: any = intersects[0].object
    let data = obj.userData
    while (obj && (!data || !data.roomNo) && obj.parent) {
      obj = obj.parent
      data = obj.userData
    }
    if (data && data.roomNo) {
      hoverRoomInfo.value = {
        x: e.clientX - rect.left + 15,
        y: e.clientY - rect.top - 20,
        floor: data.floor,
        roomNo: data.roomNo,
        faultCount: data.faultCount || 0,
        doneCount: data.doneCount || 0
      }
    }
  } else {
    mountRef.value.style.cursor = 'default'
    hoverRoomInfo.value = null
  }
}

function cleanupThree() {
  if (animFrameId) {
    cancelAnimationFrame(animFrameId)
    animFrameId = null
  }
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  if (selectionWireframe) {
    if (selectionWireframe.geometry) selectionWireframe.geometry.dispose?.()
    if (selectionWireframe.material) selectionWireframe.material.dispose?.()
    selectionWireframe = null
  }
  window.removeEventListener('resize', handleResize)
  if (threeControls) {
    threeControls.dispose()
    threeControls = null
  }
  if (threeRenderer) {
    threeRenderer.dispose()
    threeRenderer = null
  }
  threeScene = null
  threeCamera = null
  roomMeshList.length = 0
  beaconMeshList.length = 0
  if (mountRef.value) {
    mountRef.value.innerHTML = ''
  }
}

onMounted(() => {
  nextTick(() => {
    fitToContainer()
  })
  window.addEventListener('resize', fitToContainer)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', fitToContainer)
  cleanupThree()
})
</script>
<style scoped>
.admin-visual-root {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  background: #f8fafc;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
  color: #1e293b;
  overflow: hidden;
}

/* 顶部控制栏 */
.visual-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 18px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid #e2e8f0;
  gap: 16px;
  flex-shrink: 0;
}

.mode-switch-group {
  display: flex;
  background: #edf2f7;
  padding: 3px;
  border-radius: 999px;
  border: 1px solid #e2e8f0;
}

.mode-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  background: transparent;
  border: none;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.mode-btn:hover {
  color: #1e293b;
}

.mode-btn.active {
  background: #ffffff;
  color: #2563eb;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.12);
}

.btn-icon {
  font-size: 14px;
}

.floor-filter-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
}

.filter-title {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  white-space: nowrap;
}

.floor-chip-scroller {
  display: flex;
  gap: 6px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.floor-chip-scroller::-webkit-scrollbar {
  height: 3px;
}

.floor-chip-btn {
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 999px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.floor-chip-btn:hover {
  background: #e2e8f0;
  color: #0f172a;
}


.floor-chip-btn.has-fault {
  border-color: #fca5a5;
  background: #fef2f2;
  color: #e11d48;
  font-weight: 700;
}

.floor-chip-btn.has-fault:hover {
  background: #fee2e2;
  border-color: #f87171;
  color: #be123c;
}

.floor-chip-btn.has-fault.active {
  background: #ef4444;
  border-color: #dc2626;
  color: #ffffff;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.35);
}

.chip-fault-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  background-color: #ef4444;
  border-radius: 50%;
  margin-left: 3px;
  vertical-align: middle;
  box-shadow: 0 0 4px rgba(239, 68, 68, 0.6);
}

.chip-done-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  background-color: #10b981;
  border-radius: 50%;
  margin-left: 3px;
  vertical-align: middle;
  box-shadow: 0 0 4px rgba(16, 185, 129, 0.6);
}

.done-badge {
  background: #ecfdf5 !important;
  border-color: #a7f3d0 !important;
  color: #047857 !important;
  margin-left: 8px;
}

.done-dot {
  width: 8px;
  height: 8px;
  background-color: #10b981;
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.5);
}

.legend-tag.done .done-dot-legend {
  background: #10b981;
  border: 1px solid #059669;
}

.tt-desc.done {
  color: #10b981;
  font-weight: 600;
}

.floor-chip-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.25);
}

.header-badges {
  display: flex;
  align-items: center;
}

.stat-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 12px;
  background: #f1f5f9;
  color: #64748b;
  border: 1px solid #e2e8f0;
}

.stat-badge.has-faults {
  background: #fff1f2;
  border-color: #fecdd3;
  color: #e11d48;
}

.pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #10b981;
}

.stat-badge.has-faults .pulse-dot {
  background: #f43f5e;
  box-shadow: 0 0 0 3px rgba(244, 63, 94, 0.25);
  animation: pulse-dot 1.8s infinite;
}

@keyframes pulse-dot {
  0% { transform: scale(0.95); opacity: 0.8; }
  50% { transform: scale(1.3); opacity: 1; }
  100% { transform: scale(0.95); opacity: 0.8; }
}

/* 主展示区域 */
.visual-stage-container {
  flex: 1;
  display: flex;
  position: relative;
  overflow: hidden;
  height: calc(100% - 53px);
}

.canvas-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  height: 100%;
  overflow: hidden;
}

.panel-sub-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.7);
  border-bottom: 1px solid #eef2f6;
  font-size: 12px;
  color: #64748b;
  z-index: 5;
}

.sub-hint {
  font-size: 12px;
}

.zoom-action-group, .three-legend-box {
  display: flex;
  align-items: center;
  gap: 8px;
}

.square-tool-btn {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  font-size: 14px;
  font-weight: 700;
  color: #334155;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.square-tool-btn:hover {
  background: #f1f5f9;
  border-color: #94a3b8;
}

.text-tool-btn {
  padding: 4px 10px;
  font-size: 12px;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  color: #334155;
  cursor: pointer;
}

.text-tool-btn:hover {
  background: #f1f5f9;
}

.three-legend-box {
  margin-right: 12px;
}

.legend-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
}

.color-dot {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  display: inline-block;
}

.legend-tag.normal .color-dot {
  background: #93c5fd;
  border: 1px solid #60a5fa;
}

.legend-tag.fault .beacon-dot {
  background: #f43f5e;
  box-shadow: 0 0 6px #f43f5e;
}

/* 2D 平面视口 */
.plan-viewport-box {
  flex: 1;
  width: 100%;
  height: 100%;
  background: #f1f5f9;
  touch-action: none;
  cursor: grab;
  position: relative;
  overflow: hidden;
}

.plan-viewport-box:active {
  cursor: grabbing;
}

.plan-svg-canvas {
  width: 100%;
  height: 100%;
  display: block;
}

.room-svg-shape {
  cursor: pointer;
  transition: fill 0.2s, stroke 0.2s;
}

.room-svg-shape:hover {
  fill: #e0f2fe !important;
  stroke: #2563eb !important;
}

.fault-radar-dot {
  transform-box: fill-box;
  transform-origin: center;
  animation: radar-pulse 2s infinite ease-out;
}

@keyframes radar-pulse {
  0% { transform: scale(1); opacity: 1; }
  60% { transform: scale(1.6); opacity: 0.7; }
  100% { transform: scale(1); opacity: 1; }
}

/* 3D 沙盘挂载容器 */
.three-canvas-mount {
  flex: 1;
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.three-tooltip {
  position: absolute;
  pointer-events: none;
  background: rgba(15, 23, 42, 0.85);
  color: #ffffff;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 12px;
  box-shadow: 0 8px 20px rgba(0,0,0,0.25);
  z-index: 10;
  backdrop-filter: blur(6px);
  transform: translate(0, 0);
  white-space: nowrap;
}

.tt-room {
  font-weight: 700;
  margin-bottom: 2px;
}

.tt-desc {
  color: #94a3b8;
}

.tt-desc.alert {
  color: #fb7185;
  font-weight: 600;
}

.three-error-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.9);
  color: #dc2626;
  font-size: 14px;
}

/* 右侧毛玻璃工单抽屉 */
.room-order-drawer {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 340px;
  max-width: 85vw;
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-left: 1px solid rgba(226, 232, 240, 0.8);
  display: flex;
  flex-direction: column;
  box-shadow: -8px 0 24px rgba(30, 41, 59, 0.08);
  z-index: 40;
  animation: slide-in 0.25s ease-out;
}

@keyframes slide-in {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.drawer-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid #eef2f6;
}

.drawer-room-badge {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.room-code {
  font-size: 20px;
  font-weight: 800;
  color: #1e3a8a;
}

.room-floor-pill {
  font-size: 12px;
  color: #64748b;
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 999px;
}

.drawer-close-circle {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #f1f5f9;
  border: none;
  font-size: 14px;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.drawer-close-circle:hover {
  background: #fee2e2;
  color: #ef4444;
}

.drawer-scroll-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.orders-summary-bar {
  padding: 8px 12px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  background: #f0fdf4;
  color: #15803d;
  border: 1px solid #bbf7d0;
}

.orders-summary-bar.warning-status {
  background: #fff1f2;
  color: #e11d48;
  border-color: #fecdd3;
}

.order-list-cards {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.order-item-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 12px 14px;
  box-shadow: 0 2px 6px rgba(0,0,0,0.03);
  transition: all 0.2s;
}

.order-item-card:hover {
  border-color: #93c5fd;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.08);
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.ord-no {
  font-size: 12px;
  font-weight: 700;
  color: #475569;
}

.ord-status {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 600;
}

.status-1, .status-2 {
  background: #fef3c7;
  color: #d97706;
}

.status-3 {
  background: #dbeafe;
  color: #2563eb;
}

.status-4 {
  background: #dcfce7;
  color: #16a34a;
}

.card-fault {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 4px;
}

.fault-icon {
  font-size: 13px;
}

.card-desc {
  font-size: 12px;
  color: #64748b;
  line-height: 1.5;
  margin-bottom: 8px;
  background: #f8fafc;
  padding: 6px 8px;
  border-radius: 6px;
}

.card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: #94a3b8;
}

.card-priority {
  color: #ef4444;
  font-weight: 700;
  background: #fee2e2;
  padding: 1px 6px;
  border-radius: 4px;
}

.empty-room-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 10px;
  text-align: center;
}

.empty-glass-icon {
  font-size: 38px;
  margin-bottom: 12px;
  opacity: 0.8;
}

.empty-title {
  font-size: 14px;
  font-weight: 700;
  color: #334155;
  margin-bottom: 4px;
}

.empty-sub {
  font-size: 12px;
  color: #94a3b8;
}
/* ---------- V9.7.4 视觉柔化：圆角与柔和过渡（仅本组件内元素） ---------- */
[class*="panel"],
[class*="card"],
[class*="toolbar"] {
  border-radius: 16px;
}

:deep(.el-button),
:deep(.el-input__wrapper),
:deep(.el-select__wrapper),
:deep(.el-textarea__inner) {
  border-radius: 10px;
}

:deep(.el-tag) {
  border-radius: 8px;
}

:deep(.el-dialog),
:deep(.el-drawer),
:deep(.el-card) {
  border-radius: 18px;
}

:deep(.el-table) {
  border-radius: 12px;
  overflow: hidden;
}

:deep(.el-button) {
  transition: transform 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}
/* ---------- V9.7.4 移动端：头部控制条换行，楼层切片独占一行可横向滚动 ----------
   原因：单行 flex 下 .floor-filter-wrap 带 overflow:hidden，窄屏被模式切换与徽标挤到 0 宽，
   导致 2D/3D 的楼层按钮全部被裁掉（只能看第一层）。 */
@media (max-width: 767px) {
  .visual-header-bar {
    flex-wrap: wrap;
    gap: 8px 12px;
    padding: 8px 12px;
  }

  .mode-btn {
    padding: 6px 10px;
    font-size: 12px;
  }

  .header-badges {
    order: 2;
    margin-left: auto;
  }

  .floor-filter-wrap {
    order: 3;
    flex: 0 0 100%;
    overflow: visible;
    gap: 6px;
  }

  .filter-title {
    font-size: 11px;
  }

  .floor-chip-scroller {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    padding-bottom: 4px;
  }

  .floor-chip-btn {
    padding: 6px 12px;
    font-size: 12px;
  }
}
</style>
