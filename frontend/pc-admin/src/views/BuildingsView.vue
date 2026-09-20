<template>
  <AdminShell title="建筑信息管理" subtitle="维护楼栋 2D 坐标与 3D 楼宇参数">
    <section class="panel">
      <div class="toolbar">
        <el-button type="primary" @click="openCreate">＋ 新增楼栋</el-button>
        <el-input v-model="keyword" placeholder="搜索编码 / 名称" clearable style="width: 200px" />
        <span class="toolbar-tip">排序</span>
        <el-select v-model="sortBy" style="width: 150px">
          <el-option label="按 ID" value="id" />
          <el-option label="按编码" value="code" />
          <el-option label="按名称" value="name" />
          <el-option label="按楼层" value="floors" />
        </el-select>
        <el-button @click="resetFilter">重置</el-button>
        <span class="toolbar-tip">共 {{ filteredList.length }} / {{ list.length }} 栋</span>
      </div>
      <el-table :data="filteredList" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="code" label="编码" width="100" />
        <el-table-column prop="name" label="楼栋名称" min-width="160" />
        <el-table-column label="2D坐标" width="160">
          <template #default="{ row }">
            <template v-if="row.posX || row.posY">({{ row.posX }}, {{ row.posY }})</template>
            <span v-else class="pos-empty">未同步</span>
          </template>
        </el-table-column>
        <el-table-column label="尺寸" width="150">
          <template #default="{ row }">{{ row.width }} × {{ row.height }}</template>
        </el-table-column>
        <el-table-column prop="floors" label="楼层" width="80" />
        <el-table-column label="层高" width="90">
          <template #default="{ row }">{{ row.floorHeight }}</template>
        </el-table-column>
        <el-table-column prop="roomsPerFloor" label="每层房间" width="100" />
        <el-table-column label="楼层布局" width="120">
          <template #default="{ row }">
            <span class="layout-tag" :class="{ custom: !!row.layoutJson }">
              {{ row.layoutJson ? '已自定义' : '内置标准层' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="360">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openEdit(row)">编辑</el-button>
            <el-button size="small" @click="onDesignerClick(row)">布局设计</el-button>
            <el-button size="small" type="danger" plain @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length === 0" description="暂无楼栋" class="empty" />
    </section>

    <el-dialog append-to-body v-model="dialogVisible" :title="editingId ? '编辑楼栋' : '新增楼栋'" width="560px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="楼栋编码">
          <el-input v-model="form.code" placeholder="如 A3" />
        </el-form-item>
        <el-form-item label="楼栋名称">
          <el-input v-model="form.name" placeholder="如 3号宿舍楼" />
        </el-form-item>
        <el-form-item label="2D坐标 X / Y">
          <div class="inline-pair">
            <el-input-number v-model="form.posX" :precision="2" style="width: 100%" />
            <el-input-number v-model="form.posY" :precision="2" />
          </div>
        </el-form-item>
        <el-form-item label="宽 / 高">
          <div class="inline-pair">
            <el-input-number v-model="form.width" :precision="2" :min="1" />
            <el-input-number v-model="form.height" :precision="2" :min="1" />
          </div>
        </el-form-item>
        <el-form-item label="楼层数">
          <el-input-number v-model="form.floors" :min="1" />
        </el-form-item>
        <el-form-item label="层高">
          <el-input-number v-model="form.floorHeight" :precision="2" :min="0.5" :step="0.1" />
        </el-form-item>
        <el-form-item label="每层房间数">
          <el-input-number v-model="form.roomsPerFloor" :min="1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
    <LayoutDesigner v-model="designerVisible" :building="designerBuilding" @saved="load" />
  </AdminShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminBuilding } from '../api'
import { apiAdminBuildings, apiCreateBuilding, apiDeleteBuilding, apiUpdateBuilding } from '../api'
import AdminShell from '../components/AdminShell.vue'
import LayoutDesigner from '../components/LayoutDesigner.vue'
import { useIsMobile } from '../composables/useViewport'

const list = ref<AdminBuilding[]>([])
const loading = ref(false)
const keyword = ref('')
const sortBy = ref<'id' | 'code' | 'name' | 'floors'>('id')

// 列表筛选与排序：编码/名称关键字 + 按 ID / 编码 / 名称 / 楼层排序
const filteredList = computed(() => {
  const kw = keyword.value.trim()
  let arr = [...list.value]
  if (kw) arr = arr.filter((b) => String(b.code).includes(kw) || String(b.name).includes(kw))
  const key = sortBy.value
  arr.sort((a, b) => {
    if (key === 'code') return (Number(a.code) || 0) - (Number(b.code) || 0)
    if (key === 'floors') return (Number(a.floors) || 0) - (Number(b.floors) || 0)
    if (key === 'name') return String(a.name).localeCompare(String(b.name), 'zh-Hans-CN')
    return a.id - b.id
  })
  return arr
})

function resetFilter() {
  keyword.value = ''
  sortBy.value = 'id'
}
const saving = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const designerVisible = ref(false)

// V9.7.4 移动端不提供布局设计（需要鼠标与键盘）
const { isMobile } = useIsMobile()
const designerBuilding = ref<AdminBuilding | null>(null)
const form = reactive({
  code: '',
  name: '',
  posX: 0,
  posY: 0,
  width: 60,
  height: 36,
  floors: 6,
  floorHeight: 3.5,
  roomsPerFloor: 20,
})

function onDesignerClick(row: AdminBuilding) {
  // V9.7.4：布局设计是画布级编辑，移动端提示改用电脑端
  if (isMobile.value) {
    ElMessage.warning('布局设计需要在电脑端操作')
    return
  }
  openDesigner(row)
}

function openDesigner(row: AdminBuilding) {
  if (isMobile.value) {
    ElMessage.warning('布局设计需要在电脑端操作')
    return
  }
  designerBuilding.value = { ...row }
  designerVisible.value = true
}

async function load() {
  loading.value = true
  try {
    list.value = await apiAdminBuildings()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    code: '',
    name: '',
    posX: 0,
    posY: 0,
    width: 60,
    height: 36,
    floors: 6,
    floorHeight: 3.5,
    roomsPerFloor: 20,
  })
  dialogVisible.value = true
}

function openEdit(row: AdminBuilding) {
  editingId.value = row.id
  Object.assign(form, {
    code: row.code,
    name: row.name,
    posX: row.posX,
    posY: row.posY,
    width: row.width,
    height: row.height,
    floors: row.floors,
    floorHeight: row.floorHeight,
    roomsPerFloor: row.roomsPerFloor,
  })
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
      await apiUpdateBuilding(editingId.value, { ...form })
    } else {
      await apiCreateBuilding({ ...form })
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

async function remove(row: AdminBuilding) {
  try {
    await ElMessageBox.confirm(
      `确认删除楼栋 ${row.code} ${row.name}？存在工单记录的楼栋不可删除。`,
      '删除确认',
    )
  } catch {
    return
  }
  try {
    await apiDeleteBuilding(row.id)
    ElMessage.success('已删除')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

onMounted(load)
</script>

<style scoped>
.layout-tag {
  padding: 2px 9px;
  color: #5a6a85;
  font-size: 12px;
  background: var(--rv-grad-8);
  border-radius: 999px;
}

.layout-tag.custom {
  color: #2462d9;
  background: var(--rv-grad-1);
}

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
.inline-pair {
  display: flex;
  gap: 8px;
  width: 100%;
}
.empty {
  padding: 24px 0;
}
.pos-empty {
  color: #94a3b8;
  font-size: 12px;
}
</style>