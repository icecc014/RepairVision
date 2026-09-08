import axios from 'axios'

function readToken(): string {
  try {
    const raw = localStorage.getItem('rv_admin_auth')
    if (raw) {
      const auth = JSON.parse(raw) as { token?: string }
      if (auth.token) return auth.token
    }
  } catch {
    // ignore parse errors
  }
  return localStorage.getItem('rv_admin_token') || ''
}

const http = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

http.interceptors.request.use((config) => {
  const token = readToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (resp) => {
    const body = resp.data
    if (body && typeof body === 'object' && 'code' in body && body.code !== 0) {
      return Promise.reject(new Error(body.msg || '请求失败'))
    }
    return body && typeof body === 'object' && 'data' in body ? body.data : body
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('rv_admin_auth')
      localStorage.removeItem('rv_admin_token')
      if (!location.pathname.startsWith('/login')) {
        location.href = '/admin/login'
      }
    }
    const msg = error.response?.data?.msg || error.message || '网络异常'
    return Promise.reject(new Error(msg))
  },
)

export default http