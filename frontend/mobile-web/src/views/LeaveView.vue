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

function today() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

async function load() {
  try {
    const res = await apiWorkerLeaves(0, 1, 50)
    list.value = res.list
  } catch (err) {
    showToast((err as Error).message)
  }
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
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 16px;
}

.leave-form-title,
.leave-list-title {
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 800;
}

.leave-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.leave-label {
  width: 40px;
  color: #64748b;
  font-size: 13px;
}

.leave-list {
  margin-top: 6px;
}

.leave-card {
  padding: 14px;
  margin-bottom: 10px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
}

.leave-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.leave-range {
  font-size: 14px;
  font-weight: 700;
}

.leave-status {
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 700;
  border-radius: 999px;
}

.ls1 {
  color: #b45309;
  background: #fef3c7;
}

.ls2 {
  color: #15803d;
  background: #dcfce7;
}

.ls3 {
  color: #b91c1c;
  background: #fee2e2;
}

.ls4 {
  color: #64748b;
  background: #f1f5f9;
}

.leave-reason {
  margin-top: 8px;
  color: #475569;
  font-size: 13px;
}

.leave-note {
  margin-top: 6px;
  color: #94a3b8;
  font-size: 12px;
}

.leave-cancel {
  margin-top: 10px;
  padding: 6px 12px;
  color: #dc2626;
  font-size: 12px;
  font-weight: 700;
  background: #fef2f2;
  border: none;
  border-radius: 999px;
}
</style>
