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
  pendingReason?: string
  workerId?: number
  workerName?: string
  workerPhone?: string
  dispatchScore?: number
  skillScore?: number
  distanceScore?: number
  loadScore?: number
  reporterId: number
  reporterName?: string
  source: string
  createdAt: string
  updatedAt: string
}

export function apiLogin(username: string, password: string): Promise<LoginResult> {
  return http.post('/login', { username, password })
}

export interface AdminOrderPage {
  total: number
  list: OrderItem[]
}

export async function apiAdminOrders(status = 0, buildingId = 0, page = 0, size = 0, days = 0): Promise<AdminOrderPage> {
  const res = (await http.get('/admin/orders', {
    params: {
      status: status || undefined,
      buildingId: buildingId || undefined,
      page: page || undefined,
      size: size || undefined,
      days: days || undefined,
    },
  })) as { total: number; list: OrderItem[] }
  return { total: res.total || 0, list: res.list || [] }
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

export async function apiAdminFaultTypes(): Promise<AdminFaultType[]> {
  const res = (await http.get('/admin/fault-types')) as { list: AdminFaultType[] }
  return res.list || []
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

export async function apiDispatchRules(): Promise<DispatchRule[]> {
  const res = (await http.get('/admin/dispatch-rules')) as { list: DispatchRule[] }
  return res.list || []
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
  pendingReason?: string
  buildingIds?: number[]
  maxConcurrent?: number
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
  layoutJson?: string
}

export async function apiAdminUsers(params: { role?: number; status?: number; keyword?: string }): Promise<AdminUser[]> {
  const res = (await http.get('/admin/users', { params: { role: params.role || undefined, status: params.status || undefined, keyword: params.keyword || undefined } })) as { list: AdminUser[] }
  return res.list || []
}

export function apiCreateUser(payload: {
  username: string
  password: string
  name: string
  phone: string
  role: number
  buildingId?: number
  buildingIds?: number[]
  maxConcurrent?: number
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
    maxConcurrent?: number
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

export async function apiAdminBuildings(): Promise<AdminBuilding[]> {
  const res = (await http.get('/admin/buildings')) as { list: AdminBuilding[] }
  return res.list || []
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

export interface RepairRecordItem {
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
  reporterId: number
  reporterName?: string
  workerId?: number
  workerName?: string
  workerPhone?: string
  source: string
  createdAt: string
  dispatchedAt?: string
  startedAt?: string
  completedAt?: string
}

export interface RepairRecordSummary {
  total: number
  pending: number
  working: number
  done: number
  canceled: number
}

export interface RepairRecordPage {
  total: number
  summary: RepairRecordSummary
  list: RepairRecordItem[]
}

export function apiAdminRepairRecords(params: {
  buildingId?: number
  faultType?: string
  status?: number
  keyword?: string
  days?: number
  page?: number
  size?: number
}): Promise<RepairRecordPage> {
  return http.get('/admin/repair-records', {
    params: {
      buildingId: params.buildingId || undefined,
      faultType: params.faultType || undefined,
      status: params.status || undefined,
      keyword: params.keyword || undefined,
      days: params.days || undefined,
      page: params.page || 1,
      size: params.size || 20,
    },
  })
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

export function apiAdminLogs(params: { page: number; size: number; module?: string; action?: string; keyword?: string; startDate?: string; endDate?: string }): Promise<LogPage> {
  return http.get('/admin/logs', {
    params: {
      page: params.page,
      size: params.size,
      module: params.module || undefined,
      action: params.action || undefined,
      keyword: params.keyword || undefined,
      startDate: params.startDate || undefined,
      endDate: params.endDate || undefined,
    },
  })
}

export interface AdminStats {
  status: { status: number; statusText: string; count: number }[]
  buildings: { buildingId: number; buildingName: string; count: number }[]
  faults: { faultType: string; faultTypeName: string; count: number }[]
  recent: { date: string; count: number }[]
}

export function apiAdminStats(): Promise<AdminStats> {
  return http.get('/admin/stats')
}

export interface AdminReassignPayload {
  workerId: number
}

export interface BatchDispatchItem {
  orderId: number
  workerId: number
}

export interface BatchDispatchResult {
  dispatched: BatchDispatchItem[]
  remained: number
}

export function apiAdminOrderReassign(id: number, workerId: number): Promise<unknown> {
  return http.post(`/admin/orders/${id}/reassign`, { workerId })
}

export function apiAdminBatchDispatch(payload?: { buildingId?: number; orderIds?: number[] }): Promise<BatchDispatchResult> {
  return http.post('/admin/orders/batch-dispatch', payload || {})
}

export interface ScheduleItem {
  workerId: number
  workDate: string
  shiftType: string
  note?: string
}

export async function apiAdminSchedules(workerId: number, startDate: string, endDate: string): Promise<ScheduleItem[]> {
  const res = (await http.get('/admin/schedules', {
    params: { workerId: workerId || undefined, startDate, endDate },
  })) as { list: ScheduleItem[] }
  return res.list || []
}

export function apiSaveSchedules(items: ScheduleItem[]): Promise<{ list: ScheduleItem[] }> {
  return http.post('/admin/schedules', { items })
}

export function apiGenerateWeekly(weekStart: string): Promise<{ list: ScheduleItem[] }> {
  return http.post('/admin/schedules/generate', { weekStart })
}

export interface WorkerBoardItem {
  workerId: number
  name: string
  username: string
  buildingNames: string[]
  maxConcurrent: number
  todayShift: string
  activeOrders: number
  todayCompleted: number
  available: boolean
}

export async function apiAdminWorkerBoard(): Promise<WorkerBoardItem[]> {
  const res = (await http.get('/admin/worker-board')) as { list: WorkerBoardItem[] }
  return res.list || []
}

export interface SlaOverview {
  pendingTimeoutHours: number
  dispatchedTimeoutHours: number
  pendingOverdue: number
  dispatchedOverdue: number
  avgDispatchMinutes: number
  avgRepairMinutes: number
  overdueOrders: OrderItem[]
}

export function apiAdminSlaOverview(days = 3): Promise<SlaOverview> {
  return http.get('/admin/sla-overview', { params: { days } })
}

export interface FeedbackRatingCount {
  rating: number
  cnt: number
}

export interface FeedbackRecentItem {
  orderNo: string
  buildingId: number
  buildingName: string
  room: string
  workerName?: string
  rating: number
  comment: string
  createdAt: string
}

export interface AdminFeedbackStats {
  total: number
  avgRating: number
  ratings: FeedbackRatingCount[]
  recent: FeedbackRecentItem[]
}

export function apiAdminFeedbackStats(): Promise<AdminFeedbackStats> {
  return http.get('/admin/feedback-stats')
}

export interface NotificationItem {
  id: number
  type: string
  title: string
  content: string
  orderId?: number
  isRead: number
  createdAt: string
}

export interface NotificationPage {
  total: number
  unread: number
  list: NotificationItem[]
}

export function apiNotifications(page = 1, size = 20, unreadOnly = false): Promise<NotificationPage> {
  return http.get('/notifications', { params: { page, size, unreadOnly: unreadOnly ? 1 : undefined } })
}

export function apiNotificationRead(id: number): Promise<unknown> {
  return http.post(`/notifications/${id}/read`)
}

export function apiNotificationReadAll(): Promise<unknown> {
  return http.post('/notifications/read-all')
}

export interface LeaveItem {
  id: number
  workerId: number
  workerName?: string
  startDate: string
  endDate: string
  reason: string
  status: number
  statusText: string
  reviewNote?: string
  createdAt: string
}

export interface LeavePage {
  total: number
  list: LeaveItem[]
}

export function apiAdminLeaves(params: { workerId?: number; status?: number; page?: number; size?: number; days?: number }): Promise<LeavePage> {
  return http.get('/admin/leaves', {
    params: {
      workerId: params.workerId || undefined,
      status: params.status || undefined,
      page: params.page || 1,
      size: params.size || 20,
      days: params.days || undefined,
    },
  })
}

export function apiAdminLeaveReview(id: number, status: number, reviewNote: string): Promise<unknown> {
  return http.post(`/admin/leaves/${id}/review`, { status, reviewNote })
}