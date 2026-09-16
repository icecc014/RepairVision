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
  manualReview?: number
  externalMark?: number
  dispatchLocked?: number
  priority?: number
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
  category: string
  autoDispatch: number
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

export function apiCreateFaultType(payload: { code: string; name: string; sort: number; category?: string; autoDispatch?: number }): Promise<unknown> {
  return http.post('/admin/fault-types', payload)
}

export function apiUpdateFaultType(
  id: number,
  payload: { name: string; sort: number; status: number; category?: string; autoDispatch?: number },
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
  jobType?: number
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
  jobType?: number
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
    jobType?: number
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

export interface CampusLayoutData {
  id: number
  name: string
  cols: number
  rows: number
  layoutJson: string
  updatedAt: string
}

// 区域概览（校园/建筑群总平面图）
export function apiAdminCampusLayout(): Promise<CampusLayoutData> {
  return http.get('/admin/campus-layout')
}

export function apiSaveCampusLayout(payload: { name: string; cols: number; rows: number; layoutJson: string }): Promise<CampusLayoutData> {
  return http.post('/admin/campus-layout', payload)
}
export interface CampusTemplateItem {
  id: number
  name: string
  cols: number
  rows: number
  source: string
  updatedAt: string
}

export interface CampusTemplateDetail extends CampusTemplateItem {
  layoutJson: string
}

// V6.1 「我的画布」模板库
export function apiAdminCampusTemplates(): Promise<{ list: CampusTemplateItem[] }> {
  return http.get('/admin/campus-templates')
}

export function apiAdminCampusTemplate(id: number): Promise<{ template: CampusTemplateDetail }> {
  return http.get(`/admin/campus-templates/${id}`)
}

export function apiSaveCampusTemplate(payload: {
  name: string
  cols: number
  rows: number
  layoutJson: string
  overwrite?: boolean
}): Promise<{ template: CampusTemplateDetail }> {
  return http.post('/admin/campus-templates', payload)
}

export function apiRenameCampusTemplate(id: number, name: string): Promise<unknown> {
  return http.put(`/admin/campus-templates/${id}`, { name })
}

export function apiDeleteCampusTemplate(id: number): Promise<unknown> {
  return http.delete(`/admin/campus-templates/${id}`)
}

export function apiApplyCampusTemplate(id: number): Promise<CampusLayoutData> {
  return http.post(`/admin/campus-templates/${id}/apply`, {})
}

export interface AssignableWorkerItem {
  id: number
  name: string
  username: string
  jobType: number
  jobTypeText: string
  onDuty: boolean
  todayShift: string
  inProgress: number
  maxConcurrent: number
  inBuilding: boolean
  selectable: boolean
  reason: string
  buildingNames: string[]
}

export interface AssignableWorkerList {
  list: AssignableWorkerItem[]
  requiredJobType: number
  requiredJobText: string
  faultTypeName: string
}

// V6.5 手动派单 / 改派候选工人（后端按在岗 + 工种匹配计算）
export function apiAdminAssignableWorkers(orderId: number): Promise<AssignableWorkerList> {
  return http.get(`/admin/orders/${orderId}/assignable-workers`)
}

export interface CampusBackupItem {
  id: number
  name: string
  cols: number
  rows: number
  blocks: number
  buildings: number
  bytes: number
  createdAt: string
}

// V6.2 区域概览备份与恢复
export function apiAdminCampusBackups(): Promise<{ list: CampusBackupItem[]; keep: number }> {
  return http.get('/admin/campus-backups')
}

export function apiCreateCampusBackup(): Promise<unknown> {
  return http.post('/admin/campus-backups', {})
}

export function apiRestoreCampusBackup(id: number): Promise<CampusLayoutData> {
  return http.post(`/admin/campus-backups/${id}/restore`, {})
}

export function apiSetCampusBackupKeep(keep: number): Promise<unknown> {
  return http.put('/admin/campus-backups/keep', { keep })
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
  days?: number
  totalAllTime?: number
  doneAllTime?: number
  status: { status: number; statusText: string; count: number }[]
  buildings: { buildingId: number; buildingName: string; count: number }[]
  faults: { faultType: string; faultTypeName: string; count: number }[]
  recent: { date: string; count: number }[]
}

export function apiAdminStats(days = 0): Promise<AdminStats> {
  return http.get('/admin/stats', { params: { days: days || undefined } })
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

export function apiAdminOrderLock(id: number, locked: number): Promise<unknown> {
  return http.post(`/admin/orders/${id}/lock`, { locked })
}

export function apiAdminOrderPriority(id: number, priority: number): Promise<unknown> {
  return http.post(`/admin/orders/${id}/priority`, { priority })
}
export function apiAdminOrderExternal(id: number): Promise<unknown> {
  return http.post(`/admin/orders/${id}/external`, {})
}

// V6.4 管理员代为完工（演示/救急：不用逐个登录工人账号）
export function apiAdminOrderComplete(id: number): Promise<unknown> {
  return http.post(`/admin/orders/${id}/complete`, {})
}

export interface WorkSettings {
  restDaysPerWeek: number
  restMode: string
  fixedRestWeekdays: string
  morningStart: string
  morningEnd: string
  afternoonStart: string
  afternoonEnd: string
  allowForceStart: number
}

export interface DutyStatusItem {
  workerId: number
  name: string
  onDuty: boolean
  working: boolean
  enabled: boolean
  onLeave: boolean
  shiftType: string
  inWorkPeriod: boolean
  onLeave: boolean
  enabled: boolean
  reason: string
  morning: string
  afternoon: string
}

export interface DutyOverview {
  onDutyCount: number
  total: number
  morning: string
  afternoon: string
  list: DutyStatusItem[]
}

export interface CampusDistanceBuilding {
  buildingId: number
  name: string
  entryCount: number
  connected: boolean
  maxMeters: number
}

export interface CampusDistancePair {
  fromId: number
  toId: number
  meters: number
}

export interface CampusDistance {
  gridMeters: number
  maxMeters: number
  roadCells: number
  buildings: CampusDistanceBuilding[]
  pairs: CampusDistancePair[]
}

export async function apiAdminCampusDistances(): Promise<CampusDistance> {
  return http.get('/admin/campus-layout/distances')
}
export async function apiAdminWorkSettings(): Promise<WorkSettings> {
  return http.get('/admin/work-settings')
}

export async function apiUpdateWorkSettings(payload: Partial<WorkSettings>): Promise<WorkSettings> {
  return http.post('/admin/work-settings', payload)
}

export interface DispatchRuleEntry {
  key: string
  name: string
  group: string
  type: string
  unit: string
  min: number
  max: number
  step: number
  precision: number
  value: number
  defaultValue: number
  enabled: number
  remark: string
  description: string
  runtimeOnly: boolean
  registered: boolean
  saved: boolean
  custom: boolean
  updatedAt: string
}

export interface DispatchRuleGroup {
  key: string
  label: string
  items: DispatchRuleEntry[]
}

export interface DispatchRulesData {
  groups: DispatchRuleGroup[]
  paused: boolean
  pendingCount: number
  onDutyCount: number
  updatedAt: string
}

// V6.3 派单规则可视化编辑（分组 + 元数据 + 当前值）
export function apiDispatchRulesGrouped(): Promise<DispatchRulesData> {
  return http.get('/admin/dispatch-rules')
}

export function apiSaveDispatchRules(items: Array<{ key: string; value: number; enabled: number; remark?: string }>): Promise<unknown> {
  return http.put('/admin/dispatch-rules', { items })
}

export function apiResetDispatchRule(key: string): Promise<unknown> {
  return http.post(`/admin/dispatch-rules/${key}/reset`, {})
}

export function apiDeleteDispatchRule(key: string): Promise<unknown> {
  return http.delete(`/admin/dispatch-rules/${key}`)
}

export interface DispatchGuard {
  pendingCount: number
  onDutyCount: number
  longestWaitMinutes: number
  warnRatio: number
  warnMinOrders: number
  guardRatio: number
  guardMinOrders: number
  warnHours: number
  guardHours: number
  level: string
  paused: boolean
  mode: string
  message: string
  waitText: string
}

export async function apiAdminDispatchGuard(): Promise<DispatchGuard> {
  return http.get('/admin/dispatch-guard')
}

export async function apiResumeAutoDispatch(): Promise<DispatchGuard> {
  return http.post('/admin/dispatch-guard/resume', {})
}
export async function apiAdminDutyOverview(): Promise<DutyOverview> {
  return http.get('/admin/duty-overview')
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

export function apiGenerateWeekly(weekStart: string, restDaysPerWeek = 1, minPerBuilding = 1): Promise<ScheduleGenerateResult> {
  return http.post('/admin/schedules/generate', { weekStart, restDaysPerWeek, minPerBuilding })
}

export interface ScheduleGenerateResult {
  list: ScheduleItem[]
  warnings?: string[]
}

export interface AdminLeaveCreatePayload {
  workerId: number
  startDate: string
  endDate: string
  reason: string
}

// 管理员代工人登记请假（直接置为已通过）
export function apiAdminLeaveCreate(payload: AdminLeaveCreatePayload): Promise<LeaveItem> {
  return http.post('/admin/leaves', payload)
}

// 管理员撤销请假（待审批或已通过均可）
export function apiAdminLeaveCancel(id: number): Promise<unknown> {
  return http.post(`/admin/leaves/${id}/cancel`)
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
  onDuty?: boolean
  dutyReason?: string
  jobType?: number
  jobTypeText?: string
  loadMinutes?: number
  loadDeviation?: number
  currentBuildingName?: string
  currentOrderNo?: string
}

export async function apiAdminWorkerBoard(days = 1): Promise<WorkerBoardItem[]> {
  const res = (await http.get('/admin/worker-board', { params: { days } })) as { list: WorkerBoardItem[] }
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
export interface AutoAssignResult {
  buildings: number
  workers: number
  list: { workerId: number; name: string; jobTypeText: string; buildingCount: number; buildingIds: number[] }[]
}

export async function apiAutoAssignBuildings(): Promise<AutoAssignResult> {
  return http.post('/admin/worker-buildings/auto-assign', {})
}