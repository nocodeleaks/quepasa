import { ref, readonly } from 'vue'
import api from '@/services/api'

// Matches library.HeaderMasterKey on the Go backend.
const MASTER_KEY_HEADER = 'X-QUEPASA-MASTERKEY'

// In-memory store: cleared on page refresh, never persisted to localStorage or cookies.
const masterKey = ref<string | null>(null)
const verifying = ref(false)
const verifyError = ref('')

function isMasterAuthenticated(): boolean {
  return masterKey.value !== null && masterKey.value !== ''
}

function getMasterKey(): string | null {
  return masterKey.value
}

/**
 * Sends the candidate key to the backend for validation.
 * On success, stores the key in memory so subsequent API calls can include it.
 * Returns true when the key is valid.
 */
async function verifyMasterKey(candidate: string): Promise<boolean> {
  verifying.value = true
  verifyError.value = ''
  try {
    const res = await api.post('/api/master/verify', { key: candidate })
    if (res.data?.valid) {
      masterKey.value = candidate
      return true
    }
    verifyError.value = 'Invalid master key'
    return false
  } catch (err: any) {
    verifyError.value = err?.response?.data?.result || 'Verification failed'
    return false
  } finally {
    verifying.value = false
  }
}

function clearMasterKey(): void {
  masterKey.value = null
  verifyError.value = ''
}

/**
 * Returns an axios-compatible headers object with the master key header included,
 * or an empty object when no master key is in memory.
 */
function masterKeyHeaders(): Record<string, string> {
  if (!masterKey.value) return {}
  return { [MASTER_KEY_HEADER]: masterKey.value }
}

export function useMasterKey() {
  return {
    masterKey: readonly(masterKey),
    verifying: readonly(verifying),
    verifyError: readonly(verifyError),
    isMasterAuthenticated,
    getMasterKey,
    verifyMasterKey,
    clearMasterKey,
    masterKeyHeaders,
  }
}
