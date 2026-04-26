import { ref } from 'vue'
import api from '@/services/api'

const user = ref<any>(null)
const version = ref('')
const loading = ref(true)
const error = ref('')

type LoadSessionOptions = {
  allowUnauthorized?: boolean
}

async function loadSession(options: LoadSessionOptions = {}) {
  try {
    loading.value = true
    error.value = ''
    const res = await api.get('/spa/session')
    user.value = res.data.user
    version.value = res.data.version
  } catch (err: any) {
    const status = err?.response?.status
    user.value = null
    version.value = ''
    if (options.allowUnauthorized && status === 401) {
      error.value = ''
      return
    }
    error.value = err?.response?.data?.result || err.message || 'Session error'
  } finally {
    loading.value = false
  }
}

function resolveUnauthenticated() {
  user.value = null
  version.value = ''
  error.value = ''
  loading.value = false
}

function clearSession() {
  user.value = null
  version.value = ''
}

export function useSessionStore() {
  return { user, version, loading, error, loadSession, resolveUnauthenticated, clearSession }
}
