<template>
  <div class="schedule-page">
    <div class="schedule-head">
      <div>
        <div class="schedule-title">本周班次</div>
        <div class="schedule-range">{{ monday }} ~ {{ sunday }}</div>
      </div>
      <button class="rv-filter-chip" style="margin-left: auto" @click="load">↻ 刷新</button>
    </div>

    <div v-if="days.length === 0" class="rv-empty">
      <div class="rv-empty-icon">🗓️</div>
      <div class="rv-empty-text">本周暂无排班，请联系管理员</div>
    </div>

    <div v-else class="schedule-list">
      <article
        v-for="day in days"
        :key="day.date"
        class="schedule-card"
        :class="{ 'is-off': day.shiftType === 'OFF' }"
      >
        <div class="schedule-date">
          <span class="schedule-weekday">{{ day.weekday }}</span>
          <span class="schedule-daynum">{{ day.label }}</span>
        </div>
        <div class="schedule-right">
          <span class="shift-badge" :class="shiftClass(day.shiftType)">
            {{ shiftText(day.shiftType) }}
          </span>
          <span v-if="day.note" class="schedule-note">{{ day.note }}</span>
          <span v-else class="schedule-note">—</span>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showToast } from 'vant'
import type { ScheduleItem } from '../api'
import { apiWorkerSchedules } from '../api'

const map = ref<Record<string, ScheduleItem>>({})
const loading = ref(false)

const monday = computed(() => {
  const now = new Date()
  const day = (now.getDay() + 6) % 7
  const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() - day)
  return formatDate(d)
})
const sunday = computed(() => formatDate(addDays(parseDate(monday.value), 6)))

const days = computed(() => {
  const result: {
    date: string
    weekday: string
    label: string
    shiftType: string
    note?: string
  }[] = []
  const start = parseDate(monday.value)
  const weekdays = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  for (let i = 0; i < 7; i++) {
    const d = addDays(start, i)
    const date = formatDate(d)
    const item = map.value[date]
    result.push({
      date,
      weekday: weekdays[i],
      label: `${d.getMonth() + 1}/${d.getDate()}`,
      shiftType: item?.shiftType || '—',
      note: item?.note,
    })
  }
  return result
})

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const items = await apiWorkerSchedules(monday.value, sunday.value)
    const next: Record<string, ScheduleItem> = {}
    for (const item of items) {
      next[item.workDate] = item
    }
    map.value = next
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loading.value = false
  }
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

function shiftText(shift: string) {
  switch (shift) {
    case 'DAY':
      return '全天班'
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

function shiftClass(shift: string) {
  return 'shift-' + (shift === 'OFF' ? 'off' : shift === 'DAY' ? 'day' : 'part')
}

onMounted(load)
</script>

<style scoped>
.schedule-page {
  padding: 0 16px 24px;
}

.schedule-head {
  display: flex;
  align-items: center;
  margin-bottom: 14px;
}

.schedule-title {
  font-size: 17px;
  font-weight: 800;
}

.schedule-range {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}

.schedule-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.schedule-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.04);
}

.schedule-card.is-off {
  background: #f8fafc;
}

.schedule-date {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.schedule-weekday {
  font-size: 15px;
  font-weight: 800;
}

.schedule-daynum {
  color: #94a3b8;
  font-size: 12px;
}

.schedule-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.shift-badge {
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 700;
  border-radius: 999px;
}

.shift-day {
  color: #1d4ed8;
  background: #dbeafe;
}

.shift-part {
  color: #7c3aed;
  background: #ede9fe;
}

.shift-off {
  color: #64748b;
  background: #e2e8f0;
}

.schedule-note {
  color: #94a3b8;
  font-size: 12px;
}
</style>
