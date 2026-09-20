<template>
  <AdminShell title="建筑可视化" subtitle="查看全部建筑，点开任意楼栋查看 16 间户型与 3D 视图">
    <section class="panel">
      <div class="range-bar">
        <span class="range-title">报修时间窗</span>
        <el-radio-group v-model="days" size="small" @change="onDaysChange">
          <el-radio-button :value="1">1天</el-radio-button>
          <el-radio-button :value="3">3天</el-radio-button>
          <el-radio-button :value="7">7天</el-radio-button>
          <el-radio-button :value="30">30天</el-radio-button>
        </el-radio-group>
        <span class="range-hint">近 {{ days }} 天共 {{ orders.length }} 单，2D / 3D 视图与房间状态按此范围显示</span>
      </div>

      <div class="grid">
        <button
          v-for="b in buildings"
          :key="b.id"
          class="building-card"
          @click="open(b)"
        >
          <div class="card-head">
            <span class="building-title">🏢 {{ b.name }}</span>
            <div class="status-tags-group">
              <span class="stat-pill total">{{ totalOrders(b.id) }} 单报修</span>
              <span class="stat-pill pending" :class="{ unfinished: uncompletedOrders(b.id) > 0 }">
                未完成 {{ uncompletedOrders(b.id) }}
              </span>
              <span class="stat-pill completed" :class="{ finished: completedOrders(b.id) > 0 }">
                已完成 {{ completedOrders(b.id) }}
              </span>
            </div>
          </div>
          <div class="meta">{{ b.floors }} 层标准楼 · 每层 {{ b.roomsPerFloor }} 间房</div>
          <div class="action">进入楼层户型 / 3D 沙盘 ➔</div>
        </button>
      </div>
    </section>

    <el-dialog append-to-body
      v-model="dialogVisible"
      :title="selected ? (selected.name + ' · 空间可视化') : ''"
      width="1200px"
      class="visual-dialog"
      top="6vh"
      destroy-on-close
      @closed="onDialogClosed"
    >
      <AdminBuildingVisual
        v-if="selected && dialogVisible"
        :key="'visual-bldg-' + selected.id"
        :building="selected"
        :orders="orders"
      />
    </el-dialog>
  </AdminShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { AdminBuilding, OrderItem } from '../api'
import { apiAdminBuildings, apiAdminOrders } from '../api'
import AdminShell from '../components/AdminShell.vue'
import AdminBuildingVisual from '../components/AdminBuildingVisual.vue'

const buildings = ref<AdminBuilding[]>([])
const orders = ref<OrderItem[]>([])
const days = ref(3)
const selected = ref<AdminBuilding | null>(null)
const dialogVisible = ref(false)

async function loadOrders() {
  try {
    const orderPage = await apiAdminOrders(0, 0, 1, 300, days.value)
    orders.value = orderPage.list
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function onDaysChange() {
  loadOrders()
}

async function load() {
  try {
    buildings.value = await apiAdminBuildings()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
  await loadOrders()
}

function totalOrders(buildingId: number) {
  return orders.value.filter((o) => o.buildingId === buildingId).length
}

function uncompletedOrders(buildingId: number) {
  return orders.value.filter((o) => o.buildingId === buildingId && (o.status === 1 || o.status === 2 || o.status === 3)).length
}

function completedOrders(buildingId: number) {
  return orders.value.filter((o) => o.buildingId === buildingId && o.status === 4).length
}

async function open(b: AdminBuilding) {
  dialogVisible.value = false
  selected.value = null
  await loadOrders()
  selected.value = b
  dialogVisible.value = true
}

function onDialogClosed() {
  selected.value = null
}

onMounted(load)
</script>

<style scoped>
.panel { padding: 20px; background: #fff; border-radius: 14px; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(310px, 1fr)); gap: 16px; }
.building-card {
  text-align: left;
  padding: 18px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  cursor: pointer;
  transition: all .2s cubic-bezier(0.4, 0, 0.2, 1);
}
.building-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 24px rgba(37,99,235,.10);
  border-color: #93c5fd;
  background: #ffffff;
}
.card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}
.building-title { font-size: 16px; font-weight: 800; color: #1e3a8a; }

.status-tags-group {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.stat-pill {
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

.stat-pill.total {
  background: #f1f5f9;
  color: #64748b;
  border: 1px solid #e2e8f0;
}

.stat-pill.pending {
  background: #f8fafc;
  color: #94a3b8;
  border: 1px solid #e2e8f0;
}
.stat-pill.pending.unfinished {
  background: #fee2e2;
  color: #e11d48;
  border-color: #fecdd3;
}

.stat-pill.completed {
  background: #f8fafc;
  color: #94a3b8;
  border: 1px solid #e2e8f0;
}
.stat-pill.completed.finished {
  background: #ecfdf5;
  color: #059669;
  border-color: #a7f3d0;
}

.meta { margin-top: 12px; color: #64748b; font-size: 13px; }
.action { margin-top: 14px; color: #2563eb; font-size: 13px; font-weight: 600; display: inline-flex; align-items: center; }

.range-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  min-height: 32px;
}
.range-title {
  color: #475569;
  font-size: 13px;
  font-weight: 700;
  line-height: 32px;
}
.range-hint {
  color: #94a3b8;
  font-size: 12px;
  line-height: 32px;
}
.range-bar :deep(.el-radio-button__inner) {
  height: 32px;
  padding: 0 16px;
  font-size: 13px;
}
:deep(.visual-dialog .el-dialog__body) {
  padding: 0;
  height: 75vh;
  min-height: 600px;
  max-height: 820px;
  overflow: hidden;
  background: #ffffff;
}
</style>
