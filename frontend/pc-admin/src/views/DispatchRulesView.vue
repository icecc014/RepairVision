<template>
  <AdminShell title="派单规则配置" subtitle="技能/距离/负载权重、自动派单开关、超时提醒与积压保护线；改完点保存立即生效">
    <section class="panel">
      <div class="toolbar">
        <el-button type="primary" :loading="saving" @click="saveAll">保存修改</el-button>
        <el-button @click="resetAll">全部恢复默认</el-button>
        <el-button @click="openAdd">＋ 新增规则</el-button>
        <el-button text :loading="loading" @click="load">刷新</el-button>
        <div class="flex-1" />
        <el-tag :type="data.paused ? 'warning' : 'success'" effect="plain">
          {{ data.paused ? '自动派单已暂停（保护模式）' : '自动派单运行中' }}
        </el-tag>
        <el-button v-if="data.paused" size="small" type="primary" @click="resume">恢复自动派单</el-button>
      </div>

      <el-alert
        class="summary"
        :type="data.paused ? 'warning' : 'info'"
        show-icon
        :closable="false"
        :title="summaryText"
      />

      <div v-for="g in data.groups" :key="g.key" class="group">
        <div class="group-title">{{ g.label }}</div>
        <div v-for="r in g.items" :key="r.key" class="rule-row">
          <div class="rule-main">
            <div class="rule-name">
              {{ r.name }}
              <el-tag v-if="r.runtimeOnly" size="small" type="warning" effect="plain">运行态</el-tag>
              <el-tag v-else-if="r.custom" size="small" type="info" effect="plain">自定义</el-tag>
              <code class="rule-key">{{ r.key }}</code>
            </div>
            <div class="rule-desc">{{ r.remark || r.description }}</div>
          </div>
          <div class="rule-control">
            <el-switch
              v-if="r.type === 'bool'"
              v-model="draft[r.key]"
              :active-value="1"
              :inactive-value="0"
              :disabled="r.runtimeOnly"
            />
            <template v-else>
              <el-input-number
                v-model="draft[r.key]"
                :min="r.min"
                :max="r.max"
                :step="r.step"
                :precision="r.precision"
                size="small"
                style="width: 150px"
                :disabled="r.runtimeOnly"
              />
              <span class="unit">{{ r.unit || '权重' }}</span>
              <span class="range">{{ r.min }} ~ {{ r.max }}</span>
            </template>
          </div>
          <div class="rule-state">
            <el-switch
              v-model="enabled[r.key]"
              :active-value="1"
              :inactive-value="0"
              :disabled="r.runtimeOnly"
              inline-prompt
              active-text="启用"
              inactive-text="停用"
            />
          </div>
          <div class="rule-actions">
            <el-button size="small" text :disabled="r.runtimeOnly" @click="resetOne(r)">恢复默认</el-button>
            <el-button size="small" text type="danger" :disabled="r.registered" @click="removeRule(r)">删除</el-button>
          </div>
        </div>
      </div>

      <p class="foot-tip">
        说明：三条权重建议合计为 1；积压保护线取“绝对单量”与“在岗人数 × 倍数”中较大者，
        因此只有 1 人在岗时也要到 {{ guardMinOrdersHint }} 单才会暂停自动派单。
        规则保存后立即生效，引擎每次派单都会读取这里的值。
      </p>
    </section>

    <el-dialog v-model="addVisible" title="新增规则" width="580px">
      <el-radio-group v-model="addMode" size="small">
        <el-radio-button value="registered">补录系统注册规则</el-radio-button>
        <el-radio-button value="custom">自定义规则</el-radio-button>
      </el-radio-group>

      <div v-if="addMode === 'registered'" class="add-block">
        <el-select v-model="addKey" placeholder="选择尚未落库的注册规则" style="width: 100%" size="small">
          <el-option v-for="r in pendingRegistered" :key="r.key" :label="`${r.name}（${r.key}）`" :value="r.key" />
        </el-select>
        <p v-if="!pendingRegistered.length" class="foot-tip">
          系统注册的规则都已经存在，直接在上方列表里改值、启用/停用或点「恢复默认」即可。
        </p>
        <p v-else class="foot-tip">
          补录后会出现在上方对应分组里，可直接编辑；默认值 {{ addDefaultText }}。
        </p>
      </div>

      <div v-else class="add-block">
        <el-form label-width="90px">
          <el-form-item label="规则标识">
            <el-input v-model="addForm.key" placeholder="如 my_rule_weight" />
          </el-form-item>
          <el-form-item label="规则值">
            <el-input-number v-model="addForm.value" :precision="4" style="width: 100%" />
          </el-form-item>
          <el-form-item label="说明">
            <el-input v-model="addForm.remark" placeholder="这条规则的作用" />
          </el-form-item>
        </el-form>
        <el-alert
          type="warning"
          :closable="false"
          show-icon
          title="自定义规则只作记录：引擎目前不认识它，不会影响派单结果。需要真正生效时，先在服务端规则注册表登记，再由引擎读取。"
        />
      </div>

      <template #footer>
        <el-button @click="addVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitAdd">保存</el-button>
      </template>
    </el-dialog>
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { DispatchRuleEntry, DispatchRulesData } from '../api'
import {
  apiAdminDispatchGuard,
  apiCreateDispatchRule,
  apiDeleteDispatchRule,
  apiDispatchRulesGrouped,
  apiResetDispatchRule,
  apiResumeAutoDispatch,
  apiSaveDispatchRules,
} from '../api'
import AdminShell from '../components/AdminShell.vue'

const data = ref<DispatchRulesData>({ groups: [], paused: false, pendingCount: 0, onDutyCount: 0, updatedAt: '' })
// draft / enabled 用 key 索引，便于按元数据渲染控件
const draft = reactive<Record<string, number>>({})
const enabled = reactive<Record<string, number>>({})
const loading = ref(false)
const saving = ref(false)
const addVisible = ref(false)
const addMode = ref<'registered' | 'custom'>('registered')
const addKey = ref('')
const addForm = reactive({ key: '', value: 0, remark: '' })

const allRules = computed<DispatchRuleEntry[]>(() => data.value.groups.flatMap((g) => g.items))
const pendingRegistered = computed(() => allRules.value.filter((r) => r.registered && !r.saved && !r.runtimeOnly))
const addDefaultText = computed(() => {
  const rule = pendingRegistered.value.find((r) => r.key === addKey.value)
  return rule ? String(rule.defaultValue) : '—'
})
const guardMinOrdersHint = computed(() => {
  const rule = allRules.value.find((r) => r.key === 'backlog_guard_min_orders')
  return rule ? draft[rule.key] : 20
})
const summaryText = computed(() => {
  const base = `待派 ${data.value.pendingCount} 单 · 当前在岗 ${data.value.onDutyCount} 人 · 规则改动保存后立即生效（引擎每次派单都读取这里的值）`
  return data.value.paused ? base + ' · 当前处于保护模式，可点右上角「恢复自动派单」' : base
})

async function load() {
  loading.value = true
  try {
    const resp = await apiDispatchRulesGrouped()
    data.value = resp
    for (const rule of resp.groups.flatMap((g) => g.items)) {
      draft[rule.key] = Number(rule.value)
      enabled[rule.key] = Number(rule.enabled) === 0 ? 0 : 1
    }
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function saveAll() {
  const items: Array<{ key: string; value: number; enabled: number; remark?: string }> = []
  for (const rule of allRules.value) {
    if (rule.runtimeOnly) continue
    const current = draft[rule.key]
    items.push({
      key: rule.key,
      value: Number(current === undefined ? rule.value : current),
      enabled: enabled[rule.key] === 0 ? 0 : 1,
      remark: rule.remark,
    })
  }
  if (!items.length) {
    ElMessage.warning('没有可保存的规则')
    return
  }
  saving.value = true
  try {
    await apiSaveDispatchRules(items)
    ElMessage.success('规则已保存并立即生效')
    await load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

function resetOne(rule: DispatchRuleEntry) {
  draft[rule.key] = rule.defaultValue
  enabled[rule.key] = 1
  ElMessage.info(`「${rule.name}」已还原为默认值 ${rule.defaultValue}，点「保存修改」生效`)
}

function resetAll() {
  for (const rule of allRules.value) {
    if (rule.runtimeOnly) continue
    draft[rule.key] = rule.defaultValue
    enabled[rule.key] = 1
  }
  ElMessage.info('已把所有注册规则还原为默认值，点「保存修改」生效')
}

async function resume() {
  try {
    await apiResumeAutoDispatch()
    ElMessage.success('已恢复自动派单，积压工单会立即重新派发')
    await load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function removeRule(rule: DispatchRuleEntry) {
  try {
    await ElMessageBox.confirm(`删除自定义规则「${rule.key}」？`, '删除规则', { type: 'warning' })
  } catch {
    return
  }
  try {
    await apiDeleteDispatchRule(rule.key)
    ElMessage.success('已删除')
    await load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function openAdd() {
  addMode.value = pendingRegistered.value.length ? 'registered' : 'custom'
  addKey.value = pendingRegistered.value.length ? pendingRegistered.value[0].key : ''
  addForm.key = ''
  addForm.value = 0
  addForm.remark = ''
  addVisible.value = true
}

async function submitAdd() {
  saving.value = true
  try {
    if (addMode.value === 'registered') {
      const rule = pendingRegistered.value.find((r) => r.key === addKey.value)
      if (!rule) {
        ElMessage.warning('请选择要补录的注册规则')
        return
      }
      await apiCreateDispatchRule({
        ruleKey: rule.key,
        ruleValue: String(rule.defaultValue),
        enabled: 1,
        remark: rule.description,
      })
      ElMessage.success(`已补录「${rule.name}」`)
    } else {
      if (!/^[a-z][a-z0-9_]{2,31}$/.test(addForm.key.trim())) {
        ElMessage.warning('规则标识需为小写字母开头、可含数字与下划线（3~32 位）')
        return
      }
      await apiCreateDispatchRule({
        ruleKey: addForm.key.trim(),
        ruleValue: String(addForm.value),
        enabled: 1,
        remark: addForm.remark.trim(),
      })
      ElMessage.success('已新增自定义规则（仅作记录）')
    }
    addVisible.value = false
    await load()
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
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.flex-1 {
  flex: 1;
}

.summary {
  margin-bottom: 16px;
}

.group {
  margin-bottom: 18px;
}

.group-title {
  font-size: 13px;
  font-weight: 700;
  color: #33415c;
  padding-bottom: 8px;
  border-bottom: 1px solid #eef2f7;
  margin-bottom: 10px;
}

.rule-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px 96px 150px;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 10px;
}

.rule-row:hover {
  background: #f8fafc;
}

.rule-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 700;
  color: #1f2a3d;
}

.rule-key {
  font-size: 11px;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 1px 6px;
  border-radius: 6px;
}

.rule-desc {
  margin-top: 4px;
  font-size: 12px;
  color: #7b8aa0;
  line-height: 1.6;
}

.rule-control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.unit {
  font-size: 12px;
  color: #64748b;
}

.range {
  font-size: 11px;
  color: #b0bccd;
}

.rule-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
}

.add-block {
  margin-top: 14px;
}

.foot-tip {
  margin-top: 12px;
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.7;
}
/* ---------- V9.7.3 移动端：4 列规则行改单列，控件占满宽度 ---------- */
@media (max-width: 767px) {
  .toolbar {
    flex-wrap: wrap;
    gap: 8px;
  }

  .rule-row {
    grid-template-columns: 1fr;
    align-items: stretch;
    gap: 10px;
  }

  .rule-main {
    min-width: 0;
  }

  .rule-control {
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .rule-control :deep(.el-input-number),
  .rule-control :deep(.el-select) {
    width: 100% !important;
  }
}
</style>
