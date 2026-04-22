<template>
  <div class="server-page">
    <div class="server-header">
      <div class="header-top">
        <button @click="$router.back()" class="back-link hide-mobile">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
          </svg>
          Voltar
        </button>
      </div>

      <div v-if="loading" class="loading-placeholder">
        <div class="spinner"></div>
        <span>Carregando servidor...</span>
      </div>

      <div v-else-if="server" class="server-info">
        <div class="server-avatar" :class="statusClass">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor">
            <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2z" />
          </svg>
        </div>
        <div class="server-details">
          <h1>{{ formatWid(serverWid) || formatWid(server?.wid) || "Não conectado" }}</h1>
          <div class="server-meta">
            Status
            <span class="status-badge" :class="statusClass">{{ serverState || "Desconhecido" }}</span>
          </div>
        </div>
      </div>

      <div v-if="error" class="error-banner">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
        </svg>
        <span>{{ error }}</span>
      </div>
    </div>

    <div class="quick-actions" v-if="server">
      <router-link :to="`/server/${token}/qrcode`" class="action-card" :class="{ disabled: isConnected }">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M3 11h8V3H3v8zm2-6h4v4H5V5zM3 21h8v-8H3v8zm2-6h4v4H5v-4zm8-12v8h8V3h-8zm6 6h-4V5h4v4zm-6 4h2v2h-2zm2 2h2v2h-2zm-2 2h2v2h-2zm4 0h2v2h-2zm2 2h2v2h-2zm0-4h2v2h-2zm2-2h2v2h-2z" />
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">QR Code</span>
          <span class="action-desc">Conectar via QR Code</span>
        </div>
      </router-link>

      <router-link :to="`/server/${token}/paircode`" class="action-card" :class="{ disabled: isConnected }">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z" />
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">Código de Pareamento</span>
          <span class="action-desc">Conectar com código numérico</span>
        </div>
      </router-link>
    </div>

    <div class="quick-actions" v-if="server">
      <router-link :to="`/server/${token}/send`" class="action-card primary">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">Enviar Mensagem</span>
          <span class="action-desc">Envie texto, imagens e documentos</span>
        </div>
      </router-link>

      <router-link :to="`/server/${token}/messages`" class="action-card">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z" />
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">Mensagens</span>
          <span class="action-desc">Ver mensagens recebidas</span>
        </div>
      </router-link>

      <router-link :to="`/webhooks?token=${token}`" class="action-card">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M12 2a10 10 0 100 20 10 10 0 000-20zm5 14.59L15.59 18 12 14.41 8.41 18 7 16.59 10.59 13 7 9.41 8.41 8 12 11.59 15.59 8 17 9.41 13.41 13 17 16.59z" />
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">Webhooks</span>
          <span class="action-desc">Gerenciar integrações HTTP</span>
        </div>
      </router-link>

      <router-link :to="`/rabbitmq?token=${token}`" class="action-card">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M4 4h16v4H4V4zm0 6h10v4H4v-4zm0 6h16v4H4v-4zm12-6h4v4h-4v-4z" />
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">RabbitMQ</span>
          <span class="action-desc">Gerenciar integrações AMQP</span>
        </div>
      </router-link>
    </div>

    <div class="details-section" v-if="server">
      <h2>
        <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z" />
        </svg>
        Informações do Servidor
      </h2>

      <div class="details-grid">
        <div class="detail-card">
          <span class="detail-label">Token</span>
          <div class="detail-value token">
            <code>{{ token }}</code>
            <button @click="copyToken" class="copy-btn" :class="{ copied: tokenCopied }">
              <svg v-if="!tokenCopied" viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z" />
              </svg>
              <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
              </svg>
            </button>
          </div>
        </div>

        <div class="detail-card">
          <span class="detail-label">WhatsApp ID</span>
          <span class="detail-value">{{ formatWid(serverWid) || formatWid(server?.wid) || "-" }}</span>
        </div>

        <div class="detail-card">
          <span class="detail-label">Despachos Ativos</span>
          <span class="detail-value">{{ server.dispatchCount ?? 0 }}</span>
        </div>
      </div>
    </div>

    <div class="options-section" v-if="server">
      <h2>
        <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
          <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z" />
        </svg>
        Opções do Servidor
      </h2>

      <div class="options-list">
        <div class="option-card history-sync-card">
          <div class="option-info">
            <div class="option-header">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M13 3c-4.97 0-9 4.03-9 9H1l3.89 3.89.07.14L9 12H6c0-3.87 3.13-7 7-7s7 3.13 7 7-3.13 7-7 7c-1.93 0-3.68-.79-4.94-2.06l-1.42 1.42C8.27 19.99 10.51 21 13 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 5v5l4.28 2.54.72-1.21-3.5-2.08V8H12z" />
              </svg>
              <span class="option-title">Sincronização de Histórico</span>
            </div>
            <p class="option-desc">Essa configuração ainda não está exposta na API `/spa`. Use a interface clássica se precisar alterar esse valor.</p>
          </div>
          <span class="readonly-badge">Somente leitura</span>
        </div>

        <div class="option-card">
          <div class="option-info">
            <div class="option-header">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z" />
              </svg>
              <span class="option-title">Mensagens de Broadcast</span>
            </div>
            <p class="option-desc">Receber mensagens enviadas para listas de transmissão. Quando ativo, mensagens de broadcast aparecem no webhook.</p>
          </div>
          <TriStateToggle
            v-model="options.broadcasts"
            @change="updateOption('broadcasts', $event)"
            :disabled="togglingOption === 'server-broadcasts'"
          />
        </div>

        <div class="option-card">
          <div class="option-info">
            <div class="option-header">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z" />
              </svg>
              <span class="option-title">Mensagens de Grupos</span>
            </div>
            <p class="option-desc">Receber mensagens de grupos. Quando ativo, todas as mensagens de grupos serão entregues via webhook.</p>
          </div>
          <TriStateToggle
            v-model="options.groups"
            @change="updateOption('groups', $event)"
            :disabled="togglingOption === 'server-groups'"
          />
        </div>

        <div class="option-card">
          <div class="option-info">
            <div class="option-header">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M18 7l-1.41-1.41-6.34 6.34 1.41 1.41L18 7zm4.24-1.41L11.66 16.17 7.48 12l-1.41 1.41L11.66 19l12-12-1.42-1.41zM.41 13.41L6 19l1.41-1.41L1.83 12 .41 13.41z" />
              </svg>
              <span class="option-title">Confirmação de Leitura</span>
            </div>
            <p class="option-desc">Enviar confirmação de leitura automaticamente. Quando ativo, o visto azul é enviado ao receber mensagens.</p>
          </div>
          <TriStateToggle
            v-model="options.readreceipts"
            @change="updateOption('readreceipts', $event)"
            :disabled="togglingOption === 'server-readreceipts'"
          />
        </div>

        <div class="option-card">
          <div class="option-info">
            <div class="option-header">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M20.01 15.38c-1.23 0-2.42-.2-3.53-.56-.35-.12-.74-.03-1.01.24l-1.57 1.97c-2.83-1.35-5.48-3.9-6.89-6.83l1.95-1.66c.27-.28.35-.67.24-1.02-.37-1.11-.56-2.3-.56-3.53 0-.54-.45-.99-.99-.99H4.19C3.65 3 3 3.24 3 3.99 3 13.28 10.73 21 20.01 21c.71 0 .99-.63.99-1.18v-3.45c0-.54-.45-.99-.99-.99z" />
              </svg>
              <span class="option-title">Notificação de Chamadas</span>
            </div>
            <p class="option-desc">Receber notificações de chamadas via webhook. Quando ativo, você será notificado sobre chamadas recebidas.</p>
          </div>
          <TriStateToggle
            v-model="options.calls"
            @change="updateOption('calls', $event)"
            :disabled="togglingOption === 'server-calls'"
          />
        </div>
      </div>
    </div>

    <div class="danger-section" v-if="server">
      <h2>
        <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
          <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z" />
        </svg>
        Zona de Perigo
      </h2>

      <div class="danger-actions">
        <button @click="toggleServer" class="btn-warning-outline" :disabled="togglingServer">
          <svg v-if="isServerActive" viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M13 3h-2v10h2V3zm4.83 2.17l-1.42 1.42C17.99 7.86 19 9.81 19 12c0 3.87-3.13 7-7 7s-7-3.13-7-7c0-2.19 1.01-4.14 2.58-5.42L6.17 5.17C4.23 6.82 3 9.26 3 12c0 4.97 4.03 9 9 9s9-4.03 9-9c0-2.74-1.23-5.18-3.17-6.83z" />
          </svg>
          <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z" />
          </svg>
          {{ togglingServer ? "Processando..." : isServerActive ? "Desativar Servidor" : "Ativar Servidor" }}
        </button>

        <button @click="confirmDelete" class="btn-danger" :disabled="deleting">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
          </svg>
          {{ deleting ? "Excluindo..." : "Excluir Servidor" }}
        </button>
      </div>
    </div>

    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-icon danger">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor">
            <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
          </svg>
        </div>
        <h3>Excluir Servidor?</h3>
        <p>Esta ação não pode ser desfeita. O servidor e todas as configurações serão removidos permanentemente.</p>
        <div class="modal-actions">
          <button @click="showDeleteModal = false" class="btn-secondary">Cancelar</button>
          <button @click="deleteServer" class="btn-danger">Excluir</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import cableService from '@/services/cable'
import { pushToast } from '@/services/toast'
import TriStateToggle from '@/components/TriStateToggle.vue'

export default defineComponent({
  components: {
    TriStateToggle,
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const token = route.params.token as string

    const server = ref<any>(null)
    const serverState = ref('')
    const serverConnected = ref(false)
    const serverWid = ref('')
    const loading = ref(true)
    const error = ref('')
    const tokenCopied = ref(false)
    const showDeleteModal = ref(false)
    const deleting = ref(false)
    const togglingServer = ref(false)
    const togglingOption = ref('')

    const options = ref({
      broadcasts: 0,
      groups: 0,
      readreceipts: 0,
      calls: 0,
    })

    const statusClass = computed(() => {
      const state = serverState.value.toLowerCase()
      if (state === 'ready') return 'connected'
      if (state === 'connecting' || state === 'starting' || state === 'reconnecting') return 'connecting'
      return 'disconnected'
    })

    const isConnected = computed(() => serverConnected.value === true)

    const isServerActive = computed(() => {
      const activeStates = ['ready', 'connecting', 'starting', 'reconnecting']
      return activeStates.includes(serverState.value.toLowerCase())
    })

    function toTriState(value: any): number {
      if (value === 1 || value === true) return 1
      if (value === -1 || value === false) return -1
      return 0
    }

    function formatWid(wid: string | null | undefined): string {
      if (!wid) return ''

      let phone = wid.split('@')[0]
      phone = phone.split(':')[0]
      return phone
    }

    async function load() {
      loading.value = true
      error.value = ''

      try {
        const res = await api.get(`/spa/server/${token}/info`)
        const summary = res.data?.server || {}

        server.value = summary
        serverState.value = summary.state || ''
        serverConnected.value = summary.state === 'Ready' || summary.stateCode === 11
        serverWid.value = summary.wid || ''

        options.value.broadcasts = toTriState(summary.broadcasts)
        options.value.groups = toTriState(summary.groups)
        options.value.readreceipts = toTriState(summary.readReceipts ?? summary.readreceipts)
        options.value.calls = toTriState(summary.calls)
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao carregar servidor'
      } finally {
        loading.value = false
      }
    }

    async function updateOption(optionName: string, value: number) {
      togglingOption.value = `server-${optionName}`

      try {
        const payload: Record<string, number> = {}
        payload[optionName] = value

        await api.patch(`/spa/server/${token}`, payload)
        await load()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao alterar opção'
        await load()
      } finally {
        togglingOption.value = ''
      }
    }

    async function copyToken() {
      try {
        await navigator.clipboard.writeText(token)
        tokenCopied.value = true
        setTimeout(() => {
          tokenCopied.value = false
        }, 2000)
      } catch {
        // ignore clipboard fallback failures
      }
    }

    function confirmDelete() {
      showDeleteModal.value = true
    }

    async function deleteServer() {
      deleting.value = true

      try {
        await api.delete(`/spa/server/${token}`)
        router.push('/')
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao excluir servidor'
      } finally {
        deleting.value = false
        showDeleteModal.value = false
      }
    }

    async function toggleServer() {
      togglingServer.value = true

      try {
        const endpoint = isServerActive.value ? 'disable' : 'enable'
        await api.post(`/spa/server/${token}/${endpoint}`)
        await load()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao alterar estado do servidor'
      } finally {
        togglingServer.value = false
      }
    }

    const lifecycleEvents = [
      'server.connected',
      'server.disconnected',
      'server.stopped',
      'server.logged_out',
      'server.deleted',
    ]
    let cableListeners: Array<() => void> = []

    onMounted(() => {
      load()

      void cableService.connect().catch(() => {
        // The detail page can still rely on manual refresh if websocket auth fails.
      })

      cableListeners = lifecycleEvents.map((eventName) =>
        cableService.onEvent(eventName, async (payload: any) => {
          if (payload?.token !== token) {
            return
          }

          if (eventName === 'server.deleted') {
            pushToast('Servidor removido', 'info')
            router.push('/')
            return
          }

          try {
            await load()
          } catch {
            // Keep the previous state visible if the realtime refresh fails.
          }
        }),
      )
    })

    onUnmounted(() => {
      for (const unsubscribe of cableListeners) {
        unsubscribe()
      }
      cableListeners = []
      void cableService.disconnect()
    })

    return {
      confirmDelete,
      copyToken,
      deleteServer,
      error,
      formatWid,
      isConnected,
      isServerActive,
      load,
      loading,
      options,
      server,
      serverConnected,
      serverState,
      serverWid,
      showDeleteModal,
      statusClass,
      deleting,
      token,
      tokenCopied,
      toggleServer,
      togglingOption,
      togglingServer,
      updateOption,
    }
  },
})
</script>

<style scoped>
.server-page {
  margin: 0 auto;
}

.server-header {
  margin-bottom: 32px;
}

.header-top {
  margin-bottom: 24px;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  text-decoration: none;
  font-size: 14px;
}

.back-link:hover {
  color: #374151;
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

.server-info {
  display: flex;
  align-items: center;
  gap: 20px;
}

.server-avatar {
  width: 80px;
  height: 80px;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.server-avatar.connected {
  background: linear-gradient(135deg, var(--branding-primary, #7c3aed), var(--branding-secondary, #5b21b6));
}

.server-avatar.connecting {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.server-avatar.disconnected {
  background: linear-gradient(135deg, #6b7280, #4b5563);
}

.server-details h1 {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 8px;
}

.server-meta {
  display: flex;
  gap: 10px;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.connected {
  background: #f5efff;
  color: var(--branding-secondary, #5b21b6);
}

.status-badge.connecting {
  background: #fef3c7;
  color: #92400e;
}

.status-badge.disconnected {
  background: #f3f4f6;
  color: #6b7280;
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
  margin-top: 16px;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 32px;
}

.action-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: white;
  border-radius: 16px;
  text-decoration: none;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.2s;
  border: 2px solid transparent;
}

.action-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
  border-color: var(--branding-primary, #7c3aed);
}

.action-card.primary {
  background: linear-gradient(135deg, var(--branding-primary, #7c3aed), var(--branding-secondary, #5b21b6));
  color: white;
}

.action-card.primary:hover {
  border-color: white;
}

.action-card.disabled {
  opacity: 0.5;
  pointer-events: none;
}

.action-icon {
  width: 56px;
  height: 56px;
  background: rgba(124, 58, 237, 0.08);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--branding-primary, #7c3aed);
}

.action-card.primary .action-icon {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.action-info {
  display: flex;
  flex-direction: column;
}

.action-title {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.action-card.primary .action-title {
  color: white;
}

.action-desc {
  font-size: 13px;
  color: #6b7280;
  margin-top: 2px;
}

.action-card.primary .action-desc {
  color: rgba(255, 255, 255, 0.8);
}

.details-section,
.danger-section {
  background: white;
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.details-section h2,
.danger-section h2,
.options-section h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
  margin: 0 0 20px;
}

.danger-section {
  margin-top: 32px;
}

.details-section h2 svg {
  color: #3b82f6;
}

.danger-section h2 svg {
  color: #ef4444;
}

.options-section h2 svg {
  color: var(--branding-primary, #7c3aed);
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.detail-card {
  padding: 16px;
  background: #f9fafb;
  border-radius: 12px;
}

.detail-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  margin-bottom: 8px;
}

.detail-value {
  font-size: 15px;
  color: #111827;
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-value.token {
  justify-content: space-between;
}

.detail-value code {
  font-size: 12px;
  background: #e5e7eb;
  padding: 4px 8px;
  border-radius: 6px;
  word-break: break-all;
}

.copy-btn {
  padding: 6px;
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  border-radius: 6px;
}

.copy-btn:hover {
  background: #e5e7eb;
}

.copy-btn.copied {
  color: var(--branding-primary, #7c3aed);
}

.danger-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.btn-danger,
.btn-warning-outline {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-warning-outline {
  background: white;
  border: 2px solid #f59e0b;
  color: #f59e0b;
}

.btn-warning-outline:hover:not(:disabled) {
  background: #fffbeb;
}

.btn-warning-outline:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-danger {
  background: #ef4444;
  border: 2px solid #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  background: white;
  border-radius: 20px;
  padding: 32px;
  max-width: 400px;
  text-align: center;
}

.modal-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-icon.danger {
  background: #fef2f2;
  color: #ef4444;
}

.modal-content h3 {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 12px;
}

.modal-content p {
  color: #6b7280;
  margin: 0 0 24px;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.btn-secondary {
  padding: 12px 24px;
  background: #f3f4f6;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  cursor: pointer;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.options-section {
  margin-top: 32px;
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

.option-header svg {
  color: var(--branding-primary, #7c3aed);
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

.history-sync-card {
  flex-wrap: wrap;
}

.readonly-badge {
  flex-shrink: 0;
  padding: 8px 12px;
  border-radius: 999px;
  background: #f3f4f6;
  color: #6b7280;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

@media (max-width: 768px) {
  .hide-mobile {
    display: none !important;
  }

  .server-page {
    padding: 0;
  }

  .server-header {
    position: sticky;
    top: 0;
    z-index: 100;
    margin: 0;
    border-radius: 0;
    padding: 16px;
  }

  .danger-actions {
    flex-direction: column;
  }

  .danger-actions button {
    width: 100%;
  }

  .quick-actions {
    grid-template-columns: 1fr;
  }

  .details-grid {
    grid-template-columns: 1fr;
  }
}
</style>
