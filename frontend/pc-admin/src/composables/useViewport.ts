import { onBeforeUnmount, onMounted, ref } from 'vue'

// V9.7 管理端移动自适应：视口断点侦测（原生 matchMedia，不引入新依赖）
const MOBILE_QUERY = '(max-width: 767px)'

function matchNow(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
  return window.matchMedia(MOBILE_QUERY).matches
}

/**
 * 侦测当前是否处于移动档位（≤767px）。
 * 初始值在 setup 阶段同步求值，避免手机端先渲染宽屏结构再切换造成闪烁。
 */
export function useIsMobile() {
  const isMobile = ref(matchNow())
  let mql: MediaQueryList | null = null
  let onChange: ((e: MediaQueryListEvent) => void) | null = null

  onMounted(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return
    mql = window.matchMedia(MOBILE_QUERY)
    isMobile.value = mql.matches
    onChange = (e: MediaQueryListEvent) => { isMobile.value = e.matches }
    mql.addEventListener('change', onChange)
  })

  onBeforeUnmount(() => {
    if (mql && onChange) mql.removeEventListener('change', onChange)
    mql = null
    onChange = null
  })

  return { isMobile }
}