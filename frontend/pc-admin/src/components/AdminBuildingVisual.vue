<template>
  <div class="admin-visual">
    <div class="tabs">
      <button class="tab" :class="{ active: mode === 'plan' }" @click="mode = 'plan'">楼层户型</button>
      <button class="tab" :class="{ active: mode === '3d' }" @click="switch3d">3D 立体</button>
    </div>

    <div v-if="mode === 'plan'" class="plan-tab">
      <div class="plan-head">{{ building.name }} · 第 {{ floor }} 层 · 共 16 间</div>
      <svg class="plan-svg" viewBox="0 0 960 600" preserveAspectRatio="xMidYMid meet">
        <rect x="8" y="8" width="944" height="584" fill="#f8fafc" stroke="#1e293b" stroke-width="5" rx="8" />
        <rect x="24" y="48" width="72" height="152" fill="#e2e8f0" stroke="#475569" stroke-width="2" />
        <text x="60" y="110" text-anchor="middle" font-size="11" fill="#334155">楼梯间</text>
        <rect x="24" y="206" width="72" height="86" fill="#e2e8f0" stroke="#475569" stroke-width="2" />
        <text x="60" y="250" text-anchor="middle" font-size="10" fill="#334155">盥洗/卫生间</text>
        <g v-for="(r, idx) in roomsRow(true)" :key="r.no">
          <rect :x="r.x" y="48" width="96" height="150" :fill="r.active ? '#fee2e2' : '#dbeafe'" stroke="#1e3a8a" stroke-width="2" />
          <text :x="r.x + 48" y="108" text-anchor="middle" font-size="12" font-weight="bold" fill="#1e3a8a">{{ r.no }}</text>
        </g>
        <rect x="110" y="198" width="824" height="204" fill="#f1f5f9" stroke="#94a3b8" stroke-dasharray="8 6" stroke-width="2" />
        <text x="540" y="305" text-anchor="middle" font-size="16" fill="#64748b">中央走廊</text>
        <g v-for="(r, idx) in roomsRow(false)" :key="r.no">
          <rect :x="r.x" y="402" width="96" height="150" :fill="r.active ? '#fee2e2' : '#dbeafe'" stroke="#1e3a8a" stroke-width="2" />
          <text :x="r.x + 48" y="478" text-anchor="middle" font-size="12" font-weight="bold" fill="#1e3a8a">{{ r.no }}</text>
        </g>
      </svg>
    </div>

    <div v-else class="three-tab">
      <div ref="mountRef" class="three-mount"></div>
      <p v-if="webglError" class="error">{{ errorText || 'WebGL 初始化失败，可切回楼层户型' }}</p>
      <p class="hint">拖拽旋转 · 滚轮缩放</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, computed } from 'vue'
import type { AdminBuilding, OrderItem } from '../api'

const props = defineProps<{ building: AdminBuilding; orders: OrderItem[] }>()
const mode = ref<'plan' | '3d'>('plan')
const floor = ref(1)
const mountRef = ref<HTMLDivElement | null>(null)
const webglError = ref(false)
const errorText = ref('')

const activeOrders = computed(() => props.orders.filter((o) => o.buildingId === props.building.id && o.floor === floor.value && [1, 2, 3].includes(o.status)))

function roomsRow(top: boolean) {
  return Array.from({ length: 8 }, (_, i) => {
    const seq = top ? i + 1 : i + 9
    const no = `${floor.value}${String(seq).padStart(2, '0')}`
    return { no, x: 112 + i * 104, active: activeOrders.value.some((o) => o.room === no || o.room?.endsWith(String(seq).padStart(2, '0'))) }
  })
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
    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 2000)
    const renderer = new THREE.WebGLRenderer({ antialias: true })
    renderer.setSize(width, height)
    el.innerHTML = ''
    el.appendChild(renderer.domElement)
    camera.position.set(180, 170, 220)

    scene.add(new THREE.AmbientLight(0xffffff, 0.9))
    const dir = new THREE.DirectionalLight(0xffffff, 0.9)
    dir.position.set(80, 160, 60)
    scene.add(dir)

    const controls = new controlsModule.OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true
    controls.target.set(0, 50, 0)
    controls.update()

    const ground = new THREE.Mesh(new THREE.PlaneGeometry(420, 320), new THREE.MeshLambertMaterial({ color: '#12264e', side: THREE.DoubleSide }))
    ground.rotation.x = -Math.PI / 2
    ground.position.y = -2
    scene.add(ground)

    const totalFloors = Math.max(props.building.floors, 1)
    const floorH = 12
    for (let f = 0; f < totalFloors; f++) {
      const yBase = f * (floorH + 1.5)
      const slab = new THREE.Mesh(new THREE.BoxGeometry(330, 1.2, 120), new THREE.MeshLambertMaterial({ color: '#1e3a8a' }))
      slab.position.y = yBase + 0.6
      scene.add(slab)
      for (let row = 0; row < 2; row++) {
        for (let col = 0; col < 8; col++) {
          const w = 32
          const gap = 5
          const x = -160 + 6 + col * (w + gap)
          const z = row === 0 ? -46 : 24
          const mat = new THREE.MeshLambertMaterial({ color: row === 0 ? '#60a5fa' : '#34d399', transparent: true, opacity: 0.82 })
          const tile = new THREE.Mesh(new THREE.BoxGeometry(w, 1.1, 34), mat)
          tile.position.set(x, yBase + 1.1, z)
          scene.add(tile)
        }
      }
      const corridor = new THREE.Mesh(new THREE.BoxGeometry(310, 0.8, 18), new THREE.MeshLambertMaterial({ color: '#cbd5e1' }))
      corridor.position.set(0, yBase + 0.8, -10)
      scene.add(corridor)

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
.plan-head { margin-bottom: 6px; font-weight: 700; }
.plan-svg { width: 100%; height: auto; background: #fff; border-radius: 10px; }
.three-mount { width: 100%; height: 430px; }
.error { color: #dc2626; }
.hint { color: #94a3b8; text-align: center; }
</style>