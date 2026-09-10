<template>
  <el-popover placement="bottom-end" :width="420" trigger="click" @show="load">
    <template #reference>
      <el-badge :value="unread" :hidden="unread === 0" :max="99">
        <button class="bell-btn">🔔 消息</button>
      </el-badge>
    </template>

    <div class="nc-head">
      <span class="nc-title">消息中心</span>
      <el-button link type="primary" size="small" :disabled="unread === 0" @click="readAll">全部已读</el-button>
    </div>
    <div v-if="list.length === 0" class="nc-empty">暂无消息</div>
    <div v-else class="nc-list">
      <div
        v-for="item in list"
        :key="item.id"
        class="nc-item"
        :class="{ unread: item.isRead === 0 }"
        @click="markRead(item)"
      >
        <div class="nc-item-top">
          <span class="nc-item-title">{{ item.title }}</span>
          <span class="nc-item-time">{{ item.createdAt.slice(5, 16) }}</span>
        </div>
        <div class="nc-item-content">{{ item.content }}</div>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { NotificationItem } from '../api'
import { apiNotificationRead, apiNotificationReadAll, apiNotifications } from '../api'

const unread = ref(0)
const list = ref<NotificationItem[]>([])

async function load() {
  try {
    const res = await apiNotifications(1, 30)
    unread.value = res.unread
    list.value = res.list
  } catch {
    // ignore
  }
}

async function markRead(item: NotificationItem) {
  if (item.isRead === 1) return
  try {
    await apiNotificationRead(item.id)
    item.isRead = 1
    unread.value = Math.max(0, unread.value - 1)
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function readAll() {
  try {
    await apiNotificationReadAll()
    unread.value = 0
    list.value = list.value.map((x) => ({ ...x, isRead: 1 }))
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function onRefresh() {
  load()
}

onMounted(() => {
  load()
  window.addEventListener('rv-notify-refresh', onRefresh)
})

onUnmounted(() => {
  window.removeEventListener('rv-notify-refresh', onRefresh)
})
</script>

<style scoped>
.bell-btn {
  padding: 6px 12px;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
  background: #f1f5f9;
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.nc-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid #eef2f7;
}

.nc-title {
  font-weight: 800;
}

.nc-list {
  max-height: 420px;
  overflow: auto;
}

.nc-item {
  padding: 10px 6px;
  border-bottom: 1px solid #f1f5f9;
  cursor: pointer;
}

.nc-item.unread {
  background: #f8fbff;
}

.nc-item-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.nc-item-title {
  font-size: 13px;
  font-weight: 700;
}

.nc-item-time {
  color: #94a3b8;
  font-size: 11px;
}

.nc-item-content {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}

.nc-empty {
  padding: 30px 0;
  color: #94a3b8;
  text-align: center;
}
</style>
