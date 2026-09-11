<template>
  <div class="leave-page">
    <section class="leave-form">
      <div class="leave-form-title">申请请假</div>
      <div class="leave-row">
        <span class="leave-label">开始</span>
        <input v-model="form.startDate" class="rv-form-field" type="date" />
      </div>
      <div class="leave-row">
        <span class="leave-label">结束</span>
        <input v-model="form.endDate" class="rv-form-field" type="date" />
      </div>
      <div class="leave-row">
        <span class="leave-label">原因</span>
        <input v-model="form.reason" class="rv-form-field" placeholder="如：家中有事 / 身体不适" maxlength="200" />
      </div>
      <button class="rv-submit" :disabled="submitting" @click="submit">
        {{ submitting ? '提交中…' : '提交请假申请' }}
      </button>
    </section>

    <section class="leave-list">
      <div class="leave-list-title">我的请假记录</div>
      <div class="rv-filters">
        <span class="rv-days-label">时间</span>
        <button
          v-for="d in dayOptions"
          :key="d"
          class="rv-filter-chip"
          :class="{ active: days === d }"
          @click="setDays(d)"
        >
          {{ d === 30 ? '30天' : d + '天' }}
        </button>
      </div>
      <div v-if="list.length === 0" class="rv-empty">
        <div class="rv-empty-icon">🗓️</div>
        <div class="rv-empty-text">暂无请假记录</div>
      </div>
      <article v-for="item in list" :key="item.id" class="leave-card">
        <div class="leave-card-top">
          <span class="leave-range">{{ item.startDate }} ~ {{ item.endDate }}</span>
          <span class="leave-status" :class="'ls' + item.status">{{ item.statusText }}</span>
        </div>
        <div class="leave-reason">{{ item.reason }}</div>
        <div v-if="item.reviewNote" class="leave-note">审批意见：{{ item.reviewNote }}</div>
        <button v-if="item.status === 1" class="leave-cancel" @click="cancel(item)">撤销申请</button>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { LeaveItem } from '../api'
import { apiWorkerLeaveCancel, apiWorkerLeaveSubmit, apiWorkerLeaves } from '../api'

const list = ref<LeaveItem[]>([])
const submitting = ref(false)
const form = reactive({ startDate: '', endDate: '', reason: '' })
const dayOptions = [1, 3, 7, 30]
const days = ref(Number(localStorage.getItem('rv-days') || 3))

function today() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

async function load() {
  try {
    const res = await apiWorkerLeaves(0, 1, 50, days.value)
    list.value = res.list
  } catch (err) {
    showToast((err as Error).message)
  }
}

function setDays(d: number) {
  days.value = d
  localStorage.setItem('rv-days', String(d))
  load()
}

async function submit() {
  if (!form.startDate || !form.endDate || !form.reason.trim()) {
    showToast('请填写开始/结束日期与请假原因')
    return
  }
  if (form.endDate < form.startDate) {
    showToast('结束日期不能早于开始日期')
    return
  }
  submitting.value = true
  try {
    await apiWorkerLeaveSubmit({
      startDate: form.startDate,
      endDate: form.endDate,
      reason: form.reason.trim(),
    })
    showToast('请假申请已提交，等待管理员审批')
    form.reason = ''
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    submitting.value = false
  }
}

async function cancel(item: LeaveItem) {
  try {
    await showConfirmDialog({ title: '撤销请假', message: `确认撤销 ${item.startDate} ~ ${item.endDate} 的申请？` })
  } catch {
    return
  }
  try {
    await apiWorkerLeaveCancel(item.id)
    showToast('已撤销')
    load()
  } catch (err) {
    showToast((err as Error).message)
  }
}

onMounted(() => {
  form.startDate = today()
  form.endDate = today()
  load()
})
</script>

<style scoped>
.leave-page {
  padding: 0 16px 24px;
}
.leave-form {
  padding: 16px;
  margin-bottom: 16px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 18px;
  box-shadow: 0 12px 30px rgba(46, 68, 112, 0.1);
}
.leave-form-title,
.leave-list-title {
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 800;
  color: var(--rv-text);
}
.leave-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.leave-label {
  width: 40px;
  color: var(--rv-text-sub);
  font-size: 13px;
}
.leave-list {
  margin-top: 6px;
}
.leave-card {
  padding: 14px;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.65);
  border-radius: 16px;
  box-shadow: 0 10px 26px rgba(46, 68, 112, 0.09);
}
.leave-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.leave-range {
  font-size: 14px;
  font-weight: 700;
  color: var(--rv-text);
}
.leave-status {
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 700;
  border-radius: 999px;
}
.ls1 { color: #b96b1c; background: var(--rv-grad-4); }
.ls2 { color: #17865a; background: var(--rv-grad-6); }
.ls3 { color: #b34568; background: var(--rv-grad-3); }
.ls4 { color: #5a6a85; background: var(--rv-grad-8); }
.leave-reason {
  margin-top: 8px;
  color: var(--rv-text-sub);
  font-size: 13px;
}
.leave-note {
  margin-top: 6px;
  color: var(--rv-text-light);
  font-size: 12px;
}
.leave-cancel {
  margin-top: 10px;
  padding: 6px 12px;
  color: #b34568;
  font-size: 12px;
  font-weight: 700;
  background: var(--rv-grad-3);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 999px;
}
</style>
