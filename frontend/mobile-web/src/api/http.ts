import axios from 'axios'

const http = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// V9 统一登录：优先本标签页登录态（sessionStorage），回退统一登录页写入的共享登录态（localStorage）
function readToken(): string {
  return sessionStorage.getItem('rv_token') || localStorage.getItem('rv_token') || ''
}

// V9 统一登录：清理本标签页与共享登录态。仅清共享键，不影响 PC 管理端自身的 rv_admin_* 登录态
function clearAuth(): void {
  sessionStorage.removeItem('rv_token')
  sessionStorage.removeItem('rv_user')
  localStorage.removeItem('rv_token')
  localStorage.removeItem('rv_user')
  localStorage.removeItem('rv_role')
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
      clearAuth()
      // 移动端登录页路径是 /m/login，用 endsWith 判断才有效（原 startsWith('/login') 恒为真）
      if (!location.pathname.endsWith('/login')) {
        // V9 统一登录：登录态失效回到统一登录页
        location.href = '/login/'
      }
    }
    const msg = error.response?.data?.msg || error.message || '网络异常'
    return Promise.reject(new Error(msg))
  },
)

export default http