<template>
  <AdminShell title="操作日志" subtitle="记录管理员/宿管/工人的关键操作（同步写入 order_db）">
    <section class="panel">
      <div class="toolbar">
        <el-select v-model="query.module" placeholder="全部模块" clearable style="width: 140px">
          <el-option label="账号/登录" value="user" />
          <el-option label="楼栋" value="building" />
          <el-option label="维修类型" value="fault" />
          <el-option label="派单规则" value="dispatch" />
          <el-option label="工单" value="order" />
        </el-select>
        <el-select v-model="query.action" placeholder="全部动作" clearable style="width: 140px">
          <el-option label="登录" value="login" />
          <el-option label="创建" value="create" />
          <el-option label="更新" value="update" />
          <el-option label="取消" value="cancel" />
          <el-option label="开工" value="start" />
          <el-option label="完工" value="complete" />
        </el-select>
        <el-input v-model="query.keyword" placeholder="搜索用户/路径" clearable style="width: 220px" @keyup.enter="reload" />
        <el-button type="primary" @click="reload">查询</el-button>
      </div>

      <el-table :data="logs" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="createdAt" label="时间" width="170" />
        <el-table-column prop="username" label="用户" width="110">
          <template #default="{ row }">{{ row.username || '匿名' }}</template>
        </el-table-column>
        <el-table-column prop="module" label="模块" width="100" />
        <el-table-column prop="action" label="动作" width="100" />
        <el-table-column prop="method" label="方法" width="90" />
        <el-table-column prop="path" label="路径" min-width="230" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.responseCode < 400 ? 'success' : 'danger'" size="small">
              {{ row.responseCode }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="costMs" label="耗时(ms)" width="100" />
        <el-table-column prop="ip" label="IP" width="130" />
      </el-table>
      <el-empty v-if="!loading && logs.length === 0" description="暂无日志" class="empty" />
      <el-pagination
        v-if="total > 0"
        class="pager"
        background
        layout="total, prev, pager, next"
        :total="total"
        :page-size="query.size"
        :current-page="query.page"
        @current-change="onPage"
      />
    </section>
  </AdminShell>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { OperationLogItem } from '../api'
import { apiAdminLogs } from '../api'
import AdminShell from '../components/AdminShell.vue'

const logs = ref<OperationLogItem[]>([])
const total = ref(0)
const loading = ref(false)
const query = reactive({ page: 1, size: 20, module: '', action: '', keyword: '' })

async function load() {
  loading.value = true
  try {
    const data = await apiAdminLogs({
      page: query.page,
      size: query.size,
      module: query.module,
      action: query.action,
      keyword: query.keyword,
    })
    logs.value = data.list
    total.value = data.total
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reload() {
  query.page = 1
  load()
}

function onPage(page: number) {
  query.page = page
  load()
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
  gap: 12px;
  margin-bottom: 14px;
}
.empty {
  padding: 24px 0;
}
.pager {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>