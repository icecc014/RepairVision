import { defineStore } from 'pinia'
import type { UserInfo } from '../api'

function readUser(): UserInfo | null {
  try {
    // V9 统一登录：本标签页（sessionStorage）优先，回退 localStorage 中的共享登录态
    const raw = sessionStorage.getItem('rv_user') || localStorage.getItem('rv_user')
    return raw ? (JSON.parse(raw) as UserInfo) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: sessionStorage.getItem('rv_token') || localStorage.getItem('rv_token') || '',
    user: readUser(),
  }),
  actions: {
    setAuth(token: string, user: UserInfo) {
      this.token = token
      this.user = user
      sessionStorage.setItem('rv_token', token)
      sessionStorage.setItem('rv_user', JSON.stringify(user))
    },
    logout() {
      this.token = ''
      this.user = null
      sessionStorage.removeItem('rv_token')
      sessionStorage.removeItem('rv_user')
      // V9 统一登录：同步清理共享登录态（localStorage），避免残留导致重复登录/跳转
      localStorage.removeItem('rv_token')
      localStorage.removeItem('rv_user')
      localStorage.removeItem('rv_role')
    },
  },
})