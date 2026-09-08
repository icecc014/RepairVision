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

    <div class="plan-card" :class="'f' + activeFloor">
      <div class="plan-title">楼层户型 · {{ activeFloor }}F</div>
      <div class="row-north">
        <span class="side-tag">北侧</span>
        <div class="rooms">
          <button
            v-for="room in rowNorth"
            :key="room.num"
            class="room"
            :class="room.active ? 'has-fault' : ''"
            :style="{ background: room.active ? '#fee2e2' : '#dbeafe' }"
            @click="$emit('selectRoom', room)"
          >
            <span class="room-code">{{ room.num }}</span>
          </button>
        </div>
        <span class="side-tag">102-110</span>
      </div>

      <div class="corridor">
        <span class="corridor-text">中央走廊</span>
        <i class="corridor-line"></i>
      </div>

      <div class="row-south">
        <span class="side-tag">南侧</span>
        <div class="rooms">
          <button
            v-for="room in rowSouth"
            :key="room.num"
            class="room"
            :class="room.active ? 'has-fault' : ''"
            :style="{ background: room.active ? '#fee2e2' : '#dbeafe' }"
            @click="$emit('selectRoom', room)"
          >
            <span class="room-code">{{ room.num }}</span>
          </button>
        </div>
        <span class="side-tag">111-120</span>
      </div>

      <div class="plan-footer">
        <div class="stairs left">◢ 楼梯间</div>
        <div class="stairs mid">公共盥洗 / 水房</div>
        <div class="stairs right">楼梯间 ◣</div>
      </div>
    </div>
    <p class="legend-tip"><i class="red"></i> 红色 = 待处理故障；点击房间查看工单</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'

const props = defineProps<{ building: WorkerMapBuilding; orders: OrderItem[] }>()
defineEmits<{ (e: 'selectRoom', room: { num: string; floor: number; orders: OrderItem[] }): void }>()

const activeFloor = ref(1)

const activeOrders = computed(() => props.orders.filter((o) => o.buildingId === props.building.id && o.floor === activeFloor.value))

const rowNorth = computed(() => {
  const total = Math.max(props.building.roomsPerFloor, 1)
  const half = Math.ceil(total / 2)
  return Array.from({ length: half }, (_, i) => makeRoom(activeFloor.value, i + 1))
})

const rowSouth = computed(() => {
  const total = Math.max(props.building.roomsPerFloor, 1)
  const half = Math.ceil(total / 2)
  const south = total - half
  return Array.from({ length: south }, (_, i) => makeRoom(activeFloor.value, half + i + 1))
})

function makeRoom(floor: number, idx: number) {
  const num = `${floor}${String(idx).padStart(2, '0')}`
  const orders = activeOrders.value.filter((o) => o.room === num || o.room === `${floor}${String(idx).padStart(2, '0')}`)
  return { num, floor, orders, active: orders.length > 0 }
}
</script>

<style scoped>
.plan-wrap {
  padding: 0 14px;
}
.floor-tabs {
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
  overflow-x: auto;
}
.floor-tab {
  flex: 0 0 auto;
  padding: 6px 13px;
  font-size: 12px;
  background: #eef2ff;
  border: 1px solid #c7d2fe;
  border-radius: 999px;
  cursor: pointer;
}
.floor-tab.active {
  color: #fff;
  background: #2563eb;
}
.plan-card {
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #dbe4f0;
  border-radius: 12px;
}
.plan-title {
  margin-bottom: 8px;
  color: #334155;
  font-size: 13px;
  font-weight: 700;
  text-align: center;
}
.row-north,
.row-south {
  display: flex;
  align-items: center;
  gap: 6px;
}
.side-tag {
  writing-mode: vertical-rl;
  color: #94a3b8;
  font-size: 10px;
}
.rooms {
  display: flex;
  flex: 1;
  gap: 4px;
}
.room {
  flex: 1;
  min-width: 0;
  height: 42px;
  padding: 0;
  color: #1e3a8a;
  background: #dbeafe;
  border: 1px solid #bfdbfe;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.room.has-fault {
  color: #991b1b;
  background: #fee2e2;
  border-color: #fecaca;
}
.room-code {
  font-size: 10px;
  font-weight: 700;
}
.corridor {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 34px;
  margin: 6px 0;
  background: repeating-linear-gradient(90deg, #e2e8f0 0 10px, #cbd5e1 10px 20px);
  border: 1px dashed #94a3b8;
  border-radius: 6px;
}
.corridor-text {
  position: relative;
  z-index: 1;
  padding: 0 6px;
  color: #64748b;
  font-size: 11px;
  background: #eef2f7;
}
.corridor-line {
  display: none;
}
.plan-footer {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  gap: 6px;
}
.stairs {
  padding: 5px 8px;
  color: #475569;
  font-size: 10px;
  background: #e2e8f0;
  border-radius: 4px;
}
.stairs.left {
  color: #b45309;
  background: #fef3c7;
}
.stairs.right {
  color: #b45309;
  background: #fef3c7;
}
.stairs.mid {
  color: #0f766e;
  background: #ccfbf1;
}
.legend-tip {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 11px;
  text-align: center;
}
.red {
  display: inline-block;
  width: 8px;
  height: 8px;
  background: #ef4444;
  border-radius: 50%;
}
</style>