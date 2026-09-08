<template>
  <van-popup v-model:show="visible" position="bottom" round :style="{ height: '72vh' }">
    <div class="head">
      <div class="title">{{ building?.name || '楼栋 3D' }}</div>
      <div class="head-meta" v-if="building">
        {{ building.code }} · {{ building.floors }} 层 · 每层 {{ building.roomsPerFloor }} 间
      </div>
      <button class="close" @click="visible = false">✕</button>
    </div>
    <div ref="mountRef" class="three-mount"></div>
    <p v-if="webglError" class="fallback">
      当前设备不支持 3D，已降级为楼层文本列表：<br />
      <span v-for="floor in floors" :key="floor">{{ floor }} 层 · {{ building?.roomsPerFloor || 0 }} 间</span>
    </p>
    <div class="legend"><i class="dot fault"></i> 待处理故障（脉冲红点）</div>
    <div class="actions">
      <button class="btn" @click="toggleRotate">{{ rotating ? '暂停旋转' : '开始旋转' }}</button>
    </div>
  </van-popup>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'

const props = defineProps<{ building: WorkerMapBuilding | null; orders: OrderItem[] }>()
const visible = defineModel<boolean>({ default: false })
const mountRef = ref<HTMLDivElement | null>(null)
const webglError = ref(false)
const floors = ref<number[]>([])
const rotating = ref(true)

let cancel = 0
let renderer: { dispose: () => void } | null = null
let sceneGroup: any = null
let markers: any[] = []
let animateFn: ((t: number) => void) | null = null

watch(
  () => visible.value,
  async (open) => {
    if (!open || !props.building) return
    webglError.value = false
    floors.value = Array.from({ length: props.building.floors }, (_, i) => i + 1)
    await nextTick()
    await initScene()
  },
)

watch(
  () => props.building,
  async (building) => {
    if (!building || !visible.value) return
    await nextTick()
    await initScene()
  },
)

function labelTexture(THREE: any, floor: number) {
  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 64
  const ctx = canvas.getContext('2d')
  if (!ctx) return null
  ctx.clearRect(0, 0, 256, 64)
  ctx.fillStyle = 'rgba(15,23,42,0.78)'
  ctx.fillRect(16, 8, 224, 48)
  ctx.fillStyle = '#ffffff'
  ctx.font = 'bold 34px sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(`${floor}F`, 128, 32)
  const texture = new THREE.CanvasTexture(canvas)
  texture.needsUpdate = true
  return texture
}

async function initScene() {
  if (!mountRef.value || !props.building) return
  if (cancel) cancelAnimationFrame(cancel)
  mountRef.value.innerHTML = ''
  markers = []
  sceneGroup = null
  animateFn = null
  try {
    const THREE = await import('three')
    const width = mountRef.value.clientWidth || 380
    const height = 300
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x0b1e45)
    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000)
    const render = new THREE.WebGLRenderer({ antialias: true })
    renderer = render
    render.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
    render.setSize(width, height)
    mountRef.value.appendChild(render.domElement)

    const b = props.building
    const floorCount = Math.max(b.floors, 1)
    const scale = 2.8
    const boxW = Math.max(b.width, 20) * scale
    const boxD = Math.max(b.height, 20) * scale
    const floorH = 2.4
    const slabH = 0.18
    const totalH = floorCount * (floorH + slabH)

    camera.position.set(boxW * 1.05, totalH * 1.25, boxD * 1.45)
    camera.lookAt(0, totalH * 0.48, 0)
    scene.add(new THREE.AmbientLight(0xffffff, 0.9))
    const dir = new THREE.DirectionalLight(0xffffff, 0.9)
    dir.position.set(40, 80, 25)
    scene.add(dir)
    const back = new THREE.DirectionalLight(0x93c5fd, 0.4)
    back.position.set(-30, 10, -40)
    scene.add(back)

    sceneGroup = new THREE.Group()
    scene.add(sceneGroup)

    const colors = ['#3b82f6', '#60a5fa', '#93c5fd']
    const floorGroups: any[] = []
    for (let i = 0; i < floorCount; i++) {
      const y = i * (floorH + slabH)
      const mat = new THREE.MeshLambertMaterial({ color: colors[i % colors.length] })
      const box = new THREE.Mesh(new THREE.BoxGeometry(boxW, floorH, boxD), mat)
      box.position.y = y + floorH / 2
      sceneGroup.add(box)

      const slab = new THREE.Mesh(
        new THREE.BoxGeometry(boxW + 0.25, slabH, boxD + 0.25),
        new THREE.MeshLambertMaterial({ color: '#0f2557' }),
      )
      slab.position.y = y + floorH + slabH / 2
      sceneGroup.add(slab)

      // front windows
      const cols = Math.min(8, Math.max(b.roomsPerFloor, 2))
      const rows = 2
      const winMat = new THREE.MeshBasicMaterial({ color: '#dbeafe' })
      for (let r = 0; r < rows; r++) {
        for (let col = 0; col < cols; col++) {
          const wx = -boxW * 0.3 + (boxW * 0.6 * col) / Math.max(cols - 1, 1)
          const wy = y + floorH * (0.22 + 0.3 * r)
          const win = new THREE.Mesh(new THREE.BoxGeometry(0.75, 0.8, 0.06), winMat)
          win.position.set(wx, wy, boxD / 2 + 0.04)
          sceneGroup.add(win)
        }
      }

      const labelTex = labelTexture(THREE, i + 1)
      if (labelTex) {
        const labelPlane = new THREE.Mesh(
          new THREE.PlaneGeometry(1.7, 0.42),
          new THREE.MeshBasicMaterial({ map: labelTex, transparent: true }),
        )
        labelPlane.position.set(-boxW / 2 - 1.15, y + floorH / 2, 0)
        labelPlane.rotation.y = Math.PI / 2
        sceneGroup.add(labelPlane)
      }
      floorGroups.push(box)
    }

    const orderByFloor = new Map<number, number>()
    for (const o of props.orders) {
      if (o.buildingId !== b.id) continue
      orderByFloor.set(o.floor, (orderByFloor.get(o.floor) || 0) + 1)
    }
    for (const [floor, count] of orderByFloor) {
      if (floor < 1 || floor > floorCount) continue
      for (let m = 0; m < Math.min(count, 3); m++) {
        const marker = new THREE.Mesh(
          new THREE.SphereGeometry(0.45, 16, 16),
          new THREE.MeshBasicMaterial({ color: 0xef4444 }),
        )
        const y = (floor - 1) * (floorH + slabH) + floorH * 0.55
        const offsetX = count === 1 ? 0 : (m - (Math.min(count, 3) - 1) / 2) * boxW * 0.12
        marker.position.set(offsetX, y + 0.25, 0)
        markers.push(marker)
        sceneGroup.add(marker)
      }
    }

    animateFn = (t: number) => {
      if (rotating.value) sceneGroup.rotation.y = Math.sin(t / 2600) * 0.65
      for (let i = 0; i < markers.length; i++) {
        const s = 1 + 0.22 * Math.sin(t / 180 + i * 1.1)
        markers[i].scale.set(s, s, s)
      }
      render.render(scene, camera)
      cancel = requestAnimationFrame(animateFn)
    }
    animateFn(0)
  } catch (err) {
    webglError.value = true
    if (renderer) {
      renderer.dispose()
      renderer = null
    }
  }
}

function toggleRotate() {
  rotating.value = !rotating.value
}

function disposeScene() {
  if (cancel) cancelAnimationFrame(cancel)
  cancel = 0
  if (renderer) {
    renderer.dispose()
    renderer = null
  }
}

defineExpose({ disposeScene })
</script>

<style scoped>
.head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 18px 8px;
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
.three-mount {
  width: 100%;
  height: 300px;
}
.fallback {
  padding: 20px;
  color: #475569;
  font-size: 13px;
  line-height: 1.8;
}
.legend {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px 4px;
  color: #94a3b8;
  font-size: 12px;
}
.dot.fault {
  width: 10px;
  height: 10px;
  background: #ef4444;
  border-radius: 50%;
}
.actions {
  padding: 6px 18px 12px;
  text-align: center;
}
.btn {
  padding: 8px 22px;
  color: #2563eb;
  font-size: 14px;
  font-weight: 600;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 999px;
  cursor: pointer;
}
</style>