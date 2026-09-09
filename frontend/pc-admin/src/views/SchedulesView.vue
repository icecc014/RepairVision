<template>
  <AdminShell title="工人排班" subtitle="按周维护工人班次与轮休，休息日不参与自动派单">
    <section class="panel">
      <div class="toolbar">
        <div class="week-title">{{ weekLabel }}</div>
        <el-button @click="shiftWeek(-7)">上一周</el-button>
        <el-button type="primary" plain @click="shiftWeek(7)">下一周</el-button>
        <el-button @click="resetThisWeek">回到本周</el-button>
        <div style="flex: 1"></div>
        <el-button type="success" :loading="generating" @click="generateWeek">一键生成当周排班</el-button>
      </div>

      <el-table :data="rows" v-loading="loading" border stripe>
        <el-table-column label="维修工人" width="150">
          <template #default="{ row }">
            <div class="worker-name">{{ row.name }}</div>
            <div class="worker-sub">{{ row.username }}</div>
          </template>
        </el-table-column>
        <el-table-column v-for="d in weekDates" :key="d" :label="dateLabel(d)" align="center" min-width="110">
          <template #default="{ row }">
            <button class="shift-cell" :class="shiftCellClass(cellShift(row, d))" @click="openEdit(row, d)">
              {{ shiftText(cellShift(row, d)) }}
            </button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && rows.length === 0" description="暂无可排班工人" class="empty" />
      <p class="hint">点击任意日期可调整班次；生成逻辑：每名工人每周休 1 天，轮休日按周错开。</p>
    </section>

    <el-dialog v-model="editVisible" title="调整班次" width="460px">
      <template v-if="editTarget">
        <p class="edit-hint">
          {{ editTarget.name }}（{{ editTarget.username }}）· {{ dateLabel(editDate) }}
        </p>
        <el-select v-model="editShift" style="width: 100%">
          <el-option label="全天班 DAY" value="DAY" />
          <el-option label="午班 MORNING（8:00-14:00）" value="MORNING" />
          <el-option label="晚班 AFTERNOON（14:00-20:00）" value="AFTERNOON" />
          <el-option label="休息 OFF" value="OFF" />
        </el-select>
        <el-input v-model="editNote" placeholder="备注（可选）" style="margin-top: 12px" />
      </template>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminUser, ScheduleItem } from '../api'
import { apiAdminSchedules, apiAdminUsers, apiGenerateWeekly, apiSaveSchedules } from '../api'
import AdminShell from '../components/AdminShell.vue'

const workers = ref<AdminUser[]>([])
const scheduleMap = ref<Record<string, ScheduleItem>>({})
const weekStart = ref(mondayOf(new Date()))
const loading = ref(false)
const generating = ref(false)
const saving = ref(false)
const editVisible = ref(false)
const editTarget = ref<AdminUser | null>(null)
const editDate = ref('')
const editShift = ref('DAY')
const editNote = ref('')

const weekDates = computed(() => {
  const start = parseDate(weekStart.value)
  return Array.from({ length: 7 }, (_, i) => formatDate(addDays(start, i)))
})

const weekLabel = computed(() => `${dateLabel(weekDates.value[0])} ~ ${dateLabel(weekDates.value[6])}`)

const rows = computed(() => {
  const list = [...workers.value].sort((a, b) => a.id - b.id)
  return list.map((w) => ({
    ...w,
    name: w.name,
    username: w.username,
  }))
})

function cellShift(row: AdminUser, date: string) {
  return scheduleMap.value[`${row.id}|${date}`]?.shiftType || '—'
}

function shiftCellClass(shift: string) {
  const key = String(shift || '').toLowerCase()
  if (key === 'day' || key === 'morning' || key === 'afternoon' || key === 'off') {
    return 'shift-' + key
  }
  return 'shift-empty'
}function shiftText(shift: string) {
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
      return '未排'
  }
}

async function loadWorkers() {
  try {
    workers.value = await apiAdminUsers({ role: 2 })
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function loadSchedules() {
  loading.value = true
  try {
    const items = await apiAdminSchedules(0, weekDates.value[0], weekDates.value[6])
    const next: Record<string, ScheduleItem> = {}
    for (const item of items) {
      next[`${item.workerId}|${item.workDate}`] = item
    }
    scheduleMap.value = next
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function openEdit(row: AdminUser, date: string) {
  editTarget.value = row
  editDate.value = date
  const item = scheduleMap.value[`${row.id}|${date}`]
  editShift.value = item?.shiftType === 'OFF' || item?.shiftType === 'MORNING' || item?.shiftType === 'AFTERNOON' ? item.shiftType : 'DAY'
  editNote.value = item?.note || ''
  editVisible.value = true
}

async function saveEdit() {
  if (!editTarget.value) return
  saving.value = true
  try {
    await apiSaveSchedules([
      {
        workerId: editTarget.value.id,
        workDate: editDate.value,
        shiftType: editShift.value,
        note: editNote.value.trim(),
      },
    ])
    ElMessage.success('班次已保存')
    editVisible.value = false
    await loadSchedules()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

async function generateWeek() {
  try {
    await ElMessageBox.confirm(`将重新生成 ${weekLabel.value} 的全部排班并覆盖当天已有设置，是否继续？`, '一键生成')
  } catch {
    return
  }
  generating.value = true
  try {
    const res = await apiGenerateWeekly(weekStart.value)
    ElMessage.success(`已生成 ${res.list.length} 条班次记录`)
    await loadSchedules()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    generating.value = false
  }
}

function shiftWeek(days: number) {
  weekStart.value = formatDate(addDays(parseDate(weekStart.value), days))
  loadSchedules()
}

function resetThisWeek() {
  weekStart.value = mondayOf(new Date())
  loadSchedules()
}

function mondayOf(now: Date) {
  const day = (now.getDay() + 6) % 7
  return formatDate(addDays(new Date(now.getFullYear(), now.getMonth(), now.getDate()), -day))
}

function parseDate(date: string) {
  const [y, m, d] = date.split('-').map(Number)
  return new Date(y, m - 1, d)
}

function addDays(d: Date, n: number) {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate() + n)
}

function formatDate(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

function dateLabel(date: string) {
  const d = parseDate(date)
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${d.getMonth() + 1}/${d.getDate()} ${weekdays[d.getDay()]}`
}

onMounted(() => {
  loadWorkers()
  loadSchedules()
})
</script>

<style scoped>
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
  gap: 10px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.week-title {
  font-weight: 800;
  color: #1e293b;
  margin-right: 8px;
}
.worker-name {
  font-weight: 700;
  color: #1e293b;
}
.worker-sub {
  color: #94a3b8;
  font-size: 12px;
}
.shift-cell {
  min-width: 68px;
  padding: 5px 10px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.shift-day {
  color: #1d4ed8;
  background: #dbeafe;
}
.shift-morning,
.shift-afternoon {
  color: #7c3aed;
  background: #ede9fe;
}
.shift-off {
  color: #64748b;
  background: #e2e8f0;
}
.shift-empty {
  color: #94a3b8;
  background: #f1f5f9;
}
.hint {
  margin-top: 12px;
  color: #94a3b8;
  font-size: 12px;
}
.edit-hint {
  margin-bottom: 12px;
  color: #475569;
  font-size: 14px;
  font-weight: 600;
}
.empty {
  padding: 24px 0;
}
</style>
