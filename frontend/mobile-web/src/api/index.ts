import http from './http'

export interface LoginResult {
  token: string
  user: UserInfo
}

export interface UserInfo {
  id: number
  username: string
  name: string
  role: number
  buildingId?: number
}

export interface FaultType {
  id: number
  code: string
  name: string
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
  workerPhone?: string
  reporterId: number
  reporterName?: string
  source: string
  createdAt: string
  updatedAt: string
}

export function apiLogin(username: string, password: string): Promise<LoginResult> {
  return http.post('/login', { username, password })
}

export async function apiFaultTypes(): Promise<FaultType[]> {
  const res = (await http.get('/fault-types')) as { list: FaultType[] }
  return res.list || []
}

export async function apiDormOrders(status = 0): Promise<OrderItem[]> {
  const res = (await http.get('/dorm/orders', { params: { status: status || undefined } })) as {
    list: OrderItem[]
  }
  return res.list || []
}

export function apiCreateOrder(payload: {
  room: string
  floor: number
  faultType: string
  description: string
}): Promise<{ orderId: number; orderNo: string }> {
  return http.post('/dorm/orders', payload)
}

export function apiCancelOrder(id: number): Promise<unknown> {
  return http.post(`/dorm/orders/${id}/cancel`)
}

export async function apiWorkerOrders(status = 0): Promise<OrderItem[]> {
  const res = (await http.get('/worker/orders', { params: { status: status || undefined } })) as {
    list: OrderItem[]
  }
  return res.list || []
}

export function apiStartOrder(id: number): Promise<unknown> {
  return http.post(`/worker/orders/${id}/start`)
}

export function apiCompleteOrder(id: number): Promise<unknown> {
  return http.post(`/worker/orders/${id}/complete`)
}

export interface WorkerMapBuilding {
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

export interface WorkerMapData {
  buildings: WorkerMapBuilding[]
  orders: OrderItem[]
}

export function apiWorkerMapData(): Promise<WorkerMapData> {
  return http.get('/worker/map-data')
}

export function apiBatchComplete(buildingId: number, faultType: string): Promise<{ count: number }> {
  return http.post('/worker/batch-complete', { buildingId, faultType })
}

export interface ScheduleItem {
  workerId: number
  workDate: string
  shiftType: string
  note?: string
}

export async function apiWorkerSchedules(startDate: string, endDate: string): Promise<ScheduleItem[]> {
  const res = (await http.get('/worker/schedules', {
    params: { startDate, endDate },
  })) as { list: ScheduleItem[] }
  return res.list || []
}