<template>
  <AdminShell title="调度看板" subtitle="实时掌握每名工人的班次、并发与当日产出">
    <section class="board-grid">
      <div class="board-card">
        <div class="board-num">{{ list.length }}</div>
        <div class="board-label">在岗工人（启用）</div>
      </div>
      <div class="board-card">
        <div class="board-num" style="color: #16a34a">{{ availableCount }}</div>
        <div class="board-label">当前可派</div>
      </div>
      <div class="board-card">
        <div class="board-num" style="color: #d97706">{{ activeSum }}</div>
        <div class="board-label">在途工单</div>
      </div>
      <div class="board-card">
        <div class="board-num" style="color: #2563eb">{{ doneSum }}</div>
        <div class="board-label">今日已完成</div>
      </div>
    </section>

    <section class="panel">
      <div class="toolbar">
        <div class="panel-title">工人状态</div>
        <el-button size="small" @click="load">↻ 刷新</el-button>
      </div>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="工人" width="150">
          <template #default="{ row }">
            <div class="worker-name">{{ row.name }}</div>
            <div class="worker-sub">{{ row.username }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="buildingNames" label="管辖楼栋" min-width="220">
          <template #default="{ row }">{{ (row.buildingNames || []).join('、') || '—' }}</template>
        </el-table-column>
        <el-table-column label="今日班次" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="shiftTagType(row.todayShift)" size="small">{{ shiftText(row.todayShift) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="可派状态" width="100" align="center">
          <template #default="{ row }">
            <span class="status-dot" :class="row.available ? 'ok' : 'no'"></span>
            {{ row.available ? '可派' : '暂停' }}
          </template>
        </el-table-column>
        <el-table-column label="在途/最大并发" width="150" align="center">
          <template #default="{ row }">
            <el-progress
              :percentage="Math.min(100, Math.round((row.activeOrders / Math.max(row.maxConcurrent, 1)) * 100))"
              :stroke-width="10"
              style="width: 110px"
            />
            <div class="load-text">{{ row.activeOrders }} / {{ row.maxConcurrent }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="todayCompleted" label="今日完成" width="100" align="center" />
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无可展示工人" class="empty" />
    </section>
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { WorkerBoardItem } from '../api'
import { apiAdminWorkerBoard } from '../api'
import AdminShell from '../components/AdminShell.vue'

const list = ref<WorkerBoardItem[]>([])
const loading = ref(false)

const availableCount = computed(() => list.value.filter((w) => w.available).length)
const activeSum = computed(() => list.value.reduce((sum, w) => sum + w.activeOrders, 0))
const doneSum = computed(() => list.value.reduce((sum, w) => sum + w.todayCompleted, 0))

async function load() {
  loading.value = true
  try {
    list.value = await apiAdminWorkerBoard()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function shiftText(shift: string) {
  switch (shift) {
    case 'DAY':
      return '全天'
    case 'MORNING':
      return '午班'
    case 'AFTERNOON':
      return '晚班'
    case 'OFF':
      return '休息'
    default:
      return '未排班'
  }
}

function shiftTagType(shift: string) {
  if (shift === 'OFF') return 'info'
  if (shift === 'MORNING' || shift === 'AFTERNOON') return 'warning'
  return 'success'
}

onMounted(load)
</script>

<style scoped>
.board-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 18px;
}
.board-card {
  padding: 20px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
}
.board-num {
  font-size: 30px;
  font-weight: 800;
  color: #1e293b;
}
.board-label {
  margin-top: 8px;
  color: #64748b;
  font-size: 13px;
}
.panel {
  padding: 18px 20px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
}
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.panel-title {
  color: #334155;
  font-size: 15px;
  font-weight: 800;
}
.worker-name {
  font-weight: 700;
}
.worker-sub {
  color: #94a3b8;
  font-size: 12px;
}
.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 5px;
  border-radius: 50%;
}
.status-dot.ok {
  background: #22c55e;
}
.status-dot.no {
  background: #f43f5e;
}
.load-text {
  margin-top: 5px;
  color: #94a3b8;
  font-size: 12px;
}
.empty {
  padding: 24px 0;
}
</style>
