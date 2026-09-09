<template>
  <AdminShell title="数据统计看板" subtitle="工单状态、SLA 超时、处理时长、楼栋与类型分布">
    <section v-if="stats && sla" class="stats-body">
      <div class="status-grid">
        <div v-for="s in stats.status" :key="s.status" class="status-card" :class="'sc' + s.status">
          <div class="status-num">{{ s.count }}</div>
          <div class="status-label">{{ s.statusText }}</div>
        </div>
      </div>

      <div class="sla-grid">
        <div class="sla-card danger">
          <div class="sla-num">{{ sla.pendingOverdue }}</div>
          <div class="sla-label">待派超时（&gt; {{ sla.pendingTimeoutHours }}h）</div>
        </div>
        <div class="sla-card warning">
          <div class="sla-num">{{ sla.dispatchedOverdue }}</div>
          <div class="sla-label">已派未开工超时（&gt; {{ sla.dispatchedTimeoutHours }}h）</div>
        </div>
        <div class="sla-card blue">
          <div class="sla-num">{{ sla.avgDispatchMinutes }}</div>
          <div class="sla-label">平均派单时长（分钟）</div>
        </div>
        <div class="sla-card green">
          <div class="sla-num">{{ sla.avgRepairMinutes }}</div>
          <div class="sla-label">平均维修时长（分钟）</div>
        </div>
      </div>

      <div class="grid-2">
        <section class="panel">
          <div class="panel-title">近 7 日工单趋势</div>
          <div class="bar-row" v-for="d in stats.recent" :key="d.date">
            <span class="bar-label">{{ d.date.slice(5) }}</span>
            <div class="bar-track">
              <div class="bar-fill" :style="{ width: barWidth(d.count) }"></div>
            </div>
            <span class="bar-num">{{ d.count }}</span>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">楼栋工单分布</div>
          <div v-if="stats.buildings.length === 0" class="no-data">暂无数据</div>
          <div class="bar-row" v-for="b in stats.buildings" :key="b.buildingId">
            <span class="bar-label">{{ b.buildingName || '楼栋#' + b.buildingId }}</span>
            <div class="bar-track">
              <div class="bar-fill blue" :style="{ width: buildingBarWidth(b.count) }"></div>
            </div>
            <span class="bar-num">{{ b.count }}</span>
          </div>
        </section>
      </div>

      <section class="panel">
        <div class="panel-title">维修类型分布</div>
        <el-table :data="stats.faults" border stripe size="default">
          <el-table-column prop="faultTypeName" label="类型" />
          <el-table-column prop="faultType" label="编码" width="160" />
          <el-table-column prop="count" label="工单数" width="120" />
          <el-table-column label="占比" min-width="260">
            <template #default="{ row }">
              <div class="mini-bar">
                <div class="mini-fill" :style="{ width: faultWidth(row.count) }"></div>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section v-if="sla.overdueOrders.length > 0" class="panel">
        <div class="panel-title">当前 SLA 超时工单</div>
        <el-table :data="sla.overdueOrders" border stripe size="default">
          <el-table-column prop="orderNo" label="工单号" width="180" />
          <el-table-column label="位置" min-width="180">
            <template #default="{ row }">{{ row.buildingName }} {{ row.floor }}F-{{ row.room }}</template>
          </el-table-column>
          <el-table-column prop="faultTypeName" label="类型" width="110" />
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'warning' : 'danger'" size="small">{{ row.statusText }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="workerName" label="工人" width="110">
            <template #default="{ row }">{{ row.workerName || '—' }}</template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="175" />
        </el-table>
      </section>
    </section>
    <el-skeleton v-else :rows="6" animated />
  </AdminShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { AdminStats, SlaOverview } from '../api'
import { apiAdminSlaOverview, apiAdminStats } from '../api'
import AdminShell from '../components/AdminShell.vue'

const stats = ref<AdminStats | null>(null)
const sla = ref<SlaOverview | null>(null)

async function load() {
  try {
    stats.value = await apiAdminStats()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
  try {
    sla.value = await apiAdminSlaOverview()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function maxCount() {
  return Math.max(...(stats.value?.recent.map((d) => d.count) || [1]), 1)
}

function maxBuilding() {
  return Math.max(...(stats.value?.buildings.map((b) => b.count) || [1]), 1)
}

function maxFault() {
  return Math.max(...(stats.value?.faults.map((f) => f.count) || [1]), 1)
}

function barWidth(count: number) {
  return `${Math.round((count / maxCount()) * 100)}%`
}

function buildingBarWidth(count: number) {
  return `${Math.round((count / maxBuilding()) * 100)}%`
}

function faultWidth(count: number) {
  return `${Math.round((count / maxFault()) * 100)}%`
}

onMounted(load)
</script>

<style scoped>
.stats-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.status-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 14px;
}
.status-card {
  padding: 22px 16px;
  color: #fff;
  text-align: center;
  border-radius: 14px;
}
.status-num {
  font-size: 30px;
  font-weight: 800;
  line-height: 1;
}
.status-label {
  margin-top: 8px;
  font-size: 13px;
  opacity: 0.92;
}
.sc1 {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}
.sc2 {
  background: linear-gradient(135deg, #fb923c, #ea580c);
}
.sc3 {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
}
.sc4 {
  background: linear-gradient(135deg, #22c55e, #16a34a);
}
.sc5 {
  background: linear-gradient(135deg, #94a3b8, #64748b);
}
.sla-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}
.sla-card {
  padding: 20px;
  color: #fff;
  border-radius: 14px;
}
.sla-card.danger {
  background: linear-gradient(135deg, #f43f5e, #e11d48);
}
.sla-card.warning {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}
.sla-card.blue {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
}
.sla-card.green {
  background: linear-gradient(135deg, #22c55e, #16a34a);
}
.sla-num {
  font-size: 28px;
  font-weight: 800;
  line-height: 1;
}
.sla-label {
  margin-top: 10px;
  font-size: 12px;
  opacity: 0.94;
}
.grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.panel {
  padding: 18px 20px;
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
.bar-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 8px 0;
}
.bar-label {
  width: 62px;
  color: #64748b;
  font-size: 12px;
}
.bar-track,
.mini-bar {
  flex: 1;
  height: 12px;
  overflow: hidden;
  background: #f1f5f9;
  border-radius: 999px;
}
.bar-fill,
.mini-fill {
  height: 100%;
  background: linear-gradient(90deg, #60a5fa, #2563eb);
  border-radius: 999px;
  transition: width 0.4s ease;
}
.bar-fill.blue {
  background: linear-gradient(90deg, #34d399, #059669);
}
.bar-num {
  width: 36px;
  color: #334155;
  font-size: 13px;
  font-weight: 700;
  text-align: right;
}
.mini-bar {
  height: 8px;
}
.no-data {
  color: #94a3b8;
  font-size: 13px;
  text-align: center;
  padding: 20px;
}
</style>