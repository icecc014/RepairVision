<template>
  <AdminShell title="派单规则配置" subtitle="调整技能/距离/负载权重，P2 加权派单将按此计算">
    <section class="panel">
      <div class="toolbar">
        <el-button type="primary" @click="openCreate">＋ 新增规则</el-button>
        <span class="toolbar-tip">权重建议合计为 1.0；auto_dispatch_enabled=1 表示启用自动派单</span>
      </div>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="ruleKey" label="规则标识" min-width="180" />
        <el-table-column prop="remark" label="说明" min-width="200" />
        <el-table-column label="值" width="130">
          <template #default="{ row }">
            {{ row.ruleKey === 'auto_dispatch_enabled' ? (Number(row.ruleValue) ? '开启' : '关闭') : row.ruleValue }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small">
              {{ row.enabled === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="170" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无派单规则" class="empty" />
    </section>

    <el-dialog append-to-body v-model="dialogVisible" :title="editingId ? '编辑规则' : '新增规则'" width="460px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="规则标识">
          <el-input v-model="form.ruleKey" :disabled="!!editingId" placeholder="如 load_weight" />
        </el-form-item>
        <el-form-item label="规则值">
          <el-input-number
            v-model="form.ruleValue"
            :min="0"
            :max="1"
            :step="0.05"
            :precision="4"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="是否启用">
          <el-switch v-model="form.enabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.remark" placeholder="规则说明（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </AdminShell>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { DispatchRule } from '../api'
import { apiCreateDispatchRule, apiDispatchRules, apiUpdateDispatchRule } from '../api'
import AdminShell from '../components/AdminShell.vue'

const list = ref<DispatchRule[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ ruleKey: '', ruleValue: 0, enabled: 1, remark: '' })

async function load() {
  loading.value = true
  try {
    list.value = await apiDispatchRules()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  form.ruleKey = ''
  form.ruleValue = 0
  form.enabled = 1
  form.remark = ''
  dialogVisible.value = true
}

function openEdit(row: DispatchRule) {
  editingId.value = row.id
  form.ruleKey = row.ruleKey
  form.ruleValue = Number(row.ruleValue)
  form.enabled = Number(row.enabled)
  form.remark = row.remark
  dialogVisible.value = true
}

async function save() {
  if (!form.ruleKey.trim()) {
    ElMessage.warning('规则标识不能为空')
    return
  }
  saving.value = true
  const valueText = form.ruleValue.toFixed(4)
  try {
    if (editingId.value) {
      await apiUpdateDispatchRule(editingId.value, {
        ruleValue: valueText,
        enabled: form.enabled,
        remark: form.remark.trim(),
      })
    } else {
      await apiCreateDispatchRule({
        ruleKey: form.ruleKey.trim(),
        ruleValue: valueText,
        enabled: form.enabled,
        remark: form.remark.trim(),
      })
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
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
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 14px;
}

.toolbar-tip {
  color: #94a3b8;
  font-size: 12px;
}

.empty {
  padding: 24px 0;
}
</style>