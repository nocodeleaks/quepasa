import { reactive } from 'vue'

export type Toast = {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

const toasts = reactive<Toast[]>([])
let idCounter = 1

export function pushToast(message: string, type: 'success' | 'error' | 'info' = 'success', timeout = 4000) {
  const id = idCounter++
  toasts.push({ id, type, message })
  setTimeout(() => {
    const idx = toasts.findIndex((t) => t.id === id)
    if (idx !== -1) toasts.splice(idx, 1)
  }, timeout)
}

export function useToasts() {
  return toasts
}

export default { pushToast, useToasts }
