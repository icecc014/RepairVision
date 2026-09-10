<template>
  <div class="admin-shell">
    <aside class="sidebar">
      <div class="side-brand">
        <span class="side-logo">修</span>
        <div>
          <div class="side-name">RepairVision</div>
          <div class="side-sub">维修调度管理平台</div>
        </div>
      </div>

      <div class="nav-label">管理导航</div>
      <router-link class="nav-item" to="/orders" active-class="active">工单总览</router-link>
      <router-link class="nav-item" to="/dispatch-board" active-class="active">调度看板</router-link>
      <router-link class="nav-item" to="/stats" active-class="active">数据统计看板</router-link>
      <router-link class="nav-item" to="/fault-types" active-class="active">维修类型字典</router-link>
      <router-link class="nav-item" to="/dispatch-rules" active-class="active">派单规则配置</router-link>
      <router-link class="nav-item" to="/users" active-class="active">人员账号管理</router-link>
      <router-link class="nav-item" to="/schedules" active-class="active">工人排班</router-link>
      <router-link class="nav-item" to="/leaves" active-class="active">请假审批</router-link>
      <router-link class="nav-item" to="/buildings" active-class="active">建筑信息管理</router-link>
      <router-link class="nav-item" to="/logs" active-class="active">操作日志</router-link>
      <router-link class="nav-item" to="/permissions" active-class="active">权限说明</router-link>
      <router-link class="nav-item" to="/building-visual" active-class="active">建筑可视化</router-link>

      <div class="side-footer">校园维修 · 毕业设计</div>
    </aside>

    <section class="main-area">
      <header class="topbar">
        <div>
          <h2 class="page-title">{{ title }}</h2>
          <p class="page-sub">{{ subtitle }}</p>
        </div>
        <div class="user-box">
          <NotificationBell />
          <el-tag type="primary" effect="dark" size="small">管理员</el-tag>
          <span class="user-name">{{ auth.user?.name }}</span>
          <button class="logout-btn" @click="logout">退出登录</button>
        </div>
      </header>
      <main class="content">
        <slot />
      </main>
    </section>
  </div>
</template>

<script setup lang="ts">
import NotificationBell from './NotificationBell.vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

defineProps<{ title: string; subtitle?: string }>()

const router = useRouter()
const auth = useAuthStore()

function logout() {
  auth.logout()
  router.replace('/login')
}
</script>

<style scoped>
.admin-shell {
  display: grid;
  grid-template-columns: 232px 1fr;
  min-height: 100vh;
}

.sidebar {
  display: flex;
  flex-direction: column;
  padding: 20px 14px;
  color: #cbd5e1;
  background: linear-gradient(180deg, #102a6b 0%, #0f2557 55%, #0b1d47 100%);
}

.side-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px 22px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.09);
}

.side-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  color: #2563eb;
  font-size: 19px;
  font-weight: 800;
  background: #fff;
  border-radius: 11px;
}

.side-name {
  color: #fff;
  font-size: 15px;
  font-weight: 700;
}

.side-sub {
  margin-top: 2px;
  color: #8ea3cf;
  font-size: 11px;
}

.nav-label {
  margin: 22px 10px 8px;
  color: #6b82b8;
  font-size: 11px;
  letter-spacing: 1px;
}

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 11px 14px;
  margin: 2px 0;
  color: #b8c7e8;
  font-size: 14px;
  text-align: left;
  text-decoration: none;
  background: transparent;
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.nav-item.active {
  color: #fff;
  font-weight: 700;
  background: linear-gradient(90deg, rgba(37, 99, 235, 0.45), rgba(37, 99, 235, 0.08));
  box-shadow: inset 3px 0 0 #60a5fa;
}

.nav-item.disabled {
  color: #6479ad;
  cursor: not-allowed;
}

.nav-item em {
  margin-left: auto;
  padding: 1px 7px;
  color: #93b4ff;
  font-size: 10px;
  font-style: normal;
  background: rgba(37, 99, 235, 0.22);
  border-radius: 999px;
}

.side-footer {
  margin-top: auto;
  padding: 16px 8px 6px;
  color: #5d74aa;
  font-size: 11px;
}

.main-area {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 22px 28px 16px;
  background: #fff;
  border-bottom: 1px solid #eef2f7;
}

.page-title {
  margin: 0;
  font-size: 21px;
  font-weight: 800;
  color: #1e293b;
}

.page-sub {
  margin: 5px 0 0;
  color: #94a3b8;
  font-size: 13px;
}

.user-box {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-name {
  color: #475569;
  font-weight: 600;
}

.logout-btn {
  padding: 7px 14px;
  color: #dc2626;
  font-size: 13px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  cursor: pointer;
}

.content {
  padding: 22px 28px 36px;
}
</style>