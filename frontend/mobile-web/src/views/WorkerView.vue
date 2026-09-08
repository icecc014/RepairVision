<template>
  <div>
    <van-nav-bar title="工人 · 我的工单" right-text="退出" @click-right="emit('logout')" />
    <div class="toolbar">
      <van-button size="small" icon="replay" @click="load">刷新</van-button>
    </div>

    <van-empty v-if="!loading && orders.length === 0" description="暂无分配给我的工单" />
    <van-card
      v-for="item in orders"
      :key="item.id"
      class="order-card"
      :title="item.title"
      :desc="`${item.buildingName || ''} ${item.floor}层 ${item.room}室 · ${item.createdAt}`"
    >
      <template #tags>
        <van-tag :type="statusType(item.status)" plain>{{ item.statusText }}</van-tag>
        <van-tag plain>{{ item.faultTypeName }}</van-tag>
      </template>
      <template #footer>
        <van-button
          v-if="item.status === 2"
          size="mini"
          type="primary"
          :loading="actingId === item.id"
          @click="start(item)"
        >
          开工
        </van-button>
        <van-button
          v-if="item.status === 3"
          size="mini"
          type="success"
          :loading="actingId === item.id"
          @click="complete(item)"
        >
          完工
        </van-button>
      </template>
    </van-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { OrderItem } from '../api'
import { apiCompleteOrder, apiStartOrder, apiWorkerOrders } from '../api'

const emit = defineEmits<{ (e: 'logout'): void }>()

const orders = ref<OrderItem[]>([])
const loading = ref(false)
const actingId = ref<number | null>(null)

async function load() {
  loading.value = true
  try {
    orders.value = await apiWorkerOrders(0)
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function start(item: OrderItem) {
  try {
    await showConfirmDialog({ title: '开工', message: `确认开始维修 ${item.orderNo}？` })
  } catch {
    return
  }
  actingId.value = item.id
  try {
    await apiStartOrder(item.id)
    showToast('已开工')
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    actingId.value = null
  }
}

async function complete(item: OrderItem) {
  try {
    await showConfirmDialog({ title: '完工', message: `确认完成 ${item.orderNo}？` })
  } catch {
    return
  }
  actingId.value = item.id
  try {
    await apiCompleteOrder(item.id)
    showToast('已完工')
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    actingId.value = null
  }
}

function statusType(status: number) {
  if (status === 4) return 'success'
  if (status === 5) return 'default'
  if (status === 3) return 'primary'
  return 'warning'
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  padding: 8px 12px;
}
.order-card {
  margin: 8px 12px;
}
</style>