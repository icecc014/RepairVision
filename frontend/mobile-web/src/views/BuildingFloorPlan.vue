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
      <span>点击房间查看工单</span>
    </div>

    <svg class="plan-svg" viewBox="0 0 960 600" preserveAspectRatio="xMidYMid meet">
      <!-- 外墙 -->
      <rect x="8" y="8" width="944" height="584" fill="#f8fafc" stroke="#1e293b" stroke-width="5" rx="8" />
      <text x="480" y="28" text-anchor="middle" font-size="14" font-weight="bold" fill="#334155">
        {{ building.name }} · {{ activeFloor }}F 标准层（共16间）
      </text>

      <!-- 左公共区：楼梯间 + 盥洗/卫生间 -->
      <rect x="24" y="48" width="72" height="152" fill="#e2e8f0" stroke="#475569" stroke-width="2" />
      <text x="60" y="100" text-anchor="middle" font-size="11" fill="#334155">楼</text>
      <text x="60" y="115" text-anchor="middle" font-size="11" fill="#334155">梯</text>
      <text x="60" y="130" text-anchor="middle" font-size="11" fill="#334155">间</text>
      <rect x="24" y="206" width="72" height="86" fill="#e2e8f0" stroke="#475569" stroke-width="2" />
      <text x="60" y="242" text-anchor="middle" font-size="10" fill="#334155">盥洗</text>
      <text x="60" y="258" text-anchor="middle" font-size="10" fill="#334155">卫生间</text>
      <rect x="24" y="298" width="72" height="64" fill="#e2e8f0" stroke="#475569" stroke-width="2" />
      <text x="60" y="332" text-anchor="middle" font-size="10" fill="#334155">开水间</text>
      <rect x="24" y="368" width="72" height="220" fill="#e2e8f0" stroke="#475569" stroke-width="2" />
      <text x="60" y="430" text-anchor="middle" font-size="10" fill="#334155">配电/</text>
      <text x="60" y="446" text-anchor="middle" font-size="10" fill="#334155">管理间</text>

      <!-- 北侧房间 1-8 -->
      <g v-for="(room, idx) in rooms('north')" :key="room.no">
        <rect :x="room.x" y="48" width="96" height="150" :fill="room.active ? '#fee2e2' : '#dbeafe'" stroke="#1e3a8a" stroke-width="2" />
        <rect :x="room.x" y="130" width="96" height="30" fill="#bfdbfe" opacity="0.7" />
        <text :x="room.x + 48" y="160" text-anchor="middle" font-size="12" font-weight="bold" fill="#1e3a8a">{{ room.no }}</text>
        <text :x="room.x + 48" y="105" text-anchor="middle" font-size="9" fill="#475569">门 | 窗</text>
        <rect :x="room.x + 40" y="188" width="16" height="12" fill="#f59e0b" @click="select(room)" style="cursor:pointer" />
      </g>

      <!-- 中央走廊 -->
      <rect x="110" y="198" width="824" height="204" fill="#f1f5f9" stroke="#94a3b8" stroke-dasharray="8 6" stroke-width="2" />
      <text x="540" y="305" text-anchor="middle" font-size="16" fill="#64748b">中 央 走 廊</text>

      <!-- 南侧房间 9-16 -->
      <g v-for="(room, idx) in rooms('south')" :key="room.no">
        <rect :x="room.x" y="402" width="96" height="150" :fill="room.active ? '#fee2e2' : '#dbeafe'" stroke="#1e3a8a" stroke-width="2" />
        <rect :x="room.x" y="402" width="96" height="30" fill="#bfdbfe" opacity="0.7" />
        <text :x="room.x + 48" y="470" text-anchor="middle" font-size="12" font-weight="bold" fill="#1e3a8a">{{ room.no }}</text>
        <text :x="room.x + 48" y="430" text-anchor="middle" font-size="9" fill="#475569">窗 | 门</text>
        <rect :x="room.x + 40" y="390" width="16" height="12" fill="#f59e0b" @click="select(room)" style="cursor:pointer" />
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'

const props = defineProps<{ building: WorkerMapBuilding; orders: OrderItem[] }>()
const emit = defineEmits<{ (e: 'selectRoom', room: { num: string; floor: number; orders: OrderItem[] }): void }>()

const activeFloor = ref(1)
const roomStartX = 112

function rooms(side: 'north' | 'south') {
  const list: { no: string; x: number; active: boolean; orders: OrderItem[]; floor: number }[] = []
  for (let i = 1; i <= 8; i++) {
    const seq = side === 'north' ? i : i + 8
    const no = `${activeFloor.value}${String(seq).padStart(2, '0')}`
    const orders = props.orders.filter((o) => o.buildingId === props.building.id && o.floor === activeFloor.value && (o.room === no || o.room === `${seq}` || o.room?.endsWith(String(seq).padStart(2, '0'))))
    list.push({ no, x: roomStartX + (i - 1) * 104, active: orders.length > 0, orders, floor: activeFloor.value })
  }
  return list
}

function select(room: { no: string; floor: number; orders: OrderItem[] }) {
  emit('selectRoom', { num: room.no, floor: room.floor, orders: room.orders })
}
</script>

<style scoped>
.plan-wrap { padding: 4px 10px 12px; }
.floor-tabs { display: flex; gap: 6px; margin-bottom: 6px; overflow-x: auto; }
.floor-tab { flex: 0 0 auto; padding: 6px 13px; font-size: 12px; background: #eef2ff; border: 1px solid #c7d2fe; border-radius: 999px; cursor: pointer; }
.floor-tab.active { color: #fff; background: #2563eb; }
.legend-line { display: flex; gap: 14px; margin-bottom: 4px; color: #64748b; font-size: 11px; }
.red { display: inline-block; width: 8px; height: 8px; background: #ef4444; border-radius: 50%; }
.plan-svg { width: 100%; height: auto; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; }
</style>