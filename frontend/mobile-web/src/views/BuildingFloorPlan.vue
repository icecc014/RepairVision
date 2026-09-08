<template>
  <div class="unit-plan">
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
      <span><i class="red"></i> 待处理</span>
      <span>点击“户”查看工单</span>
    </div>

    <svg class="plan-svg" viewBox="0 0 900 470" preserveAspectRatio="xMidYMid meet">
      <defs>
        <pattern id="wall" width="5" height="5" patternUnits="userSpaceOnUse">
          <rect width="5" height="5" fill="#e2e8f0" />
          <line x1="0" y1="0" x2="5" y2="5" stroke="#94a3b8" stroke-width="1" />
        </pattern>
      </defs>

      <g v-for="u in units" :key="u.id" :transform="`translate(${u.x} 0)`">
        <!-- 单元楼体外框 -->
        <rect x="12" y="16" width="280" height="430" fill="#f1f5f9" stroke="#334155" stroke-width="4" rx="6" />
        <text x="152" y="36" text-anchor="middle" font-size="15" font-weight="bold" fill="#334155">{{ u.id }}单元</text>

        <!-- 楼梯间：突出/错位，位于单元中部 -->
        <rect x="125" y="55" width="54" height="230" fill="#e7edf5" stroke="#475569" stroke-width="2" />
        <rect x="135" y="52" width="16" height="240" fill="#cbd5e1" />
        <rect x="153" y="52" width="16" height="240" fill="#cbd5e1" />
        <text x="152" y="180" text-anchor="middle" font-size="12" fill="#475569">楼</text>
        <text x="152" y="195" text-anchor="middle" font-size="12" fill="#475569">梯</text>
        <text x="152" y="210" text-anchor="middle" font-size="12" fill="#475569">间</text>

        <!-- 左侧户：入户门+门厅+客厅+卧室+阳台 -->
        <g>
          <path
            d="M20,70 L20,300 L88,300 L88,328 L104,328 L104,300 L120,300 L120,440 L20,440 Z"
            fill="#dbeafe" stroke="#1e3a8a" stroke-width="2"
          />
          <path d="M112,440 L112,386 L140,386 L140,440 Z" fill="#bfdbfe" stroke="#1e3a8a" stroke-width="2" />
          <rect x="20" y="70" width="100" height="34" fill="#e0f2fe" stroke="#2563eb" stroke-width="1.5" />
          <text x="70" y="91" text-anchor="middle" font-size="10" fill="#1d4ed8">门厅</text>
          <rect x="20" y="104" width="100" height="110" fill="#eff6ff" stroke="#1e3a8a" stroke-width="1.5" />
          <text x="70" y="160" text-anchor="middle" font-size="11" fill="#1e3a8a">客厅</text>
          <rect x="20" y="214" width="100" height="100" fill="#e0f2fe" stroke="#1e3a8a" stroke-width="1.5" />
          <text x="70" y="262" text-anchor="middle" font-size="11" fill="#1e3a8a">卧室</text>
          <rect x="20" y="314" width="68" height="72" fill="#dbeafe" stroke="#1e3a8a" stroke-width="1.5" />
          <text x="54" y="352" text-anchor="middle" font-size="10" fill="#1e3a8a">厨卫</text>
          <rect x="112" y="328" width="28" height="112" fill="#e0f2fe" stroke="#0369a1" stroke-width="1.5" />
          <text x="126" y="386" text-anchor="middle" font-size="9" fill="#0369a1" transform="rotate(90 126 386)">阳台</text>
          <rect x="18" y="58" width="102" height="16" fill="#fde68a" stroke="#b45309" stroke-width="1.5" />
          <text x="69" y="70" text-anchor="middle" font-size="8" fill="#92400e">入户门</text>
          <circle v-if="u.leftHot" cx="140" cy="85" r="7" fill="#ef4444" stroke="#fff" stroke-width="2" /><rect x="16" y="58" width="130" height="385" fill="transparent" style="cursor:pointer" @click="pickUnit(u, 'left')" />
        </g>

        <!-- 右侧户 -->
        <g>
          <path
            d="M160,70 L160,300 L104,300 L104,328 L88,328 L104,328 L104,300 L104,440 L160,440 Z"
            fill="#c7f9cc" stroke="#166534" stroke-width="2"
          />
          <path d="M104,440 L104,386 L80,386 L80,440 Z" fill="#bbf7d0" stroke="#166534" stroke-width="2" />
          <rect x="160" y="70" width="100" height="34" fill="#dcfce7" stroke="#16a34a" stroke-width="1.5" />
          <text x="210" y="91" text-anchor="middle" font-size="10" fill="#15803d">门厅</text>
          <rect x="160" y="104" width="100" height="110" fill="#f0fdf4" stroke="#166534" stroke-width="1.5" />
          <text x="210" y="160" text-anchor="middle" font-size="11" fill="#166534">客厅</text>
          <rect x="160" y="214" width="100" height="100" fill="#dcfce7" stroke="#166534" stroke-width="1.5" />
          <text x="210" y="262" text-anchor="middle" font-size="11" fill="#166534">卧室</text>
          <rect x="160" y="314" width="68" height="72" fill="#c7f9cc" stroke="#166534" stroke-width="1.5" />
          <text x="194" y="352" text-anchor="middle" font-size="10" fill="#166534">厨卫</text>
          <rect x="80" y="328" width="24" height="112" fill="#bbf7d0" stroke="#047857" stroke-width="1.5" />
          <text x="92" y="386" text-anchor="middle" font-size="9" fill="#047857" transform="rotate(90 92 386)">阳台</text>
          <rect x="160" y="58" width="102" height="16" fill="#fde68a" stroke="#b45309" stroke-width="1.5" />
          <text x="211" y="70" text-anchor="middle" font-size="8" fill="#92400e">入户门</text>
          <circle v-if="u.rightHot" cx="142" cy="85" r="7" fill="#ef4444" stroke="#fff" stroke-width="2" /><rect x="160" y="58" width="110" height="385" fill="transparent" style="cursor:pointer" @click="pickUnit(u, 'right')" />
        </g>
      </g>

      <!-- 公共走廊/连廊文字 -->
      <text x="450" y="455" text-anchor="middle" font-size="11" fill="#64748b">
        单元式住宅标准层（每单元：中部楼梯 + 左/右各一户，户型含门厅/客厅/卧室/厨卫/阳台）
      </text>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { OrderItem, WorkerMapBuilding } from '../api'

const props = defineProps<{ building: WorkerMapBuilding; orders: OrderItem[] }>()
const emit = defineEmits<{ (e: 'selectRoom', room: { num: string; floor: number; orders: OrderItem[] }): void }>()

const activeFloor = ref(1)

const unitCount = computed(() => {
  const n = props.building.id % 3
  return n === 0 ? 3 : n
})

const units = computed(() => {
  return Array.from({ length: unitCount.value }, (_, i) => {
    const id = i + 1
    const floorOrders = props.orders.filter((o) => o.buildingId === props.building.id && o.floor === activeFloor.value)
    return {
      id,
      x: i * 300 + 4,
      leftHot: floorOrders.some((o) => hashRoom(o.room, id, true)),
      rightHot: floorOrders.some((o) => hashRoom(o.room, id, false)),
    }
  })
})

function hashRoom(room: string, unit: number, left: boolean) {
  const raw = room || '1'
  let h = 0
  for (let i = 0; i < raw.length; i++) h = (h * 31 + raw.charCodeAt(i)) >>> 0
  const flatIndex = (h % 6) + 1 // 1..6 对应 3 单元 x 2 户
  const target = (unit - 1) * 2 + (left ? 1 : 2)
  return flatIndex === target
}

function pickUnit(u: { id: number }, side: 'left' | 'right') {
  const floorOrders = props.orders.filter((o) => o.buildingId === props.building.id && o.floor === activeFloor.value)
  const matched = floorOrders.filter((o) => hashRoom(o.room, u.id, side === 'left'))
  const num = `${activeFloor.value}${u.id}${side === 'left' ? '01' : '02'}`
  emit('selectRoom', { num, floor: activeFloor.value, orders: matched })
}
</script>

<style scoped>
.unit-plan {
  padding: 4px 10px 12px;
}
.floor-tabs {
  display: flex;
  gap: 6px;
  margin-bottom: 6px;
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
.legend-line {
  display: flex;
  gap: 14px;
  margin-bottom: 4px;
  color: #64748b;
  font-size: 11px;
}
.red {
  display: inline-block;
  width: 8px;
  height: 8px;
  background: #ef4444;
  border-radius: 50%;
}
.plan-svg {
  width: 100%;
  height: auto;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}
</style>