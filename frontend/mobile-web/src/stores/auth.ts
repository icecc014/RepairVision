import { defineStore } from 'pinia'
import type { UserInfo } from '../api'

function readUser(): UserInfo | null {
  try {
    const raw = localStorage.getItem('rv_user')
    return raw ? (JSON.parse(raw) as UserInfo) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('rv_token') || '',
    user: readUser(),
  }),
  actions: {
    setAuth(token: string, user: UserInfo) {
      this.token = token
      this.user = user
      localStorage.setItem('rv_token', token)
      localStorage.setItem('rv_user', JSON.stringify(user))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('rv_token')
      localStorage.removeItem('rv_user')
    },
  },
})