<template>
  <div class="qrcode-page">
    <div class="page-header">
      <div class="header-content">
        <button @click="$router.back()" class="back-link hide-mobile">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
          </svg>
          Voltar
        </button>
        <h1>Conectar WhatsApp</h1>
        <p>Escaneie o QR Code com seu WhatsApp</p>
      </div>
    </div>

    <div class="content-card">
      <div v-if="error" class="error-box">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
        </svg>
        <span>{{ error }}</span>
      </div>

      <div class="qr-container">
        <div v-if="loading" class="qr-loading">
          <div class="spinner-large"></div>
          <p>Gerando QR Code...</p>
        </div>

        <div v-else-if="connected" class="qr-success">
          <svg viewBox="0 0 24 24" width="80" height="80" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L18 9l-9 9z"/>
          </svg>
          <h3>Conectado com Sucesso!</h3>
          <p>Seu WhatsApp está pareado.</p>
          <router-link :to="`/server/${token}`" class="btn-success">
            Ir para o Servidor
          </router-link>
        </div>

        <div v-else class="qr-display">
          <img v-if="qrImage" :src="qrImage" alt="QR Code" class="qr-image" />
          <div v-else class="qr-placeholder">
            <svg viewBox="0 0 24 24" width="60" height="60" fill="currentColor">
              <path d="M3 11h8V3H3v8zm2-6h4v4H5V5zM3 21h8v-8H3v8zm2-6h4v4H5v-4zm8-12v8h8V3h-8zm6 6h-4V5h4v4zm-6 4h2v2h-2zm2 2h2v2h-2zm-2 2h2v2h-2zm4 0h2v2h-2zm2 2h2v2h-2zm0-4h2v2h-2zm2-2h2v2h-2z"/>
            </svg>
            <p>Clique em "Gerar QR Code"</p>
          </div>
        </div>

        <div class="qr-actions">
          <button @click="generateQR" class="btn-primary" :disabled="loading">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
            </svg>
            {{ loading ? 'Gerando...' : 'Gerar QR Code' }}
          </button>
        </div>

        <div class="qr-instructions">
          <h4>Como conectar:</h4>
          <ol>
            <li>Abra o WhatsApp no seu celular</li>
            <li>Toque em <strong>Menu</strong> ou <strong>Configurações</strong></li>
            <li>Toque em <strong>Aparelhos Conectados</strong></li>
            <li>Toque em <strong>Conectar um aparelho</strong></li>
            <li>Aponte a câmera para o QR Code</li>
          </ol>
        </div>
      </div>

      <div class="alt-method">
        <p>Prefere usar código numérico?</p>
        <router-link :to="`/server/${token}/paircode`" class="link-primary">
          Conectar com Código de Pareamento
        </router-link>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/services/api'

export default defineComponent({
  setup() {
    const route = useRoute()
    const token = route.params.token as string

    const loading = ref(false)
    const error = ref('')
    const qrImage = ref('')
    const connected = ref(false)

    let pollInterval: any = null

    async function generateQR() {
      loading.value = true
      error.value = ''
      qrImage.value = ''

      try {
        // Use form endpoint that returns JSON with base64 QR code
        const res = await api.get(`/api/server/${token}/qrcode`)
        
        if (res.data?.connected) {
          connected.value = true
          return
        }
        
        if (res.data?.qrcode) {
          qrImage.value = res.data.qrcode
          // Start polling to check connection status
          startPolling()
        } else {
          error.value = 'Nenhum QR Code recebido'
        }
      } catch (err: any) {
        error.value = err.response?.data?.result || err.response?.data?.message || err.message || 'Erro ao gerar QR Code'
      } finally {
        loading.value = false
      }
    }

    async function checkConnection() {
      try {
        const res = await api.get(`/api/server/${token}/info`)
        if (res.data?.connected || res.data?.state === 'Ready') {
          connected.value = true
          stopPolling()
        }
      } catch {
        // ignore
      }
    }

    function startPolling() {
      stopPolling()
      pollInterval = setInterval(checkConnection, 3000)
    }

    function stopPolling() {
      if (pollInterval) {
        clearInterval(pollInterval)
        pollInterval = null
      }
    }

    onMounted(() => {
      // Auto-generate QR on mount
      generateQR()
    })

    onUnmounted(() => {
      stopPolling()
    })

    return { token, loading, error, qrImage, connected, generateQR }
  }
})
</script>

<style scoped>
.qrcode-page {
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

.qr-container {
  text-align: center;
}

.qr-loading {
  padding: 60px 0;
}

.spinner-large {
  width: 50px;
  height: 50px;
  border: 4px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.qr-loading p {
  color: #6b7280;
}

.qr-success {
  padding: 40px 0;
  color: var(--branding-primary, #7C3AED);
}

.qr-success h3 {
  font-size: 24px;
  margin: 16px 0 8px;
}

.qr-success p {
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

.btn-success:hover {
  background: var(--branding-secondary, #5B21B6);
}

.qr-display {
  padding: 24px;
  background: #f9fafb;
  border-radius: 16px;
  margin-bottom: 24px;
}

.qr-image {
  width: 256px;
  height: 256px;
  border-radius: 12px;
}

.qr-placeholder {
  padding: 60px 0;
  color: #9ca3af;
}

.qr-placeholder p {
  margin: 16px 0 0;
}

.qr-actions {
  margin-bottom: 32px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 14px 28px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(124, 58, 237, 0.25);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.qr-instructions {
  text-align: left;
  padding: 20px;
  background: #f8f6ff;
  border-radius: 12px;
  margin-bottom: 24px;
}

.qr-instructions h4 {
  font-size: 16px;
  color: var(--branding-secondary, #5B21B6);
  margin: 0 0 12px;
}

.qr-instructions ol {
  margin: 0;
  padding-left: 20px;
  color: #15803d;
}

.qr-instructions li {
  margin-bottom: 8px;
}

.qr-instructions li:last-child {
  margin-bottom: 0;
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
  color: var(--branding-primary, #7C3AED);
  text-decoration: none;
  font-weight: 600;
}

.link-primary:hover {
  text-decoration: underline;
}
</style>
