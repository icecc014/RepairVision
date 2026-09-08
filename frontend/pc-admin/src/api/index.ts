import http from './http'

export interface LoginResult {
  token: string
  user: {
    id: number
    username: string
    name: string
    role: number
    buildingId?: number
  }
}

export interface OrderItem {
  id: number
  orderNo: string
  title: string
  description: string
  buildingId: number
  buildingName: string
  room: string
  floor: number
  faultType: string
  faultTypeName: string
  status: number
  statusText: string
  workerId?: number
  workerName?: string
  reporterId: number
  reporterName?: string
  source: string
  createdAt: string
  updatedAt: string
}

export function apiLogin(username: string, password: string): Promise<LoginResult> {
  return http.post('/login', { username, password })
}

export function apiAdminOrders(status = 0, buildingId = 0): Promise<OrderItem[]> {
  return http.get('/admin/orders', {
    params: { status: status || undefined, buildingId: buildingId || undefined },
  })
}

export interface AdminFaultType {
  id: number
  code: string
  name: string
  sort: number
  status: number
}

export interface DispatchRule {
  id: number
  ruleKey: string
  ruleValue: string
  enabled: number
  remark: string
  updatedAt: string
}

export function apiAdminFaultTypes(): Promise<AdminFaultType[]> {
  return http.get('/admin/fault-types')
}

export function apiCreateFaultType(payload: { code: string; name: string; sort: number }): Promise<unknown> {
  return http.post('/admin/fault-types', payload)
}

export function apiUpdateFaultType(
  id: number,
  payload: { name: string; sort: number; status: number },
): Promise<unknown> {
  return http.put(`/admin/fault-types/${id}`, payload)
}

export function apiDeleteFaultType(id: number): Promise<unknown> {
  return http.delete(`/admin/fault-types/${id}`)
}

export function apiDispatchRules(): Promise<DispatchRule[]> {
  return http.get('/admin/dispatch-rules')
}

export function apiCreateDispatchRule(payload: {
  ruleKey: string
  ruleValue: string
  enabled: number
  remark: string
}): Promise<unknown> {
  return http.post('/admin/dispatch-rules', payload)
}

export function apiUpdateDispatchRule(
  id: number,
  payload: { ruleValue: string; enabled: number; remark: string },
): Promise<unknown> {
  return http.put(`/admin/dispatch-rules/${id}`, payload)
}