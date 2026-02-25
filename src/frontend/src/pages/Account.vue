<template>
  <div class="account-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1>
          <i class="fa fa-user-circle"></i>
          Minha Conta
        </h1>
        <p>Gerencie suas informações e configurações</p>
      </div>
    </div>

    <!-- Error -->
    <div v-if="error" class="error-box">
      <i class="fa fa-exclamation-triangle"></i>
      <span>{{ error }}</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Carregando...</p>
    </div>

    <!-- Content -->
    <div v-else-if="user" class="account-content">
      <!-- User Info Card -->
      <div class="info-card">
        <div class="card-header">
          <i class="fa fa-id-card"></i>
          <h2>Informações do Usuário</h2>
        </div>
        <div class="card-body">
          <div class="info-row">
            <span class="info-label">Email/Usuário:</span>
            <span class="info-value">{{ user.username }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Nível:</span>
            <span class="info-value">
              <span class="badge" :class="user.level === 'admin' ? 'badge-admin' : 'badge-user'">
                {{ user.level || 'user' }}
              </span>
            </span>
          </div>
        </div>
      </div>

      <!-- System Info Card -->
      <div class="info-card">
        <div class="card-header">
          <i class="fa fa-cog"></i>
          <h2>Informações do Sistema</h2>
        </div>
        <div class="card-body">
          <div class="info-row">
            <span class="info-label">Versão:</span>
            <span class="info-value">
              <code>{{ version }}</code>
            </span>
          </div>
          <div class="info-row" v-if="branding">
            <span class="info-label">Título da Aplicação:</span>
            <span class="info-value">{{ branding.title || 'QuePasa' }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">View Mode:</span>
            <span class="info-value">{{ serversViewMode || 'card' }}</span>
          </div>
        </div>
      </div>

      <!-- Branding Preview Card -->
      <div class="info-card" v-if="branding">
        <div class="card-header">
          <i class="fa fa-palette"></i>
          <h2>Branding</h2>
        </div>
        <div class="card-body">
          <div class="branding-preview">
            <div class="color-swatch" :style="{ background: branding.primaryColor }" title="Primary">
              <span>Primary</span>
            </div>
            <div class="color-swatch" :style="{ background: branding.secondaryColor }" title="Secondary">
              <span>Secondary</span>
            </div>
            <div class="color-swatch" :style="{ background: branding.accentColor }" title="Accent">
              <span>Accent</span>
            </div>
          </div>
          <div class="info-row" v-if="branding.logo">
            <span class="info-label">Logo:</span>
            <span class="info-value">
              <img :src="branding.logo" alt="Logo" class="logo-preview" />
            </span>
          </div>
        </div>
      </div>

      <!-- Master Key Card (if available) -->
      <div class="info-card" v-if="hasMasterKey">
        <div class="card-header">
          <i class="fa fa-key"></i>
          <h2>API Master Key</h2>
        </div>
        <div class="card-body">
          <div class="info-row">
            <span class="info-label">Status:</span>
            <span class="info-value">
              <span class="badge badge-success">Configurada</span>
            </span>
          </div>
          <div class="info-row" v-if="masterKey">
            <span class="info-label">Chave:</span>
            <code class="master-key" @click="copyMasterKey">
              {{ showMasterKey ? masterKey : '••••••••••••••••' }}
              <button class="toggle-btn" @click.stop="showMasterKey = !showMasterKey">
                <i :class="showMasterKey ? 'fa fa-eye-slash' : 'fa fa-eye'"></i>
              </button>
            </code>
          </div>
          <button v-if="!masterKey" class="btn-secondary" @click="loadMasterKey">
            <i class="fa fa-download"></i>
            Carregar Master Key
          </button>
        </div>
      </div>

      <!-- Actions -->
      <div class="actions-section">
        <button class="btn-primary" @click="reload">
          <i class="fa fa-sync-alt"></i>
          Recarregar
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue'
import api from '@/services/api'
import wsService from '@/services/ws'
import { pushToast } from '@/services/toast'

export default defineComponent({
  setup() {
    const user = ref<any>(null)
    const version = ref('')
    const error = ref('')
    const loading = ref(true)
    const branding = ref<any>(null)
    const serversViewMode = ref('')
    const hasMasterKey = ref(false)
    const masterKey = ref('')
    const showMasterKey = ref(false)

    async function load() {
      try {
        loading.value = true
        error.value = ''
        const res = await api.get('/api/session')
        user.value = res.data.user
        version.value = res.data.version
        branding.value = res.data.branding
        serversViewMode.value = res.data.serversViewMode
        
        // Check if master key is available
        const accountRes = await api.get('/api/account')
        hasMasterKey.value = accountRes.data?.hasMasterKey || false
        
        // Start websocket connection
        wsService.connect('/api/verify/ws')
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao carregar conta'
      } finally {
        loading.value = false
      }
    }

    async function loadMasterKey() {
      try {
        const res = await api.get('/api/account/masterkey')
        masterKey.value = res.data?.masterKey || ''
        showMasterKey.value = true
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao carregar master key', 'error')
      }
    }

    async function copyMasterKey() {
      if (!masterKey.value) return
      try {
        await navigator.clipboard.writeText(masterKey.value)
        pushToast('Master Key copiada!', 'success')
      } catch {
        pushToast('Erro ao copiar', 'error')
      }
    }

    function reload() {
      load()
    }

    onMounted(() => {
      load()
    })

    return { 
      user, version, error, loading, branding, serversViewMode,
      hasMasterKey, masterKey, showMasterKey,
      reload, loadMasterKey, copyMasterKey
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

.badge-admin {
  background: #fee2e2;
  color: #dc2626;
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
  text-shadow: 0 1px 2px rgba(0,0,0,0.3);
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
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e5e7eb;
}
</style>
