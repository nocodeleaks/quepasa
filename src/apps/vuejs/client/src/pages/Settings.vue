<template>
  <div class="settings-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile" type="button">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
        </svg>
        {{ t('back') }}
      </button>
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z" />
          </svg>
          {{ t('settings_title') }}
        </h1>
        <p class="header-desc">{{ t('settings_desc') }}</p>
      </div>
    </div>

    <div v-if="loading" class="loading-placeholder">
      <div class="spinner"></div>
      <span>{{ t('server_loading') }}</span>
    </div>

    <div v-if="error" class="error-banner">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
      </svg>
      <span>{{ error }}</span>
    </div>

    <div class="options-section" v-if="!loading">
      <div class="options-list">
        <div class="option-card history-sync-card">
          <div class="option-info">
            <div class="option-header">
              <i class="fa fa-database option-icon"></i>
              <span class="option-title">{{ t('server_store_retention_title') }}</span>
            </div>
            <p class="option-desc">{{ t('server_store_retention_desc') }}</p>
            <p class="option-desc global-note">{{ t('settings_global_note') }}{{ envRetentionLabel }}</p>
          </div>
          <div class="history-sync-edit">
            <select v-model="storeRetentionMode" class="history-days-input retention-select" :disabled="saving">
              <option value="inherit">{{ t('server_store_retention_inherit') }}</option>
              <option value="none">{{ t('server_store_retention_none') }}</option>
              <option value="forever">{{ t('server_store_retention_forever') }}</option>
              <option value="days">{{ t('server_store_retention_days') }}</option>
            </select>
            <template v-if="storeRetentionMode === 'days'">
              <input type="number" min="1" v-model.number="storeRetentionDays" class="history-days-input" :disabled="saving" />
              <span class="history-days-unit">{{ t('server_store_retention_unit') }}</span>
            </template>
          </div>
        </div>

        <div class="option-card history-sync-card">
          <div class="option-info">
            <div class="option-header">
              <i class="fa fa-filter option-icon"></i>
              <span class="option-title">{{ t('server_dispatch_types_title') }}</span>
            </div>
            <p class="option-desc">{{ t('server_dispatch_types_desc') }}</p>
            <p class="option-desc global-note">{{ t('settings_global_note') }}{{ envDispatchLabel }}</p>
            <div class="dispatch-types-grid">
              <label v-for="dt in dispatchTypeOptions" :key="dt" class="dispatch-type-item">
                <input type="checkbox" :value="dt" v-model="dispatchTypes" :disabled="saving" />
                <span>{{ dt }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>

      <div class="save-bar">
        <button class="btn-primary" @click="save" :disabled="saving">
          <i class="fa fa-check"></i>
          {{ saving ? t('processing') : t('save') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

export default defineComponent({
  setup() {
    const { t } = useLocale()

    const loading = ref(true)
    const saving = ref(false)
    const error = ref('')

    const envRetentionLabel = ref('')
    const envDispatchLabel = ref('')

    // store_retention_days: null=inherit, -1=none, 0=forever, N=days
    const storeRetentionMode = ref<'inherit' | 'none' | 'forever' | 'days'>('inherit')
    const storeRetentionDays = ref<number>(30)

    const dispatchTypeOptions = [
      'text', 'image', 'audio', 'video', 'document', 'sticker', 'location',
      'contact', 'call', 'system', 'group', 'revoke', 'poll', 'view_once', 'unhandled',
    ]
    const dispatchTypes = ref<string[]>([])

    function retentionMode(value: number | null | undefined): 'inherit' | 'none' | 'forever' | 'days' {
      if (value === null || value === undefined) return 'inherit'
      if (value === -1) return 'none'
      if (value === 0) return 'forever'
      return 'days'
    }

    function mappedRetention(): number | null {
      if (storeRetentionMode.value === 'none') return -1
      if (storeRetentionMode.value === 'forever') return 0
      if (storeRetentionMode.value === 'days') return Math.max(1, Math.floor(storeRetentionDays.value || 1))
      return null
    }

    function csvToArray(csv: string | null | undefined): string[] {
      return String(csv || '')
        .split(',')
        .map((s) => s.trim())
        .filter((s) => s.length > 0)
    }

    function applyGlobal(global: any) {
      const retention = global?.store_retention_days
      storeRetentionMode.value = retentionMode(retention)
      if (storeRetentionMode.value === 'days') storeRetentionDays.value = Number(retention)
      dispatchTypes.value = csvToArray(global?.dispatch_types)
    }

    async function load() {
      loading.value = true
      error.value = ''

      try {
        const res = await api.get('/api/settings')
        const env = res.data?.env || {}
        envRetentionLabel.value = String(env.store_retention_days ?? '-')
        envDispatchLabel.value = env.dispatch_types ? String(env.dispatch_types) : '-'
        applyGlobal(res.data?.global || {})
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || t('settings_error_load')
      } finally {
        loading.value = false
      }
    }

    async function save() {
      saving.value = true
      error.value = ''

      try {
        await api.put('/api/settings', {
          store_retention_days: mappedRetention(),
          dispatch_types: dispatchTypes.value.length ? dispatchTypes.value.join(',') : null,
        })
        await load()
        pushToast(t('settings_saved'), 'success')
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || t('settings_error_save')
        pushToast(error.value, 'error')
      } finally {
        saving.value = false
      }
    }

    onMounted(load)

    return {
      t,
      loading,
      saving,
      error,
      envRetentionLabel,
      envDispatchLabel,
      storeRetentionMode,
      storeRetentionDays,
      dispatchTypeOptions,
      dispatchTypes,
      save,
    }
  },
})
</script>

<style scoped>
.settings-page {
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  text-decoration: none;
  font-size: 14px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  margin-bottom: 16px;
}

.back-link:hover {
  color: #374151;
}

.header-content h1 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 8px;
}

.header-content h1 svg {
  color: var(--branding-primary, #7c3aed);
}

.header-desc {
  color: #6b7280;
  font-size: 14px;
  margin: 0;
}

.loading-placeholder {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #6b7280;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 3px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7c3aed);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-banner {
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

.options-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.option-info {
  flex: 1;
}

.option-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.option-header .option-icon {
  color: var(--branding-primary, #7c3aed);
  font-size: 18px;
  width: 20px;
  text-align: center;
}

.option-title {
  font-weight: 600;
  color: #374151;
  font-size: 15px;
}

.option-desc {
  color: #6b7280;
  font-size: 13px;
  line-height: 1.5;
  margin: 0;
}

.global-note {
  margin-top: 4px;
  font-style: italic;
}

.history-sync-card {
  flex-wrap: wrap;
}

.history-sync-edit {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.history-days-input {
  width: 72px;
  padding: 8px 10px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
  text-align: right;
  background: #f9fafb;
  color: #111827;
}

.history-days-unit {
  color: #6b7280;
  font-size: 13px;
}

.retention-select {
  width: auto;
  min-width: 120px;
  text-align: left;
  cursor: pointer;
}

.dispatch-types-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin-top: 12px;
}

.dispatch-type-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #374151;
  cursor: pointer;
}

.save-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  background: var(--branding-primary, #7c3aed);
  color: #fff;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-primary:hover:not(:disabled) {
  background: #6d28d9;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: default;
}

html[data-theme='dark'] .header-content h1 {
  color: #f8fafc;
}

html[data-theme='dark'] .header-desc,
html[data-theme='dark'] .option-desc,
html[data-theme='dark'] .loading-placeholder,
html[data-theme='dark'] .history-days-unit {
  color: #94a3b8;
}

html[data-theme='dark'] .option-title {
  color: #f8fafc;
}

html[data-theme='dark'] .dispatch-type-item {
  color: #cbd5e1;
}

html[data-theme='dark'] .option-card {
  background: rgba(15, 23, 42, 0.92);
  border: 1px solid rgba(71, 85, 105, 0.22);
  box-shadow: none;
}
</style>
