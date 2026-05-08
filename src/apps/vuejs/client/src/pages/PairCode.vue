<template>
  <div class="paircode-page">
    <div class="page-header">
      <div class="header-content">
        <button @click="$router.back()" class="back-link hide-mobile">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
          </svg>
          {{ t('back') }}
        </button>
        <h1>{{ t('paircode_title') }}</h1>
        <p>{{ t('paircode_subtitle') }}</p>
      </div>
    </div>

    <div class="content-card">
      <div v-if="error" class="error-box">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
        </svg>
        <span>{{ error }}</span>
      </div>

      <div v-if="connected" class="success-section">
        <div class="success-icon">
          <svg viewBox="0 0 24 24" width="80" height="80" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L18 9l-9 9z"/>
          </svg>
        </div>
        <h3>{{ t('paircode_connected_title') }}</h3>
        <p>{{ t('paircode_connected_desc') }}</p>
        <router-link :to="`/server/${token}`" class="btn-success">
          {{ t('paircode_go_server') }}
        </router-link>
      </div>

      <div v-else>
        <!-- Step 1: Enter phone number -->
        <div v-if="!pairCode" class="step-section">
          <div class="step-header">
            <div class="step-number">1</div>
            <div class="step-info">
              <h3>{{ t('paircode_step1_title') }}</h3>
              <p>{{ t('paircode_step1_desc') }}</p>
            </div>
          </div>

          <div class="phone-input-group">
            <div class="input-wrapper">
              <span class="input-prefix">+</span>
              <input 
                v-model="phone"
                type="tel"
                class="phone-input"
                :placeholder="t('paircode_phone_placeholder')"
                @keyup.enter="generateCode"
              />
            </div>
            <button @click="generateCode" class="btn-primary" :disabled="loading || !phone">
              <span v-if="loading" class="spinner"></span>
              <span v-else>{{ t('paircode_generate') }}</span>
            </button>
          </div>
        </div>

        <!-- Step 2: Show pair code -->
        <div v-if="pairCode" class="step-section">
          <div class="step-header">
            <div class="step-number completed">✓</div>
            <div class="step-info">
              <h3>{{ t('paircode_step2_number_prefix') }}+{{ phone }}</h3>
              <button @click="resetCode" class="btn-link">{{ t('paircode_change_number') }}</button>
            </div>
          </div>
        </div>

        <div v-if="pairCode" class="step-section">
          <div class="step-header">
            <div class="step-number">2</div>
            <div class="step-info">
              <h3>{{ t('paircode_step2_title') }}</h3>
              <p>{{ t('paircode_step2_desc') }}</p>
            </div>
          </div>

          <div class="code-display">
            <div class="code-box">
              <span v-for="(char, i) in formattedCode" :key="i" class="code-char" :class="{ separator: char === '-' }">
                {{ char }}
              </span>
            </div>
            <button @click="copyCode" class="btn-copy" :class="{ copied }">
              <svg v-if="!copied" viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
              </svg>
              {{ copied ? t('paircode_copied') : t('paircode_copy') }}
            </button>
          </div>

          <div class="code-timer">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm.5-13H11v6l5.25 3.15.75-1.23-4.5-2.67z"/>
            </svg>
            <span>{{ t('paircode_expires') }}</span>
          </div>

          <div class="action-buttons">
            <button @click="confirmPairing" class="btn-confirm" :disabled="checking">
              <span v-if="checking" class="spinner"></span>
              <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
              </svg>
              {{ checking ? t('paircode_checking') : t('paircode_already_done') }}
            </button>

            <button @click="generateCode" class="btn-secondary">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
              </svg>
              {{ t('paircode_new_code') }}
            </button>
          </div>
        </div>

        <div class="instructions-box">
          <h4>{{ t('paircode_help_title') }}</h4>
          <ol>
            <li>{{ t('paircode_help_step1') }}</li>
            <li>{{ t('paircode_help_step2') }}</li>
            <li>{{ t('paircode_help_step3') }}</li>
            <li>{{ t('paircode_help_step4') }}</li>
            <li>{{ t('paircode_help_step5') }}</li>
            <li>{{ t('paircode_help_step6') }}</li>
          </ol>
        </div>

        <div class="alt-method">
          <p>{{ t('paircode_alt_qr') }}</p>
          <router-link :to="`/server/${token}/qrcode`" class="link-primary">
            {{ t('paircode_qr_link') }}
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/services/api'
import { useCableSubscription } from '@/composables/useCableSubscription'
import { useLocale } from '@/i18n'

export default defineComponent({
  setup() {
    const route = useRoute()
    const token = route.params.token as string
    const { t } = useLocale()

    const phone = ref('')
    const pairCode = ref('')
    const loading = ref(false)
    const checking = ref(false)
    const error = ref('')
    const copied = ref(false)
    const connected = ref(false)

    let fallbackTimer: ReturnType<typeof setTimeout> | null = null

    const formattedCode = computed(() => {
      if (!pairCode.value) return []
      // Format as XXXX-XXXX (WhatsApp alphanumeric format)
      // Accept letters and digits, ignore other characters, then split into 4+4
      const code = pairCode.value.replace(/[^0-9A-Za-z]/g, '')
      if (code.length >= 8) {
        return (code.substring(0, 4) + '-' + code.substring(4, 8)).split('')
      }
      return code.split('')
    })

    async function generateCode() {
      if (!phone.value) return

      loading.value = true
      error.value = ''
      connected.value = false
      stopFallbackCheck()

      try {
        const res = await api.get('/api/session/paircode', {
          params: { token, phone: phone.value }
        })

        
        if (res.data?.connected || res.data?.server?.state === 'Ready' || res.data?.server?.stateCode === 11) {
          connected.value = true
          return
        }

        if (res.data?.pairCode || res.data?.paircode || res.data?.formatted) {
          pairCode.value = res.data?.pairCode || res.data?.paircode || res.data?.formatted
          scheduleFallbackCheck()
        } else {
          error.value = res.data?.result || t('paircode_error_generate')
        }
      } catch (err: any) {
        const msg = err?.response?.data?.result || err?.response?.data?.message || err.message || t('paircode_error_generate')
        error.value = msg
      } finally {
        loading.value = false
      }
    }

    function resetCode() {
      pairCode.value = ''
      stopFallbackCheck()
    }

    async function copyCode() {
      if (!pairCode.value) return
      try {
        await navigator.clipboard.writeText(pairCode.value)
        copied.value = true
        setTimeout(() => copied.value = false, 2000)
      } catch {
        // fallback
      }
    }

    async function checkConnection() {
      try {
        const res = await api.post('/api/sessions/get', { token })

        // Check the explicit connected field or state
        const isConnected = res.data?.server?.state === 'Ready' || res.data?.server?.stateCode === 11
        
        if (isConnected) {
          connected.value = true
          stopFallbackCheck()
          return true
        }
      } catch {
        // ignore errors
      }
      return false
    }

    async function confirmPairing() {
      checking.value = true
      error.value = ''
      
      // Check multiple times with short intervals
      for (let i = 0; i < 5; i++) {
        const isConnected = await checkConnection()
        if (isConnected) {
          checking.value = false
          return
        }
        // Wait 1 second between checks
        await new Promise(resolve => setTimeout(resolve, 1000))
      }
      
      checking.value = false
      error.value = t('paircode_checking')
    }

    function scheduleFallbackCheck(delay = 8000) {
      stopFallbackCheck()
      fallbackTimer = setTimeout(async () => {
        fallbackTimer = null
        if (connected.value || !pairCode.value) {
          return
        }

        await checkConnection()
      }, delay)
    }

    function stopFallbackCheck() {
      if (fallbackTimer) {
        clearTimeout(fallbackTimer)
        fallbackTimer = null
      }
    }

    useCableSubscription(
      [
        {
          event: 'server.connected',
          handler: (payload: any) => {
            if (payload?.token !== token) {
              return
            }

            connected.value = true
            error.value = ''
            stopFallbackCheck()
          },
        },
      ],
      {
        token,
        subscribeToken: true,
        onConnectError: () => {
          // Keep HTTP fallback active if websocket auth fails.
        },
      },
    )

    onUnmounted(() => {
      stopFallbackCheck()
    })

    return { 
      token, phone, pairCode, loading, checking, error, copied, connected,
      formattedCode, generateCode, resetCode, copyCode, confirmPairing, t
    }
  }
})
</script>

<style scoped>
.paircode-page {
  max-width: 600px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.header-content {
  text-align: center;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  text-decoration: none;
  font-size: 14px;
  margin-bottom: 16px;
}

.back-link:hover {
  color: #374151;
}

.page-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 8px;
}

.page-header p {
  color: #6b7280;
  margin: 0;
}

.content-card {
  background: white;
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.error-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  color: #dc2626;
  margin-bottom: 24px;
}

.success-section {
  text-align: center;
  padding: 40px 0;
}

.success-icon {
  color: var(--branding-primary, #7C3AED);
}

.success-section h3 {
  font-size: 24px;
  color: var(--branding-primary, #7C3AED);
  margin: 16px 0 8px;
}

.success-section p {
  color: #6b7280;
  margin: 0 0 24px;
}

.btn-success {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: var(--branding-primary, #7C3AED);
  color: white;
  border-radius: 10px;
  text-decoration: none;
  font-weight: 600;
}

.step-section {
  margin-bottom: 24px;
}

.step-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}

.step-number {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #7C3AED, #5B21B6);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 14px;
  flex-shrink: 0;
}

.step-number.completed {
  background: var(--branding-primary, #7C3AED);
}

.step-info h3 {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
  margin: 0 0 4px;
}

.step-info p {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
}

.btn-link {
  background: none;
  border: none;
  color: #7C3AED;
  font-size: 14px;
  cursor: pointer;
  padding: 0;
}

.btn-link:hover {
  text-decoration: underline;
}

.phone-input-group {
  display: flex;
  gap: 12px;
}

.input-wrapper {
  flex: 1;
  position: relative;
}

.input-prefix {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
  font-weight: 600;
}

.phone-input {
  width: 100%;
  padding: 14px 16px 14px 32px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 18px;
  letter-spacing: 1px;
}

.phone-input:focus {
  outline: none;
  border-color: #7C3AED;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 24px;
  background: linear-gradient(135deg, #7C3AED, #5B21B6);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.code-display {
  text-align: center;
  padding: 24px;
  background: #f0fdf4;
  border-radius: 16px;
  margin-bottom: 16px;
}

.code-box {
  display: flex;
  justify-content: center;
  gap: 4px;
  margin-bottom: 16px;
}

.code-char {
  width: 44px;
  height: 56px;
  background: white;
  border: 2px solid #7C3AED;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  font-weight: 700;
  color: #111827;
}

.code-char.separator {
  width: 20px;
  background: transparent;
  border: none;
  color: #7C3AED;
  font-size: 24px;
}

.btn-copy {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: white;
  border: 2px solid #7C3AED;
  border-radius: 10px;
  color: #7C3AED;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-copy:hover {
  background: #7C3AED;
  color: white;
}

.btn-copy.copied {
  background: var(--branding-primary, #7C3AED);
  border-color: var(--branding-primary, #7C3AED);
  color: white;
}

.code-timer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #6b7280;
  font-size: 14px;
  margin-bottom: 16px;
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: #f3f4f6;
  border: none;
  border-radius: 10px;
  color: #374151;
  font-weight: 600;
  cursor: pointer;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
}

.btn-confirm {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  max-width: 300px;
  padding: 14px 24px;
  background: linear-gradient(135deg, #10B981, #059669);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-confirm:hover:not(:disabled) {
  background: linear-gradient(135deg, #059669, #047857);
  transform: translateY(-1px);
}

.btn-confirm:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.instructions-box {
  padding: 20px;
  background: #fef3c7;
  border-radius: 12px;
  margin: 24px 0;
}

.instructions-box h4 {
  font-size: 16px;
  color: #92400e;
  margin: 0 0 12px;
}

.instructions-box ol {
  margin: 0;
  padding-left: 20px;
  color: #a16207;
}

.instructions-box li {
  margin-bottom: 8px;
}

.alt-method {
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.alt-method p {
  color: #6b7280;
  margin: 0 0 8px;
  font-size: 14px;
}

.link-primary {
  color: #7C3AED;
  text-decoration: none;
  font-weight: 600;
}

.link-primary:hover {
  text-decoration: underline;
}
</style>
