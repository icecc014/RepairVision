import { defineStore } from 'pinia'
import type { LoginResult } from '../api'

function readAuth(): LoginResult | null {
  try {
    const raw = localStorage.getItem('rv_admin_auth')
    return raw ? (JSON.parse(raw) as LoginResult) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('adminAuth', {
  state: () => readAuth() || { token: '', user: null },
  actions: {
    setAuth(auth: LoginResult) {
      this.token = auth.token
      this.user = auth.user
      localStorage.setItem('rv_admin_auth', JSON.stringify(auth))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('rv_admin_auth')
    },
  },
})