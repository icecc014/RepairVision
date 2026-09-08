<template>
  <DormView v-if="isDorm" @logout="logout" />
  <WorkerView v-else-if="isWorker" @logout="logout" />
  <div v-else>
    <van-nav-bar title="账号提示" />
    <van-empty description="移动端仅支持宿管与工人工号登录" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import DormView from './DormView.vue'
import WorkerView from './WorkerView.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const isDorm = computed(() => auth.user?.role === 3)
const isWorker = computed(() => auth.user?.role === 2)

function logout() {
  auth.logout()
  router.replace('/login')
}
</script>