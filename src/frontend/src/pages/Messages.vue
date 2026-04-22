<template>
  <div class="messages-page">
    <div class="page-header">
      <div class="header-left">
        <button @click="$router.back()" class="back-link hide-mobile">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
          </svg>
          Voltar
        </button>
        <div>
          <h1>Mensagens Recebidas</h1>
          <div v-if="serverNumber || totalMessages !== null" class="header-sub">
            <small class="text-muted">Servidor: {{ serverNumber || '—' }} — Total: {{ totalMessages !== null ? totalMessages : messages.length }} mensagens</small>
            <span v-if="wsConnected" class="ws-status ws-connected" title="WebSocket conectado - Atualizações em tempo real">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                <circle cx="12" cy="12" r="8"/>
              </svg>
              Live
            </span>
            <span v-else class="ws-status ws-disconnected" title="WebSocket desconectado">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                <circle cx="12" cy="12" r="8"/>
              </svg>
              Offline
            </span>
          </div>
        </div>
      </div>
      <div class="header-actions">
        <button @click="loadMessages" class="btn-refresh" :disabled="loading">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" :class="{ spinning: loading }">
            <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
          </svg>
          Atualizar
        </button>
      </div>
    </div>

    <div v-if="error" class="error-box">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <div class="filters-bar">
      <div class="search-box">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
        </svg>
        <input v-model="search" type="text" placeholder="Buscar por texto, nome, número ou LId..." class="search-input" />
      </div>
      <div class="filter-group">
        <label>Tipo:</label>
        <select v-model="filterType" class="filter-select">
          <option value="">Todos</option>
          <option value="text">Texto</option>
          <option value="image">Imagem</option>
          <option value="audio">Áudio</option>
          <option value="video">Vídeo</option>
          <option value="document">Documento</option>
        </select>
      </div>
      <div class="filter-group">
        <label>
          <input type="checkbox" v-model="filterGroupsOnly" class="filter-checkbox" />
          Apenas grupos
        </label>
      </div>
      <div class="messages-count">
        {{ filteredMessages.length }} mensagens
        <span v-if="totalMessages !== null" class="text-muted"> (Total: {{ totalMessages }})</span>
      </div>
    </div>

    <!-- Pagination Controls -->
    <div v-if="messages.length > 0 || (totalMessages !== null && totalMessages > 0)" class="pagination-controls">
      <button @click="goToPage(1)" :disabled="currentPage === 1 || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M18.41 16.59L13.82 12l4.59-4.59L17 6l-6 6 6 6zM6 6h2v12H6z"/>
        </svg>
      </button>
      <button @click="goToPage(currentPage - 1)" :disabled="currentPage === 1 || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z"/>
        </svg>
      </button>
      <span class="page-info">Página {{ currentPage }} de {{ totalPages }}</span>
      <button @click="goToPage(currentPage + 1)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/>
        </svg>
      </button>
      <button @click="goToPage(totalPages)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M5.59 7.41L10.18 12l-4.59 4.59L7 18l6-6-6-6zM16 6h2v12h-2z"/>
        </svg>
      </button>
      <select v-model.number="messagesPerPage" @change="changePageSize" class="page-size-select" :disabled="loading">
        <option :value="5">5 por página</option>
        <option :value="10">10 por página</option>
        <option :value="15">15 por página</option>
        <option :value="25">25 por página</option>
        <option :value="50">50 por página</option>
        <option :value="100">100 por página</option>
        <option :value="200">200 por página</option>
      </select>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="spinner-large"></div>
      <p>Carregando mensagens...</p>
    </div>

    <div v-else-if="filteredMessages.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="64" height="64" fill="currentColor">
        <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
      </svg>
      <h3>Nenhuma mensagem encontrada</h3>
      <p>As mensagens recebidas aparecerão aqui</p>
    </div>

    <div v-else class="messages-list">
      <div v-for="msg in filteredMessages" :key="msg.id" :id="'msg-'+msg.id" class="message-card" :class="{ 'from-me': msg.fromme, 'has-exceptions': msg.exceptions?.length > 0 }">
        <div class="message-header">
          <div class="message-sender">
            <div v-if="contactPicMap[msg.chat?.id]" class="sender-avatar sender-avatar-img">
              <img :src="contactPicMap[msg.chat?.id]" class="avatar-img" @error="handleAvatarError(msg.chat?.id)" />
            </div>
            <div v-else class="sender-avatar" :style="{ background: getAvatarColor(msg.chat?.id || msg.from) }">
              {{ getInitial(msg.chat?.title || msg.from) }}
            </div>
            <div class="sender-info">
              <span class="sender-name">{{ getSenderDisplayName(msg) }}</span>
              <span class="sender-phone">{{ msg.fromme ? (serverNumber || msg.chat?.id || msg.from) : (msg.chat?.id || msg.from) }}</span>
              <small v-if="msg.chat?.lid || msg.chat?.LId" class="sender-lid">LId: {{ msg.chat?.lid || msg.chat?.LId }}</small>
              <small v-if="msg.chat?.title" class="chat-title">Chat: {{ msg.chat.title }}</small>
            </div>
          </div>
          <div class="message-meta">
            <div class="message-time">
              {{ formatTime(msg.timestamp) }}
            </div>
            <div class="message-badges">
              <span v-if="msg.fromme" class="badge badge-sent">Enviada</span>
              <span v-else class="badge badge-received">Recebida</span>
              <span v-if="msg.status" class="badge badge-status">{{ msg.status }}</span>
              <span v-if="msg.edited" class="badge badge-warning">Editada</span>
              <span v-if="msg.fromhistory" class="badge badge-secondary">Histórico</span>
              <span v-if="msg.ads" class="badge badge-danger">Anúncio</span>
            </div>
            <button class="btn-small" @click="archiveChat(msg)" :disabled="archiving[msg.chat?.id]">{{ archiving[msg.chat?.id] ? 'Arquivando...' : 'Arquivar conversa' }}</button>
          </div>
        </div>

        <!-- Participant (for groups) -->
        <div v-if="msg.participant" class="message-participant">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
            <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
          </svg>
          <div>
            <div class="participant-field" style="display:flex;gap:8px;align-items:center">
              <strong>Número:</strong>
              <div style="display:flex;align-items:center;gap:8px">
                <img v-if="contactPicMap[msg.participant?.id || msg.participant?.phone]" :src="contactPicMap[msg.participant?.id || msg.participant?.phone]" alt="avatar" class="participant-avatar" @error="handleAvatarError(msg.participant?.id || msg.participant?.phone)" />
                <span>{{ msg.participant.phone }}</span>
              </div>
            </div>
            <div class="participant-field"><strong>Lid:</strong> {{ msg.participant.id }}</div>
            <div class="participant-field"><strong>Nome:</strong> {{ getParticipantDisplayName(msg) }}</div>
            <div v-if="msg.participant?.title" class="participant-field"><strong>Título:</strong> {{ msg.participant.title }}</div>
            <small v-if="msg.participant.lid || msg.participant.LId" class="participant-lid">LId: {{ msg.participant.lid || msg.participant.LId }}</small>
          </div>
        </div>

        <div class="message-body">
          <!-- Text message / Debug JSON handling -->
          <div v-if="msg.text" class="message-text">
            <div v-if="isLikelyJson(msg.text)" class="json-message">
              <div class="json-controls">
                <button class="btn-small" @click="toggleCollapse(msg.id)">{{ collapsed[msg.id] ? 'Ocultar' : 'Mostrar' }}</button>
                <button class="btn-small" @click="copyText(extractJsonFromText(msg.text))">Copiar JSON</button>
                <button v-if="msg.debug && msg.debug.event === 'ProtocolMessage' && msg.debug.reason && msg.debug.reason.includes('history sync')" class="btn-small" @click.prevent="downloadHistory(msg)" :disabled="fetchingDownload[msg.id]">
                  {{ fetchingDownload[msg.id] ? 'Baixando...' : 'Baixar mídia' }}
                </button>
              </div>
              <pre v-if="collapsed[msg.id]" class="json-block">{{ prettyJson(msg.text) }}</pre>
              <div v-else class="json-collapsed">JSON ({{ msg.id }})</div>
            </div>
            <div v-else>
              {{ msg.text }}
            </div>
          </div>

          <!-- URL Preview -->
          <div v-if="msg.url" class="message-url">
            <div class="url-preview">
              <strong>{{ msg.url.title }}</strong>
              <p v-if="msg.url.description">{{ msg.url.description }}</p>
              <a :href="msg.url.reference" target="_blank" class="url-link">{{ msg.url.reference }}</a>
            </div>
            <div v-if="urlThumbnailUrl(msg.url)" class="url-thumbnail">
              <img :src="urlThumbnailUrl(msg.url)" alt="URL thumbnail" />
            </div>
          </div>

          <!-- Ads Info -->
          <div v-if="msg.ads" class="message-ads">
            <div class="ads-header">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>
              </svg>
              <span>Anúncio: {{ msg.ads.title }} <small v-if="msg.ads.id">(#{{ msg.ads.id }})</small></span>
            </div>
            <div class="ads-meta">
              <div v-if="msg.ads.sourceId"><strong>Source ID:</strong> {{ msg.ads.sourceId }}</div>
              <div v-if="msg.ads.app"><strong>App:</strong> {{ msg.ads.app }}</div>
              <div v-if="msg.ads.type"><strong>Tipo:</strong> {{ msg.ads.type }}</div>
            </div>
            <div v-if="adThumbnailUrl(msg.ads)" class="ads-thumbnail">
              <img :src="adThumbnailUrl(msg.ads)" alt="Ad thumbnail" />
            </div>
            <a v-if="msg.ads.sourceurl" :href="msg.ads.sourceurl" target="_blank" class="ads-link">Ver anúncio</a>
          </div>

          <!-- Media attachment -->
          <div v-if="msg.attachment" class="message-attachment">
            <div v-if="isImageMessage(msg)" class="attachment-image">
              <img :src="getMediaUrl(msg)" :alt="getFilename(msg.attachment)" @error="handleImageError" />
              <div class="attachment-actions">
                <a :href="getMediaUrl(msg)" target="_blank" class="btn-download" :download="getFilename(msg.attachment)">Abrir</a>
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-download">Baixar</a>
              </div>
            </div>

            <div v-else-if="isAudioMessage(msg)" class="attachment-audio">
              <audio controls :src="getMediaUrl(msg)"></audio>
              <div class="attachment-actions">
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-download">Baixar áudio</a>
              </div>
            </div>

            <div v-else-if="isVideoMessage(msg)" class="attachment-video">
              <video controls :src="getMediaUrl(msg)"></video>
              <div class="attachment-actions">
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-download">Baixar vídeo</a>
              </div>
            </div>

            <div v-else-if="isPdf(msg.attachment)" class="attachment-file">
              <div class="file-info">
                <span class="file-name">{{ getFilename(msg.attachment) }}</span>
                <span class="file-size">{{ formatSize(getFileLength(msg.attachment)) }}</span>
              </div>
              <div class="attachment-actions">
                <a :href="getMediaUrl(msg)" target="_blank" class="btn-download">Abrir PDF</a>
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-download">Baixar</a>
              </div>
            </div>

            <div v-else class="attachment-file">
              <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
              </svg>
              <div class="file-info">
                <span class="file-name">{{ getFilename(msg.attachment) }}</span>
                <span class="file-size">{{ formatSize(getFileLength(msg.attachment)) }}</span>
              </div>
              <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-download">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                  <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
                </svg>
              </a>
            </div>

            <div class="attachment-status">
              <span v-if="isAttachmentValid(msg.attachment)" class="badge badge-success">Válido</span>
              <span v-else class="badge badge-danger">Inválido</span>
            </div>
            <div class="attachment-meta">
              <span class="mime-type">{{ getMimetype(msg.attachment) }}</span>
            </div>
          </div>

          <!-- In Reply -->
          <div v-if="msg.inreply" class="message-reply">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <path d="M10 9V5l-7 7 7 7v-4.1c5 0 8.5 1.6 11 5.1-1-5-4-10-11-11z"/>
            </svg>
            <span>Em resposta a: <a :href="'#msg-'+msg.inreply">{{ msg.inreply }}</a></span>
          </div>

          <!-- Exceptions (dispatch errors) -->
          <div v-if="msg.exceptions?.length > 0" class="message-exceptions">
            <div class="exceptions-header">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
              </svg>
              <span>Erros de Dispatch:</span>
            </div>
            <ul class="exceptions-list">
              <li v-for="(ex, i) in msg.exceptions" :key="i">{{ ex }}</li>
            </ul>
          </div>
        </div>

        <div class="message-footer">
          <span class="message-id">ID: {{ msg.id }}</span>
          <span v-if="msg.trackid" class="message-trackid">Track: {{ msg.trackid }}</span>
          <span class="message-type">{{ msg.type }}</span>
          <div class="message-actions">
            <template v-if="editing[msg.id]">
              <input v-model="editContent[msg.id]" class="edit-input" />
              <button class="btn-small me-1" @click="saveEdit(msg)">Salvar</button>
              <button class="btn-small" @click="cancelEdit(msg)">Cancelar</button>
            </template>
            <template v-else>
              <button v-if="msg.fromme" class="btn-small me-1" @click="startEdit(msg)">Editar</button>
              <button class="btn-small me-1" @click="revokeMessage(msg)">Cancelar Mensagem</button>
              <button class="btn-small" @click="sendPresence(msg)" :disabled="presenceLoading[msg.chat?.id]">{{ presenceLoading[msg.chat?.id] ? 'Enviando...' : 'Mostrar presença' }}</button>
            </template>
          </div>
        </div>
      </div>
    </div>

    
    <div v-if="totalPages > 1" class="pagination-controls">
      <button @click="goToPage(1)" :disabled="currentPage === 1 || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M18.41 16.59L13.82 12l4.59-4.59L17 6l-6 6 6 6zM6 6h2v12H6z"/>
        </svg>
      </button>
      <button @click="goToPage(currentPage - 1)" :disabled="currentPage === 1 || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z"/>
        </svg>
      </button>
      <span class="page-info">Página {{ currentPage }} de {{ totalPages }}</span>
      <button @click="goToPage(currentPage + 1)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/>
        </svg>
      </button>
      <button @click="goToPage(totalPages)" :disabled="currentPage >= totalPages || loading" class="btn-page">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M5.59 7.41L10.18 12l-4.59 4.59L7 18l6-6-6-6zM16 6h2v12h-2z"/>
        </svg>
      </button>
      <select v-model.number="messagesPerPage" @change="changePageSize" class="page-size-select" :disabled="loading">
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
import { defineComponent, ref, computed, onMounted, onUnmounted, reactive } from 'vue'
import signalRService from '../services/signalr'
import { useRoute } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

export default defineComponent({
  setup() {
    const route = useRoute()
    const token = route.params.token as string

    const messages = ref<any[]>([])
    const serverNumber = ref('')
    const totalMessages = ref<number | null>(null)
    const contacts = ref<any[]>([])
    const contactMap = reactive<Record<string, string>>({})
    const contactPicMap = reactive<Record<string, string>>({})
    const loading = ref(false)
    const error = ref('')
    const search = ref('')
    const filterType = ref('')
    const filterGroupsOnly = ref(false)
    const wsConnected = ref(false)
    const currentPage = ref(1)
    const messagesPerPage = ref(50)
    const totalPages = ref(0)

    // state to collapse/expand long or JSON debug messages
    const collapsed = reactive<Record<string, boolean>>({})
    const fetchingDownload = reactive<Record<string, boolean>>({})

    // editing state for messages
    const editing = reactive<Record<string, boolean>>({})
    const editContent = reactive<Record<string, string>>({})

    // presence and archive state
    const presenceLoading = reactive<Record<string, boolean>>({})
    const archiving = reactive<Record<string, boolean>>({})

    // contact search results (simple)
    const contactSearchResults = ref<any[]>([])

    function extractJsonFromText(text: string) {
      const first = text.indexOf('{')
      const last = text.lastIndexOf('}')
      if (first >= 0 && last > first) {
        return text.substring(first, last + 1)
      }
      return ''
    }

    function isLikelyJson(text: string) {
      try {
        const j = extractJsonFromText(text)
        if (!j) return false
        JSON.parse(j)
        return true
      } catch (e) {
        return false
      }
    }

    function prettyJson(text: string) {
      try {
        const j = extractJsonFromText(text)
        const obj = JSON.parse(j)
        return JSON.stringify(obj, null, 2)
      } catch (e) {
        return text
      }
    }

    function toggleCollapse(id: string) {
      collapsed[id] = !collapsed[id]
    }

    function copyText(t: string) {
      if (!t) return
      navigator.clipboard?.writeText(t)
    }

    const filteredMessages = computed(() => {
      let result = messages.value

      // Search filter - includes text, chat title, phone, LId, participant info
      if (search.value) {
        const s = search.value.toLowerCase()
        result = result.filter(m => {
          // Search in text content
          if (m.text && m.text.toLowerCase().includes(s)) return true
          // Search in chat info
          if (m.chat?.title && m.chat.title.toLowerCase().includes(s)) return true
          if (m.chat?.id && m.chat.id.toLowerCase().includes(s)) return true
          if (m.chat?.phone && m.chat.phone.toLowerCase().includes(s)) return true
          if (m.chat?.lid && m.chat.lid.toLowerCase().includes(s)) return true
          if (m.chat?.LId && m.chat.LId.toLowerCase().includes(s)) return true
          // Search in from field
          if (m.from && m.from.toLowerCase().includes(s)) return true
          // Search in participant info (for groups)
          if (m.participant?.id && m.participant.id.toLowerCase().includes(s)) return true
          if (m.participant?.phone && m.participant.phone.toLowerCase().includes(s)) return true
          if (m.participant?.lid && m.participant.lid.toLowerCase().includes(s)) return true
          if (m.participant?.LId && m.participant.LId.toLowerCase().includes(s)) return true
          if (m.participant?.title && m.participant.title.toLowerCase().includes(s)) return true
          return false
        })
      }

      // Type filter
      if (filterType.value) {
        result = result.filter(m => {
          if (filterType.value === 'text') return m.type === 'text' || (!m.attachment && m.text)
          if (filterType.value === 'image') return m.attachment && isImage(m.attachment)
          if (filterType.value === 'audio') return m.attachment && isAudio(m.attachment)
          if (filterType.value === 'video') return m.attachment && isVideo(m.attachment)
          if (filterType.value === 'document') return m.attachment && !isImage(m.attachment) && !isAudio(m.attachment) && !isVideo(m.attachment)
          return true
        })
      }

      // Group filter - check if chat id ends with @g.us (WhatsApp group format)
      if (filterGroupsOnly.value) {
        result = result.filter(m => {
          const chatId = m.chat?.id || m.from || ''
          return chatId.endsWith('@g.us')
        })
      }

      return result
    })

    async function loadMessages() {
      // reset fetching state map when reloading messages
      for (const k in fetchingDownload) delete fetchingDownload[k]
      loading.value = true
      error.value = ''
      try {
        const res = await api.get(`/api/server/${token}/messages`, {
          params: {
            page: currentPage.value,
            limit: messagesPerPage.value,
            timestamp: 0  // 0 = buscar TODAS as mensagens incluindo histórico
          }
        })
        messages.value = res.data?.messages || []
        serverNumber.value = res.data?.server.wid || ''
        totalMessages.value = res.data?.total ?? (res.data?.messages ? res.data.messages.length : null)
        totalPages.value = res.data?.total_pages || 0

        // Try to fetch contacts as a fallback for participant names
        try {
          const c = await api.get(`/api/server/${token}/contacts`)
          contacts.value = c.data?.contacts || []
          // build map (by id, phone and lid)
          for (const ct of contacts.value) {
            if (ct.id) contactMap[ct.id] = ct.title || ''
            if (ct.phone) contactMap[ct.phone] = ct.title || ''
            if (ct.lid) contactMap[ct.lid] = ct.title || ''
          }

          // Fetch profile pictures for unique chat and participant ids
          const uniqueIds = new Set<string>()
          for (const m of messages.value) {
            if (m.chat && m.chat.id) uniqueIds.add(m.chat.id)
            if (m.participant && (m.participant.id || m.participant.phone)) uniqueIds.add(m.participant.id || m.participant.phone)
          }

          const picPromises: Array<Promise<void>> = []
          for (const id of uniqueIds) {
            // do not fetch if already present
            if (contactPicMap[id]) continue
            const p = api.get(`/api/picinfo/${encodeURIComponent(id)}?token=${encodeURIComponent(token)}`)
              .then(res => {
                if (res && res.data && res.data.info && res.data.info.url) {
                  contactPicMap[id] = res.data.info.url
                }
              }).catch(() => {
                // ignore failures, keep initials
              })
            picPromises.push(p)
          }
          await Promise.all(picPromises)
        } catch (e) {
          // contacts are optional; ignore failures
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao carregar mensagens'
      } finally {
        loading.value = false
      }
    }

    function isImage(att: any) {
      return att?.mimetype?.startsWith('image/')
    }

    function isAudio(att: any) {
      return att?.mimetype?.startsWith('audio/')
    }

    function isVideo(att: any) {
      return att?.mimetype?.startsWith('video/')
    }

    function isImageMessage(m: any) {
      if (!m) return false
      if (m.attachment && getMimetype(m.attachment)) {
        return getMimetype(m.attachment).startsWith('image/')
      }
      return m.type === 'image'
    }

    function isAudioMessage(m: any) {
      if (!m) return false
      if (m.attachment && getMimetype(m.attachment)) {
        return getMimetype(m.attachment).startsWith('audio/')
      }
      return m.type === 'audio'
    }

    function isVideoMessage(m: any) {
      if (!m) return false
      if (m.attachment && getMimetype(m.attachment)) {
        return getMimetype(m.attachment).startsWith('video/')
      }
      return m.type === 'video'
    }

    function getMediaUrl(msg: any) {
      // Prefer explicit attachment URL when available (set by server after history download)
      if (msg?.attachment && (msg.attachment.url || msg.attachment.Url)) {
        return msg.attachment.url || msg.attachment.Url
      }
      // Use SPA download endpoint with token in path
      return `/api/server/${encodeURIComponent(token.trim())}/download/${encodeURIComponent(msg.id)}`
    }


    function handleImageError(e: Event) {
      (e.target as HTMLImageElement).style.display = 'none'
    }

    function handleAvatarError(id: string) {
      if (!id) return
      // remove cached url so initials fallback appears
      delete contactPicMap[id]
    }

    function formatTime(ts: any) {
      if (!ts) return ''
      const d = new Date(ts)
      return d.toLocaleString('pt-BR', { 
        day: '2-digit', 
        month: '2-digit', 
        hour: '2-digit', 
        minute: '2-digit' 
      })
    }

    function formatSize(bytes: number) {
      if (!bytes) return ''
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    }

    function getAvatarColor(id: string) {
      const colors = ['#7C3AED', '#5B21B6', '#8B5CF6', '#34b7f1', '#00a884']
      const hash = (id || '').split('').reduce((a, c) => a + c.charCodeAt(0), 0)
      return colors[hash % colors.length]
    }

    function getInitial(name: string) {
      return (name || '?').charAt(0).toUpperCase()
    }

    function adThumbnailUrl(ad: any) {
      if (!ad) return ''
      return ad.thumbnail?.url || ad.thumbnail || ad.thumbnailUrl || ad.ThumbnailUrl || ''
    }

    function urlThumbnailUrl(u: any) {
      if (!u) return ''
      return u.thumbnail?.url || u.thumbnail || u.thumbnailUrl || u.ThumbnailUrl || ''
    }

    function isAttachmentValid(att: any) {
      // support different field namings
      if (!att) return false
      if (typeof att.isvalidsize !== 'undefined') return !!att.isvalidsize
      if (typeof att.IsValidSize !== 'undefined') return !!att.IsValidSize
      if (typeof att.valid !== 'undefined') return !!att.valid
      if (typeof att.status === 'string') return att.status.toLowerCase() === 'valid' || att.status.toLowerCase() === 'ok'
      return true // unknown -> assume valid
    }

    function getFilename(att: any) {
      return att?.filename || att?.FileName || att?.fileName || att?.name || 'arquivo'
    }

    function getFileLength(att: any) {
      return att?.filelength || att?.FileLength || att?.FileLength || 0
    }

    function getMimetype(att: any) {
      return att?.mimetype || att?.Mimetype || att?.MimeType || ''
    }

    function isPdf(att: any) {
      const mt = getMimetype(att)
      return mt && mt.toLowerCase().includes('pdf')
    }

    function getChatDisplayName(m: any) {
      // try chat title, contact map by id, phone, lid
      const chat = m?.chat || {}
      if (chat.title && chat.title.length > 0) return chat.title
      if (chat.id && contactMap[chat.id]) return contactMap[chat.id]
      if (chat.phone && contactMap[chat.phone]) return contactMap[chat.phone]
      if (chat.lid && contactMap[chat.lid]) return contactMap[chat.lid]
      if (m?.from && contactMap[m.from]) return contactMap[m.from]
      return chat.title || chat.phone || chat.id || 'Desconhecido'
    }

    async function downloadHistory(m: any) {
      if (!m || !m.id) return
      fetchingDownload[m.id] = true
      try {
        await api.post(`/api/server/${token}/messages/${m.id}/history/download`)
        // on success reload messages to reflect new attachment
        await loadMessages()
        pushToast('Mídia do histórico baixada', 'success')
      } catch (e: any) {
        console.error('history download error', e)
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao baixar histórico', 'error')
      } finally {
        fetchingDownload[m.id] = false
      }
    }

    // Pagination functions
    function goToPage(page: number) {
      if (page < 1 || page > totalPages.value) return
      currentPage.value = page
      loadMessages() // Reload from backend with new page
    }

    function changePageSize() {
      currentPage.value = 1
      loadMessages() // Reload from backend with new page size
    }

    // Editing messages (server-sent 'fromme')
    function startEdit(m: any) {
      if (!m || !m.id) return
      editing[m.id] = true
      editContent[m.id] = m.text || ''
    }

    function cancelEdit(m: any) {
      if (!m || !m.id) return
      editing[m.id] = false
      editContent[m.id] = ''
    }

    async function saveEdit(m: any) {
      if (!m || !m.id) return
      const newText = editContent[m.id] || ''
      if (newText.trim().length === 0) {
        pushToast('Conteúdo vazio não permitido', 'error')
        return
      }
      try {
        await api.put(`/api/server/${encodeURIComponent(token.trim())}/message/${encodeURIComponent(m.id)}/edit`, { content: newText })
        editing[m.id] = false
        await loadMessages()
        pushToast('Mensagem editada', 'success')
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao editar mensagem', 'error')
      }
    }

    // Revoke message
    async function revokeMessage(m: any) {
      if (!m || !m.id) return
      if (!confirm('Deseja realmente revogar esta mensagem?')) return
      try {
        await api.delete(`/api/server/${encodeURIComponent(token.trim())}/message/${encodeURIComponent(m.id)}`)
        await loadMessages()
        pushToast('Mensagem revogada', 'success')
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao revogar mensagem', 'error')
      }
    }

    // Archive chat
    async function archiveChat(m: any) {
      if (!m || !m.chat || !m.chat.id) return
      if (!confirm('Deseja arquivar esta conversa?')) return
      archiving[m.chat.id] = true
      try {
        await api.post(`/api/server/${encodeURIComponent(token.trim())}/chat/archive`, { chatid: m.chat.id, archive: true })
        pushToast('Conversa arquivada', 'success')
        await loadMessages()
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao arquivar', 'error')
      } finally {
        archiving[m.chat.id] = false
      }
    }

    // Send presence (typing)
    async function sendPresence(m: any) {
      if (!m || !m.chat || !m.chat.id) return
      presenceLoading[m.chat.id] = true
      try {
        await api.post(`/api/server/${encodeURIComponent(token.trim())}/chat/presence`, { chatid: m.chat.id, type: 'composing' })
        pushToast('Indicador de presença enviado', 'success')
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao enviar presença', 'error')
      } finally {
        presenceLoading[m.chat.id] = false
      }
    }




    function getParticipantDisplayName(m: any) {
      const p = m?.participant || {}
      // try direct participant title
      if (p.title && p.title.length > 0) return p.title
      // try by phone, id, lid against contact map
      if (p.phone && contactMap[p.phone]) return contactMap[p.phone]
      if (p.id && contactMap[p.id]) return contactMap[p.id]
      if (p.lid && contactMap[p.lid]) return contactMap[p.lid]
      // fallback to chat title or number
      if (m?.chat && m.chat.title) return m.chat.title
      return p.title || p.phone || p.id || 'Desconhecido'
    }

    function getSenderDisplayName(m: any) {
      // If message is from this bot, show server number or label
      if (m?.fromme) {
        if (serverNumber.value && serverNumber.value.length > 0) {
          // try to get contact title for our number
          const sn = serverNumber.value
          if (contactMap[sn]) return contactMap[sn] + ' (Você)'
          return 'Você (' + sn + ')'
        }
        return 'Você'
      }

      // otherwise use chat display (group title or contact name)
      return getChatDisplayName(m)
    }

    // SignalR unsubscribe function
    let unsubscribeSignalR: (() => void) | null = null
    let wsStatusInterval: ReturnType<typeof setInterval> | null = null

    onMounted(() => {
      loadMessages()

      // Connect to SignalR for real-time message updates
      signalRService.connect(token).then(() => {
        console.log('Messages: SignalR connected')
        wsConnected.value = signalRService.isConnected()
      }).catch((err) => {
        console.error('Messages: SignalR connection error', err)
        wsConnected.value = false
      })

      // Periodically update connection status
      wsStatusInterval = setInterval(() => {
        wsConnected.value = signalRService.isConnected()
      }, 2000)

      // Listen for new messages
      unsubscribeSignalR = signalRService.onMessage((payload: any) => {
        console.log('Messages: Received new message via SignalR', payload)
        
        // Add the new message to the beginning of the list
        if (payload && payload.id) {
          // Check if message already exists
          const exists = messages.value.some(m => m.id === payload.id)
          if (!exists) {
            messages.value.unshift(payload)
            
            // Update total count
            if (totalMessages.value !== null) {
              totalMessages.value++
            }
            
            // Fetch profile picture for the new message if needed
            const chatId = payload.chat?.id
            if (chatId && !contactPicMap[chatId]) {
              api.get(`/api/picinfo/${encodeURIComponent(chatId)}?token=${encodeURIComponent(token)}`)
                .then(res => {
                  if (res?.data?.info?.url) {
                    contactPicMap[chatId] = res.data.info.url
                  }
                }).catch(() => {
                  // ignore errors
                })
            }

            // Show toast notification for new message
            const sender = payload.chat?.title || payload.from || 'Novo'
            pushToast(`Nova mensagem de ${sender}`, 'info')
          }
        }
      })
    })

    onUnmounted(() => {
      // Disconnect from SignalR when leaving the page
      if (unsubscribeSignalR) {
        unsubscribeSignalR()
        unsubscribeSignalR = null
      }
      if (wsStatusInterval) {
        clearInterval(wsStatusInterval)
        wsStatusInterval = null
      }
      signalRService.disconnect()
    })

    return {
      token, messages, serverNumber, totalMessages, loading, error, search, filterType, filterGroupsOnly,
      wsConnected, filteredMessages, currentPage, messagesPerPage, totalPages,
      loadMessages, isImage, isAudio, isVideo,
      getMediaUrl, handleImageError, handleAvatarError, formatTime, formatSize,
      getAvatarColor, getInitial, adThumbnailUrl, urlThumbnailUrl, isAttachmentValid, contactPicMap,
      // attachment helpers
      getFilename, getFileLength, getMimetype, isPdf,
      // display helpers
      getChatDisplayName, getParticipantDisplayName, getSenderDisplayName,
      // download helpers
      fetchingDownload, downloadHistory,
      // editing helpers
      editing, editContent, startEdit, cancelEdit, saveEdit,
      // revoke / archive / presence
      revokeMessage, archiveChat, sendPresence, archiving, presenceLoading,
      // message type helpers
      isImageMessage, isAudioMessage, isVideoMessage,
      // pagination helpers
      goToPage, changePageSize,
      // contact search

      // json helpers
      collapsed, isLikelyJson, extractJsonFromText, prettyJson, toggleCollapse, copyText
    }
  }
})
</script>

<style scoped>
.messages-page {
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-sub {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.ws-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  margin-left: 8px;
}

.ws-connected {
  background: #dcfce7;
  color: #166534;
}

.ws-connected svg {
  color: #22c55e;
  animation: pulse 2s infinite;
}

.ws-disconnected {
  background: #fee2e2;
  color: #991b1b;
}

.ws-disconnected svg {
  color: #ef4444;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.back-link {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  text-decoration: none;
  font-size: 14px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.back-link:hover {
  color: #374151;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: white;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  color: #374151;
  font-weight: 600;
  cursor: pointer;
}

.btn-refresh:hover:not(:disabled) {
  border-color: #7C3AED;
  color: #7C3AED;
}

.btn-refresh svg.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
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

.filters-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: white;
  border-radius: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.search-box {
  flex: 1;
  min-width: 200px;
  position: relative;
}

.search-box svg {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
}

.search-input {
  width: 100%;
  padding: 10px 12px 10px 40px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
}

.search-input:focus {
  outline: none;
  border-color: #7C3AED;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-group label {
  font-size: 14px;
  color: #6b7280;
}

.filter-select {
  padding: 10px 12px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
}

.filter-checkbox {
  margin-right: 6px;
  accent-color: #7C3AED;
}

.filter-group label {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.messages-count {
  font-size: 14px;
  color: #6b7280;
  padding: 10px 16px;
  background: #f3f4f6;
  border-radius: 8px;
}

.loading-container {
  text-align: center;
  padding: 60px 0;
}

.spinner-large {
  width: 50px;
  height: 50px;
  border: 4px solid #e5e7eb;
  border-top-color: #7C3AED;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

.loading-container p {
  color: #6b7280;
}

.empty-state {
  text-align: center;
  padding: 60px 0;
  color: #9ca3af;
}

.empty-state h3 {
  font-size: 20px;
  color: #6b7280;
  margin: 16px 0 8px;
}

.empty-state p {
  margin: 0;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.message-card.from-me {
  border-left: 4px solid #7C3AED;
}

.message-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.message-sender {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sender-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 16px;
}

.sender-info {
  display: flex;
  flex-direction: column;
}

.sender-name {
  font-weight: 600;
  color: #111827;
}

.sender-phone {
  font-size: 12px;
  color: #6b7280;
}

.message-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.message-time {
  font-size: 12px;
  color: #9ca3af;
}

.message-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  justify-content: flex-end;
}

.badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-sent {
  background: #f5efff;
  color: #7C3AED;
}

.badge-received {
  background: #ecfdf5;
  color: #059669;
}

.badge-status {
  background: #eff6ff;
  color: #2563eb;
}

.badge-warning {
  background: #fefce8;
  color: #ca8a04;
}

.badge-secondary {
  background: #f3f4f6;
  color: #6b7280;
}

.badge-danger {
  background: #fef2f2;
  color: #dc2626;
}

.message-participant {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 8px;
  padding: 6px 10px;
  background: #f9fafb;
  border-radius: 8px;
  width: fit-content;
}

.message-participant svg {
  flex-shrink: 0;
}
.message-participant .participant-field { font-size: 13px; color: #374151 }
.message-participant .participant-lid { display:block; font-size:12px; color:#6b7280; margin-top:4px }

.message-body {
  margin-bottom: 12px;
}

.message-text {
  color: #374151;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.message-attachment {
  margin-top: 12px;
}

.attachment-meta {
  margin-top: 6px;
  font-size: 11px;
  color: #9ca3af;
}

.mime-type {
  font-family: monospace;
}

.message-url {
  margin-top: 12px;
}

.url-preview {
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
  border-left: 3px solid #7C3AED;
}

.url-preview strong {
  display: block;
  color: #111827;
  margin-bottom: 4px;
}

.url-preview p {
  font-size: 13px;
  color: #6b7280;
  margin: 0 0 8px;
}

.url-link {
  font-size: 12px;
  color: #7C3AED;
  text-decoration: none;
  word-break: break-all;
}

.url-link:hover {
  text-decoration: underline;
}

.message-ads {
  margin-top: 12px;
  padding: 10px;
  background: #fef2f2;
  border-radius: 8px;
  border-left: 3px solid #dc2626;
}

.ads-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #dc2626;
  font-weight: 600;
}

.ads-link {
  display: inline-block;
  margin-top: 6px;
  font-size: 12px;
  color: #dc2626;
  text-decoration: none;
}

.ads-link:hover {
  text-decoration: underline;
}

.message-reply {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #6b7280;
  margin-top: 8px;
  padding: 6px 10px;
  background: #f3f4f6;
  border-radius: 8px;
}

.message-reply svg {
  flex-shrink: 0;
}

.message-exceptions {
  margin-top: 12px;
  padding: 10px;
  background: #fef2f2;
  border-radius: 8px;
}

.exceptions-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #dc2626;
  font-weight: 600;
  margin-bottom: 8px;
}

.exceptions-list {
  margin: 0;
  padding-left: 20px;
  font-size: 12px;
  color: #991b1b;
}

.exceptions-list li {
  margin-bottom: 4px;
}

.message-card.has-exceptions {
  border-left: 4px solid #dc2626;
}

.attachment-image img {
  max-width: 300px;
  max-height: 300px;
  border-radius: 12px;
  cursor: pointer;
}

.attachment-audio audio {
  width: 100%;
  max-width: 400px;
}

.attachment-video video {
  max-width: 100%;
  max-height: 300px;
  border-radius: 12px;
}

.attachment-file {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f3f4f6;
  border-radius: 12px;
}

.attachment-file svg {
  color: #6b7280;
}

.file-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.file-name {
  font-weight: 600;
  color: #111827;
}

.file-size {
  font-size: 12px;
  color: #6b7280;
}

.btn-download {
  padding: 8px;
  background: #7C3AED;
  border-radius: 8px;
  color: white;
}

.message-footer {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  flex-wrap: wrap;
  padding-top: 12px;
  border-top: 1px solid #f3f4f6;
}

.message-id {
  font-size: 11px;
  color: #9ca3af;
  font-family: monospace;
}

.message-trackid {
  font-size: 11px;
  color: #9ca3af;
  font-family: monospace;
}

.message-type {
  font-size: 11px;
  padding: 2px 6px;
  background: #f3f4f6;
  color: #6b7280;
  border-radius: 4px;
  font-family: monospace;
  margin-left: auto;
}
.sender-lid{ display:block; font-size:12px; color:#6b7280 }
.sender-avatar-img{ padding:0; overflow:hidden; }
.sender-avatar-img .avatar-img{ width:40px; height:40px; border-radius:50%; object-fit:cover; display:block }

/* Pagination Controls */
.pagination-controls {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 16px;
  background: white;
  border-radius: 12px;
  margin: 16px 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.btn-page {
  padding: 8px 12px;
  background: #7C3AED;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-page:hover:not(:disabled) {
  background: #6D28D9;
  transform: translateY(-1px);
}

.btn-page:disabled {
  background: #E5E7EB;
  color: #9CA3AF;
  cursor: not-allowed;
  opacity: 0.6;
}

.page-info {
  padding: 8px 16px;
  background: #F3F4F6;
  border-radius: 8px;
  color: #374151;
  font-weight: 500;
  min-width: 150px;
  text-align: center;
}

.page-size-select {
  padding: 8px 12px;
  border: 1px solid #D1D5DB;
  border-radius: 8px;
  background: white;
  color: #374151;
  font-size: 14px;
  cursor: pointer;
  transition: border-color 0.2s;
}

.page-size-select:hover:not(:disabled) {
  border-color: #7C3AED;
}

.page-size-select:disabled {
  background: #F3F4F6;
  cursor: not-allowed;
  opacity: 0.6;
}
.participant-avatar{ width:28px; height:28px; border-radius:50%; object-fit:cover }
.ads-thumbnail img, .url-thumbnail img{ max-width:200px; max-height:120px; border-radius:8px; display:block; margin-top:8px }
.attachment-status{ margin-top:8px }
.header-sub{ margin-top:6px }
.chat-title{ display:block; font-size:12px; color:#6b7280; margin-top:4px }
.json-message .json-controls{ display:flex; gap:8px; justify-content:flex-end; margin-bottom:8px }
.json-block{ background:#0b1221; color:#e6eef8; padding:12px; border-radius:8px; overflow:auto; max-height:240px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, 'Roboto Mono', 'Noto Sans Mono', monospace; font-size:12px; }
.json-collapsed{ color:#6b7280; font-style:italic }
.btn-small{ background:#eef2ff; border:1px solid #c7d2fe; padding:6px 10px; border-radius:6px; cursor:pointer; font-size:12px }
.btn-small:hover{ background:#e0e7ff }
</style>
