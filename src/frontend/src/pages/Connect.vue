<template>
  <div class="connect-page">
    <div class="connect-header">
      <button @click="$router.back()" class="back-link">
        <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
        </svg>
        Voltar
      </button>
      <h1>Conectar WhatsApp</h1>
      <p class="subtitle">Escolha o método de conexão</p>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Criando servidor...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <p>{{ error }}</p>
      <button @click="error = ''" class="btn-retry">Tentar novamente</button>
    </div>

    <div v-else class="connection-methods">
      <div class="method-card" @click="connectWith('qrcode')">
        <div class="method-icon qrcode-icon">
          <svg viewBox="0 0 24 24" width="64" height="64" fill="currentColor">
            <path d="M3 11h8V3H3v8zm2-6h4v4H5V5zM3 21h8v-8H3v8zm2-6h4v4H5v-4zM13 3v8h8V3h-8zm6 6h-4V5h4v4zM13 13h2v2h-2zM15 15h2v2h-2zM13 17h2v2h-2zM17 13h2v2h-2zM19 15h2v2h-2zM17 17h2v2h-2zM15 19h2v2h-2zM19 19h2v2h-2z"/>
          </svg>
        </div>
        <h3>QR Code</h3>
        <p>Escaneie o código QR com seu WhatsApp</p>
        <ul class="method-features">
          <li>✓ Conexão rápida</li>
          <li>✓ Use a câmera do celular</li>
          <li>✓ Método tradicional</li>
        </ul>
        <span class="method-action">Escanear QR Code →</span>
      </div>

      <div class="method-card" @click="connectWith('paircode')">
        <div class="method-icon paircode-icon">
          <svg viewBox="0 0 24 24" width="64" height="64" fill="currentColor">
            <path d="M17 7h-4v2h4c1.65 0 3 1.35 3 3s-1.35 3-3 3h-4v2h4c2.76 0 5-2.24 5-5s-2.24-5-5-5zm-6 8H7c-1.65 0-3-1.35-3-3s1.35-3 3-3h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-2zm-3-4h8v2H8z"/>
          </svg>
        </div>
        <h3>Pair Code</h3>
        <p>Digite o código de 8 dígitos no WhatsApp</p>
        <ul class="method-features">
          <li>✓ Sem necessidade de câmera</li>
          <li>✓ Use o teclado</li>
          <li>✓ Método alternativo</li>
        </ul>
        <span class="method-action">Usar Pair Code →</span>
      </div>
    </div>

    <div class="help-section">
      <h4>Como conectar?</h4>
      <ol>
        <li>Escolha um método acima</li>
        <li>Abra o WhatsApp no seu celular</li>
        <li>Vá em <strong>Configurações → Aparelhos conectados</strong></li>
        <li>Toque em <strong>Conectar um aparelho</strong></li>
        <li>Escaneie o QR Code ou digite o Pair Code</li>
      </ol>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

export default defineComponent({
  name: 'Connect',
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const error = ref('')

    const connectWith = async (method: 'qrcode' | 'paircode') => {
      loading.value = true
      error.value = ''
      
      try {
        // Create a new server first
        const response = await api.post('/api/server/create', {})
        
        if (response.data && response.data.token) {
          const token = response.data.token
          pushToast('Servidor criado com sucesso!', 'success')
          
          // Navigate to the chosen connection method
          if (method === 'qrcode') {
            router.push(`/server/${token}/qrcode`)
          } else {
            router.push(`/server/${token}/paircode`)
          }
        } else {
          throw new Error('Token não recebido do servidor')
        }
      } catch (err: any) {
        console.error('Error creating server:', err)
        error.value = err.response?.data?.message || err.message || 'Erro ao criar servidor'
        pushToast(error.value, 'error')
      } finally {
        loading.value = false
      }
    }

    return { loading, error, connectWith }
  }
})
</script>

<style scoped>
.connect-page {
  margin: 0 auto;
  padding: 2rem 1rem;
}

.connect-header {
  text-align: center;
  margin-bottom: 2rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: #6c757d;
  text-decoration: none;
  margin-bottom: 1rem;
  transition: color 0.2s;
}

.back-link:hover {
  color: #7C3AED;
}

.connect-header h1 {
  font-size: 2rem;
  font-weight: 700;
  color: #1a1a2e;
  margin: 0.5rem 0;
}

.subtitle {
  color: #6c757d;
  margin: 0;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
  color: #6c757d;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e9ecef;
  border-top-color: #7C3AED;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-state svg {
  color: #dc3545;
  margin-bottom: 1rem;
}

.btn-retry {
  margin-top: 1rem;
  padding: 0.75rem 1.5rem;
  background: #7C3AED;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.2s;
}

.btn-retry:hover {
  background: #1da851;
}

.connection-methods {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 2rem;
  margin-bottom: 3rem;
}

.method-card {
  background: white;
  border: 2px solid #e9ecef;
  border-radius: 16px;
  padding: 2rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.method-card:hover {
  border-color: #7C3AED;
  box-shadow: 0 8px 30px rgba(37, 211, 102, 0.15);
  transform: translateY(-4px);
}

.method-icon {
  width: 100px;
  height: 100px;
  margin: 0 auto 1.5rem;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qrcode-icon {
  background: linear-gradient(135deg, #7C3AED 0%, #5B21B6 100%);
  color: white;
}

.paircode-icon {
  background: linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%);
  color: white;
}

.method-card h3 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1a1a2e;
  margin: 0 0 0.5rem;
}

.method-card > p {
  color: #6c757d;
  margin: 0 0 1.5rem;
}

.method-features {
  list-style: none;
  padding: 0;
  margin: 0 0 1.5rem;
  text-align: left;
}

.method-features li {
  padding: 0.5rem 0;
  color: #495057;
  font-size: 0.95rem;
}

.method-action {
  display: inline-block;
  color: #7C3AED;
  font-weight: 600;
  font-size: 1rem;
}

.method-card:hover .method-action {
  text-decoration: underline;
}

.help-section {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 1.5rem 2rem;
}

.help-section h4 {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1a1a2e;
  margin: 0 0 1rem;
}

.help-section ol {
  margin: 0;
  padding-left: 1.5rem;
  color: #495057;
}

.help-section li {
  padding: 0.3rem 0;
}

.help-section strong {
  color: #1a1a2e;
}

@media (max-width: 768px) {
  .connect-page {
    padding: 1rem;
  }
  
  .connect-header h1 {
    font-size: 1.5rem;
  }
  
  .connection-methods {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .method-card {
    padding: 1.5rem;
  }
  
  .method-icon {
    width: 80px;
    height: 80px;
  }
  
  .method-icon svg {
    width: 48px;
    height: 48px;
  }
}
</style>
