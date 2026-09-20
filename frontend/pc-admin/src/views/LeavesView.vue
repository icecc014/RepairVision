<template>
  <AdminShell title="请假审批" subtitle="审批维修工人的请假申请，通过后自动生成休息班次">
    <section class="panel">
      <div class="toolbar">
        <el-select v-model="filter.status" placeholder="全部状态" clearable style="width: 160px" @change="load">
          <el-option label="待审批" :value="1" />
          <el-option label="已通过" :value="2" />
          <el-option label="已驳回" :value="3" />
          <el-option label="已撤销" :value="4" />
        </el-select>
        <el-radio-group v-model="days" size="small" @change="load">
          <el-radio-button :value="1">1天</el-radio-button>
          <el-radio-button :value="3">3天</el-radio-button>
          <el-radio-button :value="7">7天</el-radio-button>
          <el-radio-button :value="30">30天</el-radio-button>
        </el-radio-group>
        <el-button type="primary" @click="load">查询</el-button>
        <el-button type="primary" plain @click="openCreate">＋ 代工人登记请假</el-button>
        <div style="flex: 1"></div>
        <el-tag type="warning" effect="plain">待审批 {{ pendingCount }}</el-tag>
      </div>

      <el-table v-if="!isMobile" :data="list" v-loading="loading" border stripe>
        <el-table-column prop="workerName" label="工人" width="120">
          <template #default="{ row }">{{ row.workerName || ('#' + row.workerId) }}</template>
        </el-table-column>
        <el-table-column label="请假区间" min-width="200">
          <template #default="{ row }">{{ row.startDate }} ~ {{ row.endDate }}</template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.statusText }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reviewNote" label="审批意见" min-width="160">
          <template #default="{ row }">{{ row.reviewNote || '—' }}</template>
        </el-table-column>
        <el-table-column prop="createdAt" label="提交时间" width="170" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 1" size="small" type="success" plain @click="openReview(row, 2)">通过</el-button>
            <el-button v-if="row.status === 1" size="small" type="danger" plain @click="openReview(row, 3)">驳回</el-button>
            <el-button v-if="row.status === 2" size="small" plain @click="cancelLeave(row)">撤销</el-button>
            <span v-if="row.status !== 1 && row.status !== 2" class="muted">—</span>
          </template>
        </el-table-column>
      </el-table>

      <div v-else class="m-cards">
        <article v-for="row in list" :key="row.id || (row.workerId + row.startDate)" class="m-card">
          <header class="m-head">
            <h4 class="m-title">{{ row.workerName || ('#' + row.workerId) }}</h4>
            <el-tag :type="statusType(row.status)" size="small">{{ row.statusText }}</el-tag>
          </header>
          <dl class="m-meta">
            <div><dt>请假区间</dt><dd>{{ row.startDate }} ~ {{ row.endDate }}</dd></div>
            <div><dt>原因</dt><dd>{{ row.reason || '—' }}</dd></div>
            <div><dt>审批意见</dt><dd>{{ row.reviewNote || '—' }}</dd></div>
            <div><dt>提交时间</dt><dd>{{ row.createdAt }}</dd></div>
          </dl>
          <footer class="m-actions">
            <el-button v-if="row.status === 1" size="small" type="success" plain @click="openReview(row, 2)">通过</el-button>
            <el-button v-if="row.status === 1" size="small" type="danger" plain @click="openReview(row, 3)">驳回</el-button>
            <el-button v-if="row.status === 2" size="small" plain @click="cancelLeave(row)">撤销</el-button>
            <span v-if="row.status !== 1 && row.status !== 2" class="muted">—</span>
          </footer>
        </article>
      </div>
      <el-empty v-if="!loading && list.length === 0" description="暂无请假申请" class="empty" />
    </section>

    <el-dialog append-to-body v-model="reviewVisible" :title="reviewStatus === 2 ? '通过请假' : '驳回请假'" width="460px">
      <template v-if="reviewTarget">
        <p class="review-hint">
          {{ reviewTarget.workerName || ('#' + reviewTarget.workerId) }} · {{ reviewTarget.startDate }} ~ {{ reviewTarget.endDate }}
        </p>
        <el-input v-model="reviewNote" type="textarea" :rows="3" placeholder="审批意见（可选）" maxlength="200" />
      </template>
      <template #footer>
        <el-button @click="reviewVisible = false">取消</el-button>
        <el-button :type="reviewStatus === 2 ? 'success' : 'danger'" :loading="saving" @click="submitReview">确认</el-button>
      </template>
    </el-dialog>
    <el-dialog append-to-body v-model="createVisible" title="代工人登记请假" width="520px">
      <el-form :model="createForm" label-width="90px">
        <el-form-item label="维修工人">
          <el-select v-model="createForm.workerId" placeholder="选择工人" style="width: 100%">
            <el-option v-for="w in workers" :key="w.id" :label="`${w.name}（${w.username}）`" :value="w.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="请假区间">
          <el-date-picker
            v-model="createForm.range"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="请假原因">
          <el-input v-model="createForm.reason" type="textarea" :rows="3" maxlength="200" placeholder="如：家中有事" />
        </el-form-item>
      </el-form>
      <p class="create-tip">登记后直接生效为「已通过」，对应日期会自动写成休息班次，不再参与派单。</p>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitCreate">确认登记</el-button>
      </template>
    </el-dialog>
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminUser, LeaveItem } from '../api'
import { apiAdminLeaveCancel, apiAdminLeaveCreate, apiAdminLeaves, apiAdminLeaveReview, apiAdminUsers } from '../api'
import AdminShell from '../components/AdminShell.vue'
import { useIsMobile } from '../composables/useViewport'

// V9.7.2：≤767px 用卡片列表替代表格
const { isMobile } = useIsMobile()

const list = ref<LeaveItem[]>([])
const loading = ref(false)
const saving = ref(false)
const reviewVisible = ref(false)
const reviewTarget = ref<LeaveItem | null>(null)
const reviewStatus = ref(2)
const reviewNote = ref('')
const days = ref(3)
const filter = reactive({ status: 0 })
const createVisible = ref(false)
const workers = ref<AdminUser[]>([])
const createForm = reactive({
  workerId: undefined as number | undefined,
  range: [] as string[],
  reason: '',
})

const pendingCount = computed(() => list.value.filter((x) => x.status === 1).length)

async function loadWorkers() {
  try {
    workers.value = await apiAdminUsers({ role: 2 })
  } catch {
    // 下拉失败不阻塞列表
  }
}

function openCreate() {
  createForm.workerId = undefined
  createForm.range = []
  createForm.reason = ''
  createVisible.value = true
  if (workers.value.length === 0) loadWorkers()
}

async function submitCreate() {
  if (!createForm.workerId) {
    ElMessage.warning('请选择要登记请假的工人')
    return
  }
  if (!createForm.range || createForm.range.length !== 2) {
    ElMessage.warning('请选择请假起止日期')
    return
  }
  if (!createForm.reason.trim()) {
    ElMessage.warning('请填写请假原因')
    return
  }
  saving.value = true
  try {
    await apiAdminLeaveCreate({
      workerId: createForm.workerId,
      startDate: createForm.range[0],
      endDate: createForm.range[1],
      reason: createForm.reason.trim(),
    })
    ElMessage.success('已登记并直接通过')
    createVisible.value = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

async function cancelLeave(row: LeaveItem) {
  try {
    await ElMessageBox.confirm(`确认撤销 ${row.workerName || row.workerId} 的请假（${row.startDate} ~ ${row.endDate}）？`,
      '撤销请假')
  } catch {
    return
  }
  try {
    await apiAdminLeaveCancel(row.id)
    ElMessage.success('已撤销')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function load() {
  loading.value = true
  try {
    const res = await apiAdminLeaves({ status: filter.status, page: 1, size: 100, days: days.value })
    list.value = res.list
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function statusType(status: number) {
  if (status === 1) return 'warning'
  if (status === 2) return 'success'
  if (status === 3) return 'danger'
  return 'info'
}

function openReview(row: LeaveItem, status: number) {
  reviewTarget.value = row
  reviewStatus.value = status
  reviewNote.value = ''
  reviewVisible.value = true
}

async function submitReview() {
  if (!reviewTarget.value) return
  saving.value = true
  try {
    await apiAdminLeaveReview(reviewTarget.value.id, reviewStatus.value, reviewNote.value.trim())
    ElMessage.success(reviewStatus.value === 2 ? '已通过请假，对应日期已生成休息班次' : '已驳回')
    reviewVisible.value = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  load()
  loadWorkers()
})
</script>

<style scoped>
.create-tip {
  margin: 0;
  color: #94a3b8;
  font-size: 12px;
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
  gap: 12px;
  margin-bottom: 14px;
}
.review-hint {
  margin-bottom: 12px;
  color: #5a6a85;
  font-weight: 600;
}
.muted {
  color: #c2cbdc;
}
.empty {
  padding: 24px 0;
}
/* ---------- V9.7.2 移动端卡片（仅 ≤767px 渲染） ---------- */

.m-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.m-card {
  padding: 14px;
  background: rgba(255, 255, 255, 0.86);
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 16px;
  box-shadow: 0 10px 24px rgba(46, 68, 112, 0.08);
}

.m-head {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.m-title {
  flex: 1;
  margin: 0;
  color: var(--pc-text);
  font-size: 15px;
  font-weight: 700;
  line-height: 1.4;
}

.m-meta {
  display: grid;
  gap: 6px;
  margin: 10px 0 0;
}

.m-meta div {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.m-meta dt {
  flex: 0 0 64px;
  color: var(--pc-light);
  font-size: 12px;
}

.m-meta dd {
  margin: 0;
  color: var(--pc-sub);
  font-size: 13px;
  word-break: break-word;
}

.m-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(120, 145, 190, 0.16);
}

.m-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}
</style>
