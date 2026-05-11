<template>
  <div class="account-page">
    <div class="page-header">
      <div class="header-content">
        <h1>
          <i class="fa fa-user-circle"></i>
          {{ t('account_title') }}
        </h1>
        <p>{{ t('account_subtitle') }}</p>
      </div>
    </div>

    <div v-if="error" class="error-box">
      <i class="fa fa-exclamation-triangle"></i>
      <span>{{ error }}</span>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('loading_generic') }}</p>
    </div>

    <div v-else-if="user" class="account-content">
      <div class="info-card">
        <div class="card-header">
          <i class="fa fa-id-card"></i>
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
          <i class="fa fa-cog"></i>
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
          <i class="fa fa-palette"></i>
          <h2>{{ t('account_branding') }}</h2>
        </div>
        <div class="card-body">
          <div class="branding-preview">
            <div class="color-swatch" :style="{ background: branding.primaryColor }" :title="t('account_branding_primary')">
              <span>{{ t('account_branding_primary') }}</span>
            </div>
            <div class="color-swatch" :style="{ background: branding.secondaryColor }" :title="t('account_branding_secondary')">
              <span>{{ t('account_branding_secondary') }}</span>
            </div>
            <div class="color-swatch" :style="{ background: branding.accentColor }" :title="t('account_branding_accent')">
              <span>{{ t('account_branding_accent') }}</span>
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
          <i class="fa fa-key"></i>
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
          <i class="fa fa-users"></i>
          {{ t('account_users') }}
        </router-link>
        <router-link class="btn-secondary" to="/environment">
          <i class="fa fa-sliders-h"></i>
          {{ t('account_environment') }}
        </router-link>
        <button class="btn-primary" @click="reload">
          <i class="fa fa-sync-alt"></i>
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
  gap: 12px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px;
}

.page-header h1 i {
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

.card-header i {
  color: var(--branding-primary, #7C3AED);
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
  gap: 12px;
  margin-bottom: 16px;
}

.color-swatch {
  width: 80px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding-bottom: 6px;
  color: white;
  font-size: 10px;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
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
