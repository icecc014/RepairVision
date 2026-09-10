<template>
  <button class="rv-bell" @click="openCenter">
    🔔
    <span v-if="unread > 0" class="rv-bell-badge">{{ unread > 99 ? '99+' : unread }}</span>
  </button>

  <van-popup v-model:show="visible" position="right" :style="{ width: '86%', height: '100%' }">
    <div class="nc-head">
      <div class="nc-title">消息中心</div>
      <div class="nc-actions">
        <button class="nc-read-all" :disabled="unread === 0" @click="readAll">全部已读</button>
        <button class="nc-close" @click="visible = false">✕</button>
      </div>
    </div>
    <div v-if="list.length === 0" class="nc-empty">暂无消息</div>
    <div v-else class="nc-list">
      <article
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
      </article>
    </div>
  </van-popup>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { showToast } from 'vant'
import type { NotificationItem } from '../api'
import { apiNotificationRead, apiNotificationReadAll, apiNotifications } from '../api'

const visible = ref(false)
const unread = ref(0)
const list = ref<NotificationItem[]>([])

async function load() {
  try {
    const res = await apiNotifications(1, 30)
    unread.value = res.unread
    list.value = res.list
  } catch {
    // 通知加载失败不影响主流程
  }
}

function openCenter() {
  visible.value = true
  load()
}

async function markRead(item: NotificationItem) {
  if (item.isRead === 1) return
  try {
    await apiNotificationRead(item.id)
    item.isRead = 1
    unread.value = Math.max(0, unread.value - 1)
  } catch (err) {
    showToast((err as Error).message)
  }
}

async function readAll() {
  try {
    await apiNotificationReadAll()
    unread.value = 0
    list.value = list.value.map((x) => ({ ...x, isRead: 1 }))
  } catch (err) {
    showToast((err as Error).message)
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
.rv-bell {
  position: relative;
  padding: 6px 10px;
  font-size: 18px;
  background: #f1f5f9;
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.rv-bell-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 18px;
  padding: 0 4px;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 18px;
  text-align: center;
  background: #ef4444;
  border-radius: 999px;
}

.nc-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 16px 12px;
  border-bottom: 1px solid #eef2f7;
}

.nc-title {
  font-size: 17px;
  font-weight: 800;
}

.nc-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nc-read-all {
  padding: 4px 10px;
  color: #2563eb;
  font-size: 12px;
  background: #eff6ff;
  border: none;
  border-radius: 999px;
}

.nc-close {
  width: 26px;
  height: 26px;
  color: #64748b;
  background: #f1f5f9;
  border: none;
  border-radius: 50%;
}

.nc-list {
  padding: 8px 12px 24px;
}

.nc-item {
  padding: 12px;
  margin-bottom: 8px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
}

.nc-item.unread {
  border-color: #bfdbfe;
  background: #f8fbff;
}

.nc-item-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.nc-item-title {
  font-size: 14px;
  font-weight: 700;
}

.nc-item-time {
  color: #94a3b8;
  font-size: 11px;
}

.nc-item-content {
  margin-top: 6px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.5;
}

.nc-empty {
  padding: 60px 0;
  color: #94a3b8;
  text-align: center;
}
</style>
