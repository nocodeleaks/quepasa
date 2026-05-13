<template>
  <div class="account-page">
    <div class="page-header">
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
          </svg>
          {{ t('account_title') }}
        </h1>
        <p>{{ t('account_subtitle') }}</p>
      </div>
    </div>

    <div v-if="error" class="error-box">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('loading_generic') }}</p>
    </div>

    <div v-else-if="user" class="account-content">
      <div class="info-card">
        <div class="card-header">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm-8 2.75c1.24 0 2.25 1.01 2.25 2.25S13.24 11.25 12 11.25 9.75 10.24 9.75 9 10.76 6.75 12 6.75zM17 17H7v-.75c0-2.26 3.35-3.25 5-3.25s5 .99 5 3.25V17z"/>
          </svg>
          <h2>{{ t('account_user_info') }}</h2>
        </div>
        <div class="card-body">
          <div class="info-row">
            <span class="info-label">{{ t('account_email_user') }}</span>
            <span class="info-value">{{ user.username }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">{{ t('account_level') }}</span>
            <span class="info-value">
              <span class="badge badge-user">{{ user.level || t('account_role_user') }}</span>
            </span>
          </div>
        </div>
      </div>

      <div class="info-card">
        <div class="card-header">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.05.3-.09.63-.09.94s.02.64.07.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/>
          </svg>
          <h2>{{ t('account_system_info') }}</h2>
        </div>
        <div class="card-body">
          <div class="info-row">
            <span class="info-label">{{ t('account_version') }}</span>
            <span class="info-value">
              <code>{{ version }}</code>
            </span>
          </div>
          <div v-if="branding" class="info-row">
            <span class="info-label">{{ t('account_app_title') }}</span>
            <span class="info-value">{{ branding.title || 'QuePasa' }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">{{ t('account_servers') }}</span>
            <span class="info-value">{{ serverCount }}</span>
          </div>
        </div>
      </div>

      <div v-if="branding" class="info-card">
        <div class="card-header">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9c.83 0 1.5-.67 1.5-1.5 0-.39-.15-.74-.39-1.01-.23-.26-.38-.61-.38-.99 0-.83.67-1.5 1.5-1.5H16c2.76 0 5-2.24 5-5 0-4.42-4.03-8-9-8zm-5.5 9c-.83 0-1.5-.67-1.5-1.5S5.67 9 6.5 9 8 9.67 8 10.5 7.33 12 6.5 12zm3-4C8.67 8 8 7.33 8 6.5S8.67 5 9.5 5s1.5.67 1.5 1.5S10.33 8 9.5 8zm5 0c-.83 0-1.5-.67-1.5-1.5S13.67 5 14.5 5s1.5.67 1.5 1.5S15.33 8 14.5 8zm3 4c-.83 0-1.5-.67-1.5-1.5S16.67 9 17.5 9s1.5.67 1.5 1.5-.67 1.5-1.5 1.5z"/>
          </svg>
          <h2>{{ t('account_branding') }}</h2>
        </div>
        <div class="card-body">
          <div class="branding-preview">
            <div class="color-swatch-wrapper">
              <div class="color-swatch" :style="{ background: branding.primaryColor }"></div>
              <span class="color-label">{{ t('account_branding_primary') }}</span>
            </div>
            <div class="color-swatch-wrapper">
              <div class="color-swatch" :style="{ background: branding.secondaryColor }"></div>
              <span class="color-label">{{ t('account_branding_secondary') }}</span>
            </div>
            <div class="color-swatch-wrapper">
              <div class="color-swatch" :style="{ background: branding.accentColor }"></div>
              <span class="color-label">{{ t('account_branding_accent') }}</span>
            </div>
          </div>
          <div v-if="branding.logo" class="info-row">
            <span class="info-label">{{ t('account_logo') }}</span>
            <span class="info-value">
              <img :src="branding.logo" :alt="t('account_logo_alt')" class="logo-preview" />
            </span>
          </div>
        </div>
      </div>

      <div v-if="masterKeyConfigured" class="info-card">
        <div class="card-header">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M12.65 10C11.83 7.67 9.61 6 7 6c-3.31 0-6 2.69-6 6s2.69 6 6 6c2.61 0 4.83-1.67 5.65-4H17v4h4v-4h2v-4H12.65zM7 14c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2z"/>
          </svg>
          <h2>{{ t('account_master_key') }}</h2>
        </div>
        <div class="card-body">
          <div class="info-row">
            <span class="info-label">{{ t('account_status') }}</span>
            <span class="info-value">
              <span class="badge badge-success">{{ t('account_configured') }}</span>
            </span>
          </div>
          <div class="info-row">
            <span class="info-label">{{ t('account_visibility') }}</span>
            <span class="info-value">{{ t('account_secret_hidden') }}</span>
          </div>
        </div>
      </div>

      <div class="actions-section">
        <router-link class="btn-secondary" to="/users">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
          </svg>
          {{ t('account_users') }}
        </router-link>
        <router-link class="btn-secondary" to="/environment">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M3 17v2h6v-2H3zM3 5v2h10V5H3zm10 16v-2h8v-2h-8v-2h-2v6h2zM7 9v2H3v2h4v2h2V9H7zm14 4v-2H11v2h10zm-6-4h2V7h4V5h-4V3h-2v6z"/>
          </svg>
          {{ t('account_environment') }}
        </router-link>
        <button class="btn-primary" @click="reload">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
          </svg>
          {{ t('account_reload') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue'
import api from '@/services/api'
import { useLocale } from '@/i18n'

export default defineComponent({
  setup() {
    const user = ref<any>(null)
    const version = ref('')
    const error = ref('')
    const loading = ref(true)
    const branding = ref<any>(null)
    const serverCount = ref(0)
    const masterKeyConfigured = ref(false)
    const { t } = useLocale()

    async function load() {
      try {
        loading.value = true
        error.value = ''

        const [sessionRes, accountRes, configRes, masterKeyRes] = await Promise.all([
          api.get('/api/auth/session'),
          api.get('/api/auth/account'),
          api.get('/api/auth/config'),
          api.get('/api/auth/masterkey/status')
        ])

        user.value = accountRes.data?.user || sessionRes.data?.user
        version.value = accountRes.data?.version || sessionRes.data?.version || ''
        branding.value = configRes.data?.branding || null
        serverCount.value = accountRes.data?.serverCount || 0
        masterKeyConfigured.value = masterKeyRes.data?.configured === true
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || t('account_error_load')
      } finally {
        loading.value = false
      }
    }

    function reload() {
      load()
    }

    onMounted(() => {
      load()
    })

    return {
      user, version, error, loading, branding, serverCount,
      masterKeyConfigured,
      reload,
      t
    }
  }
})
</script>

<style scoped>
.account-page {
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px;
}

.page-header h1 svg {
  color: var(--branding-primary, #7C3AED);
}

.page-header p {
  color: #6b7280;
  margin: 0;
}

.error-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  color: #dc2626;
  margin-bottom: 24px;
}

.loading-state {
  text-align: center;
  padding: 60px 0;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.account-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.card-header i,
.card-header svg {
  color: var(--branding-primary, #7C3AED);
  flex-shrink: 0;
}

.card-header h2 {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
  margin: 0;
}

.card-body {
  padding: 20px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f3f4f6;
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 14px;
  color: #6b7280;
}

.info-value {
  font-size: 14px;
  font-weight: 500;
  color: #111827;
}

.info-value code {
  background: #f3f4f6;
  padding: 4px 8px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 13px;
}

.badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-user {
  background: #f5efff;
  color: var(--branding-secondary, #5B21B6);
}

.badge-success {
  background: #dcfce7;
  color: #16a34a;
}

.branding-preview {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.color-swatch-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.color-swatch {
  width: 72px;
  height: 56px;
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.08);
}

.color-label {
  font-size: 10px;
  font-weight: 600;
  color: #6b7280;
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.logo-preview {
  max-height: 40px;
  width: auto;
}

.master-key {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f3f4f6;
  padding: 8px 12px;
  border-radius: 8px;
  font-family: monospace;
  font-size: 13px;
  cursor: pointer;
}

.toggle-btn {
  padding: 4px;
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
}

.toggle-btn:hover {
  color: #374151;
}

.actions-section {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 8px;
  flex-wrap: wrap;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary svg {
  flex-shrink: 0;
}

.btn-secondary svg {
  flex-shrink: 0;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.25);
}

.btn-secondary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e5e7eb;
}
</style>
