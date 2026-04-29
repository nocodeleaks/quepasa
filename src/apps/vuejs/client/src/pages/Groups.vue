<template>
  <div class="groups-page">
    <div class="page-header">
      <div class="header-left">
        <router-link :to="`/server/${token}`" class="back-btn">
          <i class="fa fa-arrow-left"></i>
        </router-link>
        <h1>Grupos</h1>
        <span class="group-count" v-if="groups.length > 0">({{ groups.length }})</span>
      </div>
      <div class="header-actions">
        <div class="view-toggle">
          <button
            :class="['toggle-btn', { active: viewMode === 'card' }]"
            @click="viewMode = 'card'"
            title="Modo Card"
          >
            <i class="fa fa-th-large"></i>
          </button>
          <button
            :class="['toggle-btn', { active: viewMode === 'list' }]"
            @click="viewMode = 'list'"
            title="Modo Lista"
          >
            <i class="fa fa-list"></i>
          </button>
        </div>
        <button @click="load" class="btn-icon" title="Atualizar">
          <i class="fa fa-refresh" :class="{ 'fa-spin': loading }"></i>
        </button>
        <button @click="createGroup" class="btn-primary">
          <i class="fa fa-plus me-2"></i>Criar grupo
        </button>
      </div>
    </div>

    <div class="search-bar">
      <i class="fa fa-search"></i>
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Pesquisar grupos..."
        class="search-input"
      />
      <button v-if="searchQuery" @click="searchQuery = ''" class="clear-search">
        <i class="fa fa-times"></i>
      </button>
    </div>

    <div v-if="groups.length > 0" class="pagination-controls">
      <button @click="goToPage(1)" :disabled="currentPage === 1 || loading" class="btn-page">
        <i class="fa fa-angle-double-left"></i>
      </button>
      <button @click="prevPage" :disabled="currentPage === 1 || loading" class="btn-page">
        <i class="fa fa-angle-left"></i>
      </button>
      <span class="page-info">Pagina {{ currentPage }} de {{ totalPages }}</span>
      <button @click="nextPage" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-right"></i>
      </button>
      <button @click="goToPage(totalPages)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-double-right"></i>
      </button>
      <select v-model.number="pageSize" class="page-size-select" :disabled="loading">
        <option :value="5">5 por pagina</option>
        <option :value="10">10 por pagina</option>
        <option :value="15">15 por pagina</option>
        <option :value="25">25 por pagina</option>
        <option :value="50">50 por pagina</option>
        <option :value="100">100 por pagina</option>
        <option :value="200">200 por pagina</option>
      </select>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Carregando grupos...</p>
    </div>

    <div v-else-if="error" class="error-box">
      <i class="fa fa-exclamation-triangle"></i>
      {{ error }}
    </div>

    <div v-else-if="filteredGroups.length === 0" class="empty-state">
      <i class="fa fa-users"></i>
      <p v-if="searchQuery">Nenhum grupo encontrado para "{{ searchQuery }}"</p>
      <p v-else>Nenhum grupo encontrado</p>
    </div>

    <div v-else-if="viewMode === 'card'" class="groups-grid">
      <div
        v-for="group in displayGroups"
        :key="group.JID"
        class="group-card"
        @click="goToGroup(group.JID)"
      >
        <div class="card-avatar">
          <img
            v-if="groupPictures[group.JID]"
            :src="groupPictures[group.JID]"
            :alt="group.Name"
            @error="handleImageError(group.JID)"
          />
          <div v-else class="avatar-placeholder">
            <i class="fa fa-users"></i>
          </div>
        </div>
        <div class="card-content">
          <div class="card-header">
            <h3 class="group-name">{{ group.Name || 'Grupo sem nome' }}</h3>
            <span class="message-time" v-if="group.lastMessage">{{ formatTime(group.lastMessage.timestamp) }}</span>
          </div>
          <div class="card-body">
            <p class="last-message" v-if="group.lastMessage">
              <span class="sender-name">{{ group.lastMessage.senderName }}:</span>
              {{ truncateMessage(group.lastMessage.text) }}
            </p>
            <p class="last-message empty" v-else>
              <i class="fa fa-comment-o me-1"></i>Sem mensagens recentes
            </p>
          </div>
          <div class="card-footer">
            <span class="participants-count">
              <i class="fa fa-user me-1"></i>{{ group.Participants?.length || 0 }}
            </span>
            <span v-if="group.IsAnnounce" class="badge badge-announce" title="Apenas admins podem enviar">
              <i class="fa fa-bullhorn"></i>
            </span>
            <span v-if="group.IsParent" class="badge badge-community" title="Comunidade">
              <i class="fa fa-sitemap"></i>
            </span>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="groups-list">
      <div
        v-for="group in displayGroups"
        :key="group.JID"
        class="group-item"
        @click="goToGroup(group.JID)"
      >
        <div class="item-avatar">
          <img
            v-if="groupPictures[group.JID]"
            :src="groupPictures[group.JID]"
            :alt="group.Name"
            @error="handleImageError(group.JID)"
          />
          <div v-else class="avatar-placeholder">
            <i class="fa fa-users"></i>
          </div>
        </div>
        <div class="item-content">
          <div class="item-header">
            <h3 class="group-name">{{ group.Name || 'Grupo sem nome' }}</h3>
            <span class="message-time" v-if="group.lastMessage">{{ formatTime(group.lastMessage.timestamp) }}</span>
          </div>
          <div class="item-body">
            <p class="last-message" v-if="group.lastMessage">
              <span class="sender-name">{{ group.lastMessage.senderName }}:</span>
              {{ truncateMessage(group.lastMessage.text) }}
            </p>
            <p class="last-message empty" v-else>
              <span class="participants-info">
                <i class="fa fa-user me-1"></i>{{ group.Participants?.length || 0 }} participantes
              </span>
            </p>
          </div>
        </div>
        <div class="item-badges">
          <span v-if="group.IsAnnounce" class="badge badge-announce" title="Apenas admins podem enviar">
            <i class="fa fa-bullhorn"></i>
          </span>
          <span v-if="group.IsParent" class="badge badge-community" title="Comunidade">
            <i class="fa fa-sitemap"></i>
          </span>
        </div>
      </div>
    </div>

    <div v-if="totalPages > 1" class="pagination-controls">
      <button @click="goToPage(1)" :disabled="currentPage === 1 || loading" class="btn-page">
        <i class="fa fa-angle-double-left"></i>
      </button>
      <button @click="prevPage" :disabled="currentPage === 1 || loading" class="btn-page">
        <i class="fa fa-angle-left"></i>
      </button>
      <span class="page-info">Pagina {{ currentPage }} de {{ totalPages }}</span>
      <button @click="nextPage" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-right"></i>
      </button>
      <button @click="goToPage(totalPages)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-double-right"></i>
      </button>
      <select v-model.number="pageSize" class="page-size-select" :disabled="loading">
        <option :value="5">5 por pagina</option>
        <option :value="10">10 por pagina</option>
        <option :value="15">15 por pagina</option>
        <option :value="25">25 por pagina</option>
        <option :value="50">50 por pagina</option>
        <option :value="100">100 por pagina</option>
        <option :value="200">200 por pagina</option>
      </select>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

interface LastMessage {
  timestamp: string
  senderName: string
  text: string
}

interface Group {
  JID: string
  Name: string
  Participants: any[]
  IsAnnounce: boolean
  IsParent: boolean
  lastMessage?: LastMessage
}

export default defineComponent({
  setup() {
    const route = useRoute()
    const router = useRouter()
    const token = route.params.token as string

    const groups = ref<Group[]>([])
    const groupPictures = ref<Record<string, string>>({})
    const loading = ref(false)
    const error = ref('')
    const searchQuery = ref('')
    const viewMode = ref<'card' | 'list'>('card')
    const currentPage = ref(1)
    const pageSize = ref(24)

    const filteredGroups = computed(() => {
      if (!searchQuery.value) return groups.value
      const query = searchQuery.value.toLowerCase()
      return groups.value.filter((group) => {
        return group.Name?.toLowerCase().includes(query) || group.JID.toLowerCase().includes(query)
      })
    })

    const totalPages = computed(() => Math.ceil(filteredGroups.value.length / pageSize.value) || 1)

    const displayGroups = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value
      return filteredGroups.value.slice(start, start + pageSize.value)
    })

    async function load() {
      loading.value = true
      error.value = ''

      try {
        const res = await api.get('/api/groups', { params: { token } })
        const rawGroups = (res.data?.groups || []) as Group[]

        rawGroups.sort((a, b) => (a.Name || '').localeCompare(b.Name || ''))
        groups.value = rawGroups

        void loadGroupPictures(rawGroups.slice(0, 20))
        await loadLastMessages()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.message || 'Erro ao carregar grupos'
      } finally {
        loading.value = false
      }
    }

    watch([searchQuery, pageSize], () => {
      currentPage.value = 1
    })

    watch([filteredGroups, totalPages], () => {
      if (currentPage.value > totalPages.value) currentPage.value = totalPages.value
      if (currentPage.value < 1) currentPage.value = 1
    })

    function goToPage(page: number) {
      if (page >= 1 && page <= totalPages.value) {
        currentPage.value = page
      }
    }

    function nextPage() {
      if (currentPage.value < totalPages.value) currentPage.value += 1
    }

    function prevPage() {
      if (currentPage.value > 1) currentPage.value -= 1
    }

    async function loadGroupPictures(groupList: Group[]) {
      for (const group of groupList) {
        try {
          const res = await api.post('/api/media/pictures/info', { token, chatId: group.JID })
          if (res.data?.info?.url) {
            groupPictures.value[group.JID] = res.data.info.url
          }
        } catch {
          // ignore picture failures
        }
      }
    }

    function buildPreview(msg: any): string {
      let previewText = msg.text || ''

      if (msg.attachment) {
        const mime = msg.attachment.mimetype || ''

        if (mime.startsWith('image/')) previewText = previewText ? `[IMG] ${previewText}` : '[IMG] Imagem'
        else if (mime.startsWith('video/')) previewText = previewText ? `[VID] ${previewText}` : '[VID] Video'
        else if (mime.startsWith('audio/') || msg.type === 'ptt') previewText = previewText ? `[AUD] ${previewText}` : '[AUD] Audio'
        else if (mime.includes('pdf')) previewText = previewText ? `[PDF] ${previewText}` : '[PDF] PDF'
        else previewText = previewText ? `[ARQ] ${previewText}` : '[ARQ] Arquivo'
      }

      if (msg.type === 'sticker') {
        previewText = '[STK] Sticker'
      }

      if (!previewText && msg.inreply) {
        previewText = '[RPL] Resposta'
      }

      return previewText
    }

    async function loadLastMessages() {
      try {
        const res = await api.get('/api/messages', { params: { token } })
        const messages = res.data?.messages || []
        const lastMessages: Record<string, LastMessage> = {}

        for (const msg of messages) {
          const chatId = msg.chat?.id
          if (!chatId || !chatId.endsWith('@g.us')) continue
          if (msg.type === 'unhandled' || msg.type === 'revoked' || msg.type === 'system') continue
          if (msg.debug?.reason === 'discard') continue
          if (!msg.text && !msg.attachment && !msg.inreply) continue

          const previewText = buildPreview(msg)
          if (!previewText) continue

          if (!lastMessages[chatId] || new Date(msg.timestamp) > new Date(lastMessages[chatId].timestamp)) {
            lastMessages[chatId] = {
              timestamp: msg.timestamp,
              senderName: msg.participant?.title || msg.participant?.phone || 'Desconhecido',
              text: previewText,
            }
          }
        }

        groups.value = groups.value
          .map((group) => ({
            ...group,
            lastMessage: lastMessages[group.JID],
          }))
          .sort((a, b) => {
            if (a.lastMessage && b.lastMessage) {
              return new Date(b.lastMessage.timestamp).getTime() - new Date(a.lastMessage.timestamp).getTime()
            }
            if (a.lastMessage) return -1
            if (b.lastMessage) return 1
            return (a.Name || '').localeCompare(b.Name || '')
          })
      } catch {
        // optional enrichment
      }
    }

    function handleImageError(jid: string) {
      delete groupPictures.value[jid]
    }

    function formatTime(timestamp: string) {
      if (!timestamp) return ''

      const date = new Date(timestamp)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const oneDay = 24 * 60 * 60 * 1000

      if (diff < oneDay && date.getDate() === now.getDate()) {
        return date.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
      }
      if (diff < 2 * oneDay) return 'Ontem'
      if (diff < 7 * oneDay) return date.toLocaleDateString('pt-BR', { weekday: 'short' })
      return date.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })
    }

    function truncateMessage(text: string, maxLength = 50) {
      if (!text) return ''
      if (text.length <= maxLength) return text
      return `${text.substring(0, maxLength)}...`
    }

    function goToGroup(jid: string) {
      router.push(`/server/${token}/groups/${encodeURIComponent(jid)}`)
    }

    async function createGroup() {
      const title = prompt('Nome do grupo (<=25 caracteres):')
      if (!title) return

      const participantsRaw = prompt('Participantes (telefones separados por virgula):')
      if (!participantsRaw) return

      const participants = participantsRaw.split(',').map((value) => value.trim()).filter(Boolean)

      try {
        await api.post('/api/groups', { token, title, participants })
        pushToast('Grupo criado', 'success')
        await load()
      } catch (err: any) {
        pushToast(err?.response?.data?.result || err?.message || 'Erro ao criar grupo', 'error')
      }
    }

    onMounted(() => {
      load()
    })

    return {
      createGroup,
      currentPage,
      displayGroups,
      error,
      filteredGroups,
      formatTime,
      goToGroup,
      goToPage,
      groupPictures,
      groups,
      handleImageError,
      load,
      loading,
      nextPage,
      pageSize,
      prevPage,
      searchQuery,
      token,
      totalPages,
      truncateMessage,
      viewMode,
    }
  },
})
</script>

<style scoped>
.groups-page {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 15px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e5e7eb;
  color: #00034b;
  text-decoration: none;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #00034b;
  color: white;
}

.page-header h1 {
  margin: 0;
  font-size: 1.8rem;
  color: #111827;
}

.group-count {
  color: #6b7280;
  font-size: 1rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.view-toggle {
  display: flex;
  background: #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.toggle-btn {
  padding: 8px 12px;
  border: none;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn.active {
  background: #00034b;
  color: white;
}

.toggle-btn:hover:not(.active) {
  color: #111827;
}

.btn-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: #e5e7eb;
  color: #00034b;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: #00034b;
  color: white;
}

.btn-primary {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  background: #00034b;
  color: white;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: #000266;
  transform: translateY(-1px);
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #e5e7eb;
  border-radius: 12px;
  margin-bottom: 20px;
}

.search-bar i {
  color: #6b7280;
}

.search-input {
  flex: 1;
  border: none;
  background: transparent;
  color: #111827;
  font-size: 1rem;
  outline: none;
}

.search-input::placeholder {
  color: #6b7280;
}

.clear-search {
  padding: 4px 8px;
  border: none;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
}

.clear-search:hover {
  color: #111827;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  margin-bottom: 16px;
}

.btn-page {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  background: white;
  color: #00034b;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-page:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-page:hover:not(:disabled) {
  background: #00034b;
  color: white;
}

.page-info {
  font-size: 0.9rem;
  color: #6b7280;
  margin: 0 4px;
}

.page-size-select {
  margin-left: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 6px 10px;
  background: white;
  color: #111827;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #6b7280;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e5e7eb;
  border-top-color: #00034b;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 15px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px 20px;
  background: rgba(220, 53, 69, 0.1);
  border: 1px solid rgba(220, 53, 69, 0.3);
  border-radius: 12px;
  color: #dc3545;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #6b7280;
}

.empty-state i {
  font-size: 4rem;
  margin-bottom: 15px;
  opacity: 0.5;
}

.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 16px;
}

.group-card {
  display: flex;
  gap: 15px;
  padding: 16px;
  background: white;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #e5e7eb;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.group-card:hover {
  border-color: #00034b;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 3, 75, 0.15);
}

.card-avatar,
.item-avatar {
  flex-shrink: 0;
}

.card-avatar img,
.item-avatar img {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: linear-gradient(135deg, #00034b, #000266);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 1.5rem;
}

.card-content,
.item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.card-header,
.item-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
}

.group-name {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.message-time {
  font-size: 0.75rem;
  color: #00034b;
  white-space: nowrap;
}

.last-message {
  margin: 0;
  font-size: 0.875rem;
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.last-message.empty {
  font-style: italic;
  opacity: 0.7;
}

.sender-name {
  color: #00034b;
  font-weight: 500;
}

.card-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: auto;
}

.participants-count {
  font-size: 0.75rem;
  color: #6b7280;
}

.participants-info {
  font-size: 0.875rem;
  color: #6b7280;
}

.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  font-size: 0.7rem;
}

.badge-announce {
  background: rgba(255, 193, 7, 0.2);
  color: #d97706;
}

.badge-community {
  background: rgba(0, 3, 75, 0.1);
  color: #00034b;
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  background: white;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #e5e7eb;
}

.group-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 12px 16px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  border-bottom: 1px solid #e5e7eb;
}

.group-item:last-child {
  border-bottom: none;
}

.group-item:hover {
  background: #f3f4f6;
}

.item-avatar img {
  width: 50px;
  height: 50px;
}

.item-avatar .avatar-placeholder {
  width: 50px;
  height: 50px;
  font-size: 1.2rem;
}

.item-badges {
  display: flex;
  gap: 5px;
}

@media (max-width: 768px) {
  .groups-page {
    padding: 15px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    justify-content: space-between;
  }

  .groups-grid {
    grid-template-columns: 1fr;
  }

  .btn-primary span {
    display: none;
  }
}
</style>
