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
              <button v-if="item.workerName" class="rv-btn rv-btn-ghost" disabled style="opacity: 0.8">
                {{ item.workerName }}
              </button>
              <button v-if="canCancel(item)" class="rv-btn rv-btn-danger" @click="cancelOrder(item)">
                取消
              </button>
            </div>
          </div>
        </article>
      </div>
    </main>

    <button class="rv-fab" @click="openCreate">＋ 极简报修</button>

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

        <div class="rv-form-label">故障楼层</div>
        <input v-model.number="createForm.floor" class="rv-form-field" type="number" min="1" placeholder="例如 3" />

        <div class="rv-form-label">房间号</div>
        <input v-model="createForm.room" class="rv-form-field" placeholder="例如 301" />

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
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { FaultType, OrderItem } from '../api'
import { apiCancelOrder, apiCreateOrder, apiDormOrders, apiFaultTypes } from '../api'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits<{ (e: 'logout'): void }>()
const auth = useAuthStore()

type FilterValue = 'all' | 'todo' | 'working' | 'done' | 'canceled'

const orders = ref<OrderItem[]>([])
const faultTypes = ref<FaultType[]>([])
const loading = ref(false)
const submitting = ref(false)
const showCreate = ref(false)
const filter = ref<FilterValue>('all')
const createForm = reactive({ faultType: '', floor: 1, room: '', description: '' })

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
  try {
    orders.value = await apiDormOrders(0)
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loading.value = false
  }
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
  createForm.floor = 1
  createForm.room = ''
  createForm.description = ''
  showCreate.value = true
}

async function submitCreate() {
  if (!createForm.faultType || !createForm.room || !createForm.floor || createForm.floor < 1) {
    showToast('请完整填写维修信息')
    return
  }
  submitting.value = true
  try {
    await apiCreateOrder({
      faultType: createForm.faultType,
      floor: Number(createForm.floor),
      room: createForm.room.trim(),
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

onMounted(load)
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