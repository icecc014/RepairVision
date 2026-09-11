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
        <div style="flex: 1"></div>
        <el-tag type="warning" effect="plain">待审批 {{ pendingCount }}</el-tag>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
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
            <span v-if="row.status !== 1" class="muted">—</span>
          </template>
        </el-table-column>
      </el-table>
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
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { LeaveItem } from '../api'
import { apiAdminLeaves, apiAdminLeaveReview } from '../api'
import AdminShell from '../components/AdminShell.vue'

const list = ref<LeaveItem[]>([])
const loading = ref(false)
const saving = ref(false)
const reviewVisible = ref(false)
const reviewTarget = ref<LeaveItem | null>(null)
const reviewStatus = ref(2)
const reviewNote = ref('')
const days = ref(3)
const filter = reactive({ status: 0 })

const pendingCount = computed(() => list.value.filter((x) => x.status === 1).length)

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

onMounted(load)
</script>

<style scoped>
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
</style>
