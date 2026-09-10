<template>
  <AdminShell title="报修记录" subtitle="宿舍报修全流程台账：哪个房间、什么故障、谁报修、谁处理、当前进度">
    <section class="stat-grid">
      <div class="stat-card">
        <div class="stat-num">{{ summary.total }}</div>
        <div class="stat-label">报修记录（当前筛选）</div>
      </div>
      <div class="stat-card">
        <div class="stat-num" style="color: #b96b1c">{{ summary.pending }}</div>
        <div class="stat-label">待处理（待派单 + 已派单）</div>
      </div>
      <div class="stat-card">
        <div class="stat-num" style="color: #2462d9">{{ summary.working }}</div>
        <div class="stat-label">维修中</div>
      </div>
      <div class="stat-card">
        <div class="stat-num" style="color: #17865a">{{ summary.done }}</div>
        <div class="stat-label">已完成</div>
      </div>
    </section>

    <section class="panel filter-panel">
      <div class="panel-title">筛选条件</div>
      <div class="filter-row">
        <el-select v-model="query.buildingId" placeholder="全部楼栋" clearable style="width: 160px" @change="reload">
          <el-option v-for="b in buildings" :key="b.id" :label="b.name" :value="b.id" />
        </el-select>
        <el-select v-model="query.faultType" placeholder="全部故障类型" clearable style="width: 170px" @change="reload">
          <el-option v-for="f in faultTypes" :key="f.code" :label="f.name" :value="f.code" />
        </el-select>
        <el-select v-model="query.status" placeholder="全部状态" clearable style="width: 150px" @change="reload">
          <el-option label="待派单" :value="1" />
          <el-option label="已派单" :value="2" />
          <el-option label="维修中" :value="3" />
          <el-option label="已完成" :value="4" />
          <el-option label="已取消" :value="5" />
        </el-select>
        <el-radio-group v-model="query.days" @change="reload">
          <el-radio-button :value="1">1天</el-radio-button>
          <el-radio-button :value="3">3天</el-radio-button>
          <el-radio-button :value="7">7天</el-radio-button>
          <el-radio-button :value="30">1个月</el-radio-button>
        </el-radio-group>
        <el-input
          v-model="query.keyword"
          placeholder="搜索房间号 / 宿管 / 工人"
          clearable
          style="width: 220px"
          @keyup.enter="reload"
        />
        <el-button type="primary" @click="reload">查询</el-button>
        <el-button @click="reset">重置</el-button>
        <el-button type="success" plain :loading="exporting" @click="exportCsv">导出报修台账</el-button>
      </div>
    </section>

    <section class="panel table-panel">
      <div class="panel-title table-title">
        报修台账明细
        <el-tag type="info" effect="plain" size="small">共 {{ total }} 条</el-tag>
      </div>
      <el-table :data="records" v-loading="loading" border stripe>
        <el-table-column prop="createdAt" label="报修时间" width="170" />
        <el-table-column prop="buildingName" label="楼栋" width="130">
          <template #default="{ row }">{{ row.buildingName || '—' }}</template>
        </el-table-column>
        <el-table-column label="房间号" width="110">
          <template #default="{ row }">
            <span class="room-tag">{{ row.room }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="faultTypeName" label="故障类型" width="110">
          <template #default="{ row }">{{ row.faultTypeName || row.faultType }}</template>
        </el-table-column>
        <el-table-column prop="description" label="故障描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '—' }}</template>
        </el-table-column>
        <el-table-column prop="reporterName" label="报修宿管" width="110">
          <template #default="{ row }">{{ row.reporterName || '—' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <span class="status-badge" :class="'st' + row.status">{{ row.statusText }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="workerName" label="处理工人" width="110">
          <template #default="{ row }">{{ row.workerName || '—' }}</template>
        </el-table-column>
        <el-table-column label="完工时间" width="170">
          <template #default="{ row }">{{ row.completedAt || '—' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && records.length === 0" description="当前筛选条件下暂无报修记录" class="table-empty" />
      <div v-if="total > query.size" class="pager">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :total="total"
          :page-size="query.size"
          :current-page="query.page"
          @current-change="onPage"
        />
      </div>
      <p class="log-note">
        说明：本页只呈现业务报修信息；接口调用、耗时、IP 等技术日志已按天写入服务端 <code>logs/operation-YYYYMMDD.log</code> 文件。
      </p>
    </section>

    <el-drawer v-model="detailVisible" title="报修详情" size="480px">
      <template v-if="detailRow">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="工单号">{{ detailRow.orderNo }}</el-descriptions-item>
          <el-descriptions-item label="报修位置">
            {{ detailRow.buildingName }} · {{ detailRow.floor }} 层 {{ detailRow.room }} 室
          </el-descriptions-item>
          <el-descriptions-item label="故障类型">
            {{ detailRow.faultTypeName || detailRow.faultType }}
          </el-descriptions-item>
          <el-descriptions-item label="故障描述">{{ detailRow.description || '—' }}</el-descriptions-item>
          <el-descriptions-item label="报修宿管">{{ detailRow.reporterName || '—' }}</el-descriptions-item>
          <el-descriptions-item label="当前状态">
            <span class="status-badge" :class="'st' + detailRow.status">{{ detailRow.statusText }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="处理工人">
            {{ detailRow.workerName || '—' }}
            <span v-if="detailRow.workerPhone" class="muted">（{{ detailRow.workerPhone }}）</span>
          </el-descriptions-item>
          <el-descriptions-item label="工单来源">{{ sourceText(detailRow.source) }}</el-descriptions-item>
        </el-descriptions>

        <div class="timeline-title">处理时间线</div>
        <el-timeline>
          <el-timeline-item :timestamp="detailRow.createdAt" type="primary">宿管提交报修</el-timeline-item>
          <el-timeline-item v-if="detailRow.dispatchedAt" :timestamp="detailRow.dispatchedAt" type="warning">
            调度派单给 {{ detailRow.workerName || '维修工人' }}
          </el-timeline-item>
          <el-timeline-item v-if="detailRow.startedAt" :timestamp="detailRow.startedAt" type="primary">
            工人到场开工
          </el-timeline-item>
          <el-timeline-item
            v-if="detailRow.completedAt"
            :timestamp="detailRow.completedAt"
            type="success"
          >
            维修完成
          </el-timeline-item>
          <el-timeline-item v-if="detailRow.status === 5" :timestamp="detailRow.createdAt" type="info">
            工单已取消
          </el-timeline-item>
        </el-timeline>
      </template>
    </el-drawer>
  </AdminShell>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { AdminBuilding, AdminFaultType, RepairRecordItem, RepairRecordSummary } from '../api'
import {
  apiAdminBuildings,
  apiAdminFaultTypes,
  apiAdminRepairRecords,
} from '../api'
import AdminShell from '../components/AdminShell.vue'

const records = ref<RepairRecordItem[]>([])
const buildings = ref<AdminBuilding[]>([])
const faultTypes = ref<AdminFaultType[]>([])
const total = ref(0)
const loading = ref(false)
const exporting = ref(false)
const detailVisible = ref(false)
const detailRow = ref<RepairRecordItem | null>(null)
const summary = ref<RepairRecordSummary>({ total: 0, pending: 0, working: 0, done: 0, canceled: 0 })
const query = reactive({
  buildingId: undefined as number | undefined,
  faultType: '' as string,
  status: undefined as number | undefined,
  keyword: '',
  days: 3,
  page: 1,
  size: 20,
})

function params(page: number, size: number) {
  return {
    buildingId: query.buildingId || 0,
    faultType: query.faultType || '',
    status: query.status || 0,
    keyword: query.keyword || '',
    days: query.days,
    page,
    size,
  }
}

async function load() {
  loading.value = true
  try {
    const data = await apiAdminRepairRecords(params(query.page, query.size))
    records.value = data.list
    total.value = data.total
    summary.value = data.summary
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reload() {
  query.page = 1
  load()
}

function reset() {
  query.buildingId = undefined
  query.faultType = ''
  query.status = undefined
  query.keyword = ''
  query.days = 3
  reload()
}

function onPage(page: number) {
  query.page = page
  load()
}

function openDetail(row: RepairRecordItem) {
  detailRow.value = row
  detailVisible.value = true
}

function sourceText(source: string) {
  switch (source) {
    case 'dorm':
      return '宿管报修'
    case 'student':
      return '学生自助'
    default:
      return source || '宿管报修'
  }
}

function csvCell(value: unknown) {
  return `"${String(value ?? '').replaceAll('"', '""')}"`
}

async function exportCsv() {
  exporting.value = true
  try {
    const data = await apiAdminRepairRecords(params(1, 1000))
    if (!data.list.length) {
      ElMessage.warning('当前筛选条件下没有可导出的报修记录')
      return
    }
    const header = ['报修时间', '楼栋', '房间号', '楼层', '故障类型', '故障描述', '报修宿管', '状态', '处理工人', '派单时间', '完工时间']
    const rows = data.list.map((r) => [
      r.createdAt,
      r.buildingName,
      r.room,
      r.floor,
      r.faultTypeName || r.faultType,
      r.description,
      r.reporterName || '',
      r.statusText,
      r.workerName || '',
      r.dispatchedAt || '',
      r.completedAt || '',
    ])
    const content = [header, ...rows].map((row) => row.map(csvCell).join(',')).join('\r\n')
    const blob = new Blob(['\uFEFF' + content], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `repairvision-报修记录-${new Date().toISOString().slice(0, 10)}.csv`
    link.click()
    URL.revokeObjectURL(link.href)
    ElMessage.success(`已导出 ${data.list.length} 条报修记录`)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    exporting.value = false
  }
}

async function loadOptions() {
  try {
    const [buildingList, faultList] = await Promise.all([apiAdminBuildings(), apiAdminFaultTypes()])
    buildings.value = buildingList
    faultTypes.value = faultList
  } catch {
    // 筛选下拉失败不阻塞台账主体
  }
}

onMounted(() => {
  loadOptions()
  load()
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

.room-tag {
  display: inline-block;
  padding: 2px 10px;
  color: #2462d9;
  font-weight: 700;
  background: var(--rv-grad-1);
  border-radius: 999px;
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

.table-empty {
  padding: 30px 0;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}

.log-note {
  margin: 14px 0 0;
  color: #7c8aa3;
  font-size: 12px;
  line-height: 1.6;
}

.log-note code {
  padding: 1px 6px;
  color: #2462d9;
  background: rgba(52, 120, 246, 0.1);
  border-radius: 6px;
}

.muted {
  color: #94a3b8;
}

.timeline-title {
  margin: 22px 0 12px;
  color: #2b3445;
  font-size: 14px;
  font-weight: 800;
}
</style>
