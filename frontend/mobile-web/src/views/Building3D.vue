<template>
  <van-popup v-model:show="visible" position="bottom" round :style="{ height: '88vh', display: 'flex', flexDirection: 'column' }">
    <div class="head">
      <div class="title-wrap">
        <div class="title">{{ building?.name || '楼栋可视化' }}</div>
        <div class="head-meta" v-if="building">
          {{ building.floors }} 层 · 每层 {{ building.roomsPerFloor }} 间 · 
          <span class="meta-fault" v-if="buildingStats.faults > 0">待维修 {{ buildingStats.faults }} 间</span>
          <span class="meta-done" v-if="buildingStats.dones > 0">已修好 {{ buildingStats.dones }} 间</span>
          <span class="meta-clean" v-if="buildingStats.faults === 0 && buildingStats.dones === 0">状态正常</span>
        </div>
      </div>
      <button class="close" @click="visible = false">✕</button>
    </div>

    <div class="toolbar" v-if="building">
      <div class="mode-chips">
        <button class="chip" :class="{ active: mode === 'plan' }" @click="setMode('plan')">楼层户型</button>
        <button class="chip" :class="{ active: mode === '3d' }" @click="setMode('3d')">3D 沙盘</button>
      </div>
      <template v-if="mode === '3d'">
        <div class="tool-actions">
          <button class="chip mini" :class="{ active: transparent }" @click="toggleTransparent">透视</button>
          <button class="chip mini" :class="{ active: showLabels }" @click="toggleLabels">房间号</button>
          <button class="chip mini" :class="{ active: rotating }" @click="toggleRotate">自动旋转</button>
          <button class="chip mini" @click="reset3dCamera">复位视角</button>
        </div>
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
      <div class="stage-container">
        <div ref="mountRef" class="three-mount"></div>
        <p v-if="webglError" class="fallback">
          3D 初始化未完成：{{ loadError || '浏览器未启用 WebGL' }}<br />可切回“楼层户型”查看。
        </p>

        <!-- 悬浮选中的房间维修卡片（工人交互支持开工/完工） -->
        <transition name="slide-up">
          <div v-if="selected" class="room-sheet">
            <div class="sheet-head">
              <div class="sheet-title-group">
                <span class="room-pill">{{ selected.floorNo }}层 {{ selected.roomNo }}室</span>
                <span v-if="selected.unfinished.length > 0" class="badge-tag fault">待维修 ({{ selected.unfinished.length }})</span>
                <span v-else-if="selected.completed.length > 0" class="badge-tag done">已修好 ({{ selected.completed.length }})</span>
                <span v-else class="badge-tag normal">无报修工单</span>
              </div>
              <button class="sheet-close" @click="clearSelection">✕</button>
            </div>

            <div class="sheet-body">
              <div v-if="selected.orders.length === 0" class="sheet-empty">
                该房间当前没有派发给你的维修工单
              </div>
              <div v-else class="order-scroll-list">
                <div v-for="o in selected.orders" :key="o.id" class="order-card-item">
                  <div class="order-top">
                    <span class="type-tag">{{ o.faultTypeName || '报修' }}</span>
                    <span class="status-tag" :class="'status-' + o.status">{{ statusText(o.status, o.statusText) }}</span>
                    <span class="order-time">{{ formatTime(o.createdAt) }}</span>
                  </div>
                  <div class="order-title">{{ o.title }}</div>
                  <div class="order-desc" v-if="o.description">{{ o.description }}</div>
                  
                  <div class="order-footer">
                    <span class="order-no">单号: {{ o.orderNo }}</span>
                    <div class="action-btns">
                      <button
                        v-if="o.status === 1 || o.status === 2"
                        class="act-btn primary"
                        :disabled="actingId === o.id"
                        @click="runAction(o, 'start')"
                      >
                        {{ actingId === o.id ? '开工中…' : '开始维修' }}
                      </button>
                      <button
                        v-if="o.status === 3"
                        class="act-btn success"
                        :disabled="actingId === o.id"
                        @click="runAction(o, 'complete')"
                      >
                        {{ actingId === o.id ? '提交中…' : '确认完工' }}
                      </button>
                      <span v-if="o.status === 4" class="act-badge-done">✓ 已完工</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <div class="floor-bar" v-if="building && building.floors > 1">
        <button
          class="floor-chip"
          :class="{ active: floorFilter === 0 }"
          @click="selectFloor(0)"
        >
          全部
        </button>
        <button
          v-for="f in building.floors"
          :key="f"
          class="floor-chip"
          :class="{
            active: floorFilter === f,
            'has-fault': getFloorFaultStatus(f) === 'fault',
            'has-done': getFloorFaultStatus(f) === 'done'
          }"
          @click="selectFloor(f)"
        >
          {{ f }}F
          <i v-if="getFloorFaultStatus(f) === 'fault'" class="chip-dot fault"></i>
          <i v-else-if="getFloorFaultStatus(f) === 'done'" class="chip-dot done"></i>
        </button>
      </div>

      <div class="bottom-bar">
        <div class="legend">
          <span class="legend-item"><i class="dot fault"></i> 待处理故障</span>
          <span class="legend-item"><i class="dot done"></i> 已修好房间</span>
          <span class="legend-item"><i class="dot normal"></i> 普通房间</span>
        </div>
        <div class="tip">单指滑动旋转 · 双指缩放 · 点击房间查看详情与工单</div>
      </div>
    </template>
  </van-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onUnmounted } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'
import { apiCompleteOrder } from '../api'
import { startOrderFlow } from '../utils/startOrderFlow'
import BuildingFloorPlan from './BuildingFloorPlan.vue'
import { PLAN_DEPTH, PLAN_WIDTH, matchRoomOrders } from '../utils/floorLayout'
import { resolveFloorPlan } from '../utils/layoutGrid'
import { showConfirmDialog, showToast } from 'vant'

const props = defineProps<{ building: WorkerMapBuilding | null; orders: OrderItem[] }>()
const visible = defineModel<boolean>({ default: false })
const emit = defineEmits<{ (e: 'refresh'): void }>()

const mountRef = ref<HTMLDivElement | null>(null)
const webglError = ref(false)
const loadError = ref('')
const transparent = ref(false)
const showLabels = ref(true)
const rotating = ref(false)
const mode = ref<'plan' | '3d'>('3d')
const floorFilter = ref(0)
const actingId = ref<number | null>(null)

interface SelectedRoomData {
  roomNo: string
  floorNo: number
  orders: OrderItem[]
  unfinished: OrderItem[]
  completed: OrderItem[]
}
const selected = ref<SelectedRoomData | null>(null)

// Three.js 状态变量
let threeScene: any = null
let threeCamera: any = null
let threeRenderer: any = null
let threeControls: any = null
let animFrameId = 0
let raycaster: any = null
let mouseVec: any = null
let pointerDownPos = { x: 0, y: 0 }
let selectionWireframe: any = null
let resizeObserver: ResizeObserver | null = null

const roomMeshList: any[] = []
const beaconMeshList: any[] = []
const spriteMeshList: any[] = []

function getRoomAnalysis(floor: number, roomNo: string) {
  if (!props.building) return { unfinished: [], completed: [], status: 'normal' as const, all: [] }
  const bOrders = (props.orders || []).filter((o) => o.buildingId === props.building?.id)
  const matched = matchRoomOrders(bOrders, floor, roomNo)
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

function getFloorFaultStatus(floor: number): 'fault' | 'done' | 'normal' {
  if (!props.building) return 'normal'
  const bOrders = (props.orders || []).filter(
    (o) => o.buildingId === props.building?.id && (o.floor || 1) === floor,
  )
  if (bOrders.some((o) => o.status === 1 || o.status === 2 || o.status === 3)) {
    return 'fault'
  }
  if (bOrders.some((o) => o.status === 4)) {
    return 'done'
  }
  return 'normal'
}

const buildingStats = computed(() => {
  if (!props.building) return { faults: 0, dones: 0 }
  const bOrders = (props.orders || []).filter((o) => o.buildingId === props.building?.id)
  const faultRooms = new Set<string>()
  const doneRooms = new Set<string>()
  for (const o of bOrders) {
    const key = o.floor + '-' + (o.room || '')
    if (o.status === 1 || o.status === 2 || o.status === 3) {
      faultRooms.add(key)
    } else if (o.status === 4) {
      doneRooms.add(key)
    }
  }
  for (const k of faultRooms) {
    doneRooms.delete(k)
  }
  return { faults: faultRooms.size, dones: doneRooms.size }
})

watch(
  () => props.orders,
  () => {
    if (visible.value && mode.value === '3d' && threeScene) {
      rebuild3dScene()
      if (selected.value) {
        const analysis = getRoomAnalysis(selected.value.floorNo, selected.value.roomNo)
        selected.value = {
          roomNo: selected.value.roomNo,
          floorNo: selected.value.floorNo,
          orders: analysis.all,
          unfinished: analysis.unfinished,
          completed: analysis.completed,
        }
      }
    }
  },
  { deep: true },
)

watch(
  () => visible.value,
  async (open) => {
    if (open) {
      floorFilter.value = 0
      selected.value = null
      await nextTick()
      if (mode.value === '3d') {
        init3d()
      }
    } else {
      cleanupThree()
    }
  },
)

watch(
  () => props.building?.id,
  () => {
    floorFilter.value = 0
    selected.value = null
    if (visible.value && mode.value === '3d') {
      rebuild3dScene()
      reset3dCamera()
    }
  },
)

async function init3d() {
  if (!mountRef.value || !props.building) return
  cleanupThree()

  try {
    const THREE = await import('three')
    const controlsModule = await import('three/examples/jsm/controls/OrbitControls.js')

    const el = mountRef.value
    const width = el.clientWidth || window.innerWidth || 390
    const height = el.clientHeight || 340

    // 1. 场景与高雅柔和 BIM 浅灰蓝天空底色
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0xf1f5fb)
    threeScene = scene

    // 2. 摄像机与渲染器
    const camera = new THREE.PerspectiveCamera(42, width / height, 1, 3000)
    threeCamera = camera
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
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
    controls.autoRotate = rotating.value
    controls.autoRotateSpeed = 1.2
    threeControls = controls

    // 5. 构建立体建筑模型
    buildSceneGeometry(THREE)

    // 6. 初始视角
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

    const dom = renderer.domElement
    dom.addEventListener('pointerdown', onThreePointerDown)
    dom.addEventListener('pointerup', onThreePointerUp)

    window.addEventListener('resize', handleResize)
    if (typeof window.ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => {
        handleResize()
      })
      resizeObserver.observe(el)
    }
  } catch (err: any) {
    console.error('Three.js mobile init error:', err)
    webglError.value = true
    loadError.value = String(err?.message || err)
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
  if (!threeScene || !props.building) return
  roomMeshList.length = 0
  beaconMeshList.length = 0
  spriteMeshList.length = 0

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
    side: THREE.DoubleSide,
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

  // 2. BIM 精致微晶材质定义
  const corridorMat = new THREE.MeshStandardMaterial({
    color: 0xf8fafc,
    roughness: 0.6,
    metalness: 0.05,
    transparent: true,
    opacity: 0.75,
  })

  const stairMat = new THREE.MeshStandardMaterial({
    color: 0xe0e7ff,
    roughness: 0.5,
    metalness: 0.1,
    transparent: true,
    opacity: 0.7,
  })

  const stairEdgeMat = new THREE.LineBasicMaterial({
    color: 0x818cf8,
    linewidth: 1,
    transparent: true,
    opacity: 0.6,
  })

  const normalRoomMat = new THREE.MeshStandardMaterial({
    color: 0x93c5fd,
    roughness: 0.2,
    metalness: 0.15,
    transparent: true,
    opacity: transparent.value ? 0.22 : 0.55,
  })

  const normalEdgeMat = new THREE.LineBasicMaterial({
    color: 0x3b82f6,
    linewidth: 1,
    transparent: true,
    opacity: 0.7,
  })

  const faultRoomMat = new THREE.MeshStandardMaterial({
    color: 0xf43f5e,
    emissive: 0xe11d48,
    emissiveIntensity: 0.55,
    roughness: 0.15,
    metalness: 0.1,
    transparent: true,
    opacity: 0.85,
  })

  const faultEdgeMat = new THREE.LineBasicMaterial({
    color: 0xffe4e6,
    linewidth: 1.5,
  })

  const doneRoomMat = new THREE.MeshStandardMaterial({
    color: 0x10b981,
    emissive: 0x059669,
    emissiveIntensity: 0.35,
    roughness: 0.25,
    metalness: 0.1,
    transparent: true,
    opacity: 0.72,
  })

  const doneEdgeMat = new THREE.LineBasicMaterial({
    color: 0x6ee7b7,
    linewidth: 1.2,
  })

  for (let f = 1; f <= totalFloors; f++) {
    const isIsolated = floorFilter.value > 0
    if (isIsolated && floorFilter.value !== f) {
      continue
    }

    const yBase = (f - 1) * floorStep
    const p = resolveFloorPlan(f, props.building.roomsPerFloor, props.building.layoutJson)

    // A. 贯通走廊地面
    if (p.corridor && p.corridor.w > 0) {
      const cw = (p.corridor.w / PLAN_WIDTH) * boxW
      const cd = (p.corridor.d / PLAN_DEPTH) * boxD
      const corridor = new THREE.Mesh(new THREE.BoxGeometry(cw, 0.15, cd), corridorMat)
      corridor.position.set(0, yBase + 0.08, 0)
      corridor.receiveShadow = true
      threeScene.add(corridor)
    }

    // B. 核心筒/楼梯间
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

    // C. 房间微晶体块
    for (const room of p.rooms) {
      const w = (room.w / PLAN_WIDTH) * boxW * 0.94
      const d = (room.d / PLAN_DEPTH) * boxD * 0.94
      const x = ((room.x + room.w / 2 - PLAN_WIDTH / 2) / PLAN_WIDTH) * boxW
      const z = ((room.z + room.d / 2 - PLAN_DEPTH / 2) / PLAN_DEPTH) * boxD

      const analysis = getRoomAnalysis(f, room.no)
      const hasFault = analysis.status === 'fault'
      const isDone = analysis.status === 'done'

      const roomGeo = new THREE.BoxGeometry(w, roomHeight, d)
      const roomMesh = new THREE.Mesh(
        roomGeo,
        hasFault ? faultRoomMat : (isDone ? doneRoomMat : normalRoomMat),
      )
      roomMesh.position.set(x, yBase + roomHeight / 2, z)
      roomMesh.castShadow = true
      roomMesh.receiveShadow = true

      const edgeWire = new THREE.LineSegments(
        new THREE.EdgesGeometry(roomGeo),
        hasFault ? faultEdgeMat : (isDone ? doneEdgeMat : normalEdgeMat),
      )
      roomMesh.add(edgeWire)

      roomMesh.userData = {
        floor: f,
        roomNo: room.no,
        hasFault,
        isDone,
        orders: analysis.all,
        unfinished: analysis.unfinished,
        completed: analysis.completed,
      }

      threeScene.add(roomMesh)
      roomMeshList.push(roomMesh)

      // 3D 浮空房间号标牌 (Canvas Sprite)
      const shouldShowLabel = showLabels.value || isIsolated || hasFault || isDone
      if (shouldShowLabel) {
        const spriteCanvas = document.createElement('canvas')
        spriteCanvas.width = 128
        spriteCanvas.height = 64
        const sCtx = spriteCanvas.getContext('2d')
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

          sCtx.font = 'bold 26px -apple-system, BlinkMacSystemFont, sans-serif'
          sCtx.fillStyle = '#ffffff'
          sCtx.textAlign = 'center'
          sCtx.textBaseline = 'middle'
          sCtx.fillText(room.no, 64, 32)

          const spriteTex = new THREE.CanvasTexture(spriteCanvas)
          const spriteMat = new THREE.SpriteMaterial({
            map: spriteTex,
            transparent: true,
            depthTest: true,
            depthWrite: false,
          })
          const roomSprite = new THREE.Sprite(spriteMat)
          roomSprite.scale.set(9.5, 4.8, 1)
          roomSprite.position.set(x, yBase + roomHeight + 2.4, z)
          roomSprite.userData = roomMesh.userData
          threeScene.add(roomSprite)
          roomMeshList.push(roomSprite)
          spriteMeshList.push(roomSprite)
        }
      }

      // 待修房间：悬浮旋转水晶八面体 Beacon 光标
      if (hasFault) {
        const beaconGeo = new THREE.OctahedronGeometry(1.5, 0)
        const beaconMat = new THREE.MeshStandardMaterial({
          color: 0xff1744,
          emissive: 0xff1744,
          emissiveIntensity: 1.0,
          roughness: 0.1,
        })
        const beacon = new THREE.Mesh(beaconGeo, beaconMat)
        const baseY = yBase + roomHeight + 2.2
        beacon.position.set(x, baseY, z)
        beacon.userData = {
          baseY,
          ...roomMesh.userData,
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
  if (!selected.value) return

  const targetMesh = roomMeshList.find(
    (m) =>
      m.isMesh &&
      m.geometry?.type === 'BoxGeometry' &&
      m.userData?.roomNo === selected.value?.roomNo &&
      m.userData?.floor === selected.value?.floorNo,
  )
  if (!targetMesh) return

  const createBox = (THREE: any) => {
    const wireGeo = new THREE.EdgesGeometry(targetMesh.geometry)
    const wireMat = new THREE.LineBasicMaterial({
      color: 0x2563eb,
      linewidth: 3,
      depthTest: false,
      transparent: true,
      opacity: 0.95,
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

function rebuild3dScene() {
  if (!threeScene) return
  import('three').then((THREE) => {
    const toRemove = threeScene.children.filter(
      (child: any) =>
        child.type !== 'DirectionalLight' &&
        child.type !== 'AmbientLight' &&
        child.type !== 'HemisphereLight',
    )
    toRemove.forEach((obj: any) => {
      threeScene.remove(obj)
      if (obj.geometry) obj.geometry.dispose?.()
      if (obj.material) {
        if (Array.isArray(obj.material)) {
          obj.material.forEach((m: any) => {
            m.map?.dispose?.()
            m.dispose?.()
          })
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
  if (!threeCamera || !threeControls || !props.building) return
  const boxW = 120
  const boxD = (boxW * PLAN_DEPTH) / PLAN_WIDTH
  const totalFloors = Math.max(props.building.floors, 1)

  const roomHeight = 15.0
  const floorGap = 5.0
  const floorStep = roomHeight + floorGap

  if (floorFilter.value === 0) {
    threeCamera.position.set(boxW * 1.5, floorStep * totalFloors + 60, boxD * 1.5)
    threeControls.target.set(0, (floorStep * totalFloors) / 2, 0)
  } else {
    const targetY = (floorFilter.value - 1) * floorStep + roomHeight / 2
    threeCamera.position.set(boxW * 1.3, targetY + 45, boxD * 1.3)
    threeControls.target.set(0, targetY, 0)
  }
  threeControls.update()
}

function selectFloor(floor: number) {
  floorFilter.value = floor
  // 按照要求：切换查看楼层不重置摄像机视角
  rebuild3dScene()
}

function toggleTransparent() {
  transparent.value = !transparent.value
  rebuild3dScene()
}

function toggleLabels() {
  showLabels.value = !showLabels.value
  rebuild3dScene()
}

function toggleRotate() {
  rotating.value = !rotating.value
  if (threeControls) {
    threeControls.autoRotate = rotating.value
  }
}

function clearSelection() {
  selected.value = null
  update3dSelection()
}

function onThreePointerDown(e: PointerEvent) {
  pointerDownPos = { x: e.clientX, y: e.clientY }
}

function onThreePointerUp(e: PointerEvent) {
  const dist = Math.hypot(e.clientX - pointerDownPos.x, e.clientY - pointerDownPos.y)
  if (dist > 10) return

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
      selected.value = {
        roomNo: data.roomNo,
        floorNo: data.floor,
        orders: data.orders || [],
        unfinished: data.unfinished || [],
        completed: data.completed || [],
      }
      update3dSelection()
    }
  }
}

async function runAction(o: OrderItem, action: 'start' | 'complete') {
  try {
    await showConfirmDialog({
      title: action === 'start' ? '确认开工' : '确认完工',
      message: action === 'start' ? `开始维修 ${o.title}（${o.orderNo}）？` : `完成 ${o.title}（${o.orderNo}）？`,
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

async function setMode(next: 'plan' | '3d') {
  if (mode.value === next) return
  mode.value = next
  selected.value = null
  if (next === '3d') {
    await nextTick()
    if (visible.value && props.building && mountRef.value) {
      await init3d()
    }
  } else {
    cleanupThree()
  }
}

function onPlanSelect(room: { num: string; floor: number; orders: OrderItem[] }) {
  const analysis = getRoomAnalysis(room.floor, room.num)
  selected.value = {
    roomNo: room.num,
    floorNo: room.floor,
    orders: analysis.all,
    unfinished: analysis.unfinished,
    completed: analysis.completed,
  }
}

function statusText(status: number, defaultText?: string) {
  if (defaultText) return defaultText
  switch (status) {
    case 1:
      return '待接单'
    case 2:
      return '已派工'
    case 3:
      return '维修中'
    case 4:
      return '已完成'
    case 5:
      return '已取消'
    default:
      return '待处理'
  }
}

function formatTime(str?: string) {
  if (!str) return ''
  return str.replace('T', ' ').substring(5, 16)
}

function cleanupThree() {
  if (animFrameId) {
    cancelAnimationFrame(animFrameId)
    animFrameId = 0
  }
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  window.removeEventListener('resize', handleResize)

  if (threeRenderer) {
    const dom = threeRenderer.domElement
    if (dom) {
      dom.removeEventListener('pointerdown', onThreePointerDown)
      dom.removeEventListener('pointerup', onThreePointerUp)
    }
    threeRenderer.dispose()
    threeRenderer = null
  }
  if (threeControls) {
    threeControls.dispose()
    threeControls = null
  }
  threeScene = null
  threeCamera = null
  selectionWireframe = null
  roomMeshList.length = 0
  beaconMeshList.length = 0
  spriteMeshList.length = 0
}

onUnmounted(() => {
  cleanupThree()
})

defineExpose({ cleanupThree, reset3dCamera })
</script>

<style scoped>
.head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  padding: 14px 16px 8px;
  background: #ffffff;
  border-bottom: 1px solid #f1f5f9;
}
.title-wrap {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.title {
  font-size: 17px;
  font-weight: 800;
  color: #1e3a8a;
  letter-spacing: -0.2px;
}
.head-meta {
  color: #64748b;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.meta-fault {
  color: #e11d48;
  font-weight: 600;
  background: #fee2e2;
  padding: 1px 6px;
  border-radius: 4px;
}
.meta-done {
  color: #059669;
  font-weight: 600;
  background: #d1fae5;
  padding: 1px 6px;
  border-radius: 4px;
}
.meta-clean {
  color: #2563eb;
  font-weight: 600;
}
.close {
  width: 28px;
  height: 28px;
  color: #64748b;
  background: #f1f5f9;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 16px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  overflow-x: auto;
}
.mode-chips, .tool-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}
.chip {
  padding: 5px 12px;
  color: #475569;
  font-size: 12px;
  font-weight: 600;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.chip.active {
  color: #ffffff;
  background: #2563eb;
  border-color: #2563eb;
  box-shadow: 0 4px 10px rgba(37, 99, 235, 0.2);
}
.chip.mini {
  padding: 4px 8px;
  font-size: 11px;
  border-radius: 6px;
}
.stage-container {
  position: relative;
  flex: 1;
  min-height: 280px;
  background: #f1f5fb;
  overflow: hidden;
}
.three-mount {
  width: 100%;
  height: 100%;
  min-height: 280px;
  touch-action: none;
}
.fallback {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  padding: 16px;
  color: #64748b;
  font-size: 13px;
  text-align: center;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 8px;
}

/* 浮动选中的房间工单交互卡片 */
.room-sheet {
  position: absolute;
  bottom: 10px;
  left: 12px;
  right: 12px;
  max-height: 48%;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid #cbd5e1;
  border-radius: 14px;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.16);
  display: flex;
  flex-direction: column;
  z-index: 20;
  overflow: hidden;
}
.sheet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px 8px;
  border-bottom: 1px solid #f1f5f9;
}
.sheet-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.room-pill {
  font-size: 15px;
  font-weight: 800;
  color: #0f172a;
}
.badge-tag {
  font-size: 11px;
  padding: 2px 7px;
  border-radius: 4px;
  font-weight: 700;
}
.badge-tag.fault {
  background: #fee2e2;
  color: #e11d48;
}
.badge-tag.done {
  background: #d1fae5;
  color: #059669;
}
.badge-tag.normal {
  background: #f1f5f9;
  color: #64748b;
}
.sheet-close {
  width: 24px;
  height: 24px;
  border: none;
  background: #f1f5f9;
  border-radius: 50%;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}
.sheet-body {
  padding: 8px 12px;
  overflow-y: auto;
  max-height: 180px;
}
.sheet-empty {
  padding: 14px 0;
  color: #94a3b8;
  font-size: 13px;
  text-align: center;
}
.order-scroll-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.order-card-item {
  padding: 8px 10px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}
.order-top {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}
.type-tag {
  font-size: 11px;
  font-weight: 700;
  color: #2563eb;
  background: #eff6ff;
  padding: 1px 6px;
  border-radius: 4px;
}
.status-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
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
  background: #d1fae5;
  color: #059669;
}
.order-time {
  margin-left: auto;
  font-size: 11px;
  color: #94a3b8;
}
.order-title {
  font-size: 13px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 2px;
}
.order-desc {
  font-size: 12px;
  color: #64748b;
  line-height: 1.4;
  margin-bottom: 6px;
}
.order-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px dashed #e2e8f0;
}
.order-no {
  font-size: 11px;
  color: #94a3b8;
}
.action-btns {
  display: flex;
  align-items: center;
  gap: 6px;
}
.act-btn {
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}
.act-btn.primary {
  background: #2563eb;
  color: #ffffff;
}
.act-btn.success {
  background: #10b981;
  color: #ffffff;
}
.act-badge-done {
  font-size: 12px;
  font-weight: 700;
  color: #059669;
}

/* 楼层切换栏 */
.floor-bar {
  display: flex;
  gap: 6px;
  padding: 8px 16px;
  background: #ffffff;
  border-top: 1px solid #f1f5f9;
  border-bottom: 1px solid #f1f5f9;
  overflow-x: auto;
}
.floor-chip {
  position: relative;
  flex: 0 0 auto;
  padding: 5px 12px;
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.floor-chip.active {
  color: #ffffff;
  background: #2563eb;
  border-color: #2563eb;
}
.floor-chip.has-fault {
  border-color: #fecdd3;
  color: #e11d48;
}
.floor-chip.has-fault.active {
  background: #e11d48;
  border-color: #e11d48;
  color: #ffffff;
}
.floor-chip.has-done {
  border-color: #a7f3d0;
  color: #059669;
}
.floor-chip.has-done.active {
  background: #059669;
  border-color: #059669;
  color: #ffffff;
}
.chip-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.chip-dot.fault {
  background: #e11d48;
}
.chip-dot.done {
  background: #10b981;
}

/* 底部图例与提示 */
.bottom-bar {
  padding: 6px 16px 10px;
  background: #ffffff;
}
.legend {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  font-size: 12px;
  color: #64748b;
  margin-bottom: 4px;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.dot.fault {
  background: #f43f5e;
  box-shadow: 0 0 6px rgba(244, 63, 94, 0.6);
}
.dot.done {
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.6);
}
.dot.normal {
  background: #93c5fd;
}
.tip {
  color: #94a3b8;
  font-size: 11px;
  text-align: center;
}

/* 动画过渡 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
