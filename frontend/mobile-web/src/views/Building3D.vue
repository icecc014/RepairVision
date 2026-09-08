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
      <button class="chip" :class="{ active: transparent }" @click="toggleTransparent">透视</button>
      <button class="chip" :class="{ active: showLabels }" @click="toggleLabels">房间号</button>
      <button class="chip" :class="{ active: rotating }" @click="toggleRotate">自动旋转</button>
    </div>

    <div ref="mountRef" class="three-mount"></div>
    <p v-if="webglError" class="fallback">当前设备不支持 3D，已降级为楼层文本：{{ building?.floors }} 层。</p>

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
  </van-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'

const props = defineProps<{ building: WorkerMapBuilding | null; orders: OrderItem[] }>()
const visible = defineModel<boolean>({ default: false })
const mountRef = ref<HTMLDivElement | null>(null)
const webglError = ref(false)
const transparent = ref(true)
const showLabels = ref(false)
const rotating = ref(false)
const floorFilter = ref(0)

let cancel = 0
let renderer: { dispose: () => void } | null = null
let controls: any = null
let floorGroups: any[] = []
let markers: any[] = []
let glassMats: any[] = []
let animateFn: ((t: number) => void) | null = null

const activeCount = computed(() => {
  if (!props.building) return 0
  return props.orders.filter((o) => o.buildingId === props.building?.id).length
})

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
    if (!visible.value || !props.building) return
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
  const cols = Math.min(rooms, 10)
  const rows = Math.ceil(rooms / cols)
  return { cols, rows }
}

async function initScene() {
  if (!mountRef.value || !props.building) return
  if (cancel) cancelAnimationFrame(cancel)
  mountRef.value.innerHTML = ''
  floorGroups = []
  markers = []
  glassMats = []
  controls = null
  animateFn = null
  try {
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

      const { cols, rows } = layout(Math.max(b.roomsPerFloor, 1))
      const gap = 0.5
      const usableW = boxW - gap * (cols - 1)
      const usableD = boxD - gap * (rows - 1)
      const roomW = usableW / cols
      const roomD = usableD / rows
      const roomH = floorH * 0.72

      for (let i = 0; i < Math.max(b.roomsPerFloor, 1); i++) {
        const col = i % cols
        const row = Math.floor(i / cols)
        const cx = -boxW / 2 + roomW / 2 + col * (roomW + gap)
        const cz = -boxD / 2 + roomD / 2 + row * (roomD + gap)
        const mat = new THREE.MeshLambertMaterial({
          color: colors[(floor + i) % colors.length],
          transparent: true,
          opacity: 0.9,
        })
        const roomBox = new THREE.Mesh(new THREE.BoxGeometry(roomW, roomH, roomD), mat)
        roomBox.position.set(cx, yBase + slabH + roomH / 2, cz)
        group.add(roomBox)

        const edges = new THREE.EdgesGeometry(new THREE.BoxGeometry(roomW, roomH, roomD))
        const line = new THREE.LineSegments(
          edges,
          new THREE.LineBasicMaterial({ color: '#0f2557', transparent: true, opacity: 0.35 }),
        )
        line.position.copy(roomBox.position)
        group.add(line)

        const roomNum = `${floorNo}${String(i + 1).padStart(2, '0')}`
        const label = makeLabel(THREE, roomNum, 1.2, cx, yBase + slabH + roomH + 0.2, cz + roomD / 2 + 0.06)
        if (label) {
          label.userData = { roomNum, floorNo }
          group.add(label)
        }

        const orders = orderByFloorRoom.get(`${floorNo}:${roomNum}`) || []
        const ordersByRawRoom = orderByFloorRoom.get(`${floorNo}:${String(i + 1).padStart(2, '0')}`) || []
        const matched = orders.length > 0 ? orders : ordersByRawRoom
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
    ;(window as any).__building3d = {
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
    (window as any).__building3d = helper3d
    helper3d.updateLabels()
    helper3d.updateTransparent()  } catch (err) {
    console.error('3D init failed:', err)
    webglError.value = true
    if (renderer) {
      renderer.dispose()
      renderer = null
    }
  }
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