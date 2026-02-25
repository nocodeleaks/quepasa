<template>
  <div class="server-page">
    <!-- Header -->
    <div class="server-header">
      <div class="header-top">
        <button @click="$router.back()" class="back-link hide-mobile">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
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
            <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2z"/>
          </svg>
        </div>
        <div class="server-details">
          <h1>{{ formatWid(serverWid) || formatWid(server.wid) || 'Não conectado' }}</h1>
          <div class="server-meta">
            Status <span class="status-badge" :class="statusClass">{{ serverState || 'Desconhecido' }}</span>
          </div>
        </div>
      </div>

      <div v-if="error" class="error-banner">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
        </svg>
        <span>{{ error }}</span>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions" v-if="server">
      <router-link :to="`/server/${token}/qrcode`" class="action-card" :class="{ disabled: isConnected }">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M3 11h8V3H3v8zm2-6h4v4H5V5zM3 21h8v-8H3v8zm2-6h4v4H5v-4zm8-12v8h8V3h-8zm6 6h-4V5h4v4zm-6 4h2v2h-2zm2 2h2v2h-2zm-2 2h2v2h-2zm4 0h2v2h-2zm2 2h2v2h-2zm0-4h2v2h-2zm2-2h2v2h-2z"/>
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
            <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
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
            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
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
            <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">Mensagens</span>
          <span class="action-desc">Ver mensagens recebidas</span>
        </div>
      </router-link>

      <router-link :to="`/server/${token}/groups`" class="action-card">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
          </svg>
        </div>
        <div class="action-info">
          <span class="action-title">Grupos</span>
          <span class="action-desc">Ver e gerenciar grupos</span>
        </div>
      </router-link>
    </div>

    <!-- Server Details -->
    <div class="details-section" v-if="server">
      <h2>
        <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/>
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
                <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
              </svg>
            </button>
          </div>
        </div>

        <div class="detail-card">
          <span class="detail-label">WhatsApp ID</span>
          <span class="detail-value">{{ formatWid(serverWid) || formatWid(server.wid) || '—' }}</span>
        </div>

        <!-- <div class="detail-card">
          <span class="detail-label">Status</span>
          <span class="detail-value">
            <span class="status-dot" :class="statusClass"></span>
            {{ serverState || 'Desconhecido' }}
          </span>
        </div> -->

        <div class="detail-card">
          <span class="detail-label">Reconectar</span>
          <span class="detail-value">{{ server.reconnect ? 'Sim' : 'Não' }}</span>
        </div>
      </div>
    </div>

    <!-- Server Options (Toggles) -->
    <div class="options-section" v-if="server">
      <h2>
        <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
          <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/>
        </svg>
        Opções do Servidor
      </h2>

      <div class="options-list">
        <div class="option-card history-sync-card">
          <div class="option-info">
            <div class="option-header">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M13 3c-4.97 0-9 4.03-9 9H1l3.89 3.89.07.14L9 12H6c0-3.87 3.13-7 7-7s7 3.13 7 7-3.13 7-7 7c-1.93 0-3.68-.79-4.94-2.06l-1.42 1.42C8.27 19.99 10.51 21 13 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 5v5l4.28 2.54.72-1.21-3.5-2.08V8H12z"/>
              </svg>
              <span class="option-title">Sincronização de Histórico</span>
            </div>
            <p class="option-desc">Número de dias de histórico a sincronizar no primeiro pareamento.</p>
          </div>
          <div class="history-sync-control">
            <div class="history-input-group">
              <input type="number" min="0" max="365" v-model="options.historysync" class="history-input" placeholder="0" />
              <span class="history-input-suffix">dias</span>
            </div>
            <button class="btn-save" @click="saveHistorySync" :disabled="savingHistory">
              <svg v-if="!savingHistory" viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
              </svg>
              <span v-else class="spinner-tiny"></span>
            </button>
          </div>
        </div>

        <div class="option-card">
          <div class="option-info">
            <div class="option-header">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                <path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/>
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
                <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
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
                <path d="M18 7l-1.41-1.41-6.34 6.34 1.41 1.41L18 7zm4.24-1.41L11.66 16.17 7.48 12l-1.41 1.41L11.66 19l12-12-1.42-1.41zM.41 13.41L6 19l1.41-1.41L1.83 12 .41 13.41z"/>
              </svg>
              <span class="option-title">Confirmação de Leitura</span>
            </div>
            <p class="option-desc">Enviar confirmação de leitura automaticamente. Quando ativo, o "visto azul" é enviado ao receber mensagens.</p>
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
                <path d="M20.01 15.38c-1.23 0-2.42-.2-3.53-.56-.35-.12-.74-.03-1.01.24l-1.57 1.97c-2.83-1.35-5.48-3.9-6.89-6.83l1.95-1.66c.27-.28.35-.67.24-1.02-.37-1.11-.56-2.3-.56-3.53 0-.54-.45-.99-.99-.99H4.19C3.65 3 3 3.24 3 3.99 3 13.28 10.73 21 20.01 21c.71 0 .99-.63.99-1.18v-3.45c0-.54-.45-.99-.99-.99z"/>
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

    <!-- Danger Zone -->
    <div class="danger-section" v-if="server">
      <h2>
        <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
          <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
        </svg>
        Zona de Perigo
      </h2>

      <div class="danger-actions">
        <!-- Toggle Server (Activate/Deactivate) -->
        <button @click="toggleServer" class="btn-warning-outline" :disabled="togglingServer">
          <svg v-if="isServerActive" viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M13 3h-2v10h2V3zm4.83 2.17l-1.42 1.42C17.99 7.86 19 9.81 19 12c0 3.87-3.13 7-7 7s-7-3.13-7-7c0-2.19 1.01-4.14 2.58-5.42L6.17 5.17C4.23 6.82 3 9.26 3 12c0 4.97 4.03 9 9 9s9-4.03 9-9c0-2.74-1.23-5.18-3.17-6.83z"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z"/>
          </svg>
          {{ togglingServer ? 'Processando...' : (isServerActive ? 'Desativar Servidor' : 'Ativar Servidor') }}
        </button>

        <button v-if="isConnected" @click="disconnect" class="btn-danger-outline" :disabled="disconnecting">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M13 3h-2v10h2V3zm4.83 2.17l-1.42 1.42C17.99 7.86 19 9.81 19 12c0 3.87-3.13 7-7 7s-7-3.13-7-7c0-2.19 1.01-4.14 2.58-5.42L6.17 5.17C4.23 6.82 3 9.26 3 12c0 4.97 4.03 9 9 9s9-4.03 9-9c0-2.74-1.23-5.18-3.17-6.83z"/>
          </svg>
          {{ disconnecting ? 'Desconectando...' : 'Desconectar' }}
        </button>

        <button @click="confirmDelete" class="btn-danger" :disabled="deleting">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
          </svg>
          {{ deleting ? 'Excluindo...' : 'Excluir Servidor' }}
        </button>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-icon danger">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor">
            <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
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
import { defineComponent, ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import TriStateToggle from '@/components/TriStateToggle.vue'

export default defineComponent({
  components: {
    TriStateToggle
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const token = route.params.token as string

    const server = ref<any>(null)
    const serverState = ref<string>('')
    const serverConnected = ref<boolean>(false)
    const serverWid = ref<string>('')
    const loading = ref(true)
    const error = ref('')
    const tokenCopied = ref(false)
    const showDeleteModal = ref(false)
    const deleting = ref(false)
    const disconnecting = ref(false)
    const togglingServer = ref(false)
    
    // Toggle options - now using numeric values: -1 = off, 0 = unset, 1 = on
    const options = ref({
      broadcasts: 0,
      groups: 0,
      readreceipts: 0,
      calls: 0,
      // history sync value (string or number). Empty string means unset
      historysync: ''
    })
    const togglingOption = ref('')

    const baseUrl = computed(() => window.location.origin + '/api')

    const statusClass = computed(() => {
      const state = serverState.value?.toLowerCase() || ''
      if (state === 'ready') return 'connected'
      if (state === 'connecting' || state === 'starting') return 'connecting'
      return 'disconnected'
    })

    const isConnected = computed(() => {
      return serverConnected.value === true
    })

    const isServerActive = computed(() => {
      const state = serverState.value?.toLowerCase() || ''
      // Server is active if state is not 'stopped' and not empty
      return state !== '' && state !== 'stopped' && state !== 'disabled'
    })

    async function load() {
      loading.value = true
      error.value = ''
      try {
        const res = await api.get(`/api/server/${token}/info`)
        server.value = res.data?.server
        serverState.value = res.data?.state || ''
        serverConnected.value = res.data?.connected === true
        serverWid.value = res.data?.wid || ''
        
        // Load toggle options from server data (numeric: -1, 0, 1)
        if (res.data?.server) {
          options.value.broadcasts = toTriState(res.data.server.broadcasts)
          options.value.groups = toTriState(res.data.server.groups)
          options.value.readreceipts = toTriState(res.data.server.readreceipts)
          options.value.calls = toTriState(res.data.server.calls)
          // Load history sync (may be null)
          options.value.historysync = res.data.server.historysync == null ? '' : String(res.data.server.historysync)
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao carregar servidor'
      } finally {
        loading.value = false
      }
    }
    
    const savingHistory = ref(false)

    async function saveHistorySync() {
      savingHistory.value = true
      error.value = ''
      try {
        // Convert empty string to null to clear
        const payload: any = {}
        if (options.value.historysync === '' || options.value.historysync === null) {
          payload.historysyncdays = null
        } else {
          payload.historysyncdays = Number(options.value.historysync)
        }
        await api.post(`/api/server/${token}/update`, payload)
        await load()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao salvar history sync'
      } finally {
        savingHistory.value = false
      }
    }

    // Convert API value to tri-state number (-1, 0, 1)
    function toTriState(val: any): number {
      if (val === 1 || val === true) return 1
      if (val === -1 || val === false) return -1
      return 0 // null, undefined, 0 = unset
    }

    // Update option with new tri-state value
    async function updateOption(optionName: string, value: number) {
      togglingOption.value = `server-${optionName}`
      try {
        // Use the update endpoint with the specific value
        const payload: Record<string, number> = {}
        payload[optionName] = value
        await api.post(`/api/server/${token}/update`, payload)
        await load()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao alterar opção'
        // Reload to revert
        await load()
      } finally {
        togglingOption.value = ''
      }
    }

    async function toggleOption(key: string) {
      togglingOption.value = key
      try {
        // server-* options should use /api/command instead of /api/toggle
        if (key.startsWith('server-')) {
          const mapping: Record<string,string> = {
            'server-groups': 'groups',
            'server-broadcasts': 'broadcasts',
            'server-readreceipts': 'readreceipts',
            'server-calls': 'calls',
            'server-readupdate': 'readupdate'
          }
          const action = mapping[key]
          if (!action) throw new Error('invalid server option')
          await api.post('/api/command', { token, action })
        } else {
          await api.post('/api/toggle', { token, key })
        }
        await load() // Reload to get new values
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao alterar opção'
        await load() // Reload to revert
      } finally {
        togglingOption.value = ''
      }
    }

    // Format WID to show only phone number (remove session/server and @s.whatsapp.net)
    // Example: "554333749900:44@s.whatsapp.net" -> "554333749900"
    function formatWid(wid: string | null | undefined): string {
      if (!wid) return ''
      // Remove @s.whatsapp.net, @lid, @g.us, etc.
      let phone = wid.split('@')[0]
      // Remove session/server part (after colon)
      phone = phone.split(':')[0]
      return phone
    }

    async function copyToken() {
      try {
        await navigator.clipboard.writeText(token)
        tokenCopied.value = true
        setTimeout(() => tokenCopied.value = false, 2000)
      } catch {
        // fallback
      }
    }

    function confirmDelete() {
      showDeleteModal.value = true
    }

    async function deleteServer() {
      deleting.value = true
      try {
        await api.post('/api/delete', { token })
        router.push('/')
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao excluir servidor'
      } finally {
        deleting.value = false
        showDeleteModal.value = false
      }
    }

    async function disconnect() {
      disconnecting.value = true
      try {
        await api.post(`/bot/${token}/disconnect`, {}, {
          headers: { 'X-QUEPASA-TOKEN': token }
        })
        await load()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao desconectar'
      } finally {
        disconnecting.value = false
      }
    }

    async function toggleServer() {
      togglingServer.value = true
      try {
        const action = isServerActive.value ? 'stop' : 'start'
        await api.post('/api/command', { token, action })
        // Reload server info to get updated state
        await load()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao alterar estado do servidor'
      } finally {
        togglingServer.value = false
      }
    }

    onMounted(() => {
      load()
    })

    return {
      token, server, serverState, serverConnected, serverWid, loading, error, baseUrl, statusClass, isConnected,
      tokenCopied, copyToken, showDeleteModal, confirmDelete, deleteServer,
      deleting, disconnect, disconnecting,
      options, togglingOption, toggleOption, updateOption, toTriState,
      isServerActive, togglingServer, toggleServer,
      savingHistory, saveHistorySync, formatWid
    }
  }
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
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
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

.server-avatar.connected { background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6)); }
.server-avatar.connecting { background: linear-gradient(135deg, #f59e0b, #d97706); }
.server-avatar.disconnected { background: linear-gradient(135deg, #6b7280, #4b5563); }

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

.status-badge.connected { background: #f5efff; color: var(--branding-secondary, #5B21B6); }
.status-badge.connecting { background: #fef3c7; color: #92400e; }
.status-badge.disconnected { background: #f3f4f6; color: #6b7280; }

.version-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  background: #e0e7ff;
  color: #4338ca;
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
  border-color: var(--branding-primary, #7C3AED);
}

.action-card.primary {
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
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
  color: var(--branding-primary, #7C3AED);
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

.details-section, .endpoints-section, .danger-section {
  background: white;
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.details-section h2, .endpoints-section h2, .danger-section h2 {
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

.details-section h2 svg { color: #3b82f6; }
.endpoints-section h2 svg { color: #8b5cf6; }
.danger-section h2 svg { color: #ef4444; }

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
  color: var(--branding-primary, #7C3AED);
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.status-dot.connected { background: var(--branding-primary, #7C3AED); }
.status-dot.connecting { background: #f59e0b; }
.status-dot.disconnected { background: #9ca3af; }

.endpoints-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.endpoint-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 10px;
}

.endpoint-method {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
}

.endpoint-method.get { background: #f5efff; color: var(--branding-secondary, #5B21B6); }
.endpoint-method.post { background: #dbeafe; color: #1e40af; }

.endpoint-url {
  flex: 1;
  font-size: 13px;
  background: transparent;
  word-break: break-all;
}

.endpoint-desc {
  font-size: 12px;
  color: #6b7280;
}

.danger-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.btn-danger-outline, .btn-danger, .btn-warning-outline {
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

.btn-danger-outline {
  background: white;
  border: 2px solid #ef4444;
  color: #ef4444;
}

.btn-danger-outline:hover:not(:disabled) {
  background: #fef2f2;
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

.btn-danger-outline:disabled, .btn-danger:disabled {
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

/* Options Section */
.options-section {
  margin-top: 32px;
}

.options-section h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: #374151;
  margin: 0 0 16px;
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
  color: var(--branding-primary, #7C3AED);
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

/* Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 52px;
  height: 28px;
  flex-shrink: 0;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #d1d5db;
  transition: 0.3s;
  border-radius: 28px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 22px;
  width: 22px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.toggle-switch input:checked + .toggle-slider {
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
}

.toggle-switch input:checked + .toggle-slider:before {
  transform: translateX(24px);
}

.toggle-switch input:disabled + .toggle-slider {
  opacity: 0.5;
  cursor: not-allowed;
}

/* History Sync Control */
.history-sync-card {
  flex-wrap: wrap;
}

.history-sync-control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.history-input-group {
  display: flex;
  align-items: center;
  background: #f3f4f6;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 0.2s;
}

.history-input-group:focus-within {
  border-color: var(--branding-primary, #7C3AED);
  background: white;
}

.history-input {
  width: 60px;
  padding: 10px 12px;
  border: none;
  background: transparent;
  font-size: 15px;
  font-weight: 600;
  color: #374151;
  text-align: center;
  outline: none;
}

.history-input::placeholder {
  color: #9ca3af;
  font-weight: 400;
}

.history-input::-webkit-outer-spin-button,
.history-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.history-input-suffix {
  padding: 10px 12px 10px 0;
  color: #6b7280;
  font-size: 14px;
  font-weight: 500;
}

.btn-save {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  border: none;
  border-radius: 10px;
  color: white;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.btn-save:hover:not(:disabled) {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.3);
}

.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner-tiny {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Responsive Mobile Styles */
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
