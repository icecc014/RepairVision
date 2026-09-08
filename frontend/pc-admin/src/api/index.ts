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

export interface AdminUser {
  id: number
  username: string
  name: string
  phone: string
  role: number
  roleText: string
  buildingId?: number
  status: number
  statusText: string
  buildingIds?: number[]
  buildings?: string[]
}

export interface AdminBuilding {
  id: number
  code: string
  name: string
  posX: number
  posY: number
  width: number
  height: number
  floors: number
  floorHeight: number
  roomsPerFloor: number
}

export function apiAdminUsers(params: { role?: number; status?: number; keyword?: string }): Promise<AdminUser[]> {
  return http.get('/admin/users', { params: { role: params.role || undefined, status: params.status || undefined, keyword: params.keyword || undefined } })
}

export function apiCreateUser(payload: {
  username: string
  password: string
  name: string
  phone: string
  role: number
  buildingId?: number
  buildingIds?: number[]
}): Promise<unknown> {
  return http.post('/admin/users', payload)
}

export function apiUpdateUser(
  id: number,
  payload: {
    name: string
    phone: string
    role: number
    status: number
    buildingId?: number
    buildingIds?: number[]
  },
): Promise<unknown> {
  return http.put(`/admin/users/${id}`, payload)
}

export function apiResetPassword(id: number, password: string): Promise<unknown> {
  return http.post(`/admin/users/${id}/reset-password`, { password })
}

export function apiDeleteUser(id: number): Promise<unknown> {
  return http.delete(`/admin/users/${id}`)
}

export function apiAdminBuildings(): Promise<AdminBuilding[]> {
  return http.get('/admin/buildings')
}

export function apiCreateBuilding(payload: Partial<AdminBuilding>): Promise<unknown> {
  return http.post('/admin/buildings', payload)
}

export function apiUpdateBuilding(id: number, payload: Partial<AdminBuilding>): Promise<unknown> {
  return http.put(`/admin/buildings/${id}`, payload)
}

export function apiDeleteBuilding(id: number): Promise<unknown> {
  return http.delete(`/admin/buildings/${id}`)
}

export interface OperationLogItem {
  id: number
  userId?: number
  username?: string
  role?: number
  module: string
  action: string
  method: string
  path: string
  requestBody: string
  responseCode: number
  ip: string
  costMs: number
  createdAt: string
}

export interface LogPage {
  total: number
  list: OperationLogItem[]
}

export function apiAdminLogs(params: { page: number; size: number; module?: string; action?: string; keyword?: string }): Promise<LogPage> {
  return http.get('/admin/logs', {
    params: {
      page: params.page,
      size: params.size,
      module: params.module || undefined,
      action: params.action || undefined,
      keyword: params.keyword || undefined,
    },
  })
}