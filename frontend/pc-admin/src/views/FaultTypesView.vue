<template>
  <AdminShell title="维修类型字典" subtitle="维护宿管报修时可选择的故障类型">
    <section class="panel">
      <div class="toolbar">
        <el-button type="primary" @click="openCreate">＋ 新增维修类型</el-button>
      </div>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="code" label="编码" width="140" />
        <el-table-column prop="name" label="显示名称" min-width="160" />
        <el-table-column prop="sort" label="排序" width="90" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status === 1" size="small" type="danger" plain @click="disable(row)">
              停用
            </el-button>
            <el-button v-else size="small" type="success" plain @click="enable(row)">启用</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无维修类型" class="empty" />
    </section>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑维修类型' : '新增维修类型'" width="440px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="编码">
          <el-input v-model="form.code" :disabled="!!editingId" placeholder="如 electric" />
        </el-form-item>
        <el-form-item label="显示名称">
          <el-input v-model="form.name" placeholder="如 电维修" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminFaultType } from '../api'
import {
  apiAdminFaultTypes,
  apiCreateFaultType,
  apiDeleteFaultType,
  apiUpdateFaultType,
} from '../api'
import AdminShell from '../components/AdminShell.vue'

const list = ref<AdminFaultType[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ code: '', name: '', sort: 0 })

async function load() {
  loading.value = true
  try {
    list.value = await apiAdminFaultTypes()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  form.code = ''
  form.name = ''
  form.sort = 0
  dialogVisible.value = true
}

function openEdit(row: AdminFaultType) {
  editingId.value = row.id
  form.code = row.code
  form.name = row.name
  form.sort = row.sort
  dialogVisible.value = true
}

async function save() {
  if (!form.code.trim() || !form.name.trim()) {
    ElMessage.warning('编码和名称不能为空')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await apiUpdateFaultType(editingId.value, { name: form.name.trim(), sort: form.sort, status: 1 })
    } else {
      await apiCreateFaultType({ code: form.code.trim(), name: form.name.trim(), sort: form.sort })
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

async function disable(row: AdminFaultType) {
  try {
    await ElMessageBox.confirm(`停用后 H5 不再显示“${row.name}”，确认停用？`, '停用确认')
  } catch {
    return
  }
  try {
    await apiDeleteFaultType(row.id)
    ElMessage.success('已停用')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function enable(row: AdminFaultType) {
  try {
    await apiUpdateFaultType(row.id, { name: row.name, sort: row.sort, status: 1 })
    ElMessage.success('已启用')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
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
  margin-bottom: 14px;
}

.empty {
  padding: 24px 0;
}
</style>