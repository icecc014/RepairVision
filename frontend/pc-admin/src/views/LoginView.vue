<template>
  <div class="login-shell">
    <aside class="brand-panel">
      <div class="brand-logo">修</div>
      <h1>RepairVision</h1>
      <p class="brand-slogan">校园维修工单智能调度管理平台</p>
      <ul class="feature-list">
        <li>智能派单 · 工单全流程跟踪</li>
        <li>宿管 / 工人 / 管理员三端协同</li>
        <li>2D / 3D 楼宇可视化</li>
      </ul>
      <p class="brand-footer">毕业设计演示系统</p>
    </aside>

    <main class="form-panel">
      <div class="form-card">
        <h2 class="form-title">管理员登录</h2>
        <p class="form-subtitle">请使用管理员账号进入管理控制台</p>
        <el-form :model="form" label-position="top" @submit.prevent>
          <el-form-item label="账号">
            <el-input
              v-model="form.username"
              size="large"
              placeholder="请输入管理员账号"
              autocomplete="username"
            />
          </el-form-item>
          <el-form-item label="密码">
            <el-input
              v-model="form.password"
              type="password"
              size="large"
              show-password
              placeholder="请输入密码"
              autocomplete="current-password"
              @keyup.enter="onSubmit"
            />
          </el-form-item>
          <el-button
            type="primary"
            size="large"
            class="login-btn"
            :loading="loading"
            @click="onSubmit"
          >
            登 录
          </el-button>
        </el-form>
        <el-alert title="演示账号：admin / admin123" type="info" :closable="false" class="demo-alert" />
      </div>
    </main>
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
    if (data.user.role !== 1) {
      auth.logout()
      ElMessage.error('请使用管理员账号登录 PC 端')
      return
    }
    auth.setAuth(data)
    ElMessage.success('登录成功')
    router.replace('/orders')
  } catch (err) {
    ElMessage.error((err as Error).message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-shell {
  display: grid;
  grid-template-columns: minmax(380px, 1.05fr) minmax(420px, 0.95fr);
  min-height: 100vh;
  background: #fff;
}

.brand-panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 60px 9%;
  color: #fff;
  background: radial-gradient(1200px 600px at 20% 10%, #2f62e8 0%, #1d4ed8 45%, #102a6b 100%);
}

.brand-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 70px;
  height: 70px;
  margin-bottom: 24px;
  color: #2563eb;
  font-size: 34px;
  font-weight: 800;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.2);
}

.brand-panel h1 {
  margin: 0;
  font-size: 34px;
  font-weight: 800;
  letter-spacing: 0.5px;
}

.brand-slogan {
  margin: 10px 0 36px;
  color: rgba(255, 255, 255, 0.88);
  font-size: 15px;
}

.feature-list {
  margin: 0;
  padding: 0;
  list-style: none;
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  line-height: 2.2;
}

.feature-list li::before {
  content: '✓ ';
  font-weight: 700;
}

.brand-footer {
  margin-top: 50px;
  color: rgba(255, 255, 255, 0.55);
  font-size: 12px;
}

.form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 30px;
  background: #f8fafc;
}

.form-card {
  width: 100%;
  max-width: 420px;
  padding: 40px 38px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 18px;
  box-shadow: 0 18px 50px rgba(15, 23, 42, 0.08);
}

.form-title {
  margin: 0;
  color: #1e293b;
  font-size: 24px;
  font-weight: 800;
}

.form-subtitle {
  margin: 8px 0 26px;
  color: #94a3b8;
  font-size: 13px;
}

.login-btn {
  width: 100%;
  margin-top: 4px;
  font-weight: 700;
  letter-spacing: 2px;
}

.demo-alert {
  margin-top: 22px;
}

@media (max-width: 860px) {
  .login-shell {
    display: block;
  }

  .brand-panel {
    display: none;
  }
}
</style>