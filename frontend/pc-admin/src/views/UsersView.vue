<template>
  <AdminShell title="人员账号管理" subtitle="维护宿管、维修工人与管理员账号">
    <section class="panel">
      <div class="toolbar">
        <el-button type="primary" @click="openCreate">＋ 新增账号</el-button>
        <el-select v-model="filter.role" placeholder="全部角色" clearable style="width: 150px" @change="load">
          <el-option label="管理员" :value="1" />
          <el-option label="维修工人" :value="2" />
          <el-option label="宿管" :value="3" />
        </el-select>
        <el-select v-model="filter.status" placeholder="全部状态" clearable style="width: 130px" @change="load">
          <el-option label="启用" :value="1" />
          <el-option label="停用" :value="0" />
        </el-select>
        <el-input v-model="filter.keyword" placeholder="搜索账号/姓名" clearable style="width: 200px" @keyup.enter="load" />
        <el-button type="primary" @click="load">查询</el-button>
      </div>

      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="账号" width="130" />
        <el-table-column prop="name" label="姓名" width="130" />
        <el-table-column prop="roleText" label="角色" width="110" />
        <el-table-column prop="phone" label="手机" width="130">
          <template #default="{ row }">{{ row.phone || '—' }}</template>
        </el-table-column>
        <el-table-column label="绑定" min-width="220">
          <template #default="{ row }">
            <template v-if="row.role === 2">{{ (row.buildings || []).join('、') || '—' }}</template>
            <template v-else-if="row.role === 3 && row.buildingId">
              {{ buildingName(row.buildingId) || `#${row.buildingId}` }}
            </template>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.statusText }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="warning" plain @click="openReset(row)">重置密码</el-button>
            <el-button v-if="row.status === 1 && row.role !== 1" size="small" type="danger" plain @click="disable(row)">停用</el-button>
            <el-button v-else-if="row.status === 0" size="small" type="success" plain @click="enable(row)">启用</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无账号" class="empty" />
    </section>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑账号' : '新增账号'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="账号">
          <el-input v-model="form.username" :disabled="!!editingId" placeholder="登录账号" />
        </el-form-item>
        <el-form-item v-if="!editingId" label="初始密码">
          <el-input v-model="form.password" placeholder="至少6位" />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="form.name" placeholder="真实姓名" />
        </el-form-item>
        <el-form-item label="手机">
          <el-input v-model="form.phone" placeholder="选填" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" :disabled="isAdminRow" style="width: 100%">
            <el-option label="超级管理员" :value="1" />
            <el-option label="维修工人" :value="2" />
            <el-option label="宿管" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.role === 3" label="绑定楼栋">
          <el-select v-model="form.buildingId" style="width: 100%">
            <el-option v-for="b in buildings" :key="b.id" :label="b.code + ' ' + b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item v-else-if="form.role === 2" label="管辖楼栋">
          <el-select v-model="form.buildingIds" multiple style="width: 100%">
            <el-option v-for="b in buildings" :key="b.id" :label="b.code + ' ' + b.name" :value="b.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resetVisible" title="重置密码" width="420px">
      <el-input v-model="resetPassword" type="password" show-password placeholder="输入新密码（至少6位）" />
      <template #footer>
        <el-button @click="resetVisible = false">取消</el-button>
        <el-button type="warning" :loading="saving" @click="submitReset">重置</el-button>
      </template>
    </el-dialog>
  </AdminShell>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminBuilding, AdminUser } from '../api'
import {
  apiAdminBuildings,
  apiAdminUsers,
  apiCreateUser,
  apiDeleteUser,
  apiResetPassword,
  apiUpdateUser,
} from '../api'
import AdminShell from '../components/AdminShell.vue'

const list = ref<AdminUser[]>([])
const buildings = ref<AdminBuilding[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const resetVisible = ref(false)
const resetTarget = ref<AdminUser | null>(null)
const resetPassword = ref('')
const editingId = ref<number | null>(null)
const isAdminRow = ref(false)
const filter = reactive({ role: 0, status: 0, keyword: '' })
const form = reactive({
  username: '',
  password: '',
  name: '',
  phone: '',
  role: 3,
  buildingId: 0,
  buildingIds: [] as number[],
})

async function load() {
  loading.value = true
  try {
    list.value = await apiAdminUsers({
      role: filter.role || 0,
      status: filter.status,
      keyword: filter.keyword,
    })
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadBuildings() {
  try {
    buildings.value = await apiAdminBuildings()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function buildingName(id: number) {
  const b = buildings.value.find((x) => x.id === id)
  return b ? `${b.code} ${b.name}` : ''
}

function openCreate() {
  editingId.value = null
  isAdminRow.value = false
  Object.assign(form, {
    username: '',
    password: '',
    name: '',
    phone: '',
    role: 3,
    buildingId: buildings.value[0]?.id || 0,
    buildingIds: [] as number[],
  })
  dialogVisible.value = true
}

function openEdit(row: AdminUser) {
  editingId.value = row.id
  isAdminRow.value = row.role === 1
  Object.assign(form, {
    username: row.username,
    password: '',
    name: row.name,
    phone: row.phone || '',
    role: row.role,
    buildingId: row.buildingId || 0,
    buildingIds: row.buildingIds ? [...row.buildingIds] : [],
  })
  dialogVisible.value = true
}

async function save() {
  if (!form.name.trim() || (!editingId.value && (!form.username.trim() || !form.password))) {
    ElMessage.warning('请完整填写账号信息')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      const row = list.value.find((x) => x.id === editingId.value)
      await apiUpdateUser(editingId.value, {
        name: form.name.trim(),
        phone: form.phone.trim(),
        role: form.role,
        status: row?.status === 1 ? 1 : 1,
        buildingId: form.role === 3 ? form.buildingId : 0,
        buildingIds: form.role === 2 ? form.buildingIds : [],
      })
    } else {
      await apiCreateUser({
        username: form.username.trim(),
        password: form.password,
        name: form.name.trim(),
        phone: form.phone.trim(),
        role: form.role,
        buildingId: form.role === 3 ? form.buildingId : 0,
        buildingIds: form.role === 2 ? form.buildingIds : [],
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

function openReset(row: AdminUser) {
  resetTarget.value = row
  resetPassword.value = ''
  resetVisible.value = true
}

async function submitReset() {
  if (!resetTarget.value || resetPassword.value.length < 6) {
    ElMessage.warning('新密码至少6位')
    return
  }
  saving.value = true
  try {
    await apiResetPassword(resetTarget.value.id, resetPassword.value)
    ElMessage.success('密码已重置')
    resetVisible.value = false
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

async function disable(row: AdminUser) {
  try {
    await ElMessageBox.confirm(`确认停用账号 ${row.username}？停用后无法登录。`, '停用确认')
  } catch {
    return
  }
  try {
    await apiDeleteUser(row.id)
    ElMessage.success('已停用')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function enable(row: AdminUser) {
  try {
    await apiUpdateUser(row.id, {
      name: row.name,
      phone: row.phone || '',
      role: row.role,
      status: 1,
      buildingId: row.buildingId || 0,
      buildingIds: row.buildingIds || [],
    })
    ElMessage.success('已启用')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

onMounted(() => {
  load()
  loadBuildings()
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
  gap: 12px;
  margin-bottom: 14px;
}
.empty {
  padding: 24px 0;
}
</style>