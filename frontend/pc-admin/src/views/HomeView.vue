<template>
  <AdminShell title="工单总览" subtitle="宿管报修 → 自动派单 → 维修完工，全流程一屏掌握">
    <section class="stat-grid">
      <div class="stat-card">
        <div class="stat-num">{{ orders.length }}</div>
        <div class="stat-label">工单总数</div>
      </div>
      <div class="stat-card">
        <div class="stat-num" style="color: #d97706">{{ pendingCount }}</div>
        <div class="stat-label">待处理</div>
      </div>
      <div class="stat-card">
        <div class="stat-num" style="color: #2563eb">{{ workingCount }}</div>
        <div class="stat-label">维修中</div>
      </div>
      <div class="stat-card">
        <div class="stat-num" style="color: #16a34a">{{ doneCount }}</div>
        <div class="stat-label">已完成</div>
      </div>
    </section>

    <section class="panel filter-panel">
      <div class="panel-title">筛选条件</div>
      <div class="filter-row">
        <el-select v-model="query.status" placeholder="全部状态" clearable style="width: 170px">
          <el-option label="待派单" :value="1" />
          <el-option label="已派单" :value="2" />
          <el-option label="维修中" :value="3" />
          <el-option label="已完成" :value="4" />
          <el-option label="已取消" :value="5" />
        </el-select>
        <el-input v-model="query.buildingText" style="width: 170px" clearable placeholder="楼栋ID" />
        <el-button type="primary" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </div>
    </section>

    <section class="panel table-panel">
      <div class="panel-title table-title">
        工单列表
        <el-tag type="info" effect="plain" size="small">共 {{ visibleOrders.length }} 条</el-tag>
      </div>
      <el-table :data="visibleOrders" v-loading="loading" border stripe>
        <el-table-column prop="orderNo" label="工单号" width="180" />
        <el-table-column prop="title" label="标题" min-width="170" show-overflow-tooltip />
        <el-table-column label="位置" width="170">
          <template #default="{ row }">{{ row.buildingName }} {{ row.floor }}F-{{ row.room }}</template>
        </el-table-column>
        <el-table-column prop="faultTypeName" label="类型" width="100" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <span class="status-badge" :class="'st' + row.status">{{ row.statusText }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="workerName" label="维修工人" width="110">
          <template #default="{ row }">{{ row.workerName || '—' }}</template>
        </el-table-column>
        <el-table-column label="派单评分" width="160">
          <template #default="{ row }">
            <el-tooltip
              v-if="row.dispatchScore !== undefined && row.dispatchScore > 0"
              :content="`技能 ${row.skillScore} · 距离 ${row.distanceScore} · 负载 ${row.loadScore}`"
            >
              <span class="score-text">{{ row.dispatchScore }}</span>
            </el-tooltip>
            <span v-else class="score-empty">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="reporterName" label="报修宿管" width="120" />
        <el-table-column prop="createdAt" label="创建时间" width="175" />
      </el-table>
      <el-empty
        v-if="!loading && visibleOrders.length === 0"
        description="当前筛选条件下暂无工单"
        class="table-empty"
      />
    </section>
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { OrderItem } from '../api'
import { apiAdminOrders } from '../api'
import AdminShell from '../components/AdminShell.vue'
import { useAuthStore } from '../stores/auth'

const orders = ref<OrderItem[]>([])
const loading = ref(false)
const query = reactive({ status: 0, buildingText: '' })
const auth = useAuthStore()
let ws: WebSocket | null = null
let refreshTimer: ReturnType<typeof setTimeout> | null = null

const pendingCount = computed(() => orders.value.filter((o) => o.status === 1 || o.status === 2).length)
const workingCount = computed(() => orders.value.filter((o) => o.status === 3).length)
const doneCount = computed(() => orders.value.filter((o) => o.status === 4).length)

const visibleOrders = computed(() => {
  const status = Number(query.status || 0)
  if (status === 0) return orders.value
  return orders.value.filter((o) => o.status === status)
})

async function load() {
  loading.value = true
  const buildingId = Number(query.buildingText || 0)
  try {
    orders.value = await apiAdminOrders(0, buildingId > 0 ? buildingId : 0)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reset() {
  query.status = 0
  query.buildingText = ''
  load()
}

function scheduleRefresh() {
  if (refreshTimer) clearTimeout(refreshTimer)
  refreshTimer = setTimeout(() => load(), 350)
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
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 18px;
}

.stat-card {
  padding: 20px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
}

.stat-num {
  font-size: 30px;
  font-weight: 800;
  color: #1e293b;
  line-height: 1;
}

.stat-label {
  margin-top: 10px;
  color: #64748b;
  font-size: 13px;
}

.panel {
  padding: 18px 20px;
  margin-bottom: 18px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
}

.panel-title {
  margin-bottom: 14px;
  color: #334155;
  font-size: 14px;
  font-weight: 700;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.table-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.status-badge {
  display: inline-block;
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 999px;
}

.st1,
.st2 {
  color: #b45309;
  background: #fef3c7;
}
.st3 {
  color: #1d4ed8;
  background: #dbeafe;
}
.st4 {
  color: #15803d;
  background: #dcfce7;
}
.st5 {
  color: #64748b;
  background: #f1f5f9;
}

.table-empty {
  padding: 30px 0;
}

.score-text {
  color: #1d4ed8;
  font-weight: 700;
}

.score-empty {
  color: #cbd5e1;
}
</style>