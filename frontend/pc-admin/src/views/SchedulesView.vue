<template>
  <AdminShell title="工人排班" subtitle="双休白班：白班 08:00-12:00 / 14:00-18:00，每周休息 2 天（可配置）">
    <section class="panel">
      <div class="toolbar">
        <div class="week-title">{{ weekLabel }}</div>
        <el-button @click="shiftWeek(-7)">上一周</el-button>
        <el-button type="primary" plain @click="shiftWeek(7)">下一周</el-button>
        <el-button @click="resetThisWeek">回到本周</el-button>
        <span class="param-label">每周休息</span>
        <el-input-number v-model="restDays" :min="0" :max="3" size="small" style="width: 108px" />
        <span class="param-label">每栋最少在岗</span>
        <el-input-number v-model="minPerBuilding" :min="0" :max="5" size="small" style="width: 108px" />
        <div style="flex: 1"></div>
        <el-button @click="openSettings">排班设置</el-button>
        <el-button type="success" :loading="generating" @click="generateWeek">一键生成当周排班</el-button>
      </div>

      <div class="coverage-row">
        <el-tag type="primary" effect="plain">白班 {{ dayShiftCount() }}</el-tag>
        <el-tag type="info" effect="plain">休息 {{ offShiftCount() }}</el-tag>
        <el-tag type="success" effect="dark">当前在岗 {{ onDutyCount }} 人</el-tag>
        <el-tag type="warning" effect="dark">请假 {{ leaveCount() }}</el-tag>
        <span class="coverage-tip">{{ dutyTip }}</span>
      </div>

      <el-alert
        v-if="warnings.length"
        class="warn-box"
        type="warning"
        :closable="false"
        show-icon
        :title="`人力不足提示（${warnings.length} 条）`"
      >
        <ul class="warn-list">
          <li v-for="(w, i) in warnings" :key="i">{{ w }}</li>
        </ul>
      </el-alert>
      <el-table :data="rows" v-loading="loading" border stripe>
        <el-table-column label="维修工人" width="150">
          <template #default="{ row }">
            <div class="worker-name">{{ row.name }}</div>
            <div class="worker-sub">{{ row.username }}</div>
          </template>
        </el-table-column>
        <el-table-column v-for="d in weekDates" :key="d" :label="dateLabel(d)" align="center" min-width="110">
          <template #default="{ row }">
            <button
              class="shift-cell"
              :class="[shiftCellClass(cellShift(row, d)), { 'is-leave': isLeaveCell(row, d) }]"
              @click="openEdit(row, d)"
            >
              {{ cellText(row, d) }}
            </button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && rows.length === 0" description="暂无可排班工人" class="empty" />
      <p class="hint">点击任意日期可调整班次（白班 / 休息）。生成逻辑：先固定已批准请假，再按「每周休息天数」错峰排休（或按排班设置里的固定休息日），保证每栋楼最少在岗人数；非工作时段开工需工人确认。</p>
    </section>

    <el-dialog append-to-body v-model="settingsVisible" title="排班设置" width="520px">
      <el-form :model="settingsForm" label-width="120px">
        <el-form-item label="每周休息天数">
          <el-input-number v-model="settingsForm.restDaysPerWeek" :min="0" :max="3" />
          <div class="form-tip">默认 2 天（双休）</div>
        </el-form-item>
        <el-form-item label="轮休模式">
          <el-radio-group v-model="settingsForm.restMode">
            <el-radio value="staggered">错峰轮休（保证每栋楼在岗）</el-radio>
            <el-radio value="fixed">固定休息日</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="settingsForm.restMode === 'fixed'" label="固定休息日">
          <el-checkbox-group v-model="fixedWeekdayList">
            <el-checkbox :value="1">周一</el-checkbox>
            <el-checkbox :value="2">周二</el-checkbox>
            <el-checkbox :value="3">周三</el-checkbox>
            <el-checkbox :value="4">周四</el-checkbox>
            <el-checkbox :value="5">周五</el-checkbox>
            <el-checkbox :value="6">周六</el-checkbox>
            <el-checkbox :value="7">周日</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="上午工作时段">
          <el-time-picker v-model="morningRange" is-range format="HH:mm" value-format="HH:mm" start-placeholder="上班" end-placeholder="下班" />
        </el-form-item>
        <el-form-item label="下午工作时段">
          <el-time-picker v-model="afternoonRange" is-range format="HH:mm" value-format="HH:mm" start-placeholder="上班" end-placeholder="下班" />
        </el-form-item>
        <el-form-item label="非时段开工">
          <el-switch v-model="settingsForm.allowForceStart" :active-value="1" :inactive-value="0" />
          <div class="form-tip">开启：工人确认后可强制开工；关闭：非工作时段禁止开工</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settingsVisible = false">取消</el-button>
        <el-button type="primary" :loading="settingsSaving" @click="saveSettings">保存设置</el-button>
      </template>
    </el-dialog>
    <el-dialog append-to-body v-model="editVisible" title="调整班次" width="460px">
      <template v-if="editTarget">
        <p class="edit-hint">
          {{ editTarget.name }}（{{ editTarget.username }}）· {{ dateLabel(editDate) }}
        </p>
        <el-select v-model="editShift" style="width: 100%">
          <el-option :label="`白班 ${settings.morningStart}-${settings.morningEnd} / ${settings.afternoonStart}-${settings.afternoonEnd}`" value="DAY" />
          <el-option label="休息" value="OFF" />
          <el-option v-if="isLegacyShift(editShift)" label="白班（历史午/晚班数据）" :value="editShift" />
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
import type { AdminUser, ScheduleItem, WorkSettings } from '../api'
import {
  apiAdminDutyOverview,
  apiAdminSchedules,
  apiAdminUsers,
  apiAdminWorkSettings,
  apiGenerateWeekly,
  apiSaveSchedules,
  apiUpdateWorkSettings,
} from '../api'
import AdminShell from '../components/AdminShell.vue'

const workers = ref<AdminUser[]>([])
const scheduleMap = ref<Record<string, ScheduleItem>>({})
const weekStart = ref(mondayOf(new Date()))
const loading = ref(false)
const generating = ref(false)
const restDays = ref(2)
const minPerBuilding = ref(1)
const warnings = ref<string[]>([])
const saving = ref(false)
const editVisible = ref(false)
const editTarget = ref<AdminUser | null>(null)
const editDate = ref('')
const editShift = ref('DAY')
const editNote = ref('')
const settings = ref<WorkSettings>({
  restDaysPerWeek: 2,
  restMode: 'staggered',
  fixedRestWeekdays: '6,7',
  morningStart: '08:00',
  morningEnd: '12:00',
  afternoonStart: '14:00',
  afternoonEnd: '18:00',
  allowForceStart: 1,
})
const onDutyCount = ref(0)
const dutyTip = ref('当前在岗人数按「今日白班 ∧ 工作时段内 ∧ 未请假 ∧ 启用」统计')
const settingsVisible = ref(false)
const settingsSaving = ref(false)
const settingsForm = reactive({
  restDaysPerWeek: 2,
  restMode: 'staggered',
  allowForceStart: 1,
})
const fixedWeekdayList = ref<number[]>([6, 7])
const morningRange = ref<[string, string]>(['08:00', '12:00'])
const afternoonRange = ref<[string, string]>(['14:00', '18:00'])

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
}function shiftCount(shift: string) {
  return Object.values(scheduleMap.value).filter((x) => x.shiftType === shift).length
}
function cellNote(row: AdminUser, date: string) {
  return scheduleMap.value[`${row.id}|${date}`]?.note || ''
}

// 已批准请假在表格里单独标紫，避免与轮休混淆
function isLeaveCell(row: AdminUser, date: string) {
  return cellShift(row, date) === 'OFF' && (cellNote(row, date) || '').includes('请假')
}

function cellText(row: AdminUser, date: string) {
  return isLeaveCell(row, date) ? '请假' : shiftText(cellShift(row, date))
}

function leaveCount() {
  return Object.values(scheduleMap.value).filter((x) => x.shiftType === 'OFF' && (x.note || '').includes('请假')).length
}


function dayShiftCount() {
  return Object.values(scheduleMap.value).filter(
    (x) => x.shiftType === 'DAY' || x.shiftType === 'MORNING' || x.shiftType === 'AFTERNOON',
  ).length
}

function offShiftCount() {
  return Object.values(scheduleMap.value).filter((x) => x.shiftType === 'OFF' && !(x.note || '').includes('请假')).length
}

function isLegacyShift(shift: string) {
  return shift === 'MORNING' || shift === 'AFTERNOON'
}

async function loadSettings() {
  try {
    settings.value = await apiAdminWorkSettings()
    restDays.value = settings.value.restDaysPerWeek
    settingsForm.restDaysPerWeek = settings.value.restDaysPerWeek
    settingsForm.restMode = settings.value.restMode
    settingsForm.allowForceStart = settings.value.allowForceStart
    fixedWeekdayList.value = settings.value.fixedRestWeekdays
      .split(',')
      .map((x) => Number(x))
      .filter((x) => x >= 1 && x <= 7)
    morningRange.value = [settings.value.morningStart, settings.value.morningEnd]
    afternoonRange.value = [settings.value.afternoonStart, settings.value.afternoonEnd]
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function loadDuty() {
  try {
    const ov = await apiAdminDutyOverview()
    onDutyCount.value = ov.onDutyCount
    dutyTip.value = `工作时段 ${ov.morning} / ${ov.afternoon} · 在岗 ${ov.onDutyCount}/${ov.total} 人`
  } catch {
    // 在岗统计失败不阻塞排班表
  }
}

function openSettings() {
  settingsForm.restDaysPerWeek = settings.value.restDaysPerWeek
  settingsForm.restMode = settings.value.restMode
  settingsForm.allowForceStart = settings.value.allowForceStart
  settingsVisible.value = true
}

async function saveSettings() {
  settingsSaving.value = true
  try {
    settings.value = await apiUpdateWorkSettings({
      restDaysPerWeek: settingsForm.restDaysPerWeek,
      restMode: settingsForm.restMode,
      fixedRestWeekdays: fixedWeekdayList.value.join(','),
      morningStart: morningRange.value[0],
      morningEnd: morningRange.value[1],
      afternoonStart: afternoonRange.value[0],
      afternoonEnd: afternoonRange.value[1],
      allowForceStart: settingsForm.allowForceStart,
    })
    restDays.value = settings.value.restDaysPerWeek
    ElMessage.success('排班设置已保存')
    settingsVisible.value = false
    loadDuty()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    settingsSaving.value = false
  }
}
function shiftText(shift: string) {
  switch (shift) {
    case 'DAY':
      return '白班'
    case 'MORNING':
      return '白班*'
    case 'AFTERNOON':
      return '白班*'
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
    const res = await apiGenerateWeekly(weekStart.value, restDays.value, minPerBuilding.value)
    warnings.value = res.warnings || []
    if (warnings.value.length) {
      ElMessage.warning(`已生成 ${res.list.length} 条班次，存在 ${warnings.value.length} 条人力不足提示`)
    } else {
      ElMessage.success(`已生成 ${res.list.length} 条班次记录`)
    }
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
  loadSettings()
  loadDuty()
  loadSchedules()
})
</script>

<style scoped>
.param-label {
  color: var(--pc-sub);
  font-size: 12px;
}
.warn-box {
  margin: 10px 0 14px;
}
.warn-list {
  margin: 6px 0 0;
  padding-left: 18px;
  font-size: 12px;
  line-height: 1.8;
}
.shift-cell.is-leave {
  color: #6b3fa0;
  background: var(--rv-grad-2);
  border-color: rgba(160, 120, 220, 0.5);
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
  gap: 10px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.week-title {
  font-weight: 800;
  color: #2b3445;
  margin-right: 8px;
}
.coverage-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.coverage-tip {
  color: #8a97ad;
  font-size: 12px;
}
.worker-name {
  font-weight: 700;
  color: #2b3445;
}
.worker-sub {
  color: #8a97ad;
  font-size: 12px;
}
.shift-cell {
  min-width: 68px;
  padding: 5px 10px;
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 10px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}
.shift-cell:hover {
  transform: translateY(-1px);
}
.shift-day {
  color: #2462d9;
  background: var(--rv-grad-1);
}
.shift-morning {
  color: #b96b1c;
  background: var(--rv-grad-4);
}
.shift-afternoon {
  color: #6a4bc0;
  background: var(--rv-grad-2);
}
.shift-off {
  color: #5a6a85;
  background: var(--rv-grad-8);
}
.shift-empty {
  color: #8a97ad;
  background: rgba(255, 255, 255, 0.45);
}
.hint {
  margin-top: 12px;
  color: #8a97ad;
  font-size: 12px;
}
.edit-hint {
  margin-bottom: 12px;
  color: #5a6a85;
  font-size: 14px;
  font-weight: 600;
}
 .form-tip {
  margin-top: 4px;
  color: #8a97ad;
  font-size: 12px;
  line-height: 1.5;
}
.empty {
  padding: 24px 0;
}
</style>
