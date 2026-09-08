import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import FaultTypesView from '../views/FaultTypesView.vue'
import DispatchRulesView from '../views/DispatchRulesView.vue'
import UsersView from '../views/UsersView.vue'
import BuildingsView from '../views/BuildingsView.vue'
import LogsView from '../views/LogsView.vue'
import StatsView from '../views/StatsView.vue'
import PermissionsView from '../views/PermissionsView.vue'
import BuildingVisualView from '../views/BuildingVisualView.vue'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/', redirect: '/orders' },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/orders', name: 'orders', component: HomeView, meta: { requiresAuth: true } },
    { path: '/fault-types', name: 'fault-types', component: FaultTypesView, meta: { requiresAuth: true } },
    { path: '/dispatch-rules', name: 'dispatch-rules', component: DispatchRulesView, meta: { requiresAuth: true } },
    { path: '/users', name: 'users', component: UsersView, meta: { requiresAuth: true } },
    { path: '/buildings', name: 'buildings', component: BuildingsView, meta: { requiresAuth: true } },
    { path: '/logs', name: 'logs', component: LogsView, meta: { requiresAuth: true } },
    { path: '/stats', name: 'stats', component: StatsView, meta: { requiresAuth: true } },
    { path: '/permissions', name: 'permissions', component: PermissionsView, meta: { requiresAuth: true } },
    { path: '/building-visual', name: 'building-visual', component: BuildingVisualView, meta: { requiresAuth: true } },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  const requiresAuth = to.meta.requiresAuth as boolean | undefined
  if (requiresAuth && !auth.token) {
    return { path: '/login' }
  }
  if (to.path === '/login' && auth.token) {
    return { path: '/orders' }
  }
  return true
})

export default router