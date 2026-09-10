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
              :percentage="revealed ? Math.min(100, Math.round((row.activeOrders / Math.max(row.maxConcurrent, 1)) * 100)) : 0"
              :duration="0.9"
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
import { computed, nextTick, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { WorkerBoardItem } from '../api'
import { apiAdminWorkerBoard } from '../api'
import AdminShell from '../components/AdminShell.vue'

const list = ref<WorkerBoardItem[]>([])
const loading = ref(false)
const revealed = ref(false)

const availableCount = computed(() => list.value.filter((w) => w.available).length)
const activeSum = computed(() => list.value.reduce((sum, w) => sum + w.activeOrders, 0))
const doneSum = computed(() => list.value.reduce((sum, w) => sum + w.todayCompleted, 0))

async function load() {
  loading.value = true
  revealed.value = false
  try {
    list.value = await apiAdminWorkerBoard()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
  // 数据到齐后再让进度条从 0 涨到目标值
  await nextTick()
  revealed.value = true
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
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 10px 30px rgba(46, 68, 112, 0.08);
  animation: rv-fade-up 0.42s cubic-bezier(0.22, 1, 0.36, 1) both;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.3s ease;
}
.board-card:nth-child(1) { background: var(--rv-grad-1); }
.board-card:nth-child(2) { background: var(--rv-grad-6); }
.board-card:nth-child(3) { background: var(--rv-grad-4); }
.board-card:nth-child(4) { background: var(--rv-grad-10); }
.board-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 16px 38px rgba(46, 68, 112, 0.14);
}
.board-num {
  font-size: 30px;
  font-weight: 800;
  color: #2b3445;
}
.board-label {
  margin-top: 8px;
  color: #5a6a85;
  font-size: 13px;
}
.panel {
  padding: 18px 20px;
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 10px 30px rgba(46, 68, 112, 0.08);
}
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.panel-title {
  color: #2b3445;
  font-size: 15px;
  font-weight: 800;
}
.worker-name {
  font-weight: 700;
  color: #2b3445;
}
.worker-sub {
  color: #8a97ad;
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
  background: linear-gradient(135deg, #7fe6c8, #22b573);
  box-shadow: 0 0 0 3px rgba(34, 181, 115, 0.16);
}
.status-dot.no {
  background: linear-gradient(135deg, #ffb9cd, #e0648a);
  box-shadow: 0 0 0 3px rgba(224, 100, 138, 0.16);
}
.load-text {
  margin-top: 5px;
  color: #8a97ad;
  font-size: 12px;
}
.empty {
  padding: 24px 0;
}
</style>
