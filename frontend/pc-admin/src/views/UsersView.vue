SKIP(0):     const all = await apiAdminUsers({ role: filter.role || 0, keyword: filter.keyword })
    // 状态筛选在本地完成：接口把 status=0（停用）当作“未传”，无法直接筛停用
    list.value = filter.status === 0 || filter.status === 1
      ? all.filter((u) => u.status === filter.status)
      : all
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
  return b ? (String(b.name).startsWith(String(b.code)) ? b.name : `${b.code} ${b.name}`) : ''
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
    maxConcurrent: 3,
    jobType: 0,
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
    maxConcurrent: row.maxConcurrent || 3,
    jobType: row.jobType || 0,
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
        maxConcurrent: form.maxConcurrent,
        jobType: form.role === 2 ? form.jobType : 0,
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
        maxConcurrent: form.maxConcurrent,
        jobType: form.role === 2 ? form.jobType : 0,
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
      maxConcurrent: row.maxConcurrent || 3,
      jobType: row.jobType || 0,
    })
    ElMessage.success('已启用')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function jobTypeText(jobType?: number) {
  if (jobType === 1) return '电工'
  if (jobType === 2) return '水工'
  if (jobType === 3) return '泥瓦工'
  if (jobType === 4) return '木工'
  return '通用'
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
.form-tip {
  margin-top: 4px;
  color: #94a3b8;
  font-size: 12px;
  line-height: 1.5;
}
</style>