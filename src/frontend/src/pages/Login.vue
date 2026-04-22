<template>
  <div class="login-wrapper">
    <!-- Background decorativo -->
    <div class="login-bg">
      <div class="login-bg-shape shape-1"></div>
      <div class="login-bg-shape shape-2"></div>
      <div class="login-bg-shape shape-3"></div>
    </div>

    <div class="login-container">
      <!-- Card de login -->
      <div class="login-card">
        <!-- Header do card -->
        <div class="login-header">
          <div class="login-logo-wrapper" v-if="config.branding?.logo || config.loginLogo">
            <img :src="config.branding?.logo || config.loginLogo" alt="Logo" class="login-logo" />
          </div>
          <div class="login-logo-placeholder" v-else>
            <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor">
              <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2zm.01 1.67c2.2 0 4.26.86 5.82 2.42a8.225 8.225 0 012.41 5.83c0 4.54-3.7 8.23-8.24 8.23-1.48 0-2.93-.39-4.19-1.15l-.3-.17-3.12.82.83-3.04-.2-.32a8.188 8.188 0 01-1.26-4.38c.01-4.54 3.7-8.24 8.25-8.24zM8.53 7.33c-.16 0-.43.06-.66.31-.22.25-.87.86-.87 2.07 0 1.22.89 2.39 1 2.56.14.17 1.76 2.67 4.25 3.73.59.27 1.05.42 1.41.53.59.19 1.13.16 1.56.1.48-.07 1.46-.6 1.67-1.18.21-.58.21-1.07.15-1.18-.07-.1-.23-.16-.48-.27-.25-.14-1.47-.74-1.69-.82-.23-.08-.37-.12-.56.12-.16.25-.64.81-.78.97-.15.17-.29.19-.53.07-.26-.13-1.06-.39-2-1.23-.74-.66-1.23-1.47-1.38-1.72-.12-.24-.01-.39.11-.5.11-.11.27-.29.37-.44.13-.14.17-.25.25-.41.08-.17.04-.31-.02-.43-.06-.11-.56-1.35-.77-1.84-.2-.48-.4-.42-.56-.43-.14 0-.3-.01-.47-.01z"/>
            </svg>
          </div>
          <h1 class="login-title">{{ config.branding?.title || config.appTitle || 'QuePasa' }}</h1>
          <p class="login-subtitle" v-if="config.loginSubtitle">{{ config.loginSubtitle }}</p>
          <p class="login-subtitle" v-else>Sistema de Gerenciamento WhatsApp</p>
        </div>

        <!-- Warning -->
        <div v-if="config.loginWarning" class="login-warning">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
          </svg>
          <span>{{ config.loginWarning }}</span>
        </div>

        <!-- Formulário -->
        <form @submit.prevent="submit" class="login-form">
          <div v-if="error" class="login-error">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
            <span>{{ error }}</span>
          </div>

          <div class="form-group">
            <label for="email">Email</label>
            <div class="input-wrapper">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" class="input-icon">
                <path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/>
              </svg>
              <input 
                id="email"
                v-model="email" 
                type="email"
                class="form-input" 
                placeholder="seu@email.com"
                autocomplete="email"
                required
              />
            </div>
          </div>

          <div class="form-group">
            <label for="password">Senha</label>
            <div class="input-wrapper">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" class="input-icon">
                <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
              </svg>
              <input 
                id="password"
                v-model="password" 
                type="password" 
                class="form-input" 
                placeholder="••••••••"
                autocomplete="current-password"
                required
              />
            </div>
          </div>

          <button type="submit" class="login-button" :disabled="loading">
            <span v-if="loading" class="spinner"></span>
            <span v-else>Entrar</span>
          </button>
        </form>

        <!-- Setup link (only shown when accountSetup is enabled) -->
        <div class="setup-link" v-if="config.accountSetup">
          <router-link to="/setup" class="setup-link-text">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/>
            </svg>
            Criar conta
          </router-link>
        </div>

        <!-- Footer -->
        <div class="login-footer" v-if="config.loginFooter">
          {{ config.loginFooter }}
        </div>
        <div class="login-footer" v-else>
          <span>Powered by {{ config.branding?.title || 'QuePasa' }}</span>
        </div>
      </div>

      <!-- Info lateral -->
      <div class="login-info">
        <div class="info-content">
          <h2>Bem-vindo ao {{ config.branding?.title || 'QuePasa' }}</h2>
          <p>Plataforma completa para gerenciamento de bots WhatsApp com suporte a múltiplas conexões, webhooks e integrações.</p>
          
          <div class="features">
            <div class="feature">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/>
              </svg>
              <span>Conexão Segura</span>
            </div>
            <div class="feature">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                <path d="M21 6h-2v9H6v2c0 .55.45 1 1 1h11l4 4V7c0-.55-.45-1-1-1zm-4 6V3c0-.55-.45-1-1-1H3c-.55 0-1 .45-1 1v14l4-4h10c.55 0 1-.45 1-1z"/>
              </svg>
              <span>Multi-Servidor</span>
            </div>
            <div class="feature">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/>
              </svg>
              <span>Webhooks</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from '@/services/api'
import { useSessionStore } from '@/stores/session'

export default defineComponent({
  setup() {
    const email = ref('')
    const password = ref('')
    const error = ref('')
    const loading = ref(false)
    const router = useRouter()
    const route = useRoute()
    const session = useSessionStore()

    const config = ref<any>({})

    onMounted(async () => {
      try {
        const res = await api.get('/api/login/config')
        config.value = res.data || {}

        // Set branding title
        const title = config.value.branding?.title || config.value.appTitle || 'QuePasa'
        document.title = title

        // Apply branding colors as CSS variables
        if (config.value.branding) {
          const root = document.documentElement
          root.style.setProperty('--branding-primary', config.value.branding.primaryColor || '#7C3AED')
          root.style.setProperty('--branding-secondary', config.value.branding.secondaryColor || '#5B21B6')
          root.style.setProperty('--branding-accent', config.value.branding.accentColor || '#8B5CF6')
        }

        // Update favicon if provided
        if (config.value.branding?.favicon) {
          const favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement
          if (favicon) {
            favicon.href = config.value.branding.favicon
          } else {
            const link = document.createElement('link')
            link.rel = 'icon'
            link.href = config.value.branding.favicon
            document.head.appendChild(link)
          }
        }

        // Inject external resources
        if (config.value.fontAwesome) {
          const link = document.createElement('link')
          link.rel = 'stylesheet'
          link.href = config.value.fontAwesome
          document.head.appendChild(link)
        }
        if (config.value.googleFonts) {
          const link = document.createElement('link')
          link.rel = 'stylesheet'
          link.href = config.value.googleFonts
          document.head.appendChild(link)
        }
        if (config.value.customCss) {
          const link = document.createElement('link')
          link.rel = 'stylesheet'
          link.href = config.value.customCss
          document.head.appendChild(link)
        }
      } catch (e) {
        // ignore
      }
    })

    async function submit() {
      loading.value = true
      error.value = ''
      try {
        const payload = new URLSearchParams()
        payload.append('email', email.value)
        payload.append('password', password.value)
        await api.post('/login', payload)
        await session.loadSession()
        const redirect = (route.query.redirect as string) || '/'
        router.push(redirect)
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Falha no login'
      } finally {
        loading.value = false
      }
    }

    return { email, password, submit, error, loading, config }
  }
})
</script>

<style scoped>
/* CSS Variables for branding - can be overridden via JavaScript */
:root {
  --branding-primary: #7C3AED;
  --branding-secondary: #5B21B6;
  --branding-accent: #8B5CF6;
}

.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  position: relative;
  overflow: hidden;
  padding: 20px;
}

.login-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.login-bg-shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.1;
  animation: float 20s ease-in-out infinite;
}

.shape-1 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  top: -200px;
  right: -200px;
}

.shape-2 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, var(--branding-accent, #8B5CF6), var(--branding-primary, #7C3AED));
  bottom: -150px;
  left: -150px;
  animation-delay: -5s;
}

.shape-3 {
  width: 300px;
  height: 300px;
  background: linear-gradient(135deg, var(--branding-secondary, #5B21B6), var(--branding-primary, #7C3AED));
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -10s;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) rotate(0deg); }
  25% { transform: translate(20px, -20px) rotate(5deg); }
  50% { transform: translate(-10px, 20px) rotate(-5deg); }
  75% { transform: translate(-20px, -10px) rotate(3deg); }
}

.login-container {
  display: flex;
  max-width: 1000px;
  width: 100%;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 24px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.4);
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.login-card {
  flex: 1;
  padding: 48px;
  display: flex;
  flex-direction: column;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-logo-wrapper {
  margin-bottom: 16px;
}

.login-logo {
  max-width: 120px;
  height: auto;
}

.login-logo-placeholder {
  width: 80px;
  height: 80px;
  margin: 0 auto 16px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.login-title {
  font-size: 28px;
  font-weight: 700;
  color: #1a1a2e;
  margin: 0 0 8px;
}

.login-subtitle {
  font-size: 14px;
  color: #64748b;
  margin: 0;
}

.login-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 12px;
  color: #92400e;
  font-size: 14px;
  margin-bottom: 24px;
}

.login-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #ef4444;
  border-radius: 12px;
  color: #dc2626;
  font-size: 14px;
  margin-bottom: 16px;
}

.login-form {
  flex: 1;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.input-wrapper {
  position: relative;
}

.input-icon {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
}

.form-input {
  width: 100%;
  padding: 14px 16px 14px 48px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 16px;
  transition: all 0.2s;
  background: #f9fafb;
}

.form-input:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
  background: white;
  box-shadow: 0 0 0 4px rgba(37, 211, 102, 0.1);
}

.form-input::placeholder {
  color: #9ca3af;
}

.login-button {
  width: 100%;
  padding: 16px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.login-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(37, 211, 102, 0.3);
}

.login-button:active:not(:disabled) {
  transform: translateY(0);
}

.login-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.setup-link {
  text-align: center;
  margin-top: 16px;
}

.setup-link-text {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--branding-primary, #7C3AED);
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.setup-link-text:hover {
  color: var(--branding-secondary, #5B21B6);
  text-decoration: underline;
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

.login-footer {
  text-align: center;
  margin-top: 24px;
  font-size: 12px;
  color: #9ca3af;
}

.login-info {
  flex: 1;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  padding: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.info-content h2 {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 16px;
}

.info-content p {
  font-size: 16px;
  opacity: 0.9;
  line-height: 1.6;
  margin: 0 0 32px;
}

.features {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 12px;
  font-size: 15px;
  font-weight: 500;
}

.feature svg {
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
  }

  .login-info {
    display: none;
  }

  .login-card {
    padding: 32px 24px;
  }
}
</style>
