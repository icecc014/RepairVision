<template>
  <div class="map-page">
    <div class="map-toolbar">
      <span class="map-title">我的维修楼栋</span>
      <div class="day-chips">
        <button
          v-for="d in dayOptions"
          :key="d"
          class="mini-btn"
          :class="{ active: days === d }"
          @click="setDays(d)"
        >
          {{ d === 30 ? '30天' : d + '天' }}
        </button>
      </div>
      <button class="mini-btn" @click="load">刷新</button>
    </div>
    <div class="range-hint">近 {{ days }} 天共 {{ map.orders.length }} 单（未完工 {{ uncompletedOrderCount }} 单 · 已完工 {{ completedOrderCount }} 单）</div>

    <div class="campus-card">
      <div class="campus-head">
        <span class="campus-title">区域概览</span>
        <button class="mini-btn" @click="showCampus = !showCampus">{{ showCampus ? '收起' : '展开' }}</button>
      </div>
      <CampusOverviewMap
        v-if="showCampus"
        ref="campusRef"
        :buildings="filteredBuildings"
        :counts="orderCounts"
        :highlight-building-id="selectedBuilding?.id"
      />
    </div>

    <div v-if="!loading && map.buildings.length > 0" class="canvas-card">
      <div class="chips-head">
        <span class="chips-title">我的维修楼栋（{{ filteredBuildings.length }} 栋）</span>
        <div class="status-tabs">
          <button
            class="status-tab-btn"
            :class="{ active: statusFilter === 'all' }"
            @click="setStatusFilter('all')"
          >
            全部 {{ map.buildings.length }}
          </button>
          <button
            class="status-tab-btn danger"
            :class="{ active: statusFilter === 'uncompleted' }"
            @click="setStatusFilter('uncompleted')"
          >
            未完工 {{ uncompletedBuildingCount }}
          </button>
          <button
            class="status-tab-btn success"
            :class="{ active: statusFilter === 'completed' }"
            @click="setStatusFilter('completed')"
          >
            已完工 {{ completedBuildingCount }}
          </button>
        </div>
      </div>
      <div class="building-chips">
        <button
          v-for="b in filteredBuildings"
          :key="b.id"
          class="building-chip"
          :class="{
            active: selectedBuilding?.id === b.id,
            'is-completed': getBuildingStatus(b.id) === 'completed',
            'is-uncompleted': getBuildingStatus(b.id) === 'uncompleted'
          }"
          @click="selectAndFocus(b)"
        >
          <span class="status-indicator"></span>
          {{ b.name }} · {{ countOf(b.id) }}单
          <span v-if="getBuildingStatus(b.id) === 'completed'" class="chip-badge completed">已完工</span>
          <span v-else-if="uncompletedCountOf(b.id) > 0" class="chip-badge uncompleted">{{ uncompletedCountOf(b.id) }}待修</span>
        </button>
        <div v-if="filteredBuildings.length === 0" class="filter-empty-hint">
          当前筛选条件下暂无楼栋
        </div>
      </div>
    </div>
    <div v-else-if="!loading" class="rv-empty">
      <div class="rv-empty-icon">🗺️</div>
      <div class="rv-empty-text">当前没有派给你的工单，暂无需要前往的楼栋</div>
    </div>

    <div v-if="selectedBuilding" class="detail-card">
      <div class="detail-head">
        <div>
          <div class="detail-name">
            {{ selectedBuilding.name }}
            <span
              class="building-status-pill"
              :class="getBuildingStatus(selectedBuilding.id)"
            >
              {{ getBuildingStatus(selectedBuilding.id) === 'completed' ? '已完工' : '有待修工单' }}
            </span>
          </div>
          <div class="detail-sub">
            {{ selectedBuilding.floors }} 层 · 每层 {{ selectedBuilding.roomsPerFloor }} 间
            （待修 {{ uncompletedCountOf(selectedBuilding.id) }} 单 · 已完工 {{ completedCountOf(selectedBuilding.id) }} 单）
          </div>
        </div>
        <div class="head-actions">
          <button class="mini-btn primary" @click="open3D">3D 查看</button>
        </div>
      </div>

      <div v-if="typeGroups.length === 0" class="rv-empty small">
        <div class="rv-empty-text">
          {{ countOf(selectedBuilding.id) > 0 ? '该楼栋所有工单已维修完成 🎉' : '该楼栋暂无工单' }}
        </div>
      </div>
      <div v-else class="type-list">
        <div v-for="g in typeGroups" :key="g.type" class="type-row">
          <div class="type-info">
            <span class="type-dot"></span>
            <span class="type-name">{{ g.name }}</span>
            <span class="type-count">{{ g.orders.length }} 单待处理</span>
          </div>
          <div class="type-explain">{{ explainType(g.type) }}</div>
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
import { computed, nextTick, onMounted, onUnmounted, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { OrderItem, WorkerMapBuilding, WorkerMapData } from '../api'
import { apiBatchComplete, apiWorkerMapData } from '../api'
import Building3D from './Building3D.vue'
import CampusOverviewMap from '../components/CampusOverviewMap.vue'
import { useAuthStore } from '../stores/auth'

const map = reactive<WorkerMapData>({ buildings: [], orders: [] })
const auth = useAuthStore()
let mapWs: WebSocket | null = null
let mapTimer: ReturnType<typeof setTimeout> | null = null
const loading = ref(false)
const dayOptions = [1, 3, 7, 30]
const days = ref(Number(localStorage.getItem('rv-days') || 3))
const selectedBuilding = ref<WorkerMapBuilding | null>(null)
const show3D = ref(false)
const showCampus = ref(true)
const actingType = ref('')

// 状态筛选 tab: 'all' | 'uncompleted' | 'completed'
type StatusFilterType = 'all' | 'uncompleted' | 'completed'
const statusFilter = ref<StatusFilterType>('all')

const campusRef = ref<{ focusBuilding: (id: number, openCard?: boolean) => boolean } | null>(null)

// 统计工单数
const completedOrderCount = computed(() => map.orders.filter((o) => o.status === 4).length)
const uncompletedOrderCount = computed(() => map.orders.filter((o) => o.status !== 4).length)

// 判断楼栋状态
function getBuildingStatus(buildingId: number): 'completed' | 'uncompleted' | 'none' {
  const buildingOrders = map.orders.filter((o) => o.buildingId === buildingId)
  if (buildingOrders.length === 0) return 'none'
  const hasUncompleted = buildingOrders.some((o) => o.status !== 4)
  return hasUncompleted ? 'uncompleted' : 'completed'
}

function countOf(buildingId: number) {
  return map.orders.filter((o) => o.buildingId === buildingId).length
}

function uncompletedCountOf(buildingId: number) {
  return map.orders.filter((o) => o.buildingId === buildingId && o.status !== 4).length
}

function completedCountOf(buildingId: number) {
  return map.orders.filter((o) => o.buildingId === buildingId && o.status === 4).length
}

// 统计未完工和已完工楼栋数
const uncompletedBuildingCount = computed(() => {
  return map.buildings.filter((b) => getBuildingStatus(b.id) === 'uncompleted').length
})

const completedBuildingCount = computed(() => {
  return map.buildings.filter((b) => getBuildingStatus(b.id) === 'completed').length
})

// 根据筛选条件过滤楼栋
const filteredBuildings = computed(() => {
  if (statusFilter.value === 'all') return map.buildings
  return map.buildings.filter((b) => getBuildingStatus(b.id) === statusFilter.value)
})

function setStatusFilter(filter: StatusFilterType) {
  statusFilter.value = filter
  // 切换筛选后，如果当前选中的楼栋不在列表中，自动选中第一个
  if (filteredBuildings.value.length > 0) {
    const stillInList = filteredBuildings.value.some((b) => b.id === selectedBuilding.value?.id)
    if (!stillInList) {
      selectAndFocus(filteredBuildings.value[0])
    }
  } else {
    selectedBuilding.value = null
  }
}

// 每栋楼在当前时间窗内的工单数（传给区域概览的信息卡）
const orderCounts = computed(() => {
  const result: Record<number, number> = {}
  for (const b of map.buildings) result[b.id] = countOf(b.id)
  return result
})

// 点下方「1 · 2单」这类楼栋按钮：选中 + 展开区域概览并定位该建筑
function selectAndFocus(b: WorkerMapBuilding) {
  select(b)
  if (!showCampus.value) showCampus.value = true
  nextTick(() => {
    const ok = campusRef.value?.focusBuilding(b.id)
    if (!ok) showToast('区域概览里还没有这栋建筑的图元，可在管理端补画')
  })
}

const selectedOrders = computed<OrderItem[]>(() => {
  if (!selectedBuilding.value) return []
  return map.orders.filter((o) => o.buildingId === selectedBuilding.value?.id)
})

// 仅展示待处理的工单分组用于批量完工
const typeGroups = computed(() => {
  const groups: { type: string; name: string; orders: OrderItem[]; acting?: boolean }[] = []
  const activeOrders = selectedOrders.value.filter((o) => o.status !== 4)
  for (const o of activeOrders) {
    const g = groups.find((x) => x.type === o.faultType)
    if (g) g.orders.push(o)
    else groups.push({ type: o.faultType, name: o.faultTypeName, orders: [o], acting: actingType.value === o.faultType })
  }
  return groups
})

function explainType(type: string) {
  switch (type) {
    case 'electric':
      return '可能原因：线路接触不良 / 开关插座损坏 / 灯具故障 / 负载跳闸；建议先断电再检修。'
    case 'water':
      return '可能原因：管道接头渗漏 / 阀门老化 / 下水堵塞 / 水压异常；建议先关闭角阀避免扩大。'
    default:
      return '可能原因：设施损坏或需现场排查；建议按报修描述携带工具确认。'
  }
}

function select(b: WorkerMapBuilding) {
  selectedBuilding.value = b
}

function open3D() {
  if (!selectedBuilding.value) return
  show3D.value = true
}

// 时间窗与"我的工单"筛选保持一致，选择结果记忆在本地
function setDays(d: number) {
  days.value = d
  localStorage.setItem('rv-days', String(d))
  load()
}

async function load() {
  loading.value = true
  try {
    const data = await apiWorkerMapData(days.value)
    map.buildings = data.buildings
    map.orders = data.orders
    if (filteredBuildings.value.length > 0) {
      const stillInList = filteredBuildings.value.some((b) => b.id === selectedBuilding.value?.id)
      if (!stillInList) {
        selectedBuilding.value = filteredBuildings.value[0]
      }
    } else {
      selectedBuilding.value = map.buildings[0] || null
    }
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function batchComplete(faultType: string) {
  if (!selectedBuilding.value) return
  const group = typeGroups.value.find((g) => g.type === faultType)
  if (!group || group.orders.length === 0) return
  try {
    await showConfirmDialog({
      title: '批量完工确认',
      message: `确认将【${selectedBuilding.value.name}】下的 ${group.orders.length} 单【${group.name}】一并标记为已完工？`
    })
  } catch {
    return
  }

  actingType.value = faultType
  try {
    const orderIds = group.orders.map((o) => o.id)
    const res = await apiBatchComplete(orderIds)
    showToast(`成功完工 ${res.successCount} 单`)
    await load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    actingType.value = ''
  }
}

function connectWs() {
  if (!auth.token) return
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${proto}//${location.host}/ws/orders?token=${encodeURIComponent(auth.token)}`
  try {
    mapWs = new WebSocket(wsUrl)
    mapWs.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data)
        if (msg.type === 'order_assigned' || msg.type === 'order_status_changed') {
          if (mapTimer) clearTimeout(mapTimer)
          mapTimer = setTimeout(() => load(), 500)
        }
      } catch {}
    }
    mapWs.onclose = () => {
      mapWs = null
    }
  } catch {}
}

onMounted(() => {
  load()
  connectWs()
})

onUnmounted(() => {
  if (mapWs) {
    mapWs.close()
    mapWs = null
  }
  if (mapTimer) clearTimeout(mapTimer)
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
  background: linear-gradient(100deg, #3478f6, #22b573 60%, #a06ae8);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.mini-btn {
  padding: 6px 13px;
  color: var(--rv-primary-deep, #1e40af);
  font-size: 13px;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.62);
  border: 1px solid rgba(120, 145, 190, 0.22);
  border-radius: 999px;
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  cursor: pointer;
  transition: all 0.2s ease;
}
.mini-btn.primary {
  color: #fff;
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(52, 120, 246, 0.28);
}
.mini-btn.success {
  color: #fff;
  background: linear-gradient(135deg, #7fe6c8, #22b573);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(34, 181, 115, 0.26);
}
.chips-head {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 6px 8px 8px;
}
.chips-title {
  font-size: 13px;
  font-weight: 700;
  color: #1f2a3d;
}
.status-tabs {
  display: flex;
  gap: 6px;
  background: rgba(241, 245, 249, 0.75);
  padding: 3px;
  border-radius: 999px;
  border: 1px solid rgba(226, 232, 240, 0.8);
}
.status-tab-btn {
  flex: 1;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  background: transparent;
  border: none;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}
.status-tab-btn.active {
  background: #fff;
  color: #2563eb;
  box-shadow: 0 2px 6px rgba(37, 99, 235, 0.15);
}
.status-tab-btn.danger.active {
  background: #fff;
  color: #e11d48;
  box-shadow: 0 2px 6px rgba(225, 29, 72, 0.15);
}
.status-tab-btn.success.active {
  background: #fff;
  color: #059669;
  box-shadow: 0 2px 6px rgba(5, 150, 105, 0.15);
}
.canvas-card {
  margin: 0 14px 14px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 12px 30px rgba(46, 68, 112, 0.1);
}
.building-chips {
  display: flex;
  gap: 8px;
  padding: 6px 0;
  overflow-x: auto;
}
.building-chip {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  color: var(--rv-text-sub, #475569);
  font-size: 13px;
  font-weight: 500;
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(120, 145, 190, 0.22);
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.building-chip .status-indicator {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #94a3b8;
}
.building-chip.is-uncompleted .status-indicator {
  background: #f43f5e;
  box-shadow: 0 0 6px rgba(244, 63, 94, 0.6);
}
.building-chip.is-completed .status-indicator {
  background: #10b981;
}
.building-chip.active {
  color: #fff;
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(52, 120, 246, 0.26);
}
.building-chip.active .status-indicator {
  background: #fff;
  box-shadow: 0 0 6px rgba(255, 255, 255, 0.8);
}
.chip-badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 999px;
  font-weight: 600;
}
.chip-badge.uncompleted {
  background: rgba(244, 63, 94, 0.15);
  color: #e11d48;
}
.chip-badge.completed {
  background: rgba(16, 185, 129, 0.15);
  color: #059669;
}
.building-chip.active .chip-badge.uncompleted {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
}
.building-chip.active .chip-badge.completed {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
}
.filter-empty-hint {
  padding: 8px 12px;
  font-size: 12px;
  color: #94a3b8;
  font-style: italic;
}
.day-chips {
  display: flex;
  gap: 6px;
}
.mini-btn.active {
  color: #fff;
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  border-color: transparent;
  box-shadow: 0 8px 18px rgba(52, 120, 246, 0.28);
}
.range-hint {
  padding: 4px 14px 8px;
  color: #8a97ad;
  font-size: 12px;
}
.campus-card {
  padding: 12px 14px;
  margin: 0 14px 12px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 16px;
  box-shadow: 0 10px 26px rgba(46, 68, 112, 0.09);
}
.campus-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.campus-title {
  color: var(--rv-text, #2b3445);
  font-size: 14px;
  font-weight: 800;
}
.campus-head .mini-btn {
  padding: 4px 12px;
  color: #2462d9;
  font-size: 12px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(120, 145, 190, 0.25);
  border-radius: 999px;
}
.detail-card {
  margin: 14px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 12px 30px rgba(46, 68, 112, 0.1);
}
.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}
.detail-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 800;
  color: var(--rv-text, #1e293b);
}
.building-status-pill {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
}
.building-status-pill.uncompleted {
  background: rgba(244, 63, 94, 0.12);
  color: #e11d48;
}
.building-status-pill.completed {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}
.detail-sub {
  margin-top: 4px;
  color: var(--rv-text-sub, #64748b);
  font-size: 12px;
}
.type-list {
  margin-top: 12px;
}
.type-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  border-top: 1px solid rgba(120, 145, 190, 0.16);
}
.type-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.type-dot {
  width: 8px;
  height: 8px;
  background: linear-gradient(135deg, #ffd8a3, #f0a24b);
  border-radius: 50%;
}
.type-name {
  font-weight: 600;
  color: var(--rv-text, #1e293b);
}
.type-count {
  color: var(--rv-text-light, #94a3b8);
  font-size: 12px;
}
.type-explain {
  flex-basis: 100%;
  margin-top: 4px;
  color: var(--rv-text-light, #94a3b8);
  font-size: 11px;
  line-height: 1.5;
}
.rv-empty {
  padding: 28px 16px;
  text-align: center;
}
.rv-empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
}
.rv-empty-text {
  color: var(--rv-text-light, #94a3b8);
  font-size: 13px;
}
.rv-empty.small {
  padding: 20px;
  margin-top: 12px;
}
</style>
