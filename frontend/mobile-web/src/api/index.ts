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
  reporterId: number
  reporterName?: string
  source: string
  createdAt: string
  updatedAt: string
}

export function apiLogin(username: string, password: string): Promise<LoginResult> {
  return http.post('/login', { username, password })
}

export function apiFaultTypes(): Promise<FaultType[]> {
  return http.get('/fault-types')
}

export function apiDormOrders(status = 0): Promise<OrderItem[]> {
  return http.get('/dorm/orders', { params: { status: status || undefined } })
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

export function apiWorkerOrders(status = 0): Promise<OrderItem[]> {
  return http.get('/worker/orders', { params: { status: status || undefined } })
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