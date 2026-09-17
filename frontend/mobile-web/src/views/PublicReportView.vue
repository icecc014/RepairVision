<template>
  <div class="public-page">
    <header class="public-header">
      <h1>校园维修报修</h1>
      <p>无需登录 · 支持微信打开 · 可直接报修与查询进度</p>
    </header>

    <nav class="public-tabs">
      <button :class="{ active: tab === 'report' }" @click="tab = 'report'">我要报修</button>
      <button :class="{ active: tab === 'query' }" @click="tab = 'query'">查询进度</button>
    </nav>

    <section v-if="tab === 'report'" class="public-card">
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

    <section v-else class="public-card">
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
          {{ queryResult.buildingName }} {{ queryResult.room }} 室 · 共 {{ queryResult.orders.length }} 条记录
        </p>
        <div v-if="queryResult.orders.length === 0" class="empty">该房间暂无报修记录</div>
        <div v-for="(o, i) in queryResult.orders" :key="i" class="order-card">
          <div class="order-head">
            <span class="order-type">{{ o.faultTypeName || '维修' }}</span>
            <span class="order-status" :class="statusClass(o.statusText)">{{ o.statusText }}</span>
          </div>
          <p class="order-desc">{{ o.description }}</p>
          <p class="order-meta">报修：{{ o.createdAt }}<span v-if="o.completedAt"> · 完成：{{ o.completedAt }}</span></p>
          <p v-if="o.workerName" class="order-meta">维修工人：{{ o.workerName }}</p>
        </div>
      </div>
    </section>

    <footer class="public-footer">RepairVision · 校园维修工单智能调度系统</footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { showToast } from 'vant'
import { apiPublicBuildings, apiPublicCaptcha, apiPublicReport, apiPublicRoomOrders } from '../api'

interface PublicBuilding {
  id: number
  code: string
  name: string
}

interface HistoryItem {
  buildingId: number
  buildingName: string
  room: string
  timestamp: number
}

const HISTORY_KEY = 'repair_history'
const HISTORY_MAX = 10

const tab = ref<'report' | 'query'>('report')
const buildings = ref<PublicBuilding[]>([])
const history = ref<HistoryItem[]>([])
const captchaSvg = ref('')
const captchaId = ref('')
const submitting = ref(false)
const querying = ref(false)
const queryDone = ref(false)

const faultOptions = [
  { code: 'electric', label: '电维修' },
  { code: 'water', label: '水维修' },
  { code: 'other', label: '其他维修' },
]
const reporterOptions = [
  { value: 1, label: '学生' },
  { value: 2, label: '教师' },
  { value: 3, label: '其他' },
]

const form = reactive({
  buildingId: 0,
  room: '',
  faultType: 'electric',
  description: '',
  reporterType: 1,
  contact: '',
  captchaCode: '',
})
const query = reactive({ buildingId: 0, room: '' })
const errorMsg = ref('')
const selectedBuilding = computed(() => buildings.value.find((b) => b.id === form.buildingId) || null)
const roomHint = computed(() => {
  const b = selectedBuilding.value
  if (!b) return ''
  const per = b.roomsPerFloor > 0 ? b.roomsPerFloor : 16
  const sample = `${b.floors}01 ~ ${b.floors}${String(per).padStart(2, '0')}`
  return `${b.name}：共 ${b.floors} 层、每层 ${per} 间，房间号形如 ${sample}`
})
function fail(msg: string) {
  errorMsg.value = msg
  showToast(msg)
}
// 房间号前置校验（避免提交到服务端才返回 400）
function roomRangeError(roomValue: string): string {
  const b = selectedBuilding.value
  if (!b) return ''
  const per = b.roomsPerFloor > 0 ? b.roomsPerFloor : 16
  const num = Number(roomValue)
  const floor = Math.floor(num / 100)
  if (floor < 1 || floor > b.floors) {
    return `${b.name}共 ${b.floors} 层，房间号首位应是楼层（如 ${b.floors}01）`
  }
  const index = num - floor * 100
  if (index < 1 || index > per) {
    return `${b.name}每层 ${per} 间，${floor} 层房间号范围是 ${floor}01 ~ ${floor}${String(per).padStart(2, '0')}`
  }
  return ''
}
const queryResult = reactive({
  buildingName: '',
  room: '',
  orders: [] as Array<{ faultTypeName: string; description: string; statusText: string; createdAt: string; completedAt?: string; workerName?: string }>,
})

function readHistory(): HistoryItem[] {
  try {
    return JSON.parse(localStorage.getItem(HISTORY_KEY) || '[]') as HistoryItem[]
  } catch {
    return []
  }
}

function saveHistory(buildingId: number, buildingName: string, room: string) {
  const next = readHistory().filter((item) => !(item.buildingId === buildingId && item.room === room))
  next.unshift({ buildingId, buildingName, room, timestamp: Date.now() })
  const trimmed = next.slice(0, HISTORY_MAX)
  localStorage.setItem(HISTORY_KEY, JSON.stringify(trimmed))
  history.value = trimmed
}

function clearHistory() {
  localStorage.removeItem(HISTORY_KEY)
  history.value = []
  showToast('已清除历史记录')
}

function applyHistory(item: HistoryItem) {
  form.buildingId = item.buildingId
  form.room = item.room
  showToast('已填充 ' + item.buildingName + ' ' + item.room + ' 室')
}

function statusClass(text: string) {
  if (text.includes('完成')) return 'done'
  if (text.includes('取消')) return 'canceled'
  if (text.includes('维修')) return 'working'
  if (text.includes('派')) return 'dispatched'
  return 'pending'
}

async function loadBuildings() {
  try {
    const resp = await apiPublicBuildings()
    buildings.value = resp.list || []
  } catch (err) {
    showToast('楼栋加载失败：' + (err as Error).message)
  }
}

async function loadCaptcha() {
  try {
    const resp = await apiPublicCaptcha()
    captchaId.value = resp.captchaId
    captchaSvg.value = resp.svg
    form.captchaCode = ''
  } catch (err) {
    showToast('验证码获取失败：' + (err as Error).message)
  }
}

async function submit() {
  const room = form.room.trim()
  errorMsg.value = ''
  if (!form.buildingId) { fail('请选择报修楼栋'); return }
  if (!/^\d{3,4}$/.test(room)) { fail('请填写正确房间号，如 401'); return }
  const rangeError = roomRangeError(room)
  if (rangeError) { fail(rangeError); return }
  if (!form.faultType) { fail('请选择故障类型'); return }
  if (form.description.trim().length < 5) { fail('故障描述至少 5 个字'); return }
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
    // 控制台里也能看到具体原因（页面同时给出常驻提示条）
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
  padding: 0 14px 26px;
  background: linear-gradient(160deg, #eef4ff 0%, #f7fafc 55%, #eef7f2 100%);
  color: #1f2a3d;
}
.public-header {
  padding: 26px 6px 14px;
}
.public-header h1 {
  margin: 0 0 6px;
  font-size: 22px;
  font-weight: 800;
}
.public-header p {
  margin: 0;
  color: #5a6a85;
  font-size: 12px;
}
.public-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.public-tabs button {
  flex: 1;
  height: 40px;
  border: 1px solid #d4e0f0;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.75);
  color: #33415c;
  font-size: 14px;
  font-weight: 600;
}
.public-tabs button.active {
  border-color: #3478f6;
  background: linear-gradient(135deg, #4b86f8, #3478f6);
  color: #fff;
  box-shadow: 0 8px 18px rgba(52, 120, 246, 0.28);
}
.error-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border: 1px solid #f5c2c0;
  border-radius: 10px;
  background: #fff1f0;
  color: #c0392b;
  font-size: 13px;
  line-height: 1.6;
}
.public-card {
  padding: 14px 14px 18px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 12px 30px rgba(46, 68, 112, 0.1);
  backdrop-filter: blur(12px);
}
.history {
  margin-bottom: 14px;
  padding: 10px;
  border-radius: 12px;
  background: #f3f7ff;
}
.history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  color: #5a6a85;
  font-size: 12px;
}
.history-list {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
}
.history-item {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  min-width: 108px;
  padding: 8px 10px;
  border: 1px solid #cddcf5;
  border-radius: 10px;
  background: #fff;
  color: #2b3445;
}
.history-item b {
  font-size: 12px;
}
.history-item span {
  color: #3478f6;
  font-size: 13px;
  font-weight: 700;
}
.field {
  position: relative;
  display: block;
  margin-bottom: 14px;
}
.field-label {
  display: block;
  margin-bottom: 6px;
  color: #33415c;
  font-size: 13px;
  font-weight: 600;
}
.field-label i {
  color: #e05656;
  font-style: normal;
}
.field-input {
  width: 100%;
  height: 42px;
  padding: 0 12px;
  border: 1px solid #d4e0f0;
  border-radius: 10px;
  background: #fff;
  color: #1f2a3d;
  font-size: 14px;
}
.field-input.area {
  height: 84px;
  padding: 10px 12px;
  line-height: 1.5;
  resize: none;
}
.counter {
  position: absolute;
  right: 10px;
  bottom: 6px;
  color: #9aa8bd;
  font-size: 11px;
}
.chip-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.chip {
  padding: 8px 14px;
  border: 1px solid #d4e0f0;
  border-radius: 999px;
  background: #fff;
  color: #33415c;
  font-size: 13px;
}
.chip.active {
  border-color: #3478f6;
  background: #eaf1ff;
  color: #2462d9;
  font-weight: 700;
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
  height: 42px;
  border: 1px solid #d4e0f0;
  border-radius: 10px;
  overflow: hidden;
  background: #eef3fb;
}
.hint {
  display: block;
  margin-top: 6px;
  color: #9aa8bd;
  font-size: 11px;
}
.submit {
  width: 100%;
  height: 46px;
  margin-top: 4px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #4b86f8, #3478f6);
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  box-shadow: 0 10px 22px rgba(52, 120, 246, 0.3);
}
.submit:disabled {
  opacity: 0.7;
}
.tips {
  margin: 10px 0 0;
  color: #9aa8bd;
  font-size: 11px;
  line-height: 1.6;
}
.link {
  border: none;
  background: transparent;
  color: #3478f6;
  font-size: 12px;
}
.result {
  margin-top: 16px;
}
.result-head {
  margin: 0 0 10px;
  color: #33415c;
  font-size: 13px;
  font-weight: 700;
}
.empty {
  padding: 18px 0;
  color: #9aa8bd;
  font-size: 13px;
  text-align: center;
}
.order-card {
  margin-bottom: 10px;
  padding: 12px;
  border: 1px solid #e3ebf7;
  border-radius: 12px;
  background: #fff;
}
.order-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.order-type {
  font-size: 14px;
  font-weight: 700;
}
.order-status {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
}
.order-status.pending { background: #fff4e5; color: #b45309; }
.order-status.dispatched { background: #eaf1ff; color: #2462d9; }
.order-status.working { background: #e8f1ff; color: #1d4ed8; }
.order-status.done { background: #e7f7ec; color: #15803d; }
.order-status.canceled { background: #f1f3f7; color: #64748b; }
.order-desc {
  margin: 8px 0 4px;
  color: #33415c;
  font-size: 13px;
  line-height: 1.6;
}
.order-meta {
  margin: 0;
  color: #8b98ad;
  font-size: 11px;
  line-height: 1.8;
}
.public-footer {
  padding: 18px 0 6px;
  color: #9aa8bd;
  font-size: 11px;
  text-align: center;
}
</style>
