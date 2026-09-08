<template>
  <div>
    <van-nav-bar title="宿管 · 本栋工单" right-text="退出" @click-right="emit('logout')" />
    <div class="toolbar">
      <van-button type="primary" size="small" icon="plus" @click="openCreate">极简报修</van-button>
      <van-button size="small" icon="replay" @click="load">刷新</van-button>
    </div>

    <van-empty v-if="!loading && orders.length === 0" description="本栋暂无工单" />
    <van-card
      v-for="item in orders"
      :key="item.id"
      class="order-card"
      :title="item.title"
      :desc="`${item.buildingName || ''} ${item.floor}层 ${item.room}室 · ${item.createdAt}`"
    >
      <template #tags>
        <van-tag :type="statusType(item.status)" plain>{{ item.statusText }}</van-tag>
        <van-tag plain>{{ item.faultTypeName }}</van-tag>
        <van-tag v-if="item.workerName" plain>工人：{{ item.workerName }}</van-tag>
      </template>
      <template #footer>
        <van-button v-if="canCancel(item)" size="mini" type="danger" @click="cancelOrder(item)">
          取消工单
        </van-button>
      </template>
    </van-card>

    <van-popup v-model:show="showCreate" position="bottom" round class="create-popup">
      <div class="popup-inner">
        <van-nav-bar title="极简报修" left-arrow @click-left="showCreate = false" />
        <van-field
          readonly
          is-link
          :model-value="selectedFaultName"
          label="维修类型"
          placeholder="请选择维修类型"
          @click="showFault = true"
        />
        <van-field v-model.number="createForm.floor" type="number" label="楼层" placeholder="如 3" />
        <van-field v-model="createForm.room" label="房间号" placeholder="如 301" />
        <van-field
          v-model="createForm.description"
          rows="2"
          autosize
          type="textarea"
          label="描述"
          placeholder="故障描述（可选）"
        />
        <div class="popup-btn">
          <van-button round block type="primary" :loading="submitting" @click="submitCreate">
            提交报修
          </van-button>
        </div>
      </div>
    </van-popup>

    <van-popup v-model:show="showFault" position="bottom" round>
      <van-nav-bar title="选择维修类型" left-arrow @click-left="showFault = false" />
      <van-radio-group v-model="createForm.faultType">
        <van-cell
          v-for="ft in faultTypes"
          :key="ft.code"
          :title="ft.name"
          clickable
          @click="chooseFault(ft.code)"
        >
          <template #right-icon>
            <van-radio :name="ft.code" />
          </template>
        </van-cell>
      </van-radio-group>
    </van-popup>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import type { FaultType, OrderItem } from '../api'
import { apiCancelOrder, apiCreateOrder, apiDormOrders, apiFaultTypes } from '../api'

const emit = defineEmits<{ (e: 'logout'): void }>()

const orders = ref<OrderItem[]>([])
const faultTypes = ref<FaultType[]>([])
const loading = ref(false)
const submitting = ref(false)
const showCreate = ref(false)
const showFault = ref(false)
const createForm = reactive({ faultType: '', floor: 1, room: '', description: '' })

const selectedFaultName = computed(() => {
  const ft = faultTypes.value.find((f) => f.code === createForm.faultType)
  return ft ? ft.name : ''
})

async function load() {
  loading.value = true
  try {
    orders.value = await apiDormOrders(0)
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  if (faultTypes.value.length === 0) {
    try {
      faultTypes.value = await apiFaultTypes()
    } catch (err) {
      showToast((err as Error).message)
      return
    }
  }
  createForm.faultType = faultTypes.value[0]?.code || ''
  createForm.floor = 1
  createForm.room = ''
  createForm.description = ''
  showCreate.value = true
}

function chooseFault(code: string) {
  createForm.faultType = code
  showFault.value = false
}

async function submitCreate() {
  if (!createForm.faultType || !createForm.room || !createForm.floor) {
    showToast('请完整填写维修信息')
    return
  }
  submitting.value = true
  try {
    await apiCreateOrder({
      faultType: createForm.faultType,
      floor: Number(createForm.floor),
      room: createForm.room.trim(),
      description: createForm.description.trim(),
    })
    showToast('报修成功，已自动派单')
    showCreate.value = false
    load()
  } catch (err) {
    showToast((err as Error).message)
  } finally {
    submitting.value = false
  }
}

function canCancel(item: OrderItem) {
  return item.status === 1 || item.status === 2
}

async function cancelOrder(item: OrderItem) {
  try {
    await showConfirmDialog({ title: '取消工单', message: `确认取消 ${item.orderNo}？` })
  } catch {
    return
  }
  try {
    await apiCancelOrder(item.id)
    showToast('已取消')
    load()
  } catch (err) {
    showToast((err as Error).message)
  }
}

function statusType(status: number) {
  if (status === 4) return 'success'
  if (status === 5) return 'default'
  if (status === 3) return 'primary'
  return 'warning'
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  padding: 8px 12px;
}
.order-card {
  margin: 8px 12px;
}
.create-popup {
  max-height: 80vh;
}
.popup-inner {
  padding-bottom: 20px;
}
.popup-btn {
  margin: 16px;
}
</style>