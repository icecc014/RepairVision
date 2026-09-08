<template>
  <van-popup v-model:show="visible" position="bottom" round :style="{ height: '70vh' }">
    <div class="head">
      <div class="title">{{ building?.name || '楼栋 3D' }}</div>
      <button class="close" @click="visible = false">✕</button>
    </div>
    <div ref="mountRef" class="three-mount"></div>
    <p v-if="webglError" class="fallback">
      当前设备不支持 3D，已降级为楼层文本列表：<br />
      <span v-for="floor in floors" :key="floor">{{ floor }} 层 · {{ building?.roomsPerFloor || 0 }} 间</span>
    </p>
    <div class="legend">红色圆点 = 待处理故障（按楼层标注）</div>
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

let renderer: { dispose: () => void } | null = null
let cancel = 0

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

async function initScene() {
  if (!mountRef.value || !props.building) return
  if (cancel) cancelAnimationFrame(cancel)
  mountRef.value.innerHTML = ''
  try {
    const THREE = await import('three')
    const width = mountRef.value.clientWidth || 360
    const height = 280
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x0f2557)
    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000)
    const render = new THREE.WebGLRenderer({ antialias: true })
    renderer = render
    render.setSize(width, height)
    mountRef.value.appendChild(render.domElement)

    const b = props.building
    const floorsN = Math.max(b.floors, 1)
    const scale = 3
    const boxW = Math.max(b.width, 20) * scale
    const boxD = Math.max(b.height, 20) * scale
    const floorH = 2.6
    const totalH = floorsN * floorH
    camera.position.set(boxW * 1.1, totalH * 1.3, boxD * 1.3)
    camera.lookAt(0, totalH / 2, 0)

    scene.add(new THREE.AmbientLight(0xffffff, 0.8))
    const dir = new THREE.DirectionalLight(0xffffff, 0.7)
    dir.position.set(30, 50, 20)
    scene.add(dir)

    for (let i = 0; i < floorsN; i++) {
      const geo = new THREE.BoxGeometry(boxW, floorH, boxD)
      const mat = new THREE.MeshLambertMaterial({ color: i % 2 === 0 ? 0x3b82f6 : 0x93c5fd })
      const mesh = new THREE.Mesh(geo, mat)
      mesh.position.y = floorH / 2 + i * floorH
      scene.add(mesh)
    }

    const orderByFloor = new Map<number, number>()
    for (const o of props.orders) {
      if (o.buildingId !== b.id) continue
      orderByFloor.set(o.floor, (orderByFloor.get(o.floor) || 0) + 1)
    }
    for (const [floor, count] of orderByFloor) {
      if (floor > floorsN) continue
      const marker = new THREE.Mesh(
        new THREE.SphereGeometry(0.7, 16, 16),
        new THREE.MeshBasicMaterial({ color: 0xef4444 }),
      )
      marker.position.set(0, (floor - 0.5) * floorH + 0.1, 0)
      scene.add(marker)
    }

    const animate = () => {
      scene.rotation.y += 0.003
      render.render(scene, camera)
      cancel = requestAnimationFrame(animate)
    }
    animate()
  } catch (err) {
    webglError.value = true
    if (renderer) {
      renderer.dispose()
      renderer = null
    }
  }
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
  justify-content: space-between;
  padding: 16px 18px 8px;
}
.title {
  font-size: 17px;
  font-weight: 800;
}
.close {
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
  height: 280px;
}
.fallback {
  padding: 20px;
  color: #475569;
  font-size: 13px;
  line-height: 1.8;
}
.legend {
  padding: 8px 18px;
  color: #94a3b8;
  font-size: 12px;
}
</style>