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
      <router-link class="nav-item" to="/logs" active-class="active">报修记录</router-link>
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
  color: var(--pc-sub);
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-right: 1px solid rgba(255, 255, 255, 0.65);
  box-shadow: 6px 0 26px rgba(46, 68, 112, 0.06);
}

.side-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px 22px;
  border-bottom: 1px solid rgba(120, 145, 190, 0.18);
}

.side-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  color: #fff;
  font-size: 19px;
  font-weight: 800;
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  border-radius: 12px;
  box-shadow: 0 8px 18px rgba(52, 120, 246, 0.28);
}

.side-name {
  color: var(--pc-text);
  font-size: 15px;
  font-weight: 700;
}

.side-sub {
  margin-top: 2px;
  color: var(--pc-light);
  font-size: 11px;
}

.nav-label {
  margin: 22px 10px 8px;
  color: var(--pc-light);
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
  margin: 3px 0;
  color: var(--pc-sub);
  font-size: 14px;
  text-align: left;
  text-decoration: none;
  background: transparent;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.25s cubic-bezier(0.22, 1, 0.36, 1), color 0.25s ease, transform 0.25s ease;
}

.nav-item:hover {
  color: var(--pc-primary-deep);
  background: rgba(255, 255, 255, 0.7);
  transform: translateX(2px);
}

.nav-item.active {
  color: #fff;
  font-weight: 700;
  background: linear-gradient(135deg, #7fb2ff, #3478f6);
  box-shadow: 0 10px 20px rgba(52, 120, 246, 0.28);
}

.nav-item.disabled {
  color: #a9b4c8;
  cursor: not-allowed;
}

.nav-item em {
  margin-left: auto;
  padding: 1px 7px;
  color: var(--pc-primary-deep);
  font-size: 10px;
  font-style: normal;
  background: rgba(52, 120, 246, 0.12);
  border-radius: 999px;
}

.side-footer {
  margin-top: auto;
  padding: 16px 8px 6px;
  color: var(--pc-light);
  font-size: 11px;
}

.main-area {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 28px 16px;
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.6);
}

.page-title {
  margin: 0;
  font-size: 21px;
  font-weight: 800;
  background: linear-gradient(100deg, #3478f6, #22b573 60%, #a06ae8);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.page-sub {
  margin: 5px 0 0;
  color: var(--pc-light);
  font-size: 13px;
}

.user-box {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-name {
  color: var(--pc-sub);
  font-weight: 600;
}

.logout-btn {
  padding: 7px 14px;
  color: #b34568;
  font-size: 13px;
  background: var(--rv-grad-3);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 10px;
  cursor: pointer;
  transition: transform 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}

.logout-btn:hover {
  transform: translateY(-1px);
}

.content {
  padding: 22px 28px 36px;
  /* 只做淡入、不使用 transform：避免 .content 变成 position:fixed 弹层的包含块与层叠上下文 */
  animation: rv-content-in 0.42s cubic-bezier(0.22, 1, 0.36, 1) both;
}

@keyframes rv-content-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>