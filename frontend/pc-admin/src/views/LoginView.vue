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
  background: linear-gradient(120deg, #e0edfa, #e9e3f8, #f8e6ec);
  background-attachment: fixed;
  position: relative;
}

.login-shell::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(36% 46% at 14% 18%, rgba(180, 210, 245, 0.5), transparent 70%),
    radial-gradient(32% 42% at 86% 16%, rgba(226, 210, 250, 0.48), transparent 72%),
    radial-gradient(28% 38% at 52% 6%, rgba(248, 214, 222, 0.4), transparent 70%);
}

.brand-panel {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin: 34px;
  padding: 56px 9%;
  color: var(--pc-text);
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border: 1px solid rgba(255, 255, 255, 0.68);
  border-radius: 28px;
  box-shadow: 0 18px 44px rgba(46, 68, 112, 0.1);
  animation: rv-fade-up 0.5s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.brand-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 70px;
  height: 70px;
  margin-bottom: 24px;
  color: #fff;
  font-size: 34px;
  font-weight: 800;
  background: linear-gradient(135deg, #7fb2ff, #3478f6 60%, #9b8cf0);
  border-radius: 22px;
  box-shadow: 0 14px 30px rgba(52, 120, 246, 0.32);
}

.brand-panel h1 {
  margin: 0;
  font-size: 34px;
  font-weight: 800;
  letter-spacing: 0.5px;
  background: linear-gradient(100deg, #3478f6, #22b573 60%, #a06ae8);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.brand-slogan {
  margin: 10px 0 36px;
  color: var(--pc-sub);
  font-size: 15px;
}

.feature-list {
  margin: 0;
  padding: 0;
  list-style: none;
  color: var(--pc-sub);
  font-size: 14px;
  line-height: 2.2;
}

.feature-list li::before {
  content: '✓ ';
  color: #22b573;
  font-weight: 700;
}

.brand-footer {
  margin-top: 50px;
  color: var(--pc-light);
  font-size: 12px;
}

.form-panel {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 30px;
}

.form-card {
  width: 100%;
  max-width: 420px;
  padding: 40px 38px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 24px;
  box-shadow: 0 20px 48px rgba(46, 68, 112, 0.12);
  animation: rv-fade-up 0.55s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.form-title {
  margin: 0;
  color: var(--pc-text);
  font-size: 24px;
  font-weight: 800;
}

.form-subtitle {
  margin: 8px 0 26px;
  color: var(--pc-light);
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
  background: rgba(255, 255, 255, 0.55);
  border-radius: 12px;
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