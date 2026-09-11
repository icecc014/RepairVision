<template>
  <AdminShell title="建筑可视化" subtitle="查看全部建筑，点开任意楼栋查看 16 间户型与 3D 视图">
    <section class="panel">
      <div class="range-bar">
        <span class="range-title">报修时间窗</span>
        <el-radio-group v-model="days" size="small" @change="onDaysChange">
          <el-radio-button :value="1">1天</el-radio-button>
          <el-radio-button :value="3">3天</el-radio-button>
          <el-radio-button :value="7">7天</el-radio-button>
          <el-radio-button :value="30">1个月</el-radio-button>
        </el-radio-group>
        <span class="range-hint">近 {{ days }} 天共 {{ orders.length }} 单，2D / 3D 视图与红点按此范围显示</span>
      </div>

      <div class="grid">
        <button
          v-for="b in buildings"
          :key="b.id"
          class="building-card"
          @click="open(b)"
        >
          <div class="card-head">
            <span class="code">{{ b.code }}</span>
            <span class="badge">{{ orderCount(b.id) }} 单</span>
          </div>
          <div class="name">{{ b.name }}</div>
          <div class="meta">{{ b.floors }} 层 · 每层 {{ b.roomsPerFloor }} 间 · ({{ b.posX }}, {{ b.posY }})</div>
          <div class="action">查看户型 / 3D</div>
        </button>
      </div>
    </section>

    <el-dialog append-to-body
      v-model="dialogVisible"
      :title="selected ? `${selected.code} ${selected.name}` : ''"
      width="900px"
      top="6vh"
    >
      <AdminBuildingVisual v-if="selected" :building="selected" :orders="orders" />
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
const dialogVisible = ref(false)

async function loadOrders() {
  try {
    const orderPage = await apiAdminOrders(0, 0, 1, 200, days.value)
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

function orderCount(id: number) {
  return orders.value.filter((o) => o.buildingId === id).length
}

async function open(b: AdminBuilding) {
  selected.value = b
  dialogVisible.value = true
  // 打开弹窗时按当前时间窗刷新，确保红点与主页筛选一致
  await loadOrders()
}

onMounted(load)
</script>

<style scoped>
.panel { padding: 20px; background: #fff; border-radius: 14px; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 16px; }
.building-card { text-align: left; padding: 18px; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 14px; cursor: pointer; transition: all .15s; }
.building-card:hover { transform: translateY(-2px); box-shadow: 0 8px 20px rgba(37,99,235,.12); border-color:#93c5fd; }
.card-head { display: flex; justify-content: space-between; align-items:center; }
.code { font-size: 18px; font-weight: 800; color:#1e3a8a; }
.badge { padding:2px 10px; background:#fee2e2; color:#b91c1c; border-radius:999px; font-size:12px; }
.name { margin-top:8px; font-weight:600; color:#1e293b; }
.meta { margin-top:4px; color:#94a3b8; font-size:12px; }
.action { margin-top:12px; color:#2563eb; font-size:13px; font-weight:600; }
</style>