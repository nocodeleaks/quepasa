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
            <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
          </svg>
          {{ t('lid_mappings_title') }}
        </h1>
        <p>{{ t('lid_mappings_subtitle') }}</p>
      </div>
    </div>

    <div v-if="error" class="error-box">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <div class="grid">
      <!-- LID → Phone -->
      <div class="send-card">
        <h2>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
          </svg>
          {{ t('lid_mappings_lid_to_phone_title') }}
        </h2>

        <form @submit.prevent="lookupByLid">
          <div class="form-group">
            <label for="lid">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M17.63 5.84C17.27 5.33 16.67 5 16 5L5 5.01C3.9 5.01 3 5.9 3 7v10c0 1.1.9 1.99 2 1.99L16 19c.67 0 1.27-.33 1.63-.84L22 12l-4.37-6.16z"/>
              </svg>
              {{ t('lid_mappings_lid_label') }}
            </label>
            <input
              id="lid"
              v-model="lid"
              type="text"
              class="form-input"
              :placeholder="t('lid_mappings_lid_placeholder')"
              required
            />
          </div>

          <button type="submit" class="btn-send" :disabled="loadingLid">
            <span v-if="loadingLid" class="spinner"></span>
            <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M20.01 15.38c-1.23 0-2.42-.2-3.53-.56-.35-.12-.74-.03-1.01.24l-1.57 1.97c-2.83-1.35-5.48-3.9-6.89-6.83l1.95-1.66c.27-.28.35-.67.24-1.02-.37-1.11-.56-2.3-.56-3.53 0-.54-.45-.99-.99-.99H4.19C3.65 3 3 3.24 3 3.99 3 13.28 10.73 21 20.01 21c.71 0 .99-.63.99-1.18v-3.45c0-.54-.45-.99-.99-.99z"/>
            </svg>
            {{ loadingLid ? t('lid_mappings_loading') : t('lid_mappings_lookup_phone') }}
          </button>
        </form>

        <div v-if="lidResult" class="result-box">
          <div class="result-row">
            <span class="result-label">{{ t('lid_mappings_lid_label') }}</span>
            <span class="result-value">{{ lidResult.lid || lid }}</span>
          </div>
          <div class="result-row">
            <span class="result-label">{{ t('lid_mappings_phone_label') }}</span>
            <span class="result-value">{{ lidResult.phone || '-' }}</span>
          </div>
        </div>
      </div>

      <!-- Phone → LID -->
      <div class="send-card">
        <h2>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M6.62 10.79c1.44 2.83 3.76 5.14 6.59 6.59l2.2-2.2c.27-.27.67-.36 1.02-.24 1.12.37 2.33.57 3.57.57.55 0 1 .45 1 1V20c0 .55-.45 1-1 1-9.39 0-17-7.61-17-17 0-.55.45-1 1-1h3.5c.55 0 1 .45 1 1 0 1.25.2 2.45.57 3.57.11.35.03.74-.25 1.02l-2.2 2.2z"/>
          </svg>
          {{ t('lid_mappings_phone_to_lid_title') }}
        </h2>

        <form @submit.prevent="lookupByPhone">
          <div class="form-group">
            <label for="phone">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M20.01 15.38c-1.23 0-2.42-.2-3.53-.56-.35-.12-.74-.03-1.01.24l-1.57 1.97c-2.83-1.35-5.48-3.9-6.89-6.83l1.95-1.66c.27-.28.35-.67.24-1.02-.37-1.11-.56-2.3-.56-3.53 0-.54-.45-.99-.99-.99H4.19C3.65 3 3 3.24 3 3.99 3 13.28 10.73 21 20.01 21c.71 0 .99-.63.99-1.18v-3.45c0-.54-.45-.99-.99-.99z"/>
              </svg>
              {{ t('lid_mappings_phone_label') }}
            </label>
            <input
              id="phone"
              v-model="phone"
              type="text"
              class="form-input"
              :placeholder="t('lid_mappings_phone_placeholder')"
              required
            />
          </div>

          <button type="submit" class="btn-send" :disabled="loadingPhone">
            <span v-if="loadingPhone" class="spinner"></span>
            <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M17.63 5.84C17.27 5.33 16.67 5 16 5L5 5.01C3.9 5.01 3 5.9 3 7v10c0 1.1.9 1.99 2 1.99L16 19c.67 0 1.27-.33 1.63-.84L22 12l-4.37-6.16z"/>
            </svg>
            {{ loadingPhone ? t('lid_mappings_loading') : t('lid_mappings_lookup_lid') }}
          </button>
        </form>

        <div v-if="phoneResult" class="result-box">
          <div class="result-row">
            <span class="result-label">{{ t('lid_mappings_phone_label') }}</span>
            <span class="result-value">{{ phoneResult.phone || phone }}</span>
          </div>
          <div class="result-row">
            <span class="result-label">{{ t('lid_mappings_lid_label') }}</span>
            <span class="result-value">{{ phoneResult.lid || '-' }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="rawResponse" class="raw-box">
      <strong>{{ t('raw_response') }}</strong>
      <pre>{{ rawResponse }}</pre>
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

    const lid = ref('')
    const phone = ref('')
    const loadingLid = ref(false)
    const loadingPhone = ref(false)

    const lidResult = ref<any | null>(null)
    const phoneResult = ref<any | null>(null)
    const error = ref('')
    const rawResponse = ref('')

    const parsePayload = (data: any) => ({
      lid: data?.lid || data?.LID || data?.result?.lid || data?.result?.LID || '',
      phone: data?.phone || data?.result?.phone || '',
    })

    const lookupByLid = async () => {
      loadingLid.value = true
      error.value = ''
      try {
        const res = await api.get('/api/contacts/identifier', {
          params: {
            token: token.value,
            lid: lid.value.trim(),
          },
        })

        rawResponse.value = JSON.stringify(res.data, null, 2)
        lidResult.value = parsePayload(res.data)
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.response?.data?.message || err?.message || t('lid_mappings_error_lid')
      } finally {
        loadingLid.value = false
      }
    }

    const lookupByPhone = async () => {
      loadingPhone.value = true
      error.value = ''
      try {
        const res = await api.get('/api/contacts/identifier', {
          params: {
            token: token.value,
            phone: phone.value.trim(),
          },
        })

        rawResponse.value = JSON.stringify(res.data, null, 2)
        phoneResult.value = parsePayload(res.data)
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.response?.data?.message || err?.message || t('lid_mappings_error_phone')
      } finally {
        loadingPhone.value = false
      }
    }

    return {
      t,
      token,
      lid,
      phone,
      loadingLid,
      loadingPhone,
      lidResult,
      phoneResult,
      error,
      rawResponse,
      lookupByLid,
      lookupByPhone,
    }
  },
})
</script>

<style scoped>
.lid-page {
  max-width: 1100px;
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

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 24px;
}

.send-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.send-card h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
  margin: 0 0 20px;
}

.send-card h2 svg {
  color: var(--branding-primary, #7C3AED);
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

.btn-send {
  width: 100%;
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

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  flex-shrink: 0;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.result-box {
  margin-top: 16px;
  background: #f5efff;
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 12px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.result-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.result-label {
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  flex-shrink: 0;
  min-width: 60px;
}

.result-value {
  font-size: 14px;
  color: #111827;
  font-weight: 500;
  word-break: break-all;
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
  margin-bottom: 20px;
}

.raw-box {
  margin-top: 24px;
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
