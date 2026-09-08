<template>
  <div class="admin-shell">
    <aside class="sidebar">
      <div class="side-brand">
        <span class="side-logo">修</span>
        <div>
          <div class="side-name">RepairVision</div>
          <div class="side-sub">维修调度管理平台</div>
        </div>
      </div>

      <div class="nav-label">管理导航</div>
      <button class="nav-item active">
        <span class="nav-dot"></span>工单总览
      </button>
      <button class="nav-item disabled"><span class="nav-dot"></span>人员账号管理<em>P2</em></button>
      <button class="nav-item disabled"><span class="nav-dot"></span>建筑信息管理<em>P2</em></button>
      <button class="nav-item disabled"><span class="nav-dot"></span>维修类型字典<em>P2</em></button>
      <button class="nav-item disabled"><span class="nav-dot"></span>派单规则配置<em>P2</em></button>

      <div class="side-footer">校园维修 · 毕业设计</div>
    </aside>

    <section class="main-area">
      <header class="topbar">
        <div>
          <h2 class="page-title">工单总览</h2>
          <p class="page-sub">宿管报修 → 自动派单 → 维修完工，全流程一屏掌握</p>
        </div>
        <div class="user-box">
          <el-tag type="primary" effect="dark" size="small">管理员</el-tag>
          <span class="user-name">{{ auth.user?.name }}</span>
          <button class="logout-btn" @click="logout">退出登录</button>
        </div>
      </header>

      <main class="content">
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
            <el-select
              v-model="query.status"
              placeholder="全部状态"
              clearable
              style="width: 170px"
            >
              <el-option label="待派单" :value="1" />
              <el-option label="已派单" :value="2" />
              <el-option label="维修中" :value="3" />
              <el-option label="已完成" :value="4" />
              <el-option label="已取消" :value="5" />
            </el-select>
            <el-input
              v-model="query.buildingText"
              style="width: 170px"
              clearable
              placeholder="楼栋ID"
            />
            <el-button type="primary" @click="applyBuilding">查询</el-button>
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
              <template #default="{ row }">
                {{ row.buildingName }} {{ row.floor }}F-{{ row.room }}
              </template>
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
            <el-table-column prop="reporterName" label="报修宿管" width="120" />
            <el-table-column prop="createdAt" label="创建时间" width="175" />
          </el-table>
          <el-empty
            v-if="!loading && visibleOrders.length === 0"
            description="当前筛选条件下暂无工单"
            class="table-empty"
          />
        </section>
      </main>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { OrderItem } from '../api'
import { apiAdminOrders } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const orders = ref<OrderItem[]>([])
const loading = ref(false)
const query = reactive({ status: 0, buildingText: '' })

const pendingCount = computed(() => orders.value.filter((o) => o.status === 1 || o.status === 2).length)
const workingCount = computed(() => orders.value.filter((o) => o.status === 3).length)
const doneCount = computed(() => orders.value.filter((o) => o.status === 4).length)

const visibleOrders = computed(() => {
  const status = Number(query.status || 0)
  if (status === 0) {
    return orders.value
  }
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

function applyBuilding() {
  load()
}

function reset() {
  query.status = 0
  query.buildingText = ''
  load()
}

function logout() {
  auth.logout()
  router.replace('/login')
}

onMounted(load)
</script>

<style scoped>
.admin-shell {
  display: grid;
  grid-template-columns: 232px 1fr;
  min-height: 100vh;
}

.sidebar {
  display: flex;
  flex-direction: column;
  padding: 20px 14px;
  color: #cbd5e1;
  background: linear-gradient(180deg, #102a6b 0%, #0f2557 55%, #0b1d47 100%);
}

.side-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px 22px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.09);
}

.side-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  color: #2563eb;
  font-size: 19px;
  font-weight: 800;
  background: #fff;
  border-radius: 11px;
}

.side-name {
  color: #fff;
  font-size: 15px;
  font-weight: 700;
}

.side-sub {
  margin-top: 2px;
  color: #8ea3cf;
  font-size: 11px;
}

.nav-label {
  margin: 22px 10px 8px;
  color: #6b82b8;
  font-size: 11px;
  letter-spacing: 1px;
}

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 11px 14px;
  margin: 2px 0;
  color: #b8c7e8;
  font-size: 14px;
  text-align: left;
  background: transparent;
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.nav-item.active {
  color: #fff;
  font-weight: 700;
  background: linear-gradient(90deg, rgba(37, 99, 235, 0.45), rgba(37, 99, 235, 0.08));
  box-shadow: inset 3px 0 0 #60a5fa;
}

.nav-dot {
  width: 7px;
  height: 7px;
  background: currentColor;
  border-radius: 50%;
  opacity: 0.75;
}

.nav-item.disabled {
  color: #6479ad;
  cursor: not-allowed;
}

.nav-item em {
  margin-left: auto;
  padding: 1px 7px;
  color: #93b4ff;
  font-size: 10px;
  font-style: normal;
  background: rgba(37, 99, 235, 0.22);
  border-radius: 999px;
}

.side-footer {
  margin-top: auto;
  padding: 16px 8px 6px;
  color: #5d74aa;
  font-size: 11px;
}

.main-area {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 22px 28px 16px;
  background: #fff;
  border-bottom: 1px solid #eef2f7;
}

.page-title {
  margin: 0;
  font-size: 21px;
  font-weight: 800;
  color: #1e293b;
}

.page-sub {
  margin: 5px 0 0;
  color: #94a3b8;
  font-size: 13px;
}

.user-box {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-name {
  color: #475569;
  font-weight: 600;
}

.logout-btn {
  padding: 7px 14px;
  color: #dc2626;
  font-size: 13px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  cursor: pointer;
}

.content {
  padding: 22px 28px 36px;
}

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
</style>