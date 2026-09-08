<template>
  <div class="login-page">
    <section class="hero">
      <div class="brand-mark">修</div>
      <h1 class="brand-name">RepairVision</h1>
      <p class="slogan">校园维修工单智能调度系统</p>
      <p class="subline">智能派单 · 进度跟踪 · 可视化楼宇</p>
    </section>

    <section class="login-card">
      <form @submit.prevent="onSubmit">
        <div class="field-group">
          <label class="field-label">登录账号</label>
          <input
            v-model="form.username"
            class="field-input"
            placeholder="请输入宿管/工人工号"
            autocomplete="username"
          />
        </div>
        <div class="field-group">
          <label class="field-label">登录密码</label>
          <input
            v-model="form.password"
            type="password"
            class="field-input"
            placeholder="请输入密码"
            autocomplete="current-password"
          />
        </div>
        <button class="login-btn" type="submit" :disabled="loading">
          {{ loading ? '正在登录…' : '登 录' }}
        </button>
      </form>

      <div class="demo-box">
        <div class="demo-title">演示账号（密码 admin123）</div>
        <div class="demo-row"><span class="dot dorm">宿</span>宿管：dorm1 / dorm2</div>
        <div class="demo-row"><span class="dot worker">工</span>工人：worker1 / worker2</div>
      </div>
    </section>

    <footer class="page-footer">RepairVision · 毕业设计演示</footer>
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
  if (!form.username || !form.password) {
    showToast('请输入账号和密码')
    return
  }
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
  padding-bottom: 30px;
  background: #eef2f7;
}

.hero {
  padding: 72px 30px 90px;
  color: #fff;
  text-align: center;
  background: linear-gradient(150deg, #2563eb 0%, #1e40af 68%, #172d78 100%);
  border-radius: 0 0 36px 36px;
}

.brand-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
  color: #2563eb;
  font-size: 30px;
  font-weight: 800;
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.18);
}

.brand-name {
  margin: 0;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: 0.5px;
}

.slogan {
  margin: 10px 0 0;
  font-size: 15px;
  font-weight: 600;
  opacity: 0.95;
}

.subline {
  margin: 6px 0 0;
  font-size: 12px;
  opacity: 0.72;
}

.login-card {
  position: relative;
  z-index: 2;
  width: calc(100% - 36px);
  margin: -46px auto 0;
  padding: 26px 22px 22px;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.12);
}

.field-group {
  margin-bottom: 16px;
}

.field-label {
  display: block;
  margin-bottom: 8px;
  color: #475569;
  font-size: 13px;
  font-weight: 600;
}

.field-input {
  display: block;
  width: 100%;
  padding: 13px 14px;
  color: #1e293b;
  font-size: 15px;
  background: #f8fafc;
  border: 1.5px solid #e2e8f0;
  border-radius: 12px;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.field-input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.12);
}

.login-btn {
  display: block;
  width: 100%;
  padding: 14px;
  margin-top: 8px;
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 2px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  border: none;
  border-radius: 13px;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.28);
  cursor: pointer;
}

.login-btn:disabled {
  opacity: 0.65;
}

.demo-box {
  margin-top: 22px;
  padding: 14px;
  background: #f8fafc;
  border: 1px dashed #dbe3ef;
  border-radius: 12px;
}

.demo-title {
  margin-bottom: 8px;
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
}

.demo-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 6px 0;
  color: #475569;
  font-size: 13px;
}

.dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  border-radius: 6px;
}

.dot.dorm {
  background: #16a34a;
}

.dot.worker {
  background: #2563eb;
}

.page-footer {
  margin-top: 28px;
  color: #94a3b8;
  font-size: 12px;
  text-align: center;
}
</style>