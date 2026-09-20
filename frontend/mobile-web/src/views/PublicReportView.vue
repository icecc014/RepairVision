<template>
  <div class="public-page">
    <header class="public-header">
      <h1 class="public-title">校园维修报修</h1>
      <p class="public-sub">无需登录 · 支持微信直接打开 · 一键报修与进度透明追踪</p>
    </header>

    <nav class="public-tabs">
      <button :class="{ active: tab === 'report' }" @click="tab = 'report'">我要报修</button>
      <button :class="{ active: tab === 'query' }" @click="tab = 'query'">查询进度</button>
    </nav>

    <section v-if="tab === 'report'" class="public-card rv-card">
      <div v-if="errorMsg" class="error-banner">
        <span>{{ errorMsg }}</span>
        <button class="link" @click="errorMsg = ''">知道了</button>
      </div>
      <div v-if="history.length" class="history">
        <div class="history-head">
          <span>最近报修过的房间（点击快速填充）</span>
          <button class="link" @click="clearHistory">清除历史</button>
        </div>
        <div class="history-list">
          <button
            v-for="h in history"
            :key="h.buildingId + '-' + h.room"
            class="history-item"
            @click="applyHistory(h)"
          >
            <b>{{ h.buildingName }}</b>
            <span>{{ h.room }} 室</span>
          </button>
        </div>
      </div>

      <label class="field">
        <span class="field-label">报修楼栋 <i>*</i></span>
        <select v-model.number="form.buildingId" class="field-input">
          <option :value="0" disabled>请选择楼栋</option>
          <option v-for="b in buildings" :key="b.id" :value="b.id">{{ b.name }}</option>
        </select>
      </label>

      <label class="field">
        <span class="field-label">房间号 <i>*</i></span>
        <input
          v-model="form.room"
          class="field-input"
          inputmode="numeric"
          maxlength="4"
          placeholder="如 401（楼层 + 房间序号）"
        />
        <span v-if="roomHint" class="hint">{{ roomHint }}</span>
      </label>

      <div class="field">
        <span class="field-label">故障类型 <i>*</i></span>
        <div class="chip-row">
          <button
            v-for="f in faultOptions"
            :key="f.code"
            class="chip"
            :class="{ active: form.faultType === f.code }"
            @click="form.faultType = f.code"
          >{{ f.label }}</button>
        </div>
      </div>

      <label class="field">
        <span class="field-label">故障描述 <i>*</i></span>
        <textarea
          v-model="form.description"
          class="field-input area"
          maxlength="200"
          placeholder="如：电灯不亮，开关有火花"
        />
        <span class="counter">{{ form.description.length }}/200</span>
      </label>

      <div class="field">
        <span class="field-label">报修人身份 <i>*</i></span>
        <div class="chip-row">
          <button
            v-for="r in reporterOptions"
            :key="r.value"
            class="chip"
            :class="{ active: form.reporterType === r.value }"
            @click="form.reporterType = r.value"
          >{{ r.label }}</button>
        </div>
      </div>

      <label class="field">
        <span class="field-label">联系方式（选填）</span>
        <input v-model="form.contact" class="field-input" maxlength="32" placeholder="手机号或微信，便于联系" />
      </label>

      <div class="field">
        <span class="field-label">验证码 <i>*</i></span>
        <div class="captcha-row">
          <input v-model="form.captchaCode" class="field-input captcha-input" maxlength="4" placeholder="输入右侧字符" />
          <div class="captcha-box" v-html="captchaSvg" @click="loadCaptcha" />
        </div>
        <span class="hint">看不清？点击图片刷新</span>
      </div>

      <button class="submit" :disabled="submitting" @click="submit">
        {{ submitting ? '提交中…' : '提交报修' }}
      </button>
      <p class="tips">提交后可切到「查询进度」，用楼栋 + 房间号查看处理状态与维修工人</p>
    </section>

    <section v-else class="public-card rv-card">
      <label class="field">
        <span class="field-label">报修楼栋 <i>*</i></span>
        <select v-model.number="query.buildingId" class="field-input">
          <option :value="0" disabled>请选择楼栋</option>
          <option v-for="b in buildings" :key="b.id" :value="b.id">{{ b.name }}</option>
        </select>
      </label>
      <label class="field">
        <span class="field-label">房间号 <i>*</i></span>
        <input v-model="query.room" class="field-input" inputmode="numeric" maxlength="4" placeholder="如 401" />
      </label>
      <button class="submit" :disabled="querying" @click="runQuery">
        {{ querying ? '查询中…' : '查询进度' }}
      </button>

      <div v-if="queryDone" class="result">
        <p class="result-head">
          {{ queryResult.buildingName }} {{ queryResult.room }} 室 · 共 {{ queryResult.orders.length }} 条工单
        </p>
        <div v-if="queryResult.orders.length === 0" class="empty">
          该房间暂无报修记录
        </div>
        <div v-for="o in queryResult.orders" :key="o.id" class="order-card">
          <div class="order-head">
            <span class="order-type">{{ o.faultTypeName }}</span>
            <span class="order-status" :class="statusClass(o.status)">{{ o.statusName }}</span>
          </div>
          <div class="order-desc">{{ o.description }}</div>
          <p class="order-meta">
            报修时间：{{ o.createdAt }}<br />
            <span v-if="o.workerName">维修工人：{{ o.workerName }} ({{ o.workerPhone || '暂无电话' }})<br /></span>
            <span v-if="o.completedAt">完工时间：{{ o.completedAt }}<br /></span>
            单号：{{ o.orderNo }}
          </p>
        </div>
      </div>
    </section>

    <footer class="public-footer">
      RepairVision 校园智慧报修系统 · 公共服务通道
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { showToast } from 'vant'
import {
  apiPublicBuildings,
  apiPublicCaptcha,
  apiPublicReport,
  apiPublicRoomOrders,
  type PublicBuildingItem,
  type PublicRoomOrder,
} from '../api'

const tab = ref<'report' | 'query'>('report')

const buildings = ref<PublicBuildingItem[]>([])
const captchaId = ref('')
const captchaSvg = ref('')
const submitting = ref(false)
const querying = ref(false)
const queryDone = ref(false)
const errorMsg = ref('')

const form = reactive({
  buildingId: 0,
  room: '',
  faultType: 'electric',
  description: '',
  reporterType: 'student',
  contact: '',
  captchaCode: '',
})

const query = reactive({
  buildingId: 0,
  room: '',
})

const queryResult = reactive<{
  buildingName: string
  room: string
  orders: PublicRoomOrder[]
}>({
  buildingName: '',
  room: '',
  orders: [],
})

const faultOptions = [
  { code: 'electric', label: '电路/灯具' },
  { code: 'water', label: '水管/卫浴' },
  { code: 'hvac', label: '暖通/空调' },
  { code: 'door_window', label: '门窗/家具' },
  { code: 'network', label: '网络/弱电' },
  { code: 'other', label: '其他' },
]

const reporterOptions = [
  { value: 'student', label: '我是学生' },
  { value: 'staff', label: '教职工/宿管' },
  { value: 'other', label: '其他访客' },
]

interface HistoryItem {
  buildingId: number
  buildingName: string
  room: string
}
const HISTORY_KEY = 'rv_public_report_history'
const history = ref<HistoryItem[]>([])

function readHistory(): HistoryItem[] {
  try {
    const raw = localStorage.getItem(HISTORY_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function saveHistory(buildingId: number, buildingName: string, room: string) {
  const list = history.value.filter((h) => !(h.buildingId === buildingId && h.room === room))
  list.unshift({ buildingId, buildingName, room })
  history.value = list.slice(0, 5)
  try {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history.value))
  } catch {}
}

function clearHistory() {
  history.value = []
  try {
    localStorage.removeItem(HISTORY_KEY)
  } catch {}
}

function applyHistory(h: HistoryItem) {
  form.buildingId = h.buildingId
  form.room = h.room
  query.buildingId = h.buildingId
  query.room = h.room
}

const currentBuilding = computed(() => buildings.value.find((b) => b.id === form.buildingId))

const roomHint = computed(() => {
  if (!currentBuilding.value) return ''
  const b = currentBuilding.value
  return `该楼共 ${b.floors} 层，每层 ${b.roomsPerFloor} 间；如 3 层第 5 间可填 305`
})

function statusClass(s: number) {
  switch (s) {
    case 1:
      return 'pending'
    case 2:
      return 'dispatched'
    case 3:
      return 'working'
    case 4:
      return 'done'
    case 5:
      return 'canceled'
    default:
      return ''
  }
}

async function loadBuildings() {
  try {
    const res = await apiPublicBuildings()
    buildings.value = res.list || []
    if (buildings.value.length > 0 && form.buildingId === 0) {
      form.buildingId = buildings.value[0].id
      query.buildingId = buildings.value[0].id
    }
  } catch (err) {
    showToast('加载楼栋失败: ' + (err as Error).message)
  }
}

async function loadCaptcha() {
  try {
    const res = await apiPublicCaptcha()
    captchaId.value = res.captchaId
    captchaSvg.value = res.svg
    form.captchaCode = ''
  } catch (err) {
    showToast('获取验证码失败: ' + (err as Error).message)
  }
}

function fail(msg: string) {
  errorMsg.value = msg
  showToast(msg)
}

async function submit() {
  errorMsg.value = ''
  if (!form.buildingId) { fail('请选择报修楼栋'); return }
  const room = form.room.trim()
  if (!/^\d{3,4}$/.test(room)) {
    fail('请填写正确房间号，如 401（3 或 4 位数字）')
    return
  }
  if (!form.description.trim()) { fail('请简要描述故障现象'); return }
  if (!form.reporterType) { fail('请选择报修人身份'); return }
  if (!form.captchaCode.trim()) { fail('请输入验证码'); return }

  submitting.value = true
  try {
    const resp = await apiPublicReport({
      buildingId: form.buildingId,
      room,
      faultType: form.faultType,
      description: form.description.trim(),
      reporterType: form.reporterType,
      contact: form.contact.trim(),
      captchaId: captchaId.value,
      captchaCode: form.captchaCode.trim().toUpperCase(),
    })
    const building = buildings.value.find((b) => b.id === form.buildingId)
    saveHistory(form.buildingId, building?.name || '', room)
    if (resp.merged) {
      showToast('该房间同类报修已在处理中，已并入工单 ' + (resp.mainOrderNo || ''))
    } else if (resp.dispatched) {
      showToast('报修已提交，已派给' + (resp.workerName || '维修工人'))
    } else {
      showToast(resp.message || '报修已提交，等待处理')
    }
    form.description = ''
    await loadCaptcha()
  } catch (err) {
    const msg = (err as Error).message || '提交失败'
    console.warn('[public-report] 提交失败:', msg, { buildingId: form.buildingId, room, faultType: form.faultType })
    fail(msg)
    if (msg.includes('验证码')) await loadCaptcha()
  } finally {
    submitting.value = false
  }
}

async function runQuery() {
  const room = query.room.trim()
  if (!query.buildingId) return showToast('请选择报修楼栋')
  if (!/^\d{3,4}$/.test(room)) return showToast('请填写正确房间号，如 401')
  querying.value = true
  try {
    const resp = await apiPublicRoomOrders(query.buildingId, room)
    queryResult.buildingName = resp.buildingName
    queryResult.room = resp.room
    queryResult.orders = resp.orders || []
    queryDone.value = true
  } catch (err) {
    showToast((err as Error).message || '查询失败')
  } finally {
    querying.value = false
  }
}

onMounted(() => {
  history.value = readHistory()
  void loadBuildings()
  void loadCaptcha()
})
</script>

<style scoped>
.public-page {
  min-height: 100vh;
  padding: 0 16px 32px;
  background: var(--rv-page-bg, linear-gradient(160deg, #e0edfa 0%, #e9e3f8 52%, #f8e6ec 100%));
  color: var(--rv-text, #2b3445);
}
.public-header {
  padding: 28px 4px 16px;
  text-align: center;
}
.public-title {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 0.5px;
  background: linear-gradient(105deg, #3478f6 0%, #22b573 55%, #a06ae8 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.public-sub {
  margin: 0;
  color: var(--rv-text-sub, #5a6a85);
  font-size: 13px;
  line-height: 1.5;
}
.public-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
  background: rgba(255, 255, 255, 0.45);
  padding: 4px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: 0 4px 14px rgba(46, 68, 112, 0.05);
}
.public-tabs button {
  flex: 1;
  height: 38px;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--rv-text-sub, #5a6a85);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.22, 1, 0.36, 1);
}
.public-tabs button.active {
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  color: #fff;
  box-shadow: 0 6px 18px rgba(52, 120, 246, 0.3);
}
.error-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 14px;
  padding: 10px 14px;
  border: 1px solid rgba(224, 100, 138, 0.3);
  border-radius: 12px;
  background: rgba(254, 238, 242, 0.85);
  color: #c0392b;
  font-size: 13px;
  line-height: 1.6;
  backdrop-filter: blur(8px);
}
.public-card {
  padding: 18px 16px 20px;
  border: 1px solid rgba(255, 255, 255, 0.75);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.68) !important;
  box-shadow: 0 14px 34px rgba(46, 68, 112, 0.1);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
}
.history {
  margin-bottom: 16px;
  padding: 12px 14px;
  border-radius: 14px;
  background: rgba(240, 246, 255, 0.7);
  border: 1px solid rgba(210, 226, 250, 0.6);
}
.history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  color: var(--rv-text-sub, #5a6a85);
  font-size: 12px;
  font-weight: 600;
}
.history-list {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
}
.history-item {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  min-width: 110px;
  padding: 8px 12px;
  border: 1px solid rgba(120, 145, 190, 0.22);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.85);
  color: #2b3445;
  cursor: pointer;
  transition: all 0.2s ease;
}
.history-item:hover {
  border-color: rgba(52, 120, 246, 0.4);
  transform: translateY(-1px);
}
.history-item b {
  font-size: 12px;
  color: var(--rv-text, #2b3445);
}
.history-item span {
  color: var(--rv-primary-deep, #2462d9);
  font-size: 13px;
  font-weight: 800;
}
.field {
  position: relative;
  display: block;
  margin-bottom: 16px;
}
.field-label {
  display: block;
  margin-bottom: 7px;
  color: var(--rv-text, #2b3445);
  font-size: 13px;
  font-weight: 700;
}
.field-label i {
  color: #e0648a;
  font-style: normal;
  margin-left: 2px;
}
.field-input {
  width: 100%;
  height: 44px;
  padding: 0 14px;
  border: 1px solid rgba(120, 145, 190, 0.24);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.75);
  color: var(--rv-text, #1f2a3d);
  font-size: 14px;
  transition: all 0.22s ease;
}
.field-input:focus {
  background: rgba(255, 255, 255, 0.95);
  border-color: rgba(52, 120, 246, 0.6);
  box-shadow: 0 0 0 4px rgba(52, 120, 246, 0.12);
  outline: none;
}
.field-input.area {
  height: 90px;
  padding: 10px 14px;
  line-height: 1.5;
  resize: none;
}
.counter {
  position: absolute;
  right: 12px;
  bottom: 8px;
  color: #9aa8bd;
  font-size: 11px;
}
.chip-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.chip {
  padding: 8px 15px;
  border: 1px solid rgba(120, 145, 190, 0.22);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.7);
  color: var(--rv-text-sub, #475569);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}
.chip:hover {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(52, 120, 246, 0.35);
}
.chip.active {
  border-color: transparent;
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  color: #fff;
  font-weight: 700;
  box-shadow: 0 6px 16px rgba(52, 120, 246, 0.28);
}
.captcha-row {
  display: flex;
  gap: 10px;
  align-items: center;
}
.captcha-input {
  flex: 1;
}
.captcha-box {
  width: 120px;
  height: 44px;
  border: 1px solid rgba(120, 145, 190, 0.24);
  border-radius: 12px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.85);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.hint {
  display: block;
  margin-top: 6px;
  color: var(--rv-text-light, #8a97ad);
  font-size: 11px;
}
.submit {
  width: 100%;
  height: 46px;
  margin-top: 6px;
  border: none;
  border-radius: 14px;
  background: linear-gradient(135deg, #7fb2ff 0%, #3478f6 100%);
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  box-shadow: 0 10px 24px rgba(52, 120, 246, 0.32);
  cursor: pointer;
  transition: all 0.22s ease;
}
.submit:active:not(:disabled) {
  transform: scale(0.98);
}
.submit:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}
.tips {
  margin: 12px 0 0;
  color: var(--rv-text-light, #8a97ad);
  font-size: 12px;
  line-height: 1.6;
  text-align: center;
}
.link {
  border: none;
  background: transparent;
  color: var(--rv-primary-deep, #2462d9);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
.result {
  margin-top: 18px;
}
.result-head {
  margin: 0 0 12px;
  color: var(--rv-text, #2b3445);
  font-size: 14px;
  font-weight: 800;
}
.empty {
  padding: 24px 0;
  color: var(--rv-text-light, #8a97ad);
  font-size: 13px;
  text-align: center;
}
.order-card {
  margin-bottom: 12px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 8px 20px rgba(46, 68, 112, 0.08);
}
.order-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.order-type {
  font-size: 15px;
  font-weight: 800;
  color: var(--rv-text, #1f2a3d);
}
.order-status {
  padding: 3px 12px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
}
.order-status.pending { background: rgba(240, 162, 75, 0.15); color: #d97706; }
.order-status.dispatched { background: rgba(52, 120, 246, 0.15); color: #2563eb; }
.order-status.working { background: rgba(52, 120, 246, 0.2); color: #1d4ed8; }
.order-status.done { background: rgba(34, 181, 115, 0.15); color: #059669; }
.order-status.canceled { background: rgba(138, 151, 173, 0.18); color: #64748b; }
.order-desc {
  margin: 10px 0 6px;
  color: var(--rv-text, #33415c);
  font-size: 13px;
  line-height: 1.6;
}
.order-meta {
  margin: 0;
  color: var(--rv-text-light, #8b98ad);
  font-size: 11px;
  line-height: 1.8;
}
.public-footer {
  padding: 24px 0 8px;
  color: var(--rv-text-light, #9aa8bd);
  font-size: 12px;
  text-align: center;
}
</style>