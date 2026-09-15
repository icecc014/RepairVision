import { showConfirmDialog, showToast } from 'vant'
import { apiStartOrder } from '../api'

/**
 * startOrderFlow 统一处理"非工作时段开工"：
 * 先按正常流程开工；若后端因不在工作时段拦截（409），提示工人确认后带 force 强制开工。
 * 返回 true 表示已开工（含强制），false 表示未开工。
 */
export async function startOrderFlow(id: number): Promise<boolean> {
  try {
    const res = (await apiStartOrder(id, false)) as { warning?: string } | undefined
    if (res?.warning) {
      showToast(res.warning)
    }
    return true
  } catch (err) {
    const msg = (err as Error).message || '开工失败'
    if (!msg.includes('工作时段')) {
      showToast(msg)
      return false
    }
    try {
      await showConfirmDialog({
        title: '当前不在工作时段',
        message: msg,
        confirmButtonText: '强制开工',
        cancelButtonText: '稍后再说',
      })
    } catch {
      return false
    }
    try {
      const res = (await apiStartOrder(id, true)) as { warning?: string } | undefined
      showToast(res?.warning || '已强制开工')
      return true
    } catch (err2) {
      showToast((err2 as Error).message)
      return false
    }
  }
}