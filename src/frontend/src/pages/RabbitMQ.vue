<template>
  <div class="rabbitmq-page">
    <!-- Header -->
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile">
        <i class="fa fa-arrow-left"></i>
        Voltar
      </button>
      <div class="header-content">
        <h1>
          <i class="fa fa-server"></i>
          Configurações RabbitMQ
        </h1>
        <p v-if="data?.server?.Wid">Servidor: {{ data.server.Wid }}</p>
        <p v-else-if="data?.server?.Token">Token: {{ truncateToken(data.server.Token) }}</p>
        <p v-else>Gerencie as integrações RabbitMQ</p>
      </div>
    </div>

    <!-- Error -->
    <div v-if="error" class="error-box">
      <i class="fa fa-exclamation-triangle"></i>
      <span>{{ error }}</span>
    </div>

    <!-- No Token Warning -->
    <div v-else-if="!hasToken && !loading" class="no-token-warning">
      <div class="warning-icon">
        <i class="fa fa-exclamation-circle"></i>
      </div>
      <h2>Servidor não selecionado</h2>
      <p>Para gerenciar RabbitMQ, acesse através de um servidor específico.</p>
      <router-link to="/" class="btn-primary">
        <i class="fa fa-arrow-left"></i>
        Voltar para Servidores
      </router-link>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Carregando...</p>
    </div>

    <!-- Content -->
    <div v-else-if="data" class="rabbitmq-content">
      <!-- Add New Form -->
      <div class="add-card">
        <div class="card-header">
          <i class="fa fa-plus-circle"></i>
          <h2>Adicionar Configuração</h2>
        </div>
        <div class="card-body">
          <form @submit.prevent="createRabbit" class="add-form">
            <div class="form-row">
              <div class="form-group flex-grow">
                <label>Connection String</label>
                <input 
                  v-model="newConnectionString" 
                  type="text" 
                  class="form-input" 
                  placeholder="amqp://user:pass@host:5672/vhost"
                  required
                />
              </div>
              <div class="form-group">
                <label>Track ID</label>
                <input 
                  v-model="newTrackId" 
                  type="text" 
                  class="form-input" 
                  placeholder="Opcional"
                />
              </div>
            </div>

            <div class="form-row options-row">
              <label class="checkbox-label">
                <input type="checkbox" v-model="newBroadcasts" />
                <span>Broadcasts</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="newGroups" />
                <span>Grupos</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="newReadReceipts" />
                <span>Confirmações</span>
              </label>
            </div>

            <button type="submit" class="btn-primary" :disabled="!newConnectionString || creating">
              <i v-if="creating" class="fa fa-spinner fa-spin"></i>
              <i v-else class="fa fa-plus"></i>
              {{ creating ? 'Adicionando...' : 'Adicionar' }}
            </button>
          </form>
        </div>
      </div>

      <!-- List -->
      <div class="list-card">
        <div class="card-header">
          <i class="fa fa-list"></i>
          <h2>Configurações Ativas</h2>
          <span class="count-badge">{{ data.rabbitmq?.length || 0 }}</span>
        </div>
        <div class="card-body">
          <div v-if="!data.rabbitmq || data.rabbitmq.length === 0" class="empty-state">
            <i class="fa fa-inbox"></i>
            <p>Nenhuma configuração RabbitMQ</p>
          </div>

          <div v-else class="rabbitmq-list">
            <div v-for="r in data.rabbitmq" :key="r.ConnectionString" class="rabbitmq-item">
              <div class="item-info">
                <div class="item-main">
                  <strong class="connection-string">{{ r.ConnectionString }}</strong>
                  <span v-if="r.TrackId" class="track-id">Track: {{ r.TrackId }}</span>
                </div>
                <div class="item-meta">
                  <span v-if="r.Exchange"><i class="fa fa-exchange-alt"></i> {{ r.Exchange }}</span>
                  <span v-if="r.Queue"><i class="fa fa-stream"></i> {{ r.Queue }}</span>
                </div>
              </div>

              <div class="item-flags">
                <button 
                  class="flag-btn" 
                  :class="{ active: r.ForwardInternal }" 
                  @click="toggleRabbitFlag(r.ConnectionString, 'rabbitmq-forwardinternal')"
                  title="Forward Internal"
                >
                  <i class="fa fa-share"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="{ active: r.Broadcasts }" 
                  @click="toggleRabbitFlag(r.ConnectionString, 'rabbitmq-broadcasts')"
                  title="Broadcasts"
                >
                  <i class="fa fa-bullhorn"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="{ active: r.Groups }" 
                  @click="toggleRabbitFlag(r.ConnectionString, 'rabbitmq-groups')"
                  title="Grupos"
                >
                  <i class="fa fa-users"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="{ active: r.ReadReceipts }" 
                  @click="toggleRabbitFlag(r.ConnectionString, 'rabbitmq-readreceipts')"
                  title="Confirmações de Leitura"
                >
                  <i class="fa fa-check-double"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="{ active: r.Calls }" 
                  @click="toggleRabbitFlag(r.ConnectionString, 'rabbitmq-calls')"
                  title="Chamadas"
                >
                  <i class="fa fa-phone"></i>
                </button>
              </div>

              <button class="btn-delete" @click="confirmDelete(r.ConnectionString)" title="Remover">
                <i class="fa fa-trash"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Confirm Modal -->
    <ConfirmModal 
      :show="showConfirm" 
      title="Remover Configuração" 
      message="Tem certeza que deseja remover esta configuração RabbitMQ?" 
      @confirm="doDelete" 
      @cancel="showConfirm = false" 
      confirmLabel="Remover" 
      cancelLabel="Cancelar" 
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue'
import api from '@/services/api'
import { useRoute } from 'vue-router'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { pushToast } from '@/services/toast'

export default defineComponent({
  components: { ConfirmModal },
  setup() {
    const data = ref<any>(null)
    const error = ref('')
    const loading = ref(true)
    const creating = ref(false)
    const route = useRoute()
    const hasToken = ref(true)

    function truncateToken(token: string) {
      if (!token) return ''
      if (token.length <= 16) return token
      return token.substring(0, 8) + '...' + token.substring(token.length - 4)
    }

    async function load() {
      try {
        loading.value = true
        error.value = ''
        const token = (route.query.token as string) || ''
        
        if (!token) {
          hasToken.value = false
          loading.value = false
          return
        }
        
        hasToken.value = true
        const res = await api.get('/api/rabbitmq', { params: { token } })
        data.value = res.data
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao carregar RabbitMQ'
      } finally {
        loading.value = false
      }
    }

    // create rabbitmq
    const newConnectionString = ref('')
    const newTrackId = ref('')
    const newBroadcasts = ref(true)
    const newGroups = ref(true)
    const newReadReceipts = ref(false)
    const newExtra = ref('')

    async function createRabbit() {
      if (!newConnectionString.value) return

      creating.value = true
      try {
        const token = (route.query.token as string) || (data.value?.server?.Token || '')
        if (!token) throw new Error('Token não encontrado')
        
        let extraParsed = null
        if (newExtra.value) extraParsed = JSON.parse(newExtra.value)
        
        await api.post('/api/rabbitmq', {
          token,
          connectionString: newConnectionString.value,
          trackId: newTrackId.value,
          broadcasts: newBroadcasts.value,
          groups: newGroups.value,
          readReceipts: newReadReceipts.value,
          extra: extraParsed,
        })
        
        await load()
        
        // Clear form
        newConnectionString.value = ''
        newTrackId.value = ''
        newBroadcasts.value = true
        newGroups.value = true
        newReadReceipts.value = false
        newExtra.value = ''
        
        pushToast('Configuração RabbitMQ criada', 'success')
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao criar RabbitMQ'
        error.value = msg
        pushToast(msg, 'error')
      } finally {
        creating.value = false
      }
    }

    // confirmation flow
    const showConfirm = ref(false)
    const confirmCs = ref('')

    function confirmDelete(cs: string) {
      confirmCs.value = cs
      showConfirm.value = true
    }

    async function doDelete() {
      try {
        const token = (route.query.token as string) || (data.value?.server?.Token || '')
        if (!token) throw new Error('Token não encontrado')
        
        await api.delete('/api/rabbitmq', { 
          params: { token }, 
          data: { connectionString: confirmCs.value } 
        })
        
        showConfirm.value = false
        confirmCs.value = ''
        
        await load()
        pushToast('Configuração RabbitMQ removida', 'success')
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao remover'
        error.value = msg
        pushToast(msg, 'error')
      }
    }

    async function toggleRabbitFlag(connectionString: string, key: string) {
      try {
        const token = (route.query.token as string) || (data.value?.server?.Token || '')
        if (!token) throw new Error('Token não encontrado')
        
        await api.post('/api/toggle', { token, key, connectionString })
        await load()
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao alternar'
        pushToast(msg, 'error')
      }
    }

    onMounted(() => {
      load()
    })

    return { 
      data, error, loading, creating, hasToken,
      newConnectionString, newTrackId, newBroadcasts, newGroups, newReadReceipts, newExtra, 
      createRabbit, showConfirm, confirmDelete, doDelete, toggleRabbitFlag,
      truncateToken
    }
  }
})
</script>

<style scoped>
.rabbitmq-page {
  max-width: 1000px;
  margin: 0 auto;
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

.no-token-warning {
  text-align: center;
  padding: 60px 20px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 16px;
}

.no-token-warning .warning-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 20px;
  background: #fef3c7;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.no-token-warning .warning-icon i {
  font-size: 40px;
  color: #f59e0b;
}

.no-token-warning h2 {
  font-size: 24px;
  font-weight: 700;
  color: #92400e;
  margin: 0 0 8px;
}

.no-token-warning p {
  color: #b45309;
  margin: 0 0 24px;
}

.no-token-warning .btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border-radius: 10px;
  text-decoration: none;
  font-weight: 600;
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

.rabbitmq-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.add-card, .list-card {
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

.count-badge {
  margin-left: auto;
  background: var(--branding-primary, #7C3AED);
  color: white;
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.card-body {
  padding: 20px;
}

.add-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group.flex-grow {
  flex: 1;
}

.form-group label {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}

.form-input {
  padding: 10px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.options-row {
  gap: 20px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #374151;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--branding-primary, #7C3AED);
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
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
  align-self: flex-start;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.25);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #9ca3af;
}

.empty-state i {
  font-size: 48px;
  margin-bottom: 12px;
  display: block;
}

.rabbitmq-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rabbitmq-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-main {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.connection-string {
  font-size: 14px;
  color: #111827;
  word-break: break-all;
}

.track-id {
  font-size: 12px;
  background: #e0e7ff;
  color: #4338ca;
  padding: 2px 8px;
  border-radius: 8px;
}

.item-meta {
  display: flex;
  gap: 12px;
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
}

.item-meta i {
  margin-right: 4px;
}

.item-flags {
  display: flex;
  gap: 6px;
}

.flag-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: #e5e7eb;
  border-radius: 6px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.flag-btn:hover {
  background: #d1d5db;
}

.flag-btn.active {
  background: var(--branding-primary, #7C3AED);
  color: white;
}

.btn-delete {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: #fef2f2;
  border-radius: 8px;
  color: #dc2626;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-delete:hover {
  background: #fecaca;
}

@media (max-width: 768px) {
  .form-row {
    flex-direction: column;
  }

  .rabbitmq-item {
    flex-direction: column;
    align-items: stretch;
  }

  .item-flags {
    justify-content: center;
    padding-top: 12px;
    border-top: 1px solid #e5e7eb;
    margin-top: 12px;
  }

  .btn-delete {
    position: absolute;
    top: 12px;
    right: 12px;
  }

  .rabbitmq-item {
    position: relative;
    padding-right: 56px;
  }
}
</style>
