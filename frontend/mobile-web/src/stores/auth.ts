import { defineStore } from 'pinia'
import type { UserInfo } from '../api'

function readUser(): UserInfo | null {
  try {
    const raw = sessionStorage.getItem('rv_user')
    return raw ? (JSON.parse(raw) as UserInfo) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: sessionStorage.getItem('rv_token') || '',
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
    },
  },
})