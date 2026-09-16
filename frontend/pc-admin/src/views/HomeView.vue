<template>
  <AdminShell title="工单总览" subtitle="宿管报修 → 自动派单 → 维修完工，全流程一屏掌握">
    <section class="stat-grid">
      <div class="stat-card">
        <div class="stat-num">{{ total }}</div>
        <div class="stat-label">工单总数（当前筛选）</div>
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

    <section v-if="guard.level !== 'ok' || guard.paused" class="guard-banner" :class="guard.paused ? 'is-guard' : 'is-warn'">
      <div class="guard-text">
        <strong>{{ guard.paused ? '人工处置模式（自动派单已暂停）' : '积压预警' }}</strong>
        <span>{{ guard.message }}</span>
        <span class="guard-meta">
          待派 {{ guard.pendingCount }} 单 · 在岗 {{ guard.onDutyCount }} 人 · 最长等待 {{ guard.waitText }} ·
          预警线 {{ guard.warnRatio }}× 在岗 / 保护线 {{ guard.guardRatio }}× 在岗
        </span>
      </div>
      <el-button v-if="guard.paused" type="primary" size="small" @click="resumeDispatch">恢复自动派单</el-button>
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
        <el-radio-group v-model="query.days" @change="onDaysChange">
          <el-radio-button :value="1">1天</el-radio-button>
          <el-radio-button :value="3">3天</el-radio-button>
          <el-radio-button :value="7">7天</el-radio-button>
          <el-radio-button :value="30">30天</el-radio-button>
        </el-radio-group>
        <el-button type="primary" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
        <el-tag :type="guard.paused ? 'danger' : 'success'" effect="plain" size="small">
          派单模式：{{ guard.mode === 'manual' ? '人工处置' : '自动派单' }}
        </el-tag>
        <el-button type="success" plain :disabled="todoCount === 0" :loading="batching" @click="runBatchDispatch">
          批量派单（待派 {{ todoCount }}）
        </el-button>
      </div>
    </section>

    <section class="panel table-panel">
      <div class="panel-title table-title">
        工单列表
        <el-tag type="info" effect="plain" size="small">共 {{ visibleOrders.length }} 条</el-tag>
      </div>
      <el-table :data="visibleOrders" v-loading="loading" border stripe :row-class-name="rowClassName">
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
        <el-table-column label="位置" width="165">
          <template #default="{ row }">{{ row.buildingName }} {{ row.floor }}F-{{ row.room }}</template>
        </el-table-column>
        <el-table-column prop="faultTypeName" label="类型" width="90" />
        <el-table-column label="状态" width="176">
          <template #default="{ row }">
            <span class="status-badge" :class="'st' + row.status">{{ row.statusText }}</span>
            <el-tag
              v-if="row.manualReview === 1"
              :type="row.externalMark === 1 ? 'info' : 'danger'"
              effect="plain"
              size="small"
              class="manual-tag"
            >
              {{ row.externalMark === 1 ? '外援处理' : '待管理员处置' }}
            </el-tag>
            <el-tag v-if="row.dispatchLocked === 1" type="warning" effect="dark" size="small" class="manual-tag">🔒 已锁定</el-tag>
            <div v-if="row.status === 1 && row.pendingReason" class="pending-reason">{{ row.pendingReason }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="workerName" label="维修工人" width="100">
          <template #default="{ row }">{{ row.workerName || '—' }}</template>
        </el-table-column>
        <el-table-column label="派单评分" width="150">
          <template #default="{ row }">
            <el-tooltip
              v-if="row.dispatchScore !== undefined && row.dispatchScore > 0"
              :content="`技能 ${row.skillScore} · 路网距离 ${row.distanceScore} · 负载 ${row.loadScore}`"
            >
              <span class="score-text">{{ row.dispatchScore }}</span>
            </el-tooltip>
            <span v-else class="score-empty">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="reporterName" label="报修宿管" width="110" />
        <el-table-column label="操作" width="600" class-name="op-cell">
          <template #default="{ row }">
            <el-button size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 2 || row.status === 3"
              size="small"
              type="success"
              plain
              @click="completeOrder(row)"
            >完工</el-button>
            <el-button v-if="row.status === 1" size="small" type="warning" plain @click="toggleLock(row)">
              {{ row.dispatchLocked === 1 ? '解锁' : '锁定' }}
            </el-button>
            <el-button v-if="row.status === 1 || row.status === 2" size="small" plain @click="editPriority(row)">优先级</el-button>
            <el-button
              v-if="row.status === 1 || row.status === 2 || row.status === 3"
              size="small"
              type="primary"
              plain
              @click="openAssign(row)"
            >{{ row.status === 1 ? '手动派单' : '改派' }}</el-button>
            <el-button
              v-if="row.manualReview === 1 && row.externalMark !== 1 && row.status !== 4"
              size="small"
              type="danger"
              plain
              :loading="externalMarking === row.id"
              @click="markExternal(row)"
            >标记外援</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty
        v-if="!loading && visibleOrders.length === 0"
        description="当前筛选条件下暂无工单"
        class="table-empty"
      />
      <div v-if="total > pageSize" class="pager">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="pageChange"
        />
      </div>
    </section>

    <el-dialog append-to-body v-model="assignVisible" :title="assignTitle" width="480px">
      <template v-if="assignTarget">
        <p class="assign-hint">
          工单 {{ assignTarget.orderNo }} · {{ assignTarget.buildingName }} {{ assignTarget.floor }}F-{{ assignTarget.room }}
          · {{ assignTarget.faultTypeName }}
        </p>
        <p v-if="assignTarget.manualReview === 1" class="assign-note">
          该工单为「待管理员处置」：可协商派给内部工人，或直接标记外援处理。
        </p>
        <el-select v-model="assignWorkerId" placeholder="选择负责该楼栋的工人" style="width: 100%">
          <el-option v-for="w in assignableWorkers" :key="w.id" :label="`${w.name}（${w.username}）· 在途/并发可派`" :value="w.id" />
        </el-select>
        <p v-if="assignableWorkers.length === 0" class="assign-empty">该楼栋暂无可用工人（可能都在休息或满载）</p>
      </template>
      <template #footer>
        <el-button @click="assignVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" :disabled="!assignWorkerId" @click="submitAssign">
          确认派单
        </el-button>
      </template>
    </el-dialog>
    <el-drawer append-to-body v-model="detailVisible" title="工单详情" size="480px">
      <el-descriptions v-if="detailRow" :column="1" border>
        <el-descriptions-item label="工单号">{{ detailRow.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="标题">{{ detailRow.title }}</el-descriptions-item>
        <el-descriptions-item label="位置">{{ detailRow.buildingName }} · {{ detailRow.floor }} 层 {{ detailRow.room }} 室</el-descriptions-item>
        <el-descriptions-item label="类型">{{ detailRow.faultTypeName }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detailRow.statusText }}</el-descriptions-item>
        <el-descriptions-item v-if="detailRow.manualReview === 1" label="处置标记">
          {{ detailRow.externalMark === 1 ? '已转外援处理' : '待管理员处置（可协商派单或外援）' }}
        </el-descriptions-item>
        <el-descriptions-item label="报修宿管">{{ detailRow.reporterName }}</el-descriptions-item>
        <el-descriptions-item label="维修工人">
          {{ detailRow.workerName || '—' }}
          <span v-if="detailRow.workerPhone" style="color:#94a3b8">（{{ detailRow.workerPhone }}）</span>
        </el-descriptions-item>
        <el-descriptions-item label="故障描述">{{ detailRow.description }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detailRow.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ detailRow.updatedAt }}</el-descriptions-item>
        <el-descriptions-item v-if="detailRow.dispatchScore !== undefined && detailRow.dispatchScore > 0" label="派单解释">
          <div class="score-explain">
            <div class="score-line">
              <span class="score-key">综合得分</span>
              <el-progress :percentage="Math.round((detailRow.dispatchScore || 0) * 100)" :stroke-width="10" />
              <span class="score-val">{{ detailRow.dispatchScore }}</span>
            </div>
            <div class="score-line">
              <span class="score-key">技能 40%</span>
              <el-progress :percentage="Math.round((detailRow.skillScore || 0) * 100)" :stroke-width="8" color="#2563eb" />
              <span class="score-val">{{ detailRow.skillScore }}</span>
            </div>
            <div class="score-line">
              <span class="score-key">路网距离 30%</span>
              <el-progress :percentage="Math.round((detailRow.distanceScore || 0) * 100)" :stroke-width="8" color="#16a34a" />
              <span class="score-val">{{ detailRow.distanceScore }}</span>
            </div>
            <div class="score-line">
              <span class="score-key">负载 30%</span>
              <el-progress :percentage="Math.round((detailRow.loadScore || 0) * 100)" :stroke-width="8" color="#d97706" />
              <span class="score-val">{{ detailRow.loadScore }}</span>
            </div>
            <p class="score-tip">
              权重默认 技能 40% / 距离 30% / 负载 30%（可在「派单规则」调整）；距离按校园路网最短路计算，
              建筑未接入路网时回退楼栋坐标欧氏距离；工人负载高于人均 1.2 倍会在总分中额外扣分（最多 0.5）。
            </p>
            <p class="score-tip">
              当前派单模式：{{ guard.mode === 'manual' ? '人工处置（保护线已触发，自动派单暂停）' : '自动派单' }}
              · 在岗 {{ guard.onDutyCount }} 人 · 待派 {{ guard.pendingCount }} 单
            </p>
          </div>
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminUser, DispatchGuard, OrderItem } from '../api'
import {
  apiAdminBatchDispatch,
  apiAdminDispatchGuard,
  apiAdminOrderExternal,
  apiAdminOrderComplete,
  apiAdminOrderLock,
  apiAdminOrderPriority,
  apiAdminOrderReassign,
  apiAdminOrders,
  apiAdminStats,
  apiAdminUsers,
  apiResumeAutoDispatch,
} from '../api'
import AdminShell from '../components/AdminShell.vue'
import { useAuthStore } from '../stores/auth'

const orders = ref<OrderItem[]>([])
const statsValue = ref<{ status: { status: number; count: number }[] } | null>(null)
const total = ref(0)
const page = ref(1)
const pageSize = 20
const workers = ref<AdminUser[]>([])
const loading = ref(false)
const batching = ref(false)
const assigning = ref(false)
const externalMarking = ref<number | null>(null)
const assignVisible = ref(false)
const assignTarget = ref<OrderItem | null>(null)
const assignWorkerId = ref<number | null>(null)
const detailVisible = ref(false)
const detailRow = ref<OrderItem | null>(null)
const guard = ref<DispatchGuard>({
  pendingCount: 0,
  onDutyCount: 0,
  longestWaitMinutes: 0,
  warnRatio: 3,
  guardRatio: 5,
  warnHours: 2,
  guardHours: 4,
  level: 'ok',
  paused: false,
  mode: 'auto',
  message: '',
  waitText: '',
})

async function loadGuard() {
  try {
    guard.value = await apiAdminDispatchGuard()
  } catch {
    // 保护状态查询失败不阻塞工单列表
  }
}

async function resumeDispatch() {
  try {
    await ElMessageBox.confirm('确认已处理积压（增援 / 调班 / 外援或手动指派）并恢复自动派单？', '恢复自动派单')
  } catch {
    return
  }
  try {
    guard.value = await apiResumeAutoDispatch()
    ElMessage.success('已恢复自动派单')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}
const query = reactive({ status: 0, buildingText: '', days: 3 })
const auth = useAuthStore()
let ws: WebSocket | null = null
let refreshTimer: ReturnType<typeof setTimeout> | null = null

const pendingCount = computed(() => { const row = statsValue.value?.status.find((s) => s.status === 1); return (row?.count || 0) + (statsValue.value?.status.find((s) => s.status === 2)?.count || 0) })
const todoCount = computed(() => { const row = statsValue.value?.status.find((s) => s.status === 1); return (row?.count || 0) })
const workingCount = computed(() => { const row = statsValue.value?.status.find((s) => s.status === 3); return (row?.count || 0) })
const doneCount = computed(() => { const row = statsValue.value?.status.find((s) => s.status === 4); return (row?.count || 0) })

const visibleOrders = computed(() => {
  const status = Number(query.status || 0)
  if (status === 0) return orders.value
  return orders.value.filter((o) => o.status === status)
})

const assignableWorkers = computed(() => {
  if (!assignTarget.value) return []
  return workers.value.filter((w) => (w.buildingIds || []).includes(assignTarget.value!.buildingId))
})
const assignTitle = computed(() => {
  if (!assignTarget.value) return '手动派单'
  if (assignTarget.value.manualReview === 1) return '协商派单／手动指派'
  return assignTarget.value.status === 2 ? '改派工单' : '手动派单'
})

async function loadStats() {
  try {
    statsValue.value = await apiAdminStats(query.days)
  } catch {
    // 统计卡失败不阻塞列表
  }
}

async function load() {
  loading.value = true
  const buildingId = Number(query.buildingText || 0)
  try {
    const res = await apiAdminOrders(0, buildingId > 0 ? buildingId : 0, page.value, pageSize, query.days)
    orders.value = res.list
    total.value = res.total
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function onDaysChange() {
  page.value = 1
  load()
  loadStats()  // 统计卡与时间筛选联动
}
function pageChange(p: number) {
  page.value = p
  load()
}

async function loadWorkers() {
  try {
    workers.value = await apiAdminUsers({ role: 2 })
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function reset() {
  query.status = 0
  query.buildingText = ''
  page.value = 1
  load()
  loadStats()
}

function openDetail(row: OrderItem) {
  detailRow.value = row
  detailVisible.value = true
}
// 管理员代为完工：演示时不必逐个登录工人账号，效果与工人端完工一致
async function completeOrder(row: OrderItem) {
  try {
    await ElMessageBox.confirm(
      `把工单 ${row.orderNo}（${row.room} 室）标记为已完成？相当于该工人已完工：会记录完成时间、清除故障标记并通知宿管。`,
      '代为完工',
      { confirmButtonText: '确认完工', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    await apiAdminOrderComplete(row.id)
    ElMessage.success('已代为完工')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function toggleLock(row: OrderItem) {
  const lock = row.dispatchLocked === 1 ? 0 : 1
  try {
    await ElMessageBox.confirm(
      lock ? `锁定工单 ${row.orderNo}？锁定后不参与自动派单与批量派单。` : `解锁工单 ${row.orderNo}？解锁后会重新进入自动派单队列。`,
      lock ? '锁定工单' : '解锁工单',
    )
  } catch {
    return
  }
  try {
    await apiAdminOrderLock(row.id, lock)
    ElMessage.success(lock ? '已锁定' : '已解锁')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function editPriority(row: OrderItem) {
  let input: string
  try {
    const res = await ElMessageBox.prompt('优先级 1（最低）~ 5（最高），待派队列按优先级倒序取单', '调整优先级', {
      inputValue: String(row.priority || 1),
      inputPattern: /^[1-5]$/,
      inputErrorMessage: '请输入 1~5 的整数',
    })
    input = res.value
  } catch {
    return
  }
  try {
    await apiAdminOrderPriority(row.id, Number(input))
    ElMessage.success('优先级已更新')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}function rowClassName({ row }: { row: OrderItem }) {
  return row.manualReview === 1 ? 'row-manual' : ''
}

async function markExternal(row: OrderItem) {
  try {
    await ElMessageBox.confirm(`将工单 ${row.orderNo} 标记为外援处理？标记后不再参与自动派单。`, '标记外援')
  } catch {
    return
  }
  externalMarking.value = row.id
  try {
    await apiAdminOrderExternal(row.id)
    ElMessage.success('已标记外援处理')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    externalMarking.value = null
  }
}
function openAssign(row: OrderItem) {
  assignTarget.value = row
  assignWorkerId.value = row.workerId || null
  assignVisible.value = true
}

async function submitAssign() {
  if (!assignTarget.value || !assignWorkerId.value) return
  assigning.value = true
  try {
    await apiAdminOrderReassign(assignTarget.value.id, assignWorkerId.value)
    ElMessage.success('派单成功')
    assignVisible.value = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    assigning.value = false
  }
}

async function runBatchDispatch() {
  try {
    await ElMessageBox.confirm('将对当前全部待派工单执行容量约束的最小代价批量派单，是否继续？', '批量派单')
  } catch {
    return
  }
  batching.value = true
  try {
    const res = await apiAdminBatchDispatch({})
    ElMessage.success(`批量派单完成：成功 ${res.dispatched.length} 单，剩余 ${res.remained} 单`)
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    batching.value = false
  }
}

function scheduleRefresh() {
  if (refreshTimer) clearTimeout(refreshTimer)
  refreshTimer = setTimeout(() => {
    load()
    loadStats()
    loadGuard()
  }, 350)
}

function connectWS() {
  if (!auth.token) return
  const proto = location.protocol === 'https:' ? 'wss://' : 'ws://'
  ws = new WebSocket(`${proto}${location.host}/ws/orders?token=${encodeURIComponent(auth.token)}`)
  ws.onmessage = () => {
    scheduleRefresh()
    window.dispatchEvent(new Event('rv-notify-refresh'))
  }
  ws.onclose = () => {
    ws = null
    setTimeout(connectWS, 3000)
  }
}

onMounted(() => {
  load()
  loadStats()
  loadGuard()
  loadWorkers()
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
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 10px 30px rgba(46, 68, 112, 0.08);
  animation: rv-fade-up 0.42s cubic-bezier(0.22, 1, 0.36, 1) both;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.3s ease;
}

.stat-card:nth-child(1) { background: var(--rv-grad-1); }
.stat-card:nth-child(2) { background: var(--rv-grad-4); }
.stat-card:nth-child(3) { background: var(--rv-grad-6); }
.stat-card:nth-child(4) { background: var(--rv-grad-10); }

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 16px 38px rgba(46, 68, 112, 0.14);
}

.stat-num {
  font-size: 30px;
  font-weight: 800;
  color: #2b3445;
  line-height: 1;
}

.stat-label {
  margin-top: 10px;
  color: #5a6a85;
  font-size: 13px;
}

.panel {
  padding: 18px 20px;
  margin-bottom: 18px;
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 10px 30px rgba(46, 68, 112, 0.08);
}

.panel-title {
  margin-bottom: 14px;
  color: #2b3445;
  font-size: 14px;
  font-weight: 800;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
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
  font-weight: 700;
  border-radius: 999px;
}

.st1,
.st2 {
  color: #b96b1c;
  background: var(--rv-grad-4);
}
.st3 {
  color: #2462d9;
  background: var(--rv-grad-1);
}
.st4 {
  color: #17865a;
  background: var(--rv-grad-6);
}
.st5 {
  color: #5a6a85;
  background: var(--rv-grad-8);
}

.pending-reason {
  margin-top: 5px;
  color: #b34568;
  font-size: 11px;
  line-height: 1.4;
  max-width: 140px;
}
.row-manual :deep(td) {
  background: rgba(255, 240, 245, 0.72) !important;
}

.manual-tag {
  margin-left: 6px;
  font-weight: 700;
}
.op-cell {
  white-space: nowrap;
}

.assign-note {
  margin: -4px 0 12px;
  color: #b34568;
  font-size: 13px;
}

.table-empty {
  padding: 30px 0;
}

.score-text {
  color: #2462d9;
  font-weight: 700;
}

.score-empty {
  color: #c2cbdc;
}

.assign-hint {
  margin-bottom: 12px;
  color: #5a6a85;
  font-size: 13px;
}

.assign-empty {
  margin-top: 10px;
  color: #b34568;
  font-size: 13px;
}

.guard-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 18px;
  margin-bottom: 16px;
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.7);
}
.guard-banner.is-warn {
  background: var(--rv-grad-4);
}
.guard-banner.is-guard {
  background: var(--rv-grad-2);
}
.guard-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  color: #2b3445;
  font-size: 13px;
}
.guard-meta {
  color: #5a6a85;
  font-size: 12px;
}
.score-explain {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.score-line {
  display: flex;
  align-items: center;
  gap: 8px;
}
.score-key {
  width: 82px;
  color: #5a6a85;
  font-size: 12px;
}
.score-line :deep(.el-progress) {
  flex: 1;
}
.score-val {
  width: 52px;
  color: #2b3445;
  font-size: 12px;
  font-weight: 700;
  text-align: right;
}
.score-tip {
  margin: 4px 0 0;
  color: #8a97ad;
  font-size: 12px;
  line-height: 1.6;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}
</style>