<template>
  <div class="login-page">
    <van-nav-bar title="RepairVision 报修端" />
    <van-form @submit="onSubmit">
      <van-cell-group inset>
        <van-field
          v-model="form.username"
          name="username"
          label="账号"
          placeholder="请输入账号"
          :rules="[{ required: true, message: '请输入账号' }]"
        />
        <van-field
          v-model="form.password"
          type="password"
          name="password"
          label="密码"
          placeholder="请输入密码"
          :rules="[{ required: true, message: '请输入密码' }]"
        />
      </van-cell-group>
      <div class="submit-wrap">
        <van-button round block type="primary" native-type="submit" :loading="loading">
          登录
        </van-button>
      </div>
    </van-form>
    <van-cell-group inset class="demo-tips">
      <van-cell title="宿管演示" value="dorm1 / admin123" />
      <van-cell title="工人演示" value="worker1 / admin123" />
      <van-cell title="管理员演示" value="admin / admin123" />
    </van-cell-group>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { apiLogin } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function onSubmit() {
  loading.value = true
  try {
    const data = await apiLogin(form.username, form.password)
    auth.setAuth(data.token, data.user)
    showToast('登录成功')
    router.replace('/')
  } catch (err) {
    showToast((err as Error).message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: #f7f8fa;
}
.submit-wrap {
  margin: 24px 16px;
}
.demo-tips {
  margin-top: 8px;
}
</style>