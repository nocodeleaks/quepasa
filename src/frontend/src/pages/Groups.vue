<template>
  <div class="groups-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <router-link :to="`/server/${token}`" class="back-btn">
          <i class="fa fa-arrow-left"></i>
        </router-link>
        <h1>Grupos</h1>
        <span class="group-count" v-if="groups.length > 0">({{ groups.length }})</span>
      </div>
      <div class="header-actions">
        <!-- View Mode Toggle -->
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

    <!-- Search -->
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

    <!-- Pagination Controls -->
    <div v-if="groups.length > 0" class="pagination-controls">
      <button @click="goToPage(1)" :disabled="currentPage === 1 || loading" class="btn-page">
        <i class="fa fa-angle-double-left"></i>
      </button>
      <button @click="prevPage" :disabled="currentPage === 1 || loading" class="btn-page">
        <i class="fa fa-angle-left"></i>
      </button>
      <span class="page-info">Página {{ currentPage }} de {{ totalPages }}</span>
      <button @click="nextPage" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-right"></i>
      </button>
      <button @click="goToPage(totalPages)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-double-right"></i>
      </button>
      <select v-model.number="pageSize" class="page-size-select" :disabled="loading">
        <option :value="5">5 por página</option>
        <option :value="10">10 por página</option>
        <option :value="15">15 por página</option>
        <option :value="25">25 por página</option>
        <option :value="50">50 por página</option>
        <option :value="100">100 por página</option>
        <option :value="200">200 por página</option>
      </select>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Carregando grupos...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="error-box">
      <i class="fa fa-exclamation-triangle"></i>
      {{ error }}
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredGroups.length === 0" class="empty-state">
      <i class="fa fa-users"></i>
      <p v-if="searchQuery">Nenhum grupo encontrado para "{{ searchQuery }}"</p>
      <p v-else>Nenhum grupo encontrado</p>
    </div>

    <!-- Card View (Default) -->
    <div v-else-if="viewMode === 'card'" class="groups-grid">
      <div 
        v-for="g in displayGroups" 
        :key="g.JID" 
        class="group-card"
        @click="goToGroup(g.JID)"
      >
        <div class="card-avatar">
          <img 
            v-if="groupPictures[g.JID]" 
            :src="groupPictures[g.JID]" 
            :alt="g.Name"
            @error="handleImageError(g.JID)"
          />
          <div v-else class="avatar-placeholder">
            <i class="fa fa-users"></i>
          </div>
        </div>
        <div class="card-content">
          <div class="card-header">
            <h3 class="group-name">{{ g.Name || 'Grupo sem nome' }}</h3>
            <span class="message-time" v-if="g.lastMessage">{{ formatTime(g.lastMessage.timestamp) }}</span>
          </div>
          <div class="card-body">
            <p class="last-message" v-if="g.lastMessage">
              <span class="sender-name">{{ g.lastMessage.senderName }}:</span>
              {{ truncateMessage(g.lastMessage.text) }}
            </p>
            <p class="last-message empty" v-else>
              <i class="fa fa-comment-o me-1"></i>Sem mensagens recentes
            </p>
          </div>
          <div class="card-footer">
            <span class="participants-count">
              <i class="fa fa-user me-1"></i>{{ g.Participants?.length || 0 }}
            </span>
            <span v-if="g.IsAnnounce" class="badge badge-announce" title="Apenas admins podem enviar">
              <i class="fa fa-bullhorn"></i>
            </span>
            <span v-if="g.IsParent" class="badge badge-community" title="Comunidade">
              <i class="fa fa-sitemap"></i>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div v-else class="groups-list">
      <div 
        v-for="g in displayGroups" 
        :key="g.JID" 
        class="group-item"
        @click="goToGroup(g.JID)"
      >
        <div class="item-avatar">
          <img 
            v-if="groupPictures[g.JID]" 
            :src="groupPictures[g.JID]" 
            :alt="g.Name"
            @error="handleImageError(g.JID)"
          />
          <div v-else class="avatar-placeholder">
            <i class="fa fa-users"></i>
          </div>
        </div>
        <div class="item-content">
          <div class="item-header">
            <h3 class="group-name">{{ g.Name || 'Grupo sem nome' }}</h3>
            <span class="message-time" v-if="g.lastMessage">{{ formatTime(g.lastMessage.timestamp) }}</span>
          </div>
          <div class="item-body">
            <p class="last-message" v-if="g.lastMessage">
              <span class="sender-name">{{ g.lastMessage.senderName }}:</span>
              {{ truncateMessage(g.lastMessage.text) }}
            </p>
            <p class="last-message empty" v-else>
              <span class="participants-info">
                <i class="fa fa-user me-1"></i>{{ g.Participants?.length || 0 }} participantes
              </span>
            </p>
          </div>
        </div>
        <div class="item-badges">
          <span v-if="g.IsAnnounce" class="badge badge-announce" title="Apenas admins podem enviar">
            <i class="fa fa-bullhorn"></i>
          </span>
          <span v-if="g.IsParent" class="badge badge-community" title="Comunidade">
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
      <span class="page-info">Página {{ currentPage }} de {{ totalPages }}</span>
      <button @click="nextPage" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-right"></i>
      </button>
      <button @click="goToPage(totalPages)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <i class="fa fa-angle-double-right"></i>
      </button>
      <select v-model.number="pageSize" class="page-size-select" :disabled="loading">
        <option :value="5">5 por página</option>
        <option :value="10">10 por página</option>
        <option :value="15">15 por página</option>
        <option :value="25">25 por página</option>
        <option :value="50">50 por página</option>
        <option :value="100">100 por página</option>
        <option :value="200">200 por página</option>
      </select>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onMounted, watch } from 'vue'
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

    // Filter groups by search query
    const filteredGroups = computed(() => {
      if (!searchQuery.value) return groups.value
      const query = searchQuery.value.toLowerCase()
      return groups.value.filter(g => 
        g.Name?.toLowerCase().includes(query) ||
        g.JID.toLowerCase().includes(query)
      )
    })

    const totalPages = computed(() => {
      return Math.ceil(filteredGroups.value.length / pageSize.value) || 1
    })

    const displayGroups = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value
      const end = start + pageSize.value
      return filteredGroups.value.slice(start, end)
    })

    async function load() {
      loading.value = true
      error.value = ''
      try {
        const res = await api.get(`/api/groups/getall?token=${encodeURIComponent(token)}`)
        const rawGroups = res.data?.groups || []
        
        // Sort by name
        rawGroups.sort((a: Group, b: Group) => {
          const nameA = a.Name || ''
          const nameB = b.Name || ''
          return nameA.localeCompare(nameB)
        })
        
        groups.value = rawGroups
        
        // Load pictures in background (first 20 for performance)
        loadGroupPictures(rawGroups.slice(0, 20))
        
        // Try to get last messages
        loadLastMessages()
      } catch (e: any) {
        error.value = e?.response?.data?.result || e?.message || 'Erro ao carregar grupos'
      } finally {
        loading.value = false
      }
    }

    watch([searchQuery, pageSize], () => {
      currentPage.value = 1
    })

    watch([filteredGroups, totalPages], () => {
      if (currentPage.value > totalPages.value) {
        currentPage.value = totalPages.value
      }
      if (currentPage.value < 1) {
        currentPage.value = 1
      }
    })

    function goToPage(page: number) {
      if (page >= 1 && page <= totalPages.value) {
        currentPage.value = page
      }
    }

    function nextPage() {
      if (currentPage.value < totalPages.value) {
        currentPage.value++
      }
    }

    function prevPage() {
      if (currentPage.value > 1) {
        currentPage.value--
      }
    }

    async function loadGroupPictures(groupList: Group[]) {
      for (const g of groupList) {
        try {
          const res = await api.get(`/api/picinfo/${encodeURIComponent(g.JID)}?token=${encodeURIComponent(token)}`)
          if (res.data?.url) {
            groupPictures.value[g.JID] = res.data.url
          }
        } catch {
          // Silently ignore - will show placeholder
        }
      }
    }

    async function loadLastMessages() {
      try {
        const res = await api.get(`/api/server/${token}/messages`)
        const messages = res.data?.messages || []
        
        // Group messages by chat ID and get the last one for each group
        const lastMessages: Record<string, LastMessage> = {}
        
        for (const msg of messages) {
          const chatId = msg.chat?.id
          if (!chatId || !chatId.endsWith('@g.us')) continue
          
          // Skip unhandled/system messages
          if (msg.type === 'unhandled' || msg.type === 'revoked' || msg.type === 'system') continue
          
          // Skip protocol messages (SenderKeyDistribution, etc)
          if (msg.debug?.reason === 'discard') continue
          
          // Must have some content: text, attachment, or be a reply
          if (!msg.text && !msg.attachment && !msg.inreply) continue
          
          if (!lastMessages[chatId] || new Date(msg.timestamp) > new Date(lastMessages[chatId].timestamp)) {
            // Determine message preview text
            let previewText = msg.text || ''
            
            // Handle attachments with emojis
            if (msg.attachment) {
              const mime = msg.attachment.mimetype || ''
              let mediaPrefix = ''
              if (mime.startsWith('image/')) mediaPrefix = '📷 '
              else if (mime.startsWith('video/')) mediaPrefix = '🎥 '
              else if (mime.startsWith('audio/') || msg.type === 'ptt') mediaPrefix = '🎵 '
              else if (mime.includes('pdf')) mediaPrefix = '📄 '
              else mediaPrefix = '📎 '
              
              if (previewText) {
                previewText = mediaPrefix + previewText
              } else {
                if (mime.startsWith('image/')) previewText = '📷 Imagem'
                else if (mime.startsWith('video/')) previewText = '🎥 Vídeo'
                else if (mime.startsWith('audio/') || msg.type === 'ptt') previewText = '🎵 Áudio'
                else if (mime.includes('pdf')) previewText = '📄 PDF'
                else previewText = '📎 Arquivo'
              }
            }
            
            // Handle stickers
            if (msg.type === 'sticker') {
              previewText = '🏷️ Sticker'
            }
            
            // Handle replies without text
            if (!previewText && msg.inreply) {
              previewText = '↩️ Resposta'
            }
            
            if (!previewText) continue // Skip if still no content
            
            lastMessages[chatId] = {
              timestamp: msg.timestamp,
              senderName: msg.participant?.title || msg.participant?.phone || 'Desconhecido',
              text: previewText
            }
          }
        }
        
        // Assign last messages to groups
        groups.value = groups.value.map(g => ({
          ...g,
          lastMessage: lastMessages[g.JID]
        }))
        
        // Re-sort by last message time (most recent first)
        groups.value.sort((a, b) => {
          if (a.lastMessage && b.lastMessage) {
            return new Date(b.lastMessage.timestamp).getTime() - new Date(a.lastMessage.timestamp).getTime()
          }
          if (a.lastMessage) return -1
          if (b.lastMessage) return 1
          return (a.Name || '').localeCompare(b.Name || '')
        })
      } catch {
        // Silently ignore - messages are optional
      }
    }

    function handleImageError(jid: string) {
      delete groupPictures.value[jid]
    }

    function formatTime(timestamp: string): string {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const oneDay = 24 * 60 * 60 * 1000
      
      if (diff < oneDay && date.getDate() === now.getDate()) {
        return date.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
      } else if (diff < 2 * oneDay) {
        return 'Ontem'
      } else if (diff < 7 * oneDay) {
        return date.toLocaleDateString('pt-BR', { weekday: 'short' })
      } else {
        return date.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })
      }
    }

    function truncateMessage(text: string, maxLength: number = 50): string {
      if (!text) return ''
      if (text.length <= maxLength) return text
      return text.substring(0, maxLength) + '...'
    }

    function goToGroup(jid: string) {
      router.push(`/server/${token}/groups/${encodeURIComponent(jid)}`)
    }

    async function createGroup() {
      const title = prompt('Nome do grupo (<=25 caracteres):')
      if (!title) return
      const participantsRaw = prompt('Participantes (telefone separados por vírgula):')
      if (!participantsRaw) return
      const participants = participantsRaw.split(',').map(s => s.trim()).filter(Boolean)
      try {
        await api.post('/api/groups/create', { title, participants, token })
        pushToast('Grupo criado', 'success')
        await load()
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao criar grupo', 'error')
      }
    }

    onMounted(() => { load() })

    return { 
      token, 
      groups, 
      groupPictures,
      loading, 
      error, 
      searchQuery,
      viewMode,
      filteredGroups,
      displayGroups,
      currentPage,
      totalPages,
      pageSize,
      createGroup,
      load,
      goToGroup,
      goToPage,
      nextPage,
      prevPage,
      handleImageError,
      formatTime,
      truncateMessage
    }
  }
})
</script>

<style scoped>
.groups-page {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

/* Header */
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
  color: #00034B;
  text-decoration: none;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #00034B;
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
  background: #00034B;
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
  color: #00034B;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: #00034B;
  color: white;
}

.btn-primary {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  background: #00034B;
  color: white;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: #000266;
  transform: translateY(-1px);
}

/* Search Bar */
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

/* Pagination */
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
  color: #00034B;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-page:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-page:hover:not(:disabled) {
  background: #00034B;
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

/* Loading */
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
  border-top-color: #00034B;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 15px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Error */
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

/* Empty State */
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

/* Card View */
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
  border-color: #00034B;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 3, 75, 0.15);
}

.card-avatar, .item-avatar {
  flex-shrink: 0;
}

.card-avatar img, .item-avatar img {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: linear-gradient(135deg, #00034B, #000266);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 1.5rem;
}

.card-content, .item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.card-header, .item-header {
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
  color: #00034B;
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
  color: #00034B;
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
  color: #00034B;
}

/* List View */
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

/* Responsive */
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
