<template>
  <div class="login-page">
    <el-card class="login-card">
      <template #header>
        <div class="card-title">RepairVision PC 管理端</div>
      </template>
      <el-form :model="form" label-width="64px" @submit.prevent>
        <el-form-item label="账号">
          <el-input v-model="form.username" placeholder="请输入管理员账号" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" class="login-btn" @click="onSubmit">
            登录
          </el-button>
        </el-form-item>
      </el-form>
      <el-alert title="演示账号：admin / admin123" type="info" :closable="false" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { apiLogin } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function onSubmit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入账号和密码')
    return
  }
  loading.value = true
  try {
    const data = await apiLogin(form.username, form.password)
    auth.setAuth(data)
    if (data.user.role !== 1) {
      auth.logout()
      ElMessage.error('请使用管理员账号登录PC端')
      return
    }
    ElMessage.success('登录成功')
    router.replace('/')
  } catch (err) {
    ElMessage.error((err as Error).message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
}
.login-card {
  width: 420px;
}
.card-title {
  text-align: center;
  font-weight: 600;
}
.login-btn {
  width: 100%;
}
</style>