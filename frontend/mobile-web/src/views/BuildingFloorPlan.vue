<template>
  <div class="plan-wrap">
    <div class="floor-tabs">
      <button
        v-for="f in building.floors"
        :key="f"
        class="floor-tab"
        :class="{ active: activeFloor === f }"
        @click="activeFloor = f"
      >
        {{ f }}F
      </button>
    </div>

    <div class="legend-line">
      <span><i class="red"></i> 待处理工单</span>
      <span><i class="blue"></i> 普通房间</span>
      <span>点击房间查看工单</span>
    </div>

    <svg class="plan-svg" :viewBox="`0 0 ${PLAN_WIDTH} ${PLAN_DEPTH}`" preserveAspectRatio="xMidYMid meet">
      <!-- 外墙 -->
      <rect x="1" y="1" :width="PLAN_WIDTH - 2" :height="PLAN_DEPTH - 2" fill="#f8fafc" stroke="#1e293b" stroke-width="1.6" rx="1.5" />

      <!-- 贯通走廊 -->
      <rect :x="plan.corridor.x" :y="plan.corridor.z" :width="plan.corridor.w" :height="plan.corridor.d" fill="#eef2f7" stroke="#94a3b8" stroke-width="0.6" stroke-dasharray="3 2" />
      <text :x="plan.corridor.x + plan.corridor.w / 2" :y="PLAN_DEPTH / 2" text-anchor="middle" font-size="4" fill="#94a3b8" transform="rotate(90, 50, 88)">过道</text>

      <!-- 两处核心筒：封闭防火楼梯 + 公共区域 -->
      <g v-for="core in plan.cores" :key="core.index">
        <rect :x="0" :y="core.z" :width="32" :height="core.d" fill="#e2e8f0" stroke="#475569" stroke-width="0.7" />
        <text x="16" :y="core.z + core.d / 2 - 1.6" text-anchor="middle" font-size="3.4" fill="#334155">封闭防火</text>
        <text x="16" :y="core.z + core.d / 2 + 3.4" text-anchor="middle" font-size="3.4" fill="#334155">楼梯间①</text>
        <rect :x="32" :y="core.z" :width="68" :height="core.d" fill="#e8eef7" stroke="#64748b" stroke-width="0.7" stroke-dasharray="2 1.6" />
        <text x="66" :y="core.z + core.d / 2 + 1.4" text-anchor="middle" font-size="3.8" fill="#64748b">公共区域</text>
      </g>

      <!-- 房间（北区/中区/南区，按设计 2 列排布） -->
      <g v-for="room in plan.rooms" :key="room.no">
        <rect
          :x="room.x"
          :y="room.z"
          :width="room.w"
          :height="room.d"
          :fill="roomOrders(room)[0] ? '#fee2e2' : '#dbeafe'"
          stroke="#1e3a8a"
          stroke-width="0.8"
          @click="select(room)"
          style="cursor: pointer"
        />
        <text
          :x="room.x + room.w / 2"
          :y="room.z + room.d / 2 + 1.4"
          text-anchor="middle"
          font-size="4.6"
          font-weight="bold"
          fill="#1e3a8a"
        >
          {{ room.no }}
        </text>
        <circle
          v-if="roomOrders(room).length"
          :cx="room.x + room.w - 5"
          :cy="room.z + 5"
          r="3.4"
          fill="#ef4444"
          @click="select(room)"
          style="cursor: pointer"
        />
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'
import {
  PLAN_DEPTH,
  PLAN_WIDTH,
  buildFloorPlan,
  buildGridRooms,
  matchRoomOrders,
  supportsCorridorLayout,
  type PlanRoom,
} from '../utils/floorLayout'

const props = defineProps<{ building: WorkerMapBuilding; orders: OrderItem[] }>()
const emit = defineEmits<{ (e: 'selectRoom', room: { num: string; floor: number; orders: OrderItem[] }): void }>()

const activeFloor = ref(1)

const plan = computed(() => {
  if (supportsCorridorLayout(props.building.roomsPerFloor)) {
    return buildFloorPlan(activeFloor.value, props.building.roomsPerFloor)
  }
  return {
    floor: activeFloor.value,
    rooms: buildGridRooms(activeFloor.value, props.building.roomsPerFloor || 8),
    corridor: { x: 0, z: 0, w: 0, d: 0 },
    cores: [],
  }
})

function roomOrders(room: PlanRoom) {
  return matchRoomOrders(
    props.orders.filter((o) => o.buildingId === props.building.id),
    activeFloor.value,
    room.no,
  )
}

function select(room: PlanRoom) {
  emit('selectRoom', { num: room.no, floor: activeFloor.value, orders: roomOrders(room) })
}
</script>

<style scoped>
.plan-wrap { padding: 4px 10px 12px; }
.floor-tabs { display: flex; gap: 6px; margin-bottom: 6px; overflow-x: auto; }
.floor-tab { flex: 0 0 auto; padding: 6px 13px; font-size: 12px; background: #eef2ff; border: 1px solid #c7d2fe; border-radius: 999px; cursor: pointer; }
.floor-tab.active { color: #fff; background: #2563eb; }
.legend-line { display: flex; gap: 14px; margin-bottom: 4px; color: #64748b; font-size: 11px; }
.red { display: inline-block; width: 8px; height: 8px; background: #ef4444; border-radius: 50%; }
.blue { display: inline-block; width: 8px; height: 8px; background: #dbeafe; border: 1px solid #1e3a8a; border-radius: 2px; }
.plan-svg { width: 100%; height: auto; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; }
</style>