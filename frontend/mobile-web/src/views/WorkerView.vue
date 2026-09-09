<template>
  <div class="page">
    <header class="rv-header">
      <div>
        <div class="rv-header-title">维修工工作台</div>
        <div class="rv-header-sub">{{ auth.user?.name }} · 我的工单</div>
      </div>
      <button class="rv-logout" @click="emit('logout')">退出</button>
    </header>

    <div class="mode-tabs">
      <button class="mode-tab" :class="{ active: tab === 'orders' }" @click="tab = 'orders'">我的工单</button>
      <button class="mode-tab" :class="{ active: tab === 'map' }" @click="tab = 'map'">报修地图</button>
      <button class="mode-tab" :class="{ active: tab === 'schedule' }" @click="tab = 'schedule'">我的班次</button>
    </div>

    <MapView v-if="tab === 'map'" />
    <WorkerSchedule v-else-if="tab === 'schedule'" />

    <main v-else class="rv-content">
      <section class="rv-stats">
        <div class="rv-stat">
          <div class="rv-stat-num" style="color: #d97706">{{ todoCount }}</div>
          <div class="rv-stat-label">待开工</div>
        </div>
        <div class="rv-stat">
          <div class="rv-stat-num" style="color: #2563eb">{{ workingCount }}</div>
          <div class="rv-stat-label">维修中</div>
        </div>
        <div class="rv-stat">
          <div class="rv-stat-num" style="color: #16a34a">{{ doneCount }}</div>
          <div class="rv-stat-label">已完成</div>
        </div>
      </section>

      <div class="rv-filters">
        <button
          v-for="chip in chips"
          :key="chip.value"
          class="rv-filter-chip"
          :class="{ active: filter === chip.value }"
          @click="filter = chip.value"
        >
          {{ chip.label }}
        </button>
        <button class="rv-filter-chip" style="margin-left: auto" @click="load">↻ 刷新</button>
      </div>

      <div v-if="visibleOrders.length === 0" class="rv-empty">
        <div class="rv-empty-icon">🔧</div>
        <div class="rv-empty-text">当前筛选下暂无工单</div>
      </div>

      <div v-else class="rv-order-list">
        <article v-for="item in visibleOrders" :key="item.id" class="rv-order-card">
          <div class="rv-order-top">
            <span class="rv-type">{{ item.faultTypeName }}</span>
            <span class="rv-status" :class="'s' + item.status">{{ item.statusText }}</span>
          </div>
          <h3 class="rv-order-title">{{ item.title }}</h3>
          <p v-if="item.description && item.description !== '无补充说明'" class="rv-order-desc">
            {{ item.description }}
          </p>
          <p class="rv-order-desc">{{ item.buildingName }} · {{ item.floor }} 层 {{ item.room }} 室</p>
          <div class="rv-order-meta">
            <span class="rv-order-time">派发 {{ item.createdAt }}</span>
            <div class="rv-order-actions">
              <button
                v-if="item.status === 2"
                class="rv-btn rv-btn-primary"
                :disabled="actingId === item.id"
                @click="start(item)"
              >
                {{ actingId === item.id ? '开工中…' : '开工' }}
              </button>
              <button
                v-if="item.status === 3"
                class="rv-btn rv-btn-success"
                :disabled="actingId === item.id"
                @click="complete(item)"
              >
                {{ actingId === item.id ? '提交中…' : '完工' }}
              </button>
              <span v-if="item.status === 5" class="rv-order-time">已取消</span>
            </div>
          </div>
        </article>
      </div>
      <button
        v-if="!loading && orders.length < total"
        class="rv-load-more"
        :disabled="loadingMore"
        @click="loadMore"
      >
        {{ loadingMore ? '加载中…' : '加载更多' }}
      </button>    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { OrderItem } from '../api'
import { apiCompleteOrder, apiStartOrder, apiWorkerOrderPage } from '../api'
import MapView from './MapView.vue'
import WorkerSchedule from './WorkerSchedule.vue'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits<{ (e: 'logout'): void }>()
const auth = useAuthStore()

type FilterValue = 'all' | 'today' | 'todo' | 'working' | 'done'

const orders = ref<OrderItem[]>([])
const tab = ref<'orders' | 'map' | 'schedule'>('orders')
let ws: WebSocket | null = null
let refreshTimer: ReturnType<typeof setTimeout> | null = null
const loading = ref(false)
const actingId = ref<number | null>(null)
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loadingMore = ref(false)
const filter = ref<FilterValue>('all')

const todoCount = computed(() => orders.value.filter((o) => o.status === 2).length)
const workingCount = computed(() => orders.value.filter((o) => o.status === 3).length)
const doneCount = computed(() => orders.value.filter((o) => o.status === 4).length)
function todayPrefix() {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

const todayCount = computed(() => orders.value.filter((o) => o.createdAt.startsWith(todayPrefix())).length)

const chips = computed(() => [
  { label: `全部 ${orders.value.length}`, value: 'all' as FilterValue },
  { label: `今日 ${todayCount.value}`, value: 'today' as FilterValue },
  { label: `待开工 ${todoCount.value}`, value: 'todo' as FilterValue },
  { label: `维修中 ${workingCount.value}`, value: 'working' as FilterValue },
  { label: `已完成 ${doneCount.value}`, value: 'done' as FilterValue },
])

const visibleOrders = computed(() => {
  switch (filter.value) {
    case 'today':
      return orders.value.filter((o) => o.createdAt.startsWith(todayPrefix()))
    case 'todo':
      return orders.value.filter((o) => o.status === 2)
    case 'working':
      return orders.value.filter((o) => o.status === 3)
    case 'done':
      return orders.value.filter((o) => o.status === 4)
    default:
      return orders.value
  }
})

async function load() {
  loading.value = true
  page.value = 1
  try {
    const res = await apiWorkerOrderPage(0, 1, pageSize)
    orders.value = res.list
    total.value = res.total
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || orders.value.length >= total.value) return
  loadingMore.value = true
  try {
    const next = page.value + 1
    const res = await apiWorkerOrderPage(0, next, pageSize)
    const seen = new Set(orders.value.map((o) => o.id))
    for (const item of res.list) {
      if (!seen.has(item.id)) {
        orders.value.push(item)
        seen.add(item.id)
      }
    }
    total.value = res.total
    page.value = next
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loadingMore.value = false
  }
}

async function start(item: OrderItem) {
  try {
    await showConfirmDialog({ title: '确认开工', message: `确认开始维修工单 ${item.orderNo}？` })
  } catch {
    return
  }
  actingId.value = item.id
  try {
    await apiStartOrder(item.id)
    showToast('已开工，请尽快处理')
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    actingId.value = null
  }
}

async function complete(item: OrderItem) {
  try {
    await showConfirmDialog({ title: '确认完工', message: `确认完成工单 ${item.orderNo}？` })
  } catch {
    return
  }
  actingId.value = item.id
  try {
    await apiCompleteOrder(item.id)
    showToast('维修完成')
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    actingId.value = null
  }
}

function scheduleRefresh() {
  if (refreshTimer) clearTimeout(refreshTimer)
  refreshTimer = setTimeout(() => load(), 300)
}

function connectWS() {
  if (!auth.token) return
  const proto = location.protocol === 'https:' ? 'wss://' : 'ws://'
  ws = new WebSocket(`${proto}${location.host}/ws/orders?token=${encodeURIComponent(auth.token)}`)
  ws.onmessage = () => scheduleRefresh()
  ws.onclose = () => {
    ws = null
    setTimeout(connectWS, 3000)
  }
}

onMounted(() => {
  load()
  connectWS()
})

onUnmounted(() => {
  if (refreshTimer) clearTimeout(refreshTimer)
  if (ws) ws.close()
})
</script>

<style scoped>
</style>
<style scoped>
.page .rv-header {
  padding-bottom: 16px;
}
.page .rv-content {
  margin-top: 0;
}
.mode-tabs {
  display: flex;
  gap: 8px;
  padding: 10px 14px 0;
}
.mode-tab {
  padding: 8px 16px;
  color: #64748b;
  font-size: 14px;
  font-weight: 600;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  cursor: pointer;
}
.mode-tab.active {
  color: #fff;
  background: #2563eb;
  border-color: #2563eb;
}
</style>

.rv-load-more {
  display: block;
  width: 100%;
  padding: 11px;
  margin: 12px 0 4px;
  color: #2563eb;
  font-size: 14px;
  font-weight: 700;
  background: #eff6ff;
  border: none;
  border-radius: 12px;
}
