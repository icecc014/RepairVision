import axios from 'axios'

const http = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// V9 统一登录：优先本标签页登录态（sessionStorage），回退统一登录页写入的共享登录态（localStorage）
function readToken(): string {
  return sessionStorage.getItem('rv_token') || localStorage.getItem('rv_token') || ''
}

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
    if (error.response?.status === 401 || error.response?.status === 403) {
      sessionStorage.removeItem('rv_token')
      sessionStorage.removeItem('rv_user')
      if (!location.pathname.startsWith('/login')) {
        location.href = '/m/login'
      }
    }
    const msg = error.response?.data?.msg || error.message || '网络异常'
    return Promise.reject(new Error(msg))
  },
)

export default http