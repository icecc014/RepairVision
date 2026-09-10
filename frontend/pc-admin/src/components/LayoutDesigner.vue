<template>
  <el-dialog append-to-body
    :model-value="modelValue"
    :title="`楼层布局设计 · ${building?.name || ''}`"
    width="980px"
    top="5vh"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
  >
    <div class="designer">
      <aside class="tools">
        <div class="tool-title">绘制工具栏</div>
        <button
          v-for="t in CELL_TYPES"
          :key="t.code"
          class="tool-btn"
          :class="{ active: brush === t.code }"
          @click="brush = t.code"
        >
          <i class="chip" :style="{ background: t.color }" />
          <span>{{ t.label }}</span>
          <em>{{ t.hint }}</em>
        </button>

        <div class="tool-title">画布尺寸</div>
        <div class="size-row">
          <el-input-number v-model="cols" :min="4" :max="16" size="small" @change="applySize" />
          <span class="times">×</span>
          <el-input-number v-model="rows" :min="4" :max="20" size="small" @change="applySize" />
        </div>

        <div class="tool-title">快捷操作</div>
        <el-button size="small" @click="loadTemplate">一键生成标准层</el-button>
        <el-button size="small" @click="clearAll">清空画布</el-button>
        <el-button size="small" type="danger" plain @click="clearCustom">清除自定义布局</el-button>
      </aside>

      <section class="canvas-area">
        <div class="hint-line">
          点击或按住拖动即可绘制；房间号按「楼层 × 100 + 顺序」自动生成（如 4 层第 1 间 = 401）。
        </div>
        <div
          class="grid"
          :style="{ gridTemplateColumns: `repeat(${cols}, 1fr)` }"
          @pointerup="painting = false"
          @pointerleave="painting = false"
        >
          <button
            v-for="(cell, i) in flatCells"
            :key="i"
            class="cell"
            :class="[`c${cell.code}`, { snapshot: cell.snapshot }]"
            @pointerdown.prevent="paint(i)"
            @pointerenter="paint(i)"
          >
            <span v-if="cell.code === '1'" class="cell-no">{{ cell.no }}</span>
            <span v-else-if="cell.code === '3'" class="cell-tag">楼梯</span>
            <span v-else-if="cell.code === '4'" class="cell-tag">公共</span>
          </button>
        </div>

        <div class="summary">
          <span>房间 <b>{{ roomCount }}</b> 间</span>
          <span>过道 <b>{{ typeCount('2') }}</b> 格</span>
          <span>楼梯 <b>{{ typeCount('3') }}</b> 格</span>
          <span>公共区 <b>{{ typeCount('4') }}</b> 格</span>
          <span>楼层 <b>{{ building?.floors || 0 }}</b> 层</span>
        </div>
        <p class="tips">
          保存后，该楼栋的 2D 平面图与 3D 楼宇都会按这份布局逐层渲染；其它楼栋不受影响。
        </p>
      </section>
    </div>

    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存布局</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminBuilding } from '../api'
import { apiUpdateBuilding } from '../api'
import {
  CELL_TYPES,
  buildPlanFromLayout,
  countRooms,
  defaultLayout,
  emptyLayout,
  parseLayout,
  serializeLayout,
  type CellCode,
  type LayoutGrid,
} from '../utils/layoutGrid'

const props = defineProps<{ modelValue: boolean; building: AdminBuilding | null }>()
const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'saved'): void
}>()

const grid = ref<LayoutGrid>(defaultLayout())
const cols = ref(defaultLayout().cols)
const rows = ref(defaultLayout().rows)
const brush = ref<CellCode>('1')
const painting = ref(false)
const saving = ref(false)

watch(
  () => [props.modelValue, props.building?.id] as const,
  () => {
    if (!props.modelValue) return
    const parsed = parseLayout(props.building?.layoutJson)
    const base = parsed || defaultLayout()
    grid.value = { version: base.version, cols: base.cols, rows: base.rows, cells: [...base.cells] }
    cols.value = base.cols
    rows.value = base.rows
    brush.value = '1'
  },
  { immediate: true },
)

const flatCells = computed(() => {
  const numbers = new Map<string, string>()
  const plan = buildPlanFromLayout(grid.value, 1)
  plan.rooms.forEach((room, i) => {
    const c = i % grid.value.cols
    const r = Math.floor(i / grid.value.cols)
    numbers.set(`${r}-${c}`, String(room.no).slice(1))
  })
  let index = 0
  const out: { code: string; no: string; snapshot: boolean }[] = []
  for (let r = 0; r < grid.value.rows; r++) {
    const line = (grid.value.cells[r] || '').padEnd(grid.value.cols, '0')
    for (let c = 0; c < grid.value.cols; c++) {
      const code = line[c] || '0'
      index++
      out.push({ code, no: numbers.get(`${r}-${c}`) || String(index), snapshot: code === '1' })
    }
  }
  return out
})

const roomCount = computed(() => countRooms(grid.value))

function typeCount(code: CellCode) {
  let total = 0
  for (const line of grid.value.cells) {
    for (const ch of line) if (ch === code) total++
  }
  return total
}

function paint(index: number) {
  painting.value = true
  const r = Math.floor(index / grid.value.cols)
  const c = index % grid.value.cols
  const line = (grid.value.cells[r] || '').padEnd(grid.value.cols, '0')
  const next = line.slice(0, c) + brush.value + line.slice(c + 1)
  grid.value.cells[r] = next
}

function applySize() {
  const c = Math.max(4, Math.min(16, Number(cols.value) || 6))
  const r = Math.max(4, Math.min(20, Number(rows.value) || 10))
  const cells: string[] = []
  for (let i = 0; i < r; i++) {
    const old = (grid.value.cells[i] || '').padEnd(c, '0').slice(0, c)
    cells.push(old)
  }
  grid.value = { version: grid.value.version, cols: c, rows: r, cells }
  cols.value = c
  rows.value = r
}

function loadTemplate() {
  const tpl = defaultLayout(cols.value, rows.value)
  grid.value = tpl
}

function clearAll() {
  grid.value = emptyLayout(cols.value, rows.value)
}

// 清除自定义布局：楼栋回到内置标准层模板渲染
async function clearCustom() {
  if (!props.building) return
  try {
    await ElMessageBox.confirm('清除后该楼栋将回到内置标准层模板渲染，是否继续？', '清除自定义布局')
  } catch {
    return
  }
  saving.value = true
  try {
    await apiUpdateBuilding(props.building.id, { ...props.building, layoutJson: '' })
    ElMessage.success('已清除自定义布局')
    emit('saved')
    emit('update:modelValue', false)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

async function save() {
  if (!props.building) return
  if (roomCount.value === 0) {
    ElMessage.warning('请至少绘制一间房间')
    return
  }
  saving.value = true
  try {
    await apiUpdateBuilding(props.building.id, {
      code: props.building.code,
      name: props.building.name,
      posX: props.building.posX,
      posY: props.building.posY,
      width: props.building.width,
      height: props.building.height,
      floors: props.building.floors,
      floorHeight: props.building.floorHeight,
      roomsPerFloor: roomCount.value,
      layoutJson: serializeLayout(grid.value),
    })
    ElMessage.success(`布局已保存：每层 ${roomCount.value} 间房`)
    emit('saved')
    emit('update:modelValue', false)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.designer {
  display: grid;
  grid-template-columns: 210px 1fr;
  gap: 18px;
}

.tools {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px;
  background: var(--rv-grad-2);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 16px;
}

.tool-title {
  margin: 6px 0 2px;
  color: #5a6a85;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.tool-btn {
  display: grid;
  grid-template-columns: 14px 1fr;
  gap: 8px;
  align-items: center;
  padding: 8px 10px;
  color: #33415c;
  font-size: 13px;
  text-align: left;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.22s ease;
}

.tool-btn em {
  grid-column: 2;
  color: #94a3b8;
  font-size: 11px;
  font-style: normal;
}

.tool-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 18px rgba(46, 68, 112, 0.12);
}

.tool-btn.active {
  color: #2462d9;
  background: var(--rv-grad-1);
  box-shadow: 0 8px 18px rgba(52, 120, 246, 0.18);
}

.chip {
  width: 14px;
  height: 14px;
  border: 1px solid #9fb2d4;
  border-radius: 4px;
}

.size-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.times {
  color: #94a3b8;
}

.canvas-area {
  min-width: 0;
}

.hint-line {
  margin-bottom: 10px;
  color: #5a6a85;
  font-size: 12px;
}

.grid {
  display: grid;
  gap: 3px;
  padding: 10px;
  background: linear-gradient(135deg, #eef3fc, #e6ecfa);
  border: 1px solid #c8d5ea;
  border-radius: 14px;
  max-height: 52vh;
  overflow: auto;
  touch-action: none;
}

.cell {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  color: #3d5687;
  font-size: 12px;
  font-weight: 700;
  background: transparent;
  border: 1px dashed rgba(150, 170, 205, 0.55);
  border-radius: 6px;
  cursor: pointer;
  transition: transform 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;
}

.cell.c1 {
  background: linear-gradient(135deg, #dcebff, #c7dcf7);
  border: 1px solid #5f7bb5;
  cursor: pointer;
}

.cell.c2 {
  background: linear-gradient(180deg, #f3f7fd, #e9f0fa);
  border: 1px dashed #a9b8d4;
}

.cell.c3 {
  background: linear-gradient(135deg, #e6ecf7, #d3dcec);
  border: 1px solid #7d8db3;
}

.cell.c4 {
  background: linear-gradient(135deg, #eaf2fd, #dde8f8);
  border: 1px dashed #9aa9c6;
}

.cell:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(46, 68, 112, 0.16);
}

.cell-no {
  pointer-events: none;
}

.cell-tag {
  color: #64748b;
  font-size: 10px;
  pointer-events: none;
}

.summary {
  display: flex;
  gap: 18px;
  margin-top: 12px;
  color: #5a6a85;
  font-size: 13px;
}

.summary b {
  color: #2462d9;
}

.tips {
  margin: 8px 0 0;
  color: #94a3b8;
  font-size: 12px;
}
</style>
