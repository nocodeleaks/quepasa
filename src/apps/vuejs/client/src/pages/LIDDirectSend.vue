<template>
  <div class="lid-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
        </svg>
        {{ t('back') }}
      </button>
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
          </svg>
          {{ t('lid_direct_title') }}
        </h1>
        <p>{{ t('lid_direct_subtitle') }}</p>
      </div>
    </div>

    <div class="send-card">
      <div v-if="error" class="error-box">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
        </svg>
        <span>{{ error }}</span>
      </div>

      <div v-if="success" class="success-box">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 7l-9 9z"/>
        </svg>
        <div>
          <strong>{{ t('lid_direct_success_title') }}</strong>
          <div class="success-details">
            <span>{{ t('lid_direct_message_id_label') }}: {{ success.id }}</span>
            <span>{{ t('lid_direct_chatid_label') }}: {{ success.chatid }}</span>
            <span v-if="success.trackid">{{ t('lid_direct_trackid_label_short') }}: {{ success.trackid }}</span>
          </div>
        </div>
      </div>

      <form @submit.prevent="sendDirectLid">
        <div class="form-group">
          <label for="chatid">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
            </svg>
            {{ t('lid_direct_recipient_label') }}
          </label>
          <input
            id="chatid"
            v-model="chatid"
            type="text"
            class="form-input"
            :placeholder="t('lid_direct_recipient_placeholder')"
            required
          />
        </div>

        <div class="form-group">
          <label for="text">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
            </svg>
            {{ t('lid_direct_text_label') }}
          </label>
          <textarea
            id="text"
            v-model="text"
            class="form-textarea"
            rows="5"
            required
          ></textarea>
        </div>

        <div class="form-group">
          <label for="inreply">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M10 9V5l-7 7 7 7v-4.1c5 0 8.5 1.6 11 5.1-1-5-4-10-11-11z"/>
            </svg>
            {{ t('lid_direct_inreply_label') }}
          </label>
          <input
            id="inreply"
            v-model="inreply"
            type="text"
            class="form-input"
          />
        </div>

        <div class="form-group">
          <label for="trackid">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M17.63 5.84C17.27 5.33 16.67 5 16 5L5 5.01C3.9 5.01 3 5.9 3 7v10c0 1.1.9 1.99 2 1.99L16 19c.67 0 1.27-.33 1.63-.84L22 12l-4.37-6.16z"/>
            </svg>
            {{ t('lid_direct_trackid_label') }}
          </label>
          <input
            id="trackid"
            v-model="trackid"
            type="text"
            class="form-input"
          />
        </div>

        <div class="form-actions">
          <button type="submit" class="btn-send" :disabled="sending">
            <span v-if="sending" class="spinner"></span>
            <svg v-else viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
            </svg>
            {{ sending ? t('lid_direct_sending') : t('lid_direct_send_button') }}
          </button>

          <RouterLink :to="`/server/${encodeURIComponent(token)}/lid/mappings`" class="btn-secondary">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
            </svg>
            {{ t('lid_direct_open_mappings') }}
          </RouterLink>
        </div>
      </form>

      <div v-if="rawResponse" class="raw-box">
        <strong>{{ t('raw_response') }}</strong>
        <pre>{{ rawResponse }}</pre>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useLocale } from '@/i18n'
import api from '@/services/api'

export default defineComponent({
  setup() {
    const route = useRoute()
    const { t } = useLocale()
    const token = ref(String(route.params.token || ''))
    const chatid = ref('')
    const text = ref('')
    const inreply = ref('')
    const trackid = ref('')
    const sending = ref(false)
    const error = ref('')
    const rawResponse = ref('')
    const success = ref<any | null>(null)

    const sendDirectLid = async () => {
      sending.value = true
      error.value = ''
      success.value = null
      rawResponse.value = ''

      try {
        const payload: Record<string, string> = {
          token: token.value,
          chatid: chatid.value.trim(),
          text: text.value,
        }

        if (inreply.value.trim()) payload.inreply = inreply.value.trim()
        if (trackid.value.trim()) payload.trackid = trackid.value.trim()

        const res = await api.post('/api/messages/lid/direct', payload)
        rawResponse.value = JSON.stringify(res.data, null, 2)

        const result = res.data?.result || res.data?.data || {}
        success.value = {
          id: result?.id || '',
          chatid: result?.chatid || chatid.value,
          trackid: result?.trackid || trackid.value,
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.response?.data?.message || err?.message || t('lid_direct_error_send')
      } finally {
        sending.value = false
      }
    }

    return {
      t,
      token,
      chatid,
      text,
      inreply,
      trackid,
      sending,
      error,
      rawResponse,
      success,
      sendDirectLid,
    }
  },
})
</script>

<style scoped>
.lid-page {
  max-width: 680px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
}

.header-content h1 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 6px;
}

.header-content h1 svg {
  color: var(--branding-primary, #7C3AED);
}

.header-content p {
  color: #6b7280;
  margin: 0;
  font-size: 14px;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #334155;
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 10px;
  padding: 6px 12px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.back-link:hover {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #312e81;
}

.send-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
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
  margin-bottom: 16px;
}

.success-box {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 16px;
  background: #f5efff;
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 12px;
  color: var(--branding-secondary, #5B21B6);
  margin-bottom: 16px;
}

.success-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-top: 4px;
  font-size: 13px;
  opacity: 0.85;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.form-group label svg {
  color: #6b7280;
}

.form-input {
  width: 100%;
  padding: 14px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 16px;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
}

.form-textarea {
  width: 100%;
  padding: 14px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 16px;
  resize: vertical;
  min-height: 120px;
  box-sizing: border-box;
  font-family: inherit;
  transition: all 0.2s;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
}

.form-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 8px;
}

.btn-send {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 14px 20px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-send:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(124, 58, 237, 0.25);
}

.btn-send:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 14px 20px;
  background: #f3f4f6;
  border: 2px solid transparent;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  color: #374151;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  flex-shrink: 0;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.raw-box {
  margin-top: 20px;
  background: #0b1220;
  color: #e5e7eb;
  padding: 16px;
  border-radius: 12px;
  font-size: 13px;
}

.raw-box pre {
  margin: 8px 0 0;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
