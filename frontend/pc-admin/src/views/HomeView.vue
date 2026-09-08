<template>
  <el-container class="admin-layout">
    <el-header class="admin-header">
      <span class="header-title">RepairVision · 工单总览（只读）</span>
      <span class="header-user">
        {{ auth.user?.name }}
        <el-button size="small" type="danger" plain @click="logout">退出</el-button>
      </span>
    </el-header>
    <el-main>
      <el-card class="filter-card" shadow="never">
        <div class="filter-row">
          <el-select v-model="query.status" placeholder="全部状态" clearable style="width: 160px">
            <el-option label="待派单" :value="1" />
            <el-option label="已派单" :value="2" />
            <el-option label="维修中" :value="3" />
            <el-option label="已完成" :value="4" />
            <el-option label="已取消" :value="5" />
          </el-select>
          <el-input-number
            v-model="query.buildingId"
            :min="0"
            placeholder="楼栋ID"
            style="width: 160px"
          />
          <el-button type="primary" @click="load">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </div>
      </el-card>

      <el-card shadow="never">
        <el-table :data="orders" v-loading="loading" border stripe>
          <el-table-column prop="orderNo" label="工单号" width="180" />
          <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
          <el-table-column label="位置" width="150">
            <template #default="{ row }">
              {{ row.buildingName }} {{ row.floor }}F-{{ row.room }}
            </template>
          </el-table-column>
          <el-table-column prop="faultTypeName" label="类型" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)" size="small">{{ row.statusText }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="workerName" label="维修工人" width="110" />
          <el-table-column prop="reporterName" label="报修宿管" width="120" />
          <el-table-column prop="createdAt" label="创建时间" width="170" />
        </el-table>
        <el-empty v-if="!loading && orders.length === 0" description="暂无数据" />
      </el-card>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { OrderItem } from '../api'
import { apiAdminOrders } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const orders = ref<OrderItem[]>([])
const loading = ref(false)
const query = reactive({ status: 0, buildingId: 0 })

async function load() {
  loading.value = true
  try {
    orders.value = await apiAdminOrders(query.status, query.buildingId)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reset() {
  query.status = 0
  query.buildingId = 0
  load()
}

function statusType(status: number) {
  if (status === 4) return 'success'
  if (status === 5) return 'info'
  if (status === 3) return 'warning'
  return 'primary'
}

function logout() {
  auth.logout()
  router.replace('/login')
}

onMounted(load)
</script>

<style scoped>
.admin-layout {
  min-height: 100vh;
}
.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #eee;
}
.header-title {
  font-size: 18px;
  font-weight: 600;
}
.header-user {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #606266;
}
.filter-card {
  margin-bottom: 12px;
}
.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>