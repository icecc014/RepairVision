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
    router.replace(data.user.role === 3 ? '/dorm' : '/worker')
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
  background: linear-gradient(160deg, #e0edfa 0%, #e9e3f8 52%, #f8e6ec 100%);
  background-attachment: fixed;
}

.hero {
  padding: 64px 26px 72px;
  color: var(--rv-text);
  text-align: center;
  animation: rv-fade-up 0.5s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.brand-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
  color: #fff;
  font-size: 30px;
  font-weight: 800;
  background: linear-gradient(135deg, #7fb2ff, #3478f6 60%, #9b8cf0);
  border-radius: 20px;
  box-shadow: 0 14px 30px rgba(52, 120, 246, 0.32);
}

.brand-name {
  margin: 0;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: 0.5px;
  background: linear-gradient(100deg, #3478f6, #22b573 60%, #a06ae8);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.slogan {
  margin: 10px 0 0;
  color: var(--rv-text-sub);
  font-size: 15px;
  font-weight: 600;
}

.subline {
  margin: 6px 0 0;
  color: var(--rv-text-light);
  font-size: 12px;
}

.login-card {
  position: relative;
  z-index: 2;
  width: calc(100% - 36px);
  margin: -34px auto 0;
  padding: 26px 22px 22px;
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 22px;
  box-shadow: 0 18px 42px rgba(46, 68, 112, 0.12);
  animation: rv-fade-up 0.55s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.field-group {
  margin-bottom: 16px;
}

.field-label {
  display: block;
  margin-bottom: 8px;
  color: var(--rv-text-sub);
  font-size: 13px;
  font-weight: 600;
}

.field-input {
  display: block;
  width: 100%;
  padding: 13px 14px;
  color: var(--rv-text);
  font-size: 15px;
  background: rgba(255, 255, 255, 0.55);
  border: 1.5px solid rgba(120, 145, 190, 0.2);
  border-radius: 13px;
  outline: none;
  transition: border-color 0.25s ease, box-shadow 0.25s ease, background 0.25s ease;
}

.field-input:focus {
  background: rgba(255, 255, 255, 0.85);
  border-color: rgba(52, 120, 246, 0.55);
  box-shadow: 0 0 0 4px rgba(52, 120, 246, 0.14);
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
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  border: none;
  border-radius: 14px;
  box-shadow: 0 12px 24px rgba(52, 120, 246, 0.3);
  cursor: pointer;
  transition: transform 0.2s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.2s ease;
}

.login-btn:active:not(:disabled) {
  transform: scale(0.985);
}

.login-btn:disabled {
  opacity: 0.65;
}

.demo-box {
  margin-top: 22px;
  padding: 14px;
  background: rgba(255, 255, 255, 0.5);
  border: 1px dashed rgba(120, 145, 190, 0.35);
  border-radius: 14px;
}

.demo-title {
  margin-bottom: 8px;
  color: var(--rv-text-sub);
  font-size: 12px;
  font-weight: 600;
}

.demo-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 6px 0;
  color: var(--rv-text-sub);
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
  background: linear-gradient(135deg, #7fe6c8, #22b573);
}

.dot.worker {
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
}

.page-footer {
  margin-top: 28px;
  color: var(--rv-text-light);
  font-size: 12px;
  text-align: center;
}
</style>