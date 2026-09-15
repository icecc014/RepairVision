<template>
  <div class="campus-wrap">
    <div v-if="!layout" class="campus-empty">
      {{ loading ? '正在加载校园概览…' : '管理员尚未绘制区域概览' }}
    </div>
    <template v-else>
      <svg class="campus-svg" :viewBox="`0 0 ${layout.cols} ${layout.rows}`" preserveAspectRatio="xMidYMid meet">
        <defs>
          <linearGradient id="campusBg" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#f2f6fb" />
            <stop offset="100%" stop-color="#e8eef8" />
          </linearGradient>
        </defs>
        <rect x="0.2" y="0.2" :width="layout.cols - 0.4" :height="layout.rows - 0.4" fill="url(#campusBg)" stroke="#c7d3e8" stroke-width="0.25" rx="0.8" />
        <g v-for="b in layout.blocks" :key="b.id">
          <rect
            :x="b.col + 0.12"
            :y="b.row + 0.12"
            :width="Math.max(0.2, b.colSpan - 0.24)"
            :height="Math.max(0.2, b.rowSpan - 0.24)"
            :fill="fillOf(b.kind)"
            :stroke="strokeOf(b.kind)"
            :stroke-width="highlightBuildingId && b.buildingId === highlightBuildingId ? 0.5 : 0.22"
            rx="0.35"
          />
          <text
            v-if="b.kind !== 'road'"
            :x="b.col + b.colSpan / 2"
            :y="b.row + b.rowSpan / 2 + 0.5"
            text-anchor="middle"
            :font-size="labelSize(b)"
            fill="#33415c"
          >
            {{ displayLabel(b) }}
          </text>
        </g>
      </svg>
      <div class="campus-legend">
        <span><i class="lg building" />建筑</span>
        <span><i class="lg road" />道路</span>
        <span><i class="lg green" />广场/绿地</span>
        <span><i class="lg gate" />校门</span>
        <span v-if="buildingCount > 0">共 {{ buildingCount }} 处建筑</span>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { apiCampusMap } from '../api'

interface CampusBlock {
  id: string
  kind: string
  row: number
  col: number
  rowSpan: number
  colSpan: number
  label?: string
  buildingId?: number
}

const props = defineProps<{
  buildings?: { id: number; name: string }[]
  highlightBuildingId?: number
  autoLoad?: boolean
}>()

const loading = ref(false)
const layout = ref<{ cols: number; rows: number; blocks: CampusBlock[] } | null>(null)

const KIND_TEXT: Record<string, string> = {
  building: '建筑',
  road: '道路',
  green: '广场',
  gate: '校门',
  custom: '设施',
}

const buildingCount = computed(() => (layout.value ? layout.value.blocks.filter((b) => b.kind === 'building').length : 0))

function fillOf(kind: string) {
  switch (kind) {
    case 'building':
      return '#dcebff'
    case 'road':
      return '#efe8da'
    case 'green':
      return '#ddf0e3'
    case 'gate':
      return '#f7e3ef'
    default:
      return '#f2e8ff'
  }
}

function strokeOf(kind: string) {
  switch (kind) {
    case 'building':
      return '#5f7bb5'
    case 'road':
      return '#c3b598'
    case 'green':
      return '#7fb894'
    case 'gate':
      return '#c084a5'
    default:
      return '#a985c8'
  }
}

function labelSize(b: CampusBlock) {
  const base = Math.min(b.rowSpan, b.colSpan)
  return Math.max(0.9, Math.min(2.2, base * 0.55))
}

// 建筑名称以"建筑信息管理"中的最新名称优先，其次用概览内保存的名称
function displayLabel(b: CampusBlock) {
  if (b.buildingId) {
    const hit = props.buildings?.find((x) => x.id === b.buildingId)
    if (hit) return hit.name
  }
  return b.label || KIND_TEXT[b.kind] || ''
}

async function load() {
  loading.value = true
  try {
    const data = await apiCampusMap()
    const raw = data.layoutJson ? JSON.parse(data.layoutJson) : null
    if (raw && Array.isArray(raw.blocks) && raw.blocks.length > 0) {
      layout.value = {
        cols: Number(raw.cols) || 40,
        rows: Number(raw.rows) || 30,
        blocks: raw.blocks.filter((b: CampusBlock) => b && b.kind),
      }
    } else {
      layout.value = null
    }
  } catch {
    layout.value = null
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (props.autoLoad !== false) load()
})

defineExpose({ load })
</script>

<style scoped>
.campus-wrap {
  padding: 4px 0;
}
.campus-svg {
  width: 100%;
  max-height: 42vh;
  border-radius: 12px;
  background: #fff;
}
.campus-empty {
  padding: 18px 10px;
  color: var(--rv-text-light, #94a3b8);
  font-size: 13px;
  text-align: center;
}
.campus-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
  color: var(--rv-text-sub, #5a6a85);
  font-size: 11px;
}
.campus-legend .lg {
  display: inline-block;
  width: 10px;
  height: 10px;
  margin-right: 4px;
  border-radius: 2px;
  vertical-align: -1px;
}
.lg.building { background: #dcebff; border: 1px solid #5f7bb5; }
.lg.road { background: #efe8da; border: 1px solid #c3b598; }
.lg.green { background: #ddf0e3; border: 1px solid #7fb894; }
.lg.gate { background: #f7e3ef; border: 1px solid #c084a5; }
</style>
