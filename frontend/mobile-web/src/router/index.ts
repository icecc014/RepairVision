import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'

const router = createRouter({
  history: createWebHistory('/m/'),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/dorm', name: 'dorm', component: HomeView, meta: { requiresAuth: true, role: 3 } },
    { path: '/worker', name: 'worker', component: HomeView, meta: { requiresAuth: true, role: 2 } },
  ],
})

function homePath(auth: ReturnType<typeof useAuthStore>) {
  if (auth.user?.role === 3) return '/dorm'
  if (auth.user?.role === 2) return '/worker'
  return '/login'
}

router.beforeEach((to) => {
  const auth = useAuthStore()
  const requiresAuth = to.meta.requiresAuth as boolean | undefined
  const requiredRole = to.meta.role as number | undefined
  if (requiresAuth && !auth.token) {
    return { path: '/login' }
  }
  if (requiredRole && auth.user?.role !== requiredRole) {
    return { path: '/login' }
  }
  if (to.path === '/login' && auth.token && auth.user) {
    return homePath(auth)
  }
  return true
})

export default router