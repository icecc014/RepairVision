<template>
  <div class="page">
    <header class="rv-header">
      <div>
        <div class="rv-header-title">宿管工作台</div>
        <div class="rv-header-sub">{{ auth.user?.name }} · 本栋工单管理</div>
      </div>
      <button class="rv-logout" @click="emit('logout')">退出</button>
    </header>

    <main class="rv-content">
      <section class="rv-stats">
        <div class="rv-stat">
          <div class="rv-stat-num" style="color: #d97706">{{ pendingCount }}</div>
          <div class="rv-stat-label">待处理</div>
        </div>
        <div class="rv-stat">
          <div class="rv-stat-num" style="color: #2563eb">{{ workingCount }}</div>
          <div class="rv-stat-label">维修中</div>
        </div>
        <div class="rv-stat">
          <div class="rv-stat-num" style="color: #16a34a">{{ completedCount }}</div>
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
      </div>

      <div v-if="visibleOrders.length === 0" class="rv-empty">
        <div class="rv-empty-icon">🗂️</div>
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
            <span class="rv-order-time">报修 {{ item.createdAt }}</span>
            <div class="rv-order-actions">
              <button class="rv-btn rv-btn-ghost" @click="openDetail(item)">详情</button>
              <button v-if="item.workerName" class="rv-btn rv-btn-ghost" disabled style="opacity: 0.8">
                {{ item.workerName }}
              </button>
              <button v-if="item.status === 4 && !item.rated" class="rv-btn rv-btn-success" @click="openFeedback(item)">评价</button>
              <span v-else-if="item.status === 4 && item.rating" class="rv-order-time">★ {{ item.rating }}</span>
              <button v-if="canCancel(item)" class="rv-btn rv-btn-danger" @click="cancelOrder(item)">
                取消
              </button>
            </div>
          </div>
        </article>
      </div>
    </main>

      <button
        v-if="!loading && orders.length < total"
        class="rv-load-more"
        :disabled="loadingMore"
        @click="loadMore"
      >
        {{ loadingMore ? '加载中…' : '加载更多' }}
      </button>    <button class="rv-fab" @click="openCreate">＋ 极简报修</button>

    <van-popup
      v-model:show="showCreate"
      position="bottom"
      round
      :style="{ maxHeight: '86vh' }"
    >
      <div class="sheet-head">
        <div class="sheet-title">极简报修</div>
        <button class="sheet-close" @click="showCreate = false">✕</button>
      </div>
      <div class="rv-sheet-body">
        <div class="rv-form-label">维修类型</div>
        <div class="rv-chip-row">
          <button
            v-for="ft in faultTypes"
            :key="ft.code"
            class="rv-filter-chip"
            :class="{ active: createForm.faultType === ft.code }"
            @click="createForm.faultType = ft.code"
          >
            {{ ft.name }}
          </button>
        </div>

        <div class="rv-form-label">房间号</div>
        <input v-model="createForm.room" class="rv-form-field" placeholder="只需填房间号，如 401 / 301" />
        <p v-if="previewFloor > 0" class="floor-hint">将自动报修为 {{ previewFloor }} 层</p>

        <div class="rv-form-label">故障描述（可选）</div>
        <textarea
          v-model="createForm.description"
          class="rv-form-field"
          rows="2"
          placeholder="简单描述一下故障情况"
        />

        <button class="rv-submit" :disabled="submitting" @click="submitCreate">
          {{ submitting ? '提交中…' : '提交并自动派单' }}
        </button>
      </div>
    </van-popup>
  </div>
    <van-popup v-model:show="showDetail" position="bottom" round :style="{ maxHeight: '78vh' }">
      <div v-if="detail" style="padding: 18px 18px 26px">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:14px">
          <div style="font-size:18px;font-weight:800">工单详情</div>
          <button class="sheet-close" @click="showDetail = false">✕</button>
        </div>
        <div style="font-size:16px;font-weight:700;margin-bottom:8px">{{ detail.title }}</div>
        <div style="color:#94a3b8;font-size:12px;margin-bottom:16px">{{ detail.orderNo }} · {{ detail.createdAt }}</div>
        <div class="rv-form-label">位置</div>
        <div style="color:#334155;font-size:14px;margin-bottom:12px">{{ detail.buildingName }} · {{ detail.floor }} 层 {{ detail.room }} 室</div>
        <div class="rv-form-label">维修类型</div>
        <div style="color:#334155;font-size:14px;margin-bottom:12px">{{ detail.faultTypeName }}</div>
        <div class="rv-form-label">状态</div>
        <div style="color:#334155;font-size:14px;margin-bottom:12px">{{ detail.statusText }}</div>
        <div class="rv-form-label">故障描述</div>
        <div style="color:#475569;font-size:14px;margin-bottom:12px;white-space:pre-wrap">{{ detail.description || '无补充说明' }}</div>
        <div class="rv-form-label">处理工人</div>
        <div style="color:#334155;font-size:14px;margin-bottom:14px">{{ detail.workerName || '待派单' }}</div>
        <a
          v-if="detail.workerName && detail.workerPhone"
          :href="`tel:${detail.workerPhone}`"
          style="display:block;text-align:center;padding:12px;color:#fff;background:#2563eb;border-radius:12px;font-weight:700;text-decoration:none"
        >
          联系 {{ detail.workerName }}：{{ detail.workerPhone }}
        </a>
      </div>
    </van-popup>
    <van-popup v-model:show="feedbackVisible" position="bottom" round :style="{ maxHeight: '70vh' }">
      <div v-if="feedbackTarget" style="padding: 18px 18px 24px">
        <div style="font-size:17px;font-weight:800;margin-bottom:6px">服务评价</div>
        <div style="color:#94a3b8;font-size:12px;margin-bottom:14px">{{ feedbackTarget.orderNo }} · {{ feedbackTarget.title }}</div>
        <div style="display:flex;align-items:center;gap:14px;margin-bottom:16px">
          <span class="rv-form-label" style="margin:0">维修质量</span>
          <van-rate v-model="feedbackRating" :count="5" color="#f59e0b" void-icon="star" void-color="#e2e8f0" />
        </div>
        <div class="rv-form-label">评价内容（可选）</div>
        <textarea v-model="feedbackComment" class="rv-form-field" rows="3" maxlength="500" placeholder="说说本次维修服务怎么样" />
        <button class="rv-submit" :disabled="feedbackSubmitting" @click="submitFeedback">
          {{ feedbackSubmitting ? '提交中…' : '提交评价' }}
        </button>
      </div>
    </van-popup>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { FaultType, OrderItem } from '../api'
import { apiCancelOrder, apiCreateOrder, apiDormFeedback, apiDormOrderPage, apiFaultTypes } from '../api'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits<{ (e: 'logout'): void }>()
const auth = useAuthStore()

type FilterValue = 'all' | 'todo' | 'working' | 'done' | 'canceled'

const orders = ref<OrderItem[]>([])
let ws: WebSocket | null = null
let refreshTimer: ReturnType<typeof setTimeout> | null = null
const faultTypes = ref<FaultType[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loadingMore = ref(false)
const submitting = ref(false)
const showCreate = ref(false)
const showDetail = ref(false)
const detail = ref<OrderItem | null>(null)
const feedbackVisible = ref(false)
const feedbackTarget = ref<OrderItem | null>(null)
const feedbackRating = ref(5)
const feedbackComment = ref('')
const feedbackSubmitting = ref(false)
const filter = ref<FilterValue>('all')
const createForm = reactive({ faultType: '', room: '', description: '' })

const previewFloor = computed(() => {
  const room = createForm.room.trim()
  if (!room) return 0
  const n = Number(room.charAt(0))
  return n >= 1 && n <= 9 ? n : 0
})
const pendingCount = computed(() => orders.value.filter((o) => o.status === 1 || o.status === 2).length)
const workingCount = computed(() => orders.value.filter((o) => o.status === 3).length)
const completedCount = computed(() => orders.value.filter((o) => o.status === 4).length)
const canceledCount = computed(() => orders.value.filter((o) => o.status === 5).length)

const chips = computed(() => [
  { label: `全部 ${orders.value.length}`, value: 'all' as FilterValue },
  { label: `待处理 ${pendingCount.value}`, value: 'todo' as FilterValue },
  { label: `维修中 ${workingCount.value}`, value: 'working' as FilterValue },
  { label: `已完成 ${completedCount.value}`, value: 'done' as FilterValue },
  { label: `已取消 ${canceledCount.value}`, value: 'canceled' as FilterValue },
])

const visibleOrders = computed(() => {
  switch (filter.value) {
    case 'todo':
      return orders.value.filter((o) => o.status === 1 || o.status === 2)
    case 'working':
      return orders.value.filter((o) => o.status === 3)
    case 'done':
      return orders.value.filter((o) => o.status === 4)
    case 'canceled':
      return orders.value.filter((o) => o.status === 5)
    default:
      return orders.value
  }
})

async function load() {
  loading.value = true
  page.value = 1
  try {
    const res = await apiDormOrderPage(0, 1, pageSize)
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
    const res = await apiDormOrderPage(0, next, pageSize)
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

function openFeedback(item: OrderItem) {
  feedbackTarget.value = item
  feedbackRating.value = 5
  feedbackComment.value = ''
  feedbackVisible.value = true
}

async function submitFeedback() {
  if (!feedbackTarget.value) return
  feedbackSubmitting.value = true
  try {
    await apiDormFeedback(feedbackTarget.value.id, feedbackRating.value, feedbackComment.value.trim())
    showToast('评价成功，感谢反馈')
    feedbackVisible.value = false
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    feedbackSubmitting.value = false
  }
}
function openDetail(item: OrderItem) {
  detail.value = item
  showDetail.value = true
}
async function openCreate() {
  if (faultTypes.value.length === 0) {
    try {
      faultTypes.value = await apiFaultTypes()
    } catch (err) {
      showToast((err as Error).message)
      return
    }
  }
  createForm.faultType = faultTypes.value[0]?.code || ''
  createForm.room = ''
  createForm.description = ''
  showCreate.value = true
}

async function submitCreate() {
  const room = createForm.room.trim()
  if (!createForm.faultType || !room) {
    showToast('请选择维修类型并填写房间号')
    return
  }
  const derivedFloor = Number(room.charAt(0))
  if (!/^\d{2,4}$/.test(room) || derivedFloor < 1 || derivedFloor > 9) {
    showToast('房间号格式不正确，如 401 表示 4 层 01 房')
    return
  }
  submitting.value = true
  try {
    await apiCreateOrder({
      faultType: createForm.faultType,
      floor: derivedFloor,
      room: room,
      description: createForm.description.trim(),
    })
    showToast('报修成功，已自动派单')
    showCreate.value = false
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    submitting.value = false
  }
}

function canCancel(item: OrderItem) {
  return item.status === 1 || item.status === 2
}

async function cancelOrder(item: OrderItem) {
  try {
    await showConfirmDialog({ title: '取消工单', message: `确认取消 ${item.orderNo}？` })
  } catch {
    return
  }
  try {
    await apiCancelOrder(item.id)
    showToast('已取消')
    load()
  } catch (err) {
    showToast((err as Error).message)
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
.sheet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 18px 0;
}

.sheet-title {
  font-size: 17px;
  font-weight: 800;
}

.sheet-close {
  width: 28px;
  height: 28px;
  color: #64748b;
  font-size: 14px;
  background: #f1f5f9;
  border: none;
  border-radius: 50%;
  cursor: pointer;
}
</style>

.floor-hint {
  margin: 6px 2px 0;
  color: #2563eb;
  font-size: 12px;
}

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
