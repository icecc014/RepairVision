<template>
  <van-popup v-model:show="visible" position="bottom" round :style="{ height: '82vh' }">
    <div class="head">
      <div class="title">{{ building?.name || '楼栋 3D' }}</div>
      <div class="head-meta" v-if="building">
        {{ building.code }} · {{ building.floors }} 层 · 每层 {{ building.roomsPerFloor }} 间 · 活动 {{ activeCount }}
      </div>
      <button class="close" @click="visible = false">✕</button>
    </div>

    <div class="toolbar" v-if="building">
      <button class="chip" :class="{ active: mode === 'plan' }" @click="setMode('plan')">楼层户型</button>
      <button class="chip" :class="{ active: mode === '3d' }" @click="setMode('3d')">3D 立体</button>
      <template v-if="mode === '3d'">
        <button class="chip" :class="{ active: transparent }" @click="toggleTransparent">透视</button>
        <button class="chip" :class="{ active: showLabels }" @click="toggleLabels">房间号</button>
        <button class="chip" :class="{ active: rotating }" @click="toggleRotate">自动旋转</button>
      </template>
    </div>

    <BuildingFloorPlan
      v-if="mode === 'plan' && building"
      :building="building"
      :orders="orders"
      @select-room="onPlanSelect"
      @refresh="emit('refresh')"
    />

    <template v-if="mode === '3d'">
      <div ref="mountRef" class="three-mount"></div>
      <p v-if="webglError" class="fallback">3D 初始化未完成：{{ loadError || '浏览器未启用 WebGL' }}<br />可切回“楼层户型”查看。</p>
      <div class="floor-bar" v-if="building && building.floors > 1">
        <button class="floor-chip" :class="{ active: floorFilter === 0 }" @click="selectFloor(0)">全部</button>
        <button
          v-for="f in building.floors"
          :key="f"
          class="floor-chip"
          :class="{ active: floorFilter === f }"
          @click="selectFloor(f)"
        >
          {{ f }}F
        </button>
      </div>
      <div class="legend"><i class="dot fault"></i> 待处理故障</div>
      <p class="tip">手指拖拽旋转 · 双指缩放 · 点击楼层只看该层</p>
    </template>

    <div v-if="selected" class="room-panel">
      <div class="room-panel-head">
        <div>
          <span class="room-no">{{ selected.roomNo }}</span>
          <span class="room-floor">{{ selected.floorNo }} 层 · {{ selected.orders.length }} 个待处理工单</span>
        </div>
        <button class="panel-close" @click="selected = null">关闭</button>
      </div>
      <div v-if="selected.orders.length === 0" class="panel-empty">该房间暂无工单</div>
      <div v-for="o in selected.orders" :key="o.id" class="room-order">
        <div class="order-row">
          <span class="type">{{ o.faultTypeName }}</span>
          <span class="status" :class="'s' + o.status">{{ o.statusText }}</span>
        </div>
        <div class="order-title">{{ o.title }}</div>
        <div class="order-meta">报修 {{ o.createdAt }} · 工人 {{ o.workerName || '待派' }}</div>
        <div class="order-actions">
          <button v-if="o.status === 2" class="mini" @click="runAction(o, 'start')">开工</button>
          <button v-if="o.status === 3" class="mini done" @click="runAction(o, 'complete')">完工</button>
        </div>
      </div>
    </div>

  </van-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'
import { apiCompleteOrder, apiStartOrder } from '../api'
import BuildingFloorPlan from './BuildingFloorPlan.vue'
import { PLAN_DEPTH, PLAN_WIDTH, buildFloorPlan, buildGridRooms, supportsCorridorLayout } from '../utils/floorLayout'
import { showConfirmDialog, showToast } from 'vant'

const props = defineProps<{ building: WorkerMapBuilding | null; orders: OrderItem[] }>()
const visible = defineModel<boolean>({ default: false })
const emit = defineEmits<{ (e: 'refresh'): void }>()
const mountRef = ref<HTMLDivElement | null>(null)
const webglError = ref(false)
const loadError = ref('')
const transparent = ref(true)
const showLabels = ref(false)
const rotating = ref(false)
const mode = ref<'plan' | '3d'>('plan')
const floorFilter = ref(0)
const selected = ref<{ roomNo: string; floorNo: number; orders: OrderItem[] } | null>(null)

let cancel = 0
let renderer: { dispose: () => void } | null = null
let controls: any = null
let floorGroups: any[] = []
let markers: any[] = []
let roomMeshes: any[] = []
let allRoomOrders = new Map<string, OrderItem[]>()
let glassMats: any[] = []
let animateFn: ((t: number) => void) | null = null

const activeCount = computed(() => {
  if (!props.building) return 0
  return props.orders.filter((o) => o.buildingId === props.building?.id).length
})

watch(
  () => props.orders,
  async () => {
    if (!visible.value || !props.building || mode.value !== '3d') return
    await nextTick()
    await initScene()
  },
)
watch(
  () => visible.value,
  async (open) => {
    if (!open || !props.building) return
    floorFilter.value = 0
    await nextTick()
    await initScene()
  },
)

watch(
  () => props.building,
  async () => {
    if (!visible.value || !props.building || mode.value !== '3d') return
    await nextTick()
    await initScene()
  },
)

function textTexture(THREE: any, text: string, bg = 'rgba(15,23,42,0.88)') {
  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 64
  const ctx = canvas.getContext('2d')
  if (!ctx) return null
  ctx.clearRect(0, 0, 256, 64)
  ctx.fillStyle = bg
  ctx.fillRect(4, 4, 248, 56)
  ctx.strokeStyle = '#93c5fd'
  ctx.lineWidth = 2
  ctx.strokeRect(4, 4, 248, 56)
  ctx.fillStyle = '#ffffff'
  ctx.font = 'bold 30px sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(text, 128, 34)
  const tex = new THREE.CanvasTexture(canvas)
  tex.needsUpdate = true
  return tex
}

function makeLabel(THREE: any, text: string, width: number, x: number, y: number, z: number, rotY = 0) {
  const tex = textTexture(THREE, text)
  if (!tex) return
  const plane = new THREE.Mesh(
    new THREE.PlaneGeometry(width, width * 0.28),
    new THREE.MeshBasicMaterial({ map: tex, transparent: true }),
  )
  plane.position.set(x, y, z)
  plane.rotation.y = rotY
  return plane
}

function layout(rooms: number) {
  const cols = Math.min(rooms, 8)
  const rows = Math.ceil(rooms / cols)
  return { cols, rows }
}

async function initScene() {
  if (!mountRef.value || !props.building) return
  if (cancel) cancelAnimationFrame(cancel)
  mountRef.value.innerHTML = ''
  floorGroups = []
  markers = []
  roomMeshes = []
  allRoomOrders.clear()
  glassMats = []
  controls = null
  animateFn = null
  try {
    const testCanvas = document.createElement('canvas')
    const testGL = testCanvas.getContext('webgl2') || testCanvas.getContext('webgl')
    if (!testGL) {
      webglError.value = true
      loadError.value = '浏览器未启用 WebGL'
      return
    }
    const THREE = await import('three')
    const width = mountRef.value.clientWidth || 390
    const height = 320
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x0b1e45)
    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 2000)
    const render = new THREE.WebGLRenderer({ antialias: true })
    renderer = render
    render.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
    render.setSize(width, height)
    mountRef.value.appendChild(render.domElement)

    const b = props.building
    const floorCount = Math.max(b.floors, 1)
    const scale = 2.4
    const boxW = Math.max(b.width, 20) * scale
    const boxD = Math.max(b.height, 20) * scale
    const floorH = 3.1
    const slabH = 0.16
    const totalH = floorCount * (floorH + slabH)

    camera.position.set(boxW * 1.15, totalH * 0.95, boxD * 1.45)
    scene.add(new THREE.AmbientLight(0xffffff, 0.95))
    const dir = new THREE.DirectionalLight(0xffffff, 0.85)
    dir.position.set(45, 100, 30)
    scene.add(dir)
    const fill = new THREE.DirectionalLight(0xbfdbfe, 0.5)
    fill.position.set(-40, 40, -60)
    scene.add(fill)

    const ground = new THREE.Mesh(
      new THREE.PlaneGeometry(boxW * 3.2, boxD * 3.2),
      new THREE.MeshLambertMaterial({ color: '#12264e', side: THREE.DoubleSide }),
    )
    ground.rotation.x = -Math.PI / 2
    ground.position.y = -0.35
    scene.add(ground)
    const grid = new THREE.GridHelper(Math.max(boxW, boxD) * 2.8, 14, 0x2f6be0, 0x1d3f8f)
    grid.position.y = -0.3
    scene.add(grid)

    const controlsModule = await import('three/examples/jsm/controls/OrbitControls.js')
    controls = new controlsModule.OrbitControls(camera, render.domElement)
    controls.enableDamping = true
    controls.autoRotate = rotating.value
    controls.maxPolarAngle = Math.PI / 2.02
    controls.minDistance = boxW * 0.5
    controls.maxDistance = boxW * 5
    controls.target.set(0, totalH * 0.5, 0)
    controls.update()

    const colors = ['#3b82f6', '#5b9bf8', '#73aefb', '#8cc0fd']
    const roomMatByFloor: THREE.Material[] = []

    const orderByFloorRoom = new Map<string, OrderItem[]>()
    for (const o of props.orders) {
      if (o.buildingId !== b.id) continue
      const key = `${o.floor}:${o.room || ''}`
      if (!orderByFloorRoom.has(key)) orderByFloorRoom.set(key, [])
      orderByFloorRoom.get(key)!.push(o)
    }

    for (let floor = 0; floor < floorCount; floor++) {
      const floorNo = floor + 1
      const group = new THREE.Group()
      const yBase = floor * (floorH + slabH)

      const slab = new THREE.Mesh(
        new THREE.BoxGeometry(boxW, slabH, boxD),
        new THREE.MeshLambertMaterial({ color: '#1e3a8a' }),
      )
      slab.position.y = yBase + slabH / 2
      group.add(slab)

      const plan = supportsCorridorLayout(b.roomsPerFloor)
        ? buildFloorPlan(floorNo, b.roomsPerFloor)
        : { floor: floorNo, rooms: buildGridRooms(floorNo, Math.max(b.roomsPerFloor, 1)), corridor: { x: 0, z: 0, w: 0, d: 0 }, cores: [] }
      const roomH = 0.8
      const tileColors = ['#3b82f6', '#60a5fa', '#7cb3f8', '#94c3fa', '#38bdf8']

      for (const room of plan.rooms) {
        const roomW = (room.w / PLAN_WIDTH) * boxW * 0.92
        const roomD = (room.d / PLAN_DEPTH) * boxD * 0.92
        const cx = ((room.x + room.w / 2 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW
        const cz = ((room.z + room.d / 2 - PLAN_DEPTH / 2) / PLAN_DEPTH) * boxD
        const roomNum = room.no
        const i = room.index - 1

        const tileMat = new THREE.MeshLambertMaterial({
          color: tileColors[(floor + i) % tileColors.length],
          side: THREE.DoubleSide,
          transparent: true,
          opacity: 0.92,
        })
        const tile = new THREE.Mesh(new THREE.PlaneGeometry(roomW * 0.94, roomD * 0.94), tileMat)
        tile.rotation.x = -Math.PI / 2
        tile.position.set(cx, yBase + slabH + 0.06, cz)
        tile.userData = { roomNo: roomNum, floorNo }
        group.add(tile)
        roomMeshes.push(tile)

        const outlineGeo = new THREE.EdgesGeometry(new THREE.BoxGeometry(roomW * 0.94, 0.04, roomD * 0.94))
        const outline = new THREE.LineSegments(
          outlineGeo,
          new THREE.LineBasicMaterial({ color: "#0f2557", transparent: true, opacity: 0.75 }),
        )
        outline.position.set(cx, yBase + slabH + 0.08, cz)
        group.add(outline)

        const label = makeLabel(THREE, roomNum, 1.1, cx, yBase + slabH + 0.85, cz)
        if (label) {
          label.userData = { roomNum, floorNo }
          group.add(label)
        }

        const orders = orderByFloorRoom.get(`${floorNo}:${roomNum}`) || []
        const suffix = roomNum.slice(String(floorNo).length)
        const ordersByRawRoom =
          orderByFloorRoom.get(`${floorNo}:${suffix}`) ||
          orderByFloorRoom.get(`${floorNo}:${String(i + 1).padStart(2, '0')}`) ||
          []
        const matched = orders.length > 0 ? orders : ordersByRawRoom
        allRoomOrders.set(`${floorNo}:${roomNum}`, matched)
        if (matched.length > 0) {
          const marker = new THREE.Mesh(
            new THREE.SphereGeometry(0.55, 18, 18),
            new THREE.MeshBasicMaterial({ color: 0xef4444 }),
          )
          marker.position.set(cx, yBase + slabH + roomH + 0.75, cz)
          marker.userData = { roomNum, floorNo, count: matched.length }
          markers.push(marker)
          group.add(marker)

          const active = makeLabel(THREE, `${roomNum} ${matched[0].faultTypeName || '维修'}`, 1.8, cx, yBase + slabH + roomH + 1.45, cz, 0)
          if (active) group.add(active)
        }
      }

      // 贯通走廊
      if (plan.corridor && plan.corridor.w > 0) {
        const corridorW = (plan.corridor.w / PLAN_WIDTH) * boxW
        const corridorD = (plan.corridor.d / PLAN_DEPTH) * boxD
        const corridorCx = ((plan.corridor.x + plan.corridor.w / 2 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW
        const corridor = new THREE.Mesh(
          new THREE.PlaneGeometry(corridorW * 0.96, corridorD * 0.98),
          new THREE.MeshLambertMaterial({ color: '#e2e8f0' }),
        )
        corridor.rotation.x = -Math.PI / 2
        corridor.position.set(corridorCx, yBase + slabH + 0.05, 0)
        group.add(corridor)
        const corridorLabel = makeLabel(THREE, '走廊', 1.6, corridorCx, yBase + slabH + 0.55, 0)
        if (corridorLabel) group.add(corridorLabel)
      }

      // 两处核心筒：封闭防火楼梯 + 公共区域
      for (const core of plan.cores) {
        const stairW = (32 / PLAN_WIDTH) * boxW
        const coreD = (core.d / PLAN_DEPTH) * boxD
        const coreCz = ((core.z + core.d / 2 - PLAN_DEPTH / 2) / PLAN_DEPTH) * boxD
        const stairX = ((16 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW
        const stair = new THREE.Mesh(
          new THREE.BoxGeometry(stairW * 0.96, floorH * 0.96, coreD * 0.96),
          new THREE.MeshLambertMaterial({ color: '#94a3b8', transparent: true, opacity: 0.75 }),
        )
        stair.position.set(stairX, yBase + slabH + floorH / 2, coreCz)
        group.add(stair)
        const stairLabel = makeLabel(THREE, '楼梯间', 1.4, stairX, yBase + slabH + floorH + 0.5, coreCz)
        if (stairLabel) group.add(stairLabel)

        const publicW = ((PLAN_WIDTH - 32) / PLAN_WIDTH) * boxW
        const publicCx = ((32 + (PLAN_WIDTH - 32) / 2 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW
        const publicArea = new THREE.Mesh(
          new THREE.PlaneGeometry(publicW * 0.95, coreD * 0.95),
          new THREE.MeshLambertMaterial({ color: '#dbeafe' }),
        )
        publicArea.rotation.x = -Math.PI / 2
        publicArea.position.set(publicCx, yBase + slabH + 0.06, coreCz)
        group.add(publicArea)
        const publicLabel = makeLabel(THREE, '公共区域', 1.8, publicCx, yBase + slabH + 0.55, coreCz)
        if (publicLabel) group.add(publicLabel)
      }

      const floorEdge = new THREE.EdgesGeometry(new THREE.BoxGeometry(boxW, 0.04, boxD))
      const lineMat = new THREE.LineBasicMaterial({ color: '#93c5fd', transparent: true, opacity: 0.45 })
      const outline = new THREE.LineSegments(floorEdge, lineMat)
      outline.position.y = yBase + slabH + 0.02
      group.add(outline)

      const glassMat = new THREE.MeshBasicMaterial({
        color: '#bfdbfe',
        transparent: true,
        opacity: transparent.value ? 0.05 : 0.28,
        side: THREE.DoubleSide,
      })
      glassMats.push(glassMat)
      const glass = new THREE.Mesh(new THREE.BoxGeometry(boxW + 0.25, floorH, boxD + 0.25), glassMat)
      glass.position.y = yBase + slabH + floorH / 2
      glass.renderOrder = 10
      group.add(glass)

      floorGroups.push(group)
      scene.add(group)
    }
    const namePlate = makeLabel(THREE, `${b.code} ${b.name}`, 4.2, 0, totalH + 1.0, 0)
    if (namePlate) {
      namePlate.position.set(0, totalH + 1.0, 0)
      scene.add(namePlate)
    }


    function applyFloorFilter() {
      for (let i = 0; i < floorGroups.length; i++) {
        const visibleFloor = floorFilter.value === 0 || floorFilter.value === i + 1
        floorGroups[i].visible = visibleFloor
      }
      markers.forEach((m) => {
        if (floorFilter.value === 0 || m.userData?.floorNo === floorFilter.value) {
          m.visible = true
        } else {
          m.visible = false
        }
      })
    }
    applyFloorFilter()

    const raycaster = new THREE.Raycaster()
    const pointer = new THREE.Vector2()
    const canvasEl = render.domElement
    canvasEl.style.touchAction = 'pan-y'
    canvasEl.addEventListener('click', (e: MouseEvent) => {
      const rect = canvasEl.getBoundingClientRect()
      pointer.x = ((e.clientX - rect.left) / rect.width) * 2 - 1
      pointer.y = -((e.clientY - rect.top) / rect.height) * 2 + 1
      raycaster.setFromCamera(pointer, camera)
      const targets = [...markers, ...roomMeshes]
      const hits = raycaster.intersectObjects(targets, false)
      if (hits.length > 0) {
        const obj = hits[0].object as any
        const roomNo = obj.userData?.roomNo as string | undefined
        const floorNo = obj.userData?.floorNo as number | undefined
        if (roomNo && floorNo) {
          selected.value = {
            roomNo,
            floorNo,
            orders: allRoomOrders.get(`${floorNo}:${roomNo}`) || [],
          }
        }
      } else {
        selected.value = null
      }
    })
    animateFn = (t: number) => {
      if (controls) {
        controls.autoRotate = rotating.value
        controls.update()
      }
      for (let i = 0; i < markers.length; i++) {
        const s = 1 + 0.18 * Math.sin(t / 190 + i * 0.8)
        markers[i].scale.set(s, s, s)
      }
      render.render(scene, camera)
      cancel = requestAnimationFrame(animateFn)
    }
    animateFn(0)

    // store methods
    const helper3d = {
      applyFloorFilter,
      updateTransparent() {
        glassMats.forEach((m: any) => {
          m.opacity = transparent.value ? 0.05 : 0.28
        })
      },
      updateLabels() {
        floorGroups.forEach((g: any) => {
          g.children.forEach((child: any) => {
            if (child.userData && child.userData.roomNum && child.geometry && child.geometry.type === 'PlaneGeometry') {
              child.visible = showLabels.value
            }
          })
        })
      },
    }
    ;(window as any).__building3d = helper3d
    helper3d.updateLabels()
    helper3d.updateTransparent()
  } catch (err) {
    console.error('3D init failed:', err)
    loadError.value = (err as Error)?.message || String(err)
    webglError.value = true
    if (renderer) {
      renderer.dispose()
      renderer = null
    }
  }
}

async function runAction(o: OrderItem, action: 'start' | 'complete') {
  try {
    await showConfirmDialog({
      title: action === 'start' ? '确认开工' : '确认完工',
      message: `${action === 'start' ? '开始维修' : '完成'} ${o.title}（${o.orderNo}）？`,
    })
  } catch {
    return
  }
  try {
    if (action === 'start') {
      await apiStartOrder(o.id)
      showToast('已开工')
    } else {
      await apiCompleteOrder(o.id)
      showToast('已完工')
    }
    selected.value = null
    emit('refresh')
  } catch (err) {
    showToast((err as Error).message)
  }
}
async function setMode(next: 'plan' | '3d') {
  if (mode.value === next) return
  mode.value = next
  selected.value = null
  if (next === '3d') {
    await nextTick()
    if (visible.value && props.building && mountRef.value) {
      await initScene()
    }
  } else {
    disposeScene()
  }
}

function onPlanSelect(room: { num: string; floor: number; orders: OrderItem[] }) {
  selected.value = { roomNo: room.num, floorNo: room.floor, orders: room.orders || [] }
}
function selectFloor(floor: number) {
  floorFilter.value = floor
  const helper = (window as any).__building3d
  if (helper) helper.applyFloorFilter()
}

function toggleTransparent() {
  transparent.value = !transparent.value
  const helper = (window as any).__building3d
  if (helper) helper.updateTransparent()
}

function toggleLabels() {
  showLabels.value = !showLabels.value
  const helper = (window as any).__building3d
  if (helper) helper.updateLabels()
}

function toggleRotate() {
  rotating.value = !rotating.value
  if (controls) controls.autoRotate = rotating.value
}

function disposeScene() {
  if (cancel) cancelAnimationFrame(cancel)
  cancel = 0
  if (renderer) {
    renderer.dispose()
    renderer = null
  }
  controls = null
  delete (window as any).__building3d
}

defineExpose({ disposeScene })
</script>

<style scoped>
.head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px 6px;
}
.title {
  font-size: 17px;
  font-weight: 800;
}
.head-meta {
  color: #64748b;
  font-size: 12px;
}
.close {
  margin-left: auto;
  width: 28px;
  height: 28px;
  color: #64748b;
  background: #f1f5f9;
  border: none;
  border-radius: 50%;
  cursor: pointer;
}
.toolbar {
  display: flex;
  gap: 8px;
  padding: 8px 16px;
  overflow-x: auto;
}
.chip {
  flex: 0 0 auto;
  padding: 6px 14px;
  color: #475569;
  font-size: 13px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  cursor: pointer;
}
.chip.active {
  color: #fff;
  background: #2563eb;
  border-color: #2563eb;
}
.three-mount {
  width: 100%;
  height: 320px;
}
.floor-bar {
  display: flex;
  gap: 6px;
  padding: 8px 16px 0;
  overflow-x: auto;
}
.floor-chip {
  flex: 0 0 auto;
  padding: 6px 12px;
  color: #64748b;
  font-size: 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
}
.floor-chip.active {
  color: #fff;
  background: #1d4ed8;
  border-color: #1d4ed8;
}
.fallback {
  padding: 20px;
  color: #475569;
  font-size: 13px;
}
.legend {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px 0;
  color: #94a3b8;
  font-size: 12px;
}
.dot.fault {
  width: 10px;
  height: 10px;
  background: #ef4444;
  border-radius: 50%;
}
.tip {
  margin: 4px 0 12px;
  color: #94a3b8;
  font-size: 12px;
  text-align: center;
}
</style>

.room-panel {
  margin: 10px 16px;
  padding: 12px 14px;
  background: #0f2557;
  border: 1px solid #3b5ca8;
  border-radius: 12px;
}
.room-panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.room-no {
  color: #fff;
  font-size: 18px;
  font-weight: 800;
}
.room-floor {
  margin-left: 8px;
  color: #93c5fd;
  font-size: 12px;
}
.panel-close {
  padding: 5px 12px;
  color: #fff;
  background: rgba(255,255,255,0.12);
  border: 1px solid rgba(255,255,255,0.25);
  border-radius: 999px;
  cursor: pointer;
}
.panel-empty {
  padding: 12px 0;
  color: #94a3b8;
  font-size: 13px;
}
.room-order {
  margin-top: 8px;
  padding: 8px 0;
  border-top: 1px dashed rgba(255,255,255,0.15);
}
.order-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.type {
  color: #bfdbfe;
  font-size: 12px;
  background: rgba(59,130,246,0.25);
  padding: 2px 8px;
  border-radius: 999px;
}
.status {
  font-size: 12px;
  font-weight: 600;
}
.status.s1, .status.s2 {
  color: #fbbf24;
}
.status.s3 {
  color: #60a5fa;
}
.status.s4 {
  color: #4ade80;
}
.status.s5 {
  color: #94a3b8;
}
.order-title {
  margin-top: 4px;
  color: #e2e8f0;
  font-size: 14px;
  font-weight: 600;
}
.order-meta {
  margin-top: 3px;
  color: #94a3b8;
  font-size: 12px;
}
.order-actions {
  margin-top: 6px;
  display: flex;
  gap: 8px;
}
.mini {
  padding: 4px 12px;
  color: #fff;
  font-size: 12px;
  background: #2563eb;
  border: none;
  border-radius: 999px;
  cursor: pointer;
}
.mini.done {
  background: #16a34a;
}
