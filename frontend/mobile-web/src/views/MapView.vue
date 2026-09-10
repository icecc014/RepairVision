<template>
  <div class="map-page">
    <div class="map-toolbar">
      <span class="map-title">2D 楼栋总览</span>
      <button class="mini-btn" @click="load">刷新</button>
    </div>

    <div v-if="!loading && map.buildings.length > 0" class="canvas-card">
      <v-stage :config="stageConfig">
        <v-layer>
          <v-rect v-for="shape in buildingShapes" :key="shape.id" :config="shape.config" />
          <v-text v-for="text in labelShapes" :key="text.id" :config="text.config" />
          <v-group v-for="b in map.buildings" :key="'badge' + b.id" :config="{}">
            <v-circle v-if="countOf(b.id) > 0" :config="badgeConfig(b)" />
            <v-text v-if="countOf(b.id) > 0" :config="badgeTextConfig(b)" />
          </v-group>
        </v-layer>
      </v-stage>
      <div class="map-hint">点击楼栋查看工单，可进入 3D</div>
      <div class="building-chips">
        <button
          v-for="b in map.buildings"
          :key="b.id"
          class="building-chip"
          :class="{ active: selectedBuilding?.id === b.id }"
          @click="select(b)"
        >
          {{ b.code }} · {{ countOf(b.id) }}单
        </button>
      </div>
    </div>
    <div v-else-if="!loading" class="rv-empty">
      <div class="rv-empty-icon">🗺️</div>
      <div class="rv-empty-text">暂无可管理的楼栋</div>
    </div>

    <div v-if="selectedBuilding" class="detail-card">
      <div class="detail-head">
        <div>
          <div class="detail-name">{{ selectedBuilding.code }} {{ selectedBuilding.name }}</div>
          <div class="detail-sub">{{ selectedBuilding.floors }} 层 · 每层 {{ selectedBuilding.roomsPerFloor }} 间</div>
        </div>
        <div class="head-actions">
          <button class="mini-btn primary" @click="open3D">3D 查看</button>
        </div>
      </div>

      <div v-if="typeGroups.length === 0" class="rv-empty small">
        <div class="rv-empty-text">该楼栋暂无待处理工单</div>
      </div>
      <div v-else class="type-list">
        <div v-for="g in typeGroups" :key="g.type" class="type-row">
          <div class="type-info">
            <span class="type-dot"></span>
            <span class="type-name">{{ g.name }}</span>
            <span class="type-count">{{ g.orders.length }} 单</span>
          </div>
          <button class="mini-btn success" :disabled="g.acting" @click="batchComplete(g.type)">
            {{ g.acting ? '提交中' : '按类型完工' }}
          </button>
        </div>
      </div>
    </div>

    <Building3D v-model="show3D" :building="selectedBuilding" :orders="map.orders" @refresh="load" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { OrderItem, WorkerMapBuilding, WorkerMapData } from '../api'
import { apiBatchComplete, apiWorkerMapData } from '../api'
import Building3D from './Building3D.vue'
import { useAuthStore } from '../stores/auth'

const map = reactive<WorkerMapData>({ buildings: [], orders: [] })
const auth = useAuthStore()
let mapWs: WebSocket | null = null
let mapTimer: ReturnType<typeof setTimeout> | null = null
const loading = ref(false)
const selectedBuilding = ref<WorkerMapBuilding | null>(null)
const show3D = ref(false)
const actingType = ref('')

const stageWidth = Math.max(Math.min((window.innerWidth || 390) - 28, 420), 300)
const stageHeight = 330

const stageConfig = { width: stageWidth, height: stageHeight, x: 14, y: 10 }

function countOf(buildingId: number) {
  return map.orders.filter((o) => o.buildingId === buildingId).length
}

function scaleBounds() {
  const buildings = map.buildings
  if (buildings.length === 0) return { minX: 0, minY: 0, rangeX: 1, rangeY: 1, pad: 36, cw: stageWidth - 72, ch: stageHeight - 72 }
  let minX = Infinity
  let maxX = -Infinity
  let minY = Infinity
  let maxY = -Infinity
  for (const b of buildings) {
    minX = Math.min(minX, b.posX - b.width / 2)
    maxX = Math.max(maxX, b.posX + b.width / 2)
    minY = Math.min(minY, b.posY - b.height / 2)
    maxY = Math.max(maxY, b.posY + b.height / 2)
  }
  const pad = 40
  return {
    minX,
    minY,
    rangeX: Math.max(maxX - minX, 1),
    rangeY: Math.max(maxY - minY, 1),
    pad,
    cw: stageWidth - pad * 2,
    ch: stageHeight - pad * 2,
  }
}

const buildingShapes = computed(() => {
  const s = scaleBounds()
  return map.buildings.map((b) => ({
    id: b.id,
    config: {
      x: s.pad + ((b.posX - b.width / 2 - s.minX) / s.rangeX) * s.cw,
      y: s.pad + ((b.posY - b.height / 2 - s.minY) / s.rangeY) * s.ch,
      width: Math.max((b.width / s.rangeX) * s.cw, 18),
      height: Math.max((b.height / s.rangeY) * s.ch, 14),
      fill: selectedBuilding.value?.id === b.id ? '#60a5fa' : '#bfdbfe',
      stroke: selectedBuilding.value?.id === b.id ? '#2563eb' : '#93c5fd',
      strokeWidth: selectedBuilding.value?.id === b.id ? 3 : 1.5,
      cornerRadius: 6,
      shadowColor: '#2563eb',
      shadowBlur: selectedBuilding.value?.id === b.id ? 12 : 0,
      onClick: () => select(b),
    },
  }))
})

const labelShapes = computed(() => {
  const s = scaleBounds()
  return map.buildings.map((b) => ({
    id: 'label' + b.id,
    config: {
      x: s.pad + ((b.posX - s.minX) / s.rangeX) * s.cw - 30,
      y: s.pad + ((b.posY - b.height / 2 - s.minY) / s.rangeY) * s.ch - 12,
      width: 60,
      text: b.code,
      fontSize: 11,
      fontStyle: 'bold',
      fill: '#1e3a8a',
      align: 'center',
    },
  }))
})

function badgePos(b: WorkerMapBuilding) {
  const s = scaleBounds()
  return {
    x: s.pad + ((b.posX + b.width / 2 - s.minX) / s.rangeX) * s.cw,
    y: s.pad + ((b.posY - b.height / 2 - s.minY) / s.rangeY) * s.ch,
  }
}

function badgeConfig(b: WorkerMapBuilding) {
  const p = badgePos(b)
  return { x: p.x + 8, y: p.y - 4, radius: 13, fill: '#ef4444', stroke: '#fff', strokeWidth: 2 }
}

function badgeTextConfig(b: WorkerMapBuilding) {
  const p = badgePos(b)
  return {
    x: p.x - 4,
    y: p.y - 10,
    width: 25,
    text: String(countOf(b.id)),
    fontSize: 12,
    fontStyle: 'bold',
    fill: '#fff',
    align: 'center',
  }
}

const selectedOrders = computed<OrderItem[]>(() => {
  if (!selectedBuilding.value) return []
  return map.orders.filter((o) => o.buildingId === selectedBuilding.value?.id)
})

const typeGroups = computed(() => {
  const groups: { type: string; name: string; orders: OrderItem[] }[] = []
  for (const o of selectedOrders.value) {
    const g = groups.find((x) => x.type === o.faultType)
    if (g) g.orders.push(o)
    else groups.push({ type: o.faultType, name: o.faultTypeName, orders: [o] })
  }
  return groups
})

function select(b: WorkerMapBuilding) {
  selectedBuilding.value = b
}

function open3D() {
  if (!selectedBuilding.value) return
  show3D.value = true
}

async function load() {
  loading.value = true
  try {
    const data = await apiWorkerMapData()
    map.buildings = data.buildings
    map.orders = data.orders
    selectedBuilding.value = map.buildings[0] || null
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function batchComplete(faultType: string) {
  if (!selectedBuilding.value) return
  try {
    await showConfirmDialog({
      title: '批量完工',
      message: `确认将 ${selectedBuilding.value.name} 的${typeGroups.value.find((g) => g.type === faultType)?.name || faultType}全部完工？`,
    })
  } catch {
    return
  }
  actingType.value = faultType
  try {
    const res = await apiBatchComplete(selectedBuilding.value.id, faultType)
    showToast(`已完工 ${res.count} 单`)
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    actingType.value = ''
  }
}

function scheduleMapRefresh() {
  if (mapTimer) clearTimeout(mapTimer)
  mapTimer = setTimeout(() => load(), 350)
}

function connectMapWS() {
  if (!auth.token) return
  const proto = location.protocol === 'https:' ? 'wss://' : 'ws://'
  mapWs = new WebSocket(`${proto}${location.host}/ws/orders?token=${encodeURIComponent(auth.token)}`)
  mapWs.onmessage = () => scheduleMapRefresh()
  mapWs.onclose = () => {
    mapWs = null
    setTimeout(connectMapWS, 3000)
  }
}

onMounted(() => {
  load()
  connectMapWS()
})

onUnmounted(() => {
  if (mapTimer) clearTimeout(mapTimer)
  if (mapWs) mapWs.close()
})
</script>

<style scoped>
.map-page {
  padding-bottom: 16px;
}
.map-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
}
.map-title {
  font-size: 16px;
  font-weight: 800;
}
.mini-btn {
  padding: 6px 13px;
  color: #2563eb;
  font-size: 13px;
  font-weight: 600;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 999px;
  cursor: pointer;
}
.mini-btn.primary {
  color: #fff;
  background: #2563eb;
  border-color: #2563eb;
}
.mini-btn.success {
  color: #fff;
  background: #16a34a;
  border-color: #16a34a;
}
.canvas-card {
  padding: 6px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.06);
}
.map-hint {
  margin: 0 8px 8px;
  color: #94a3b8;
  font-size: 12px;
  text-align: center;
}
.detail-card {
  margin: 14px;
  padding: 16px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.06);
}
.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}
.detail-name {
  font-size: 16px;
  font-weight: 800;
}
.detail-sub {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}
.type-list {
  margin-top: 12px;
}
.type-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  border-top: 1px solid #f1f5f9;
}
.type-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.type-dot {
  width: 8px;
  height: 8px;
  background: #f59e0b;
  border-radius: 50%;
}
.type-name {
  font-weight: 600;
}
.type-count {
  color: #94a3b8;
  font-size: 12px;
}
.rv-empty.small {
  padding: 20px;
  margin-top: 12px;
}
</style>
.building-chips {
  display: flex;
  gap: 8px;
  padding: 8px 14px;
  overflow-x: auto;
}
.building-chip {
  flex: 0 0 auto;
  padding: 7px 13px;
  color: #475569;
  font-size: 13px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  cursor: pointer;
}
.building-chip.active {
  color: #fff;
  background: #2563eb;
  border-color: #2563eb;
}
