import { ref } from 'vue'
import api from '@/services/api'

const user = ref<any>(null)
const version = ref('')
const loading = ref(true)
const error = ref('')

async function loadSession() {
  try {
    loading.value = true
    error.value = ''
    const res = await api.get('/api/session')
    user.value = res.data.user
    version.value = res.data.version
  } catch (err: any) {
    user.value = null
    version.value = ''
    error.value = err?.response?.data?.result || err.message || 'Session error'
  } finally {
    loading.value = false
  }
}

function clearSession() {
  user.value = null
  version.value = ''
}

export function useSessionStore() {
  return { user, version, loading, error, loadSession, clearSession }
}
