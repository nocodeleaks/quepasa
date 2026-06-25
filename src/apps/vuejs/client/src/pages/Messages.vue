<template>
  <div class="messages-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
        </svg>
        {{ t('messages_back') }}
      </button>
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
          </svg>
          {{ t('messages_title') }}
        </h1>
        <div class="header-sub">
          <small v-if="serverNumber || totalMessages !== null" class="text-muted">{{ t('messages_server_info', [serverNumber || '\u2014', totalMessages !== null ? totalMessages : messages.length]) }}</small>
          <span v-if="wsConnected" class="ws-status ws-connected" :title="t('messages_ws_connected')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <circle cx="12" cy="12" r="8"/>
            </svg>
            {{ t('messages_live') }}
          </span>
          <span v-else class="ws-status ws-disconnected" :title="t('messages_ws_disconnected')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <circle cx="12" cy="12" r="8"/>
            </svg>
            {{ t('messages_offline') }}
            <button
              @click="syncPaused = !syncPaused"
              class="btn-sync-toggle"
              :title="syncPaused ? t('messages_sync_resume') : t('messages_sync_pause')"
            >
              <svg v-if="syncPaused" viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
                <path d="M8 5v14l11-7z"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
              </svg>
            </button>
          </span>
        </div>
      </div>
      <div class="header-actions">
        <button @click="loadMessages" class="btn-refresh" :disabled="loading">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" :class="{ spinning: loading }">
            <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
          </svg>
          {{ t('messages_refresh') }}
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
        <input v-model="search" type="text" :placeholder="t('messages_search_placeholder')" class="search-input" />
      </div>
      <div class="filter-group">
        <label>{{ t('messages_filter_type_label') }}</label>
        <select v-model="filterType" class="filter-select">
          <option value="">{{ t('messages_filter_all') }}</option>
          <option value="text">{{ t('messages_filter_text') }}</option>
          <option value="image">{{ t('messages_filter_image') }}</option>
          <option value="audio">{{ t('messages_filter_audio') }}</option>
          <option value="video">{{ t('messages_filter_video') }}</option>
          <option value="document">{{ t('messages_filter_document') }}</option>
        </select>
      </div>
      <div class="filter-group">
        <label>
          <input type="checkbox" v-model="filterGroupsOnly" class="filter-checkbox" />
          {{ t('messages_groups_only') }}
        </label>
      </div>
      <div class="messages-count">
        {{ t('messages_count', [filteredMessages.length]) }}
        <span v-if="totalMessages !== null" class="text-muted"> {{ t('messages_total', [totalMessages]) }}</span>
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
      <span class="page-info">{{ t('messages_page_of', [currentPage, totalPages]) }}</span>
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
        <option :value="5">{{ t('messages_per_page_5') }}</option>
        <option :value="10">{{ t('messages_per_page_10') }}</option>
        <option :value="15">{{ t('messages_per_page_15') }}</option>
        <option :value="25">{{ t('messages_per_page_25') }}</option>
        <option :value="50">{{ t('messages_per_page_50') }}</option>
        <option :value="100">{{ t('messages_per_page_100') }}</option>
        <option :value="200">{{ t('messages_per_page_200') }}</option>
      </select>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="spinner-large"></div>
      <p>{{ t('messages_loading') }}</p>
    </div>

    <div v-else-if="filteredMessages.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="64" height="64" fill="currentColor">
        <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
      </svg>
      <h3>{{ t('messages_empty_title') }}</h3>
      <p>{{ t('messages_empty_desc') }}</p>
    </div>

    <div v-else class="messages-list">
      <div v-for="msg in filteredMessages" :key="msg.id" :id="'msg-'+msg.id"
        class="mc" :class="{ 'mc--me': msg.fromme, 'mc--err': msg.exceptions?.length > 0 }">

        <!-- HEADER -->
        <div class="mc-header">
          <div v-if="contactPicMap[msg.chat?.id]" class="mc-avatar mc-avatar--img">
            <img :src="contactPicMap[msg.chat?.id]" @error="handleAvatarError(msg.chat?.id)" />
          </div>
          <div v-else class="mc-avatar" :style="{ background: getAvatarColor(msg.chat?.id || msg.from) }">
            {{ getInitial(msg.chat?.title || msg.from) }}
          </div>

          <div class="mc-identity">
            <div class="mc-identity-main">
              <span class="mc-name">{{ getSenderDisplayName(msg) }}</span>
              <span class="mc-phone">{{ msg.fromme ? (serverNumber || msg.chat?.id || msg.from) : (msg.chat?.id || msg.from) }}</span>
              <span v-if="msg.chat?.title" class="mc-chat">{{ msg.chat.title }}</span>
            </div>
            <!-- participante ao lado direito -->
            <div v-if="msg.participant" class="mc-participant">
              <img v-if="contactPicMap[msg.participant?.id || msg.participant?.phone]"
                :src="contactPicMap[msg.participant?.id || msg.participant?.phone]"
                class="mc-participant-avatar"
                @error="handleAvatarError(msg.participant?.id || msg.participant?.phone)" />
              <div v-else class="mc-participant-avatar mc-participant-avatar--initial"
                :style="{ background: getAvatarColor(msg.participant?.id || msg.participant?.phone) }">
                {{ getInitial(msg.participant?.title || msg.participant?.phone) }}
              </div>
              <div class="mc-participant-info">
                <span class="mc-participant-name">{{ getParticipantDisplayName(msg) }}</span>
                <span class="mc-participant-phone">{{ msg.participant.phone }}</span>
                <span v-if="msg.participant.id" class="mc-participant-lid">{{ msg.participant.id }}</span>
              </div>
            </div>
          </div>

          <div class="mc-meta">
            <span class="mc-time">{{ formatTime(msg.timestamp) }}</span>
            <div class="mc-badges">
              <span v-if="msg.fromme" class="badge badge-sent">{{ t('messages_badge_sent') }}</span>
              <span v-else class="badge badge-received">{{ t('messages_badge_received') }}</span>
              <span v-if="msg.status === 'read'" class="badge badge-read">✓✓ {{ t('messages_badge_read') }}</span>
              <span v-else-if="msg.status" class="badge badge-status">{{ msg.status }}</span>
              <span v-if="msg.fromhistory" class="badge badge-history">{{ t('messages_badge_history') }}</span>
              <span v-if="msg.isforwarded || msg.forwarded" class="badge badge-forward">{{ t('messages_badge_forwarded') }}</span>
              <span v-if="msg.isbroadcast || msg.broadcast" class="badge badge-broadcast">{{ t('messages_badge_broadcast') }}</span>
              <span v-if="msg.edited" class="badge badge-warning">{{ t('messages_badge_edited') }}</span>
              <span v-if="msg.ads" class="badge badge-danger">{{ t('messages_badge_ad') }}</span>
            </div>
            <button class="mc-btn-archive btn-small" @click="archiveChat(msg)" :disabled="archiving[msg.chat?.id]">
              {{ archiving[msg.chat?.id] ? t('messages_archiving_btn') : t('messages_archive_btn') }}
            </button>
          </div>
        </div>

        <!-- BODY -->
        <div class="mc-body">
          <!-- resposta a -->
          <div v-if="msg.inreply" class="mc-inreply">
            ↩ {{ t('messages_in_reply') }} <a :href="'#msg-'+msg.inreply">{{ msg.inreply }}</a>
          </div>

          <!-- texto / json -->
          <div v-if="msg.text" class="mc-text">
            <div v-if="isLikelyJson(msg.text)" class="json-message">
              <div class="json-controls">
                <button class="btn-small" @click="toggleCollapse(msg.id)">{{ collapsed[msg.id] ? t('messages_json_hide') : t('messages_json_show') }}</button>
                <button class="btn-small" @click="copyText(extractJsonFromText(msg.text))">{{ t('messages_json_copy') }}</button>
                <button v-if="msg.debug?.event === 'ProtocolMessage' && msg.debug?.reason?.includes('history sync')"
                  class="btn-small" @click.prevent="downloadHistory(msg)" :disabled="fetchingDownload[msg.id]">
                  {{ fetchingDownload[msg.id] ? t('messages_downloading') : t('messages_download_media') }}
                </button>
              </div>
              <pre v-if="collapsed[msg.id]" class="json-block">{{ prettyJson(msg.text) }}</pre>
              <div v-else class="json-collapsed">{{ t('messages_json_summary', [msg.id]) }}</div>
            </div>
            <span v-else>{{ msg.text }}</span>
          </div>

          <!-- anúncio -->
          <div v-if="msg.ads" class="mc-ads">
            <img v-if="adThumbnailUrl(msg.ads)" :src="adThumbnailUrl(msg.ads)" class="mc-ads-thumb" />
            <div class="mc-ads-info">
              <strong>{{ msg.ads.title }}</strong>
              <span v-if="msg.ads.app || msg.ads.type" class="mc-ads-meta">
                {{ [msg.ads.app, msg.ads.type].filter(Boolean).join(' · ') }}
              </span>
              <a v-if="msg.ads.sourceurl" :href="msg.ads.sourceurl" target="_blank" class="mc-ads-link">{{ t('messages_ad_link') }}</a>
            </div>
          </div>

          <!-- url preview -->
          <div v-if="msg.url" class="mc-url">
            <img v-if="urlThumbnailUrl(msg.url)" :src="urlThumbnailUrl(msg.url)" class="mc-url-thumb" />
            <div class="mc-url-info">
              <strong>{{ msg.url.title }}</strong>
              <span v-if="msg.url.description" class="mc-url-desc">{{ msg.url.description }}</span>
              <a :href="msg.url.reference" target="_blank" class="mc-url-link">{{ msg.url.reference }}</a>
            </div>
          </div>

          <!-- mídia -->
          <div v-if="msg.attachment" class="mc-media">
            <div v-if="isImageMessage(msg)">
              <img :src="getMediaUrl(msg)" :alt="getFilename(msg.attachment)" @error="handleImageError" class="mc-img" />
              <div class="mc-media-actions">
                <a :href="getMediaUrl(msg)" target="_blank" class="btn-small">{{ t('messages_open') }}</a>
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-small">{{ t('messages_download') }}</a>
              </div>
            </div>
            <div v-else-if="isAudioMessage(msg)">
              <audio controls :src="getMediaUrl(msg)" class="mc-audio"></audio>
              <div class="mc-media-actions">
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-small">{{ t('messages_download_audio') }}</a>
              </div>
            </div>
            <div v-else-if="isVideoMessage(msg)">
              <video controls :src="getMediaUrl(msg)" class="mc-video"></video>
              <div class="mc-media-actions">
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-small">{{ t('messages_download_video') }}</a>
              </div>
            </div>
            <div v-else class="mc-file">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor" class="mc-file-icon"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
              <div class="mc-file-info">
                <span class="mc-file-name">{{ getFilename(msg.attachment) }}</span>
                <span class="mc-file-meta">{{ getMimetype(msg.attachment) }}<template v-if="getFileLength(msg.attachment)"> · {{ formatSize(getFileLength(msg.attachment)) }}</template></span>
              </div>
              <div class="mc-media-actions">
                <a v-if="isPdf(msg.attachment)" :href="getMediaUrl(msg)" target="_blank" class="btn-small">{{ t('messages_open_pdf') }}</a>
                <a :href="getMediaUrl(msg)" :download="getFilename(msg.attachment)" class="btn-small">{{ t('messages_download') }}</a>
              </div>
            </div>
          </div>

          <!-- erros -->
          <div v-if="msg.exceptions?.length > 0" class="mc-exceptions">
            ⚠ {{ t('messages_dispatch_errors') }}
            <ul>
              <li v-for="(ex, i) in msg.exceptions" :key="i">{{ ex }}</li>
            </ul>
          </div>
        </div>

        <!-- FOOTER -->
        <div class="mc-footer">
          <span class="mc-id">{{ msg.id }}</span>
          <span v-if="msg.trackid" class="mc-trackid">{{ msg.trackid }}</span>
          <div class="mc-actions">
            <span class="mc-type">{{ msg.type }}</span>
            <template v-if="editing[msg.id]">
              <input v-model="editContent[msg.id]" class="edit-input" />
              <button class="btn-small" @click="saveEdit(msg)">{{ t('messages_save_edit') }}</button>
              <button class="btn-small" @click="cancelEdit(msg)">{{ t('messages_cancel_edit') }}</button>
            </template>
            <template v-else>
              <button v-if="msg.fromme" class="btn-small" @click="startEdit(msg)">{{ t('messages_edit') }}</button>
              <button v-if="msg.type !== 'system'" class="btn-small btn-revoke" @click="revokeMessage(msg)">{{ t('messages_revoke_btn') }}</button>
              <button class="btn-small" @click="sendPresence(msg)" :disabled="presenceLoading[msg.chat?.id]">
                {{ presenceLoading[msg.chat?.id] ? t('messages_sending') : t('messages_show_presence') }}
              </button>
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
      <span class="page-info">{{ t('messages_page_of', [currentPage, totalPages]) }}</span>
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
        <option :value="5">{{ t('messages_per_page_5') }}</option>
        <option :value="10">{{ t('messages_per_page_10') }}</option>
        <option :value="15">{{ t('messages_per_page_15') }}</option>
        <option :value="25">{{ t('messages_per_page_25') }}</option>
        <option :value="50">{{ t('messages_per_page_50') }}</option>
        <option :value="100">{{ t('messages_per_page_100') }}</option>
        <option :value="200">{{ t('messages_per_page_200') }}</option>
      </select>
    </div>

  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onMounted, onUnmounted, reactive } from 'vue'
import { useCableSubscription } from '@/composables/useCableSubscription'
import { useRoute } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

export default defineComponent({
  setup() {
    const { t, locale } = useLocale()
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
    const syncPaused = ref(false)
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

    function canLoadProfilePicture(id: any) {
      const value = String(id || '').trim()
      if (!value) return false

      const normalized = value.toLowerCase()
      if (normalized === 'system' || normalized === 'readreceipt') {
        return false
      }

      return true
    }

    async function loadMessages() {
      // reset fetching state map when reloading messages
      for (const k in fetchingDownload) delete fetchingDownload[k]
      loading.value = true
      error.value = ''
      try {
        const res = await api.get('/api/messages', {
          params: {
            token,
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
          const c = await api.get('/api/contacts', { params: { token } })
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
            if (canLoadProfilePicture(m.chat?.id)) uniqueIds.add(m.chat.id)
            if (canLoadProfilePicture(m.participant?.id || m.participant?.phone)) uniqueIds.add(m.participant.id || m.participant.phone)
          }

          const picPromises: Array<Promise<void>> = []
          for (const id of uniqueIds) {
            // do not fetch if already present
            if (contactPicMap[id]) continue
            const p = api.post('/api/media/pictures/info', { token, chatId: id })
              .then((res: any) => {
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
        error.value = err?.response?.data?.result || err.message || t('messages_error_load')
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
      return `/api/media/messages?token=${encodeURIComponent(token.trim())}&messageid=${encodeURIComponent(msg.id)}`
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
      return d.toLocaleString(locale.value, {
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
      return att?.filename || att?.FileName || att?.fileName || att?.name || t('messages_file_default')
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
      return chat.title || chat.phone || chat.id || t('unknown')
    }

    async function downloadHistory(m: any) {
      if (!m || !m.id) return
      fetchingDownload[m.id] = true
      try {
        await api.post('/api/media/download', { token, messageId: m.id })
        // on success reload messages to reflect new attachment
        await loadMessages()
        pushToast(t('messages_history_downloaded'), 'success')
      } catch (e: any) {
        console.error('history download error', e)
        pushToast(e?.response?.data?.result || e?.message || t('messages_error_download_history'), 'error')
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
        pushToast(t('messages_empty_content'), 'error')
        return
      }
      try {
        await api.patch('/api/messages', { token: token.trim(), messageId: m.id, content: newText })
        editing[m.id] = false
        await loadMessages()
        pushToast(t('messages_edited_msg'), 'success')
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || t('messages_error_edit_msg'), 'error')
      }
    }

    // Revoke message
    async function revokeMessage(m: any) {
      if (!m || !m.id) return
      if (m.type === 'system') return
      if (!confirm(t('messages_confirm_revoke'))) return
      try {
        await api.delete('/api/messages', { data: { token: token.trim(), messageid: m.id } })
        await loadMessages()
        pushToast(t('messages_revoked'), 'success')
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || t('messages_error_revoke'), 'error')
      }
    }

    // Archive chat
    async function archiveChat(m: any) {
      if (!m || !m.chat || !m.chat.id) return
      if (!confirm(t('messages_confirm_archive'))) return
      archiving[m.chat.id] = true
      try {
        await api.post('/api/chats/archive', { token: token.trim(), chatid: m.chat.id, archive: true })
        pushToast(t('messages_archived'), 'success')
        await loadMessages()
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || t('messages_error_archive'), 'error')
      } finally {
        archiving[m.chat.id] = false
      }
    }

    // Send presence (typing)
    async function sendPresence(m: any) {
      if (!m || !m.chat || !m.chat.id) return
      presenceLoading[m.chat.id] = true
      try {
        await api.post('/api/chats/presence', { token: token.trim(), chatid: m.chat.id, type: 'text', duration: 10000 })
        pushToast(t('messages_presence_sent'), 'success')
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || t('messages_error_presence'), 'error')
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
      return p.title || p.phone || p.id || t('unknown')
    }

    function getSenderDisplayName(m: any) {
      // If message is from this bot, show server number or label
      if (m?.fromme) {
        if (serverNumber.value && serverNumber.value.length > 0) {
          const sn = serverNumber.value
          if (contactMap[sn]) return `${contactMap[sn]} (${t('messages_you')})`
          return `${t('messages_you')} (${sn})`
        }
        return t('messages_you')
      }

      // otherwise use chat display (group title or contact name)
      return getChatDisplayName(m)
    }
    let wsStatusInterval: ReturnType<typeof setInterval> | null = null
    let syncInterval: ReturnType<typeof setInterval> | null = null

    async function syncMessagesFallback() {
      if (loading.value) return
      await loadMessages()
    }

    const cable = useCableSubscription(
      [
        {
          event: 'server.message',
          handler: (payload: any) => {
            if (payload?.token !== token || !payload?.message?.id) {
              return
            }

            const incoming = payload.message

            // Read receipts: id is the literal string "readreceipt" and
            // the actual referenced message id lives in the text field.
            // Update the existing message status instead of adding a new card.
            if (incoming.id === 'readreceipt') {
              const targetId = incoming.text  // original message id
              if (targetId) {
                const idx = messages.value.findIndex(m => m.id === targetId)
                if (idx !== -1) {
                  messages.value[idx] = { ...messages.value[idx], status: 'read' }
                }
              }
              return
            }

            const exists = messages.value.some(m => m.id === incoming.id)
            if (!exists) {
              messages.value.unshift(incoming)

              if (totalMessages.value !== null) {
                totalMessages.value++
              }

              const chatId = incoming.chat?.id
              if (canLoadProfilePicture(chatId) && !contactPicMap[chatId]) {
                api.post('/api/media/pictures/info', { token, chatId })
                  .then((res: any) => {
                    if (res?.data?.info?.url) {
                      contactPicMap[chatId] = res.data.info.url
                    }
                  }).catch(() => {
                    // ignore errors
                  })
              }

              const sender = incoming.chat?.title || incoming.from || t('unknown')
              pushToast(t('messages_new_message_from', [sender]), 'info')
            }
          },
        },
      ],
      {
        token,
        subscribeToken: true,
        onConnectError: (err: unknown) => {
          console.error('Messages: cable connection error', err)
          wsConnected.value = false
        },
      },
    )

    onMounted(() => {
      loadMessages()

      wsStatusInterval = setInterval(() => {
        wsConnected.value = cable.isConnected()
      }, 2000)

      // Fallback polling only when WebSocket is disconnected and user hasn't paused it.
      syncInterval = setInterval(() => {
        if (!cable.isConnected() && !syncPaused.value) {
          syncMessagesFallback()
        }
      }, 10000)
    })

    onUnmounted(() => {
      if (wsStatusInterval) {
        clearInterval(wsStatusInterval)
        wsStatusInterval = null
      }

      if (syncInterval) {
        clearInterval(syncInterval)
        syncInterval = null
      }
    })

    return {
      token, messages, serverNumber, totalMessages, loading, error, search, filterType, filterGroupsOnly,
      wsConnected, syncPaused, filteredMessages, currentPage, messagesPerPage, totalPages,
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
      collapsed, isLikelyJson, extractJsonFromText, prettyJson, toggleCollapse, copyText,
      t
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
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-content {
  flex: 1;
}

.header-content h1 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 24px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px;
}

.header-content h1 svg {
  color: var(--branding-primary, #7C3AED);
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

.btn-sync-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid currentColor;
  border-radius: 4px;
  padding: 1px 4px;
  cursor: pointer;
  color: inherit;
  opacity: 0.7;
  margin-left: 4px;
  line-height: 1;
}

.btn-sync-toggle:hover {
  opacity: 1;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #334155;
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 10px;
  padding: 6px 12px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.back-link:hover {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #312e81;
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

/* ── LIST ── */
.messages-list { display: flex; flex-direction: column; gap: 8px; }

/* ── CARD ── */
.mc {
  background: white;
  border-radius: 12px;
  padding: 12px 14px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.06);
  border-left: 3px solid transparent;
}
.mc--me  { border-left-color: #7C3AED; }
.mc--err { border-left-color: #dc2626; }

/* ── HEADER: avatar | identidade | meta ── */
.mc-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.mc-avatar {
  width: 36px; height: 36px;
  border-radius: 50%;
  flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  color: white; font-weight: 700; font-size: 14px;
  overflow: hidden;
}
.mc-avatar--img img { width: 36px; height: 36px; object-fit: cover; display: block; }
.mc-identity {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
}
.mc-identity-main {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}
.mc-name  { font-weight: 600; font-size: 13px; color: #111827; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mc-phone { font-size: 11px; color: #6b7280; font-family: monospace; }
.mc-chat  { font-size: 11px; color: #7C3AED; }
.mc-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  flex-shrink: 0;
}
.mc-time { font-size: 11px; color: #9ca3af; white-space: nowrap; }
.mc-badges { display: flex; flex-wrap: wrap; gap: 3px; justify-content: flex-end; }
.mc-btn-archive { white-space: nowrap; }

/* ── BADGES ── */
.badge {
  font-size: 10px; padding: 1px 5px; border-radius: 3px;
  font-weight: 600; text-transform: uppercase; letter-spacing: .02em;
}
.badge-sent      { background: #f5efff; color: #7C3AED; }
.badge-received  { background: #ecfdf5; color: #059669; }
.badge-read      { background: #eff6ff; color: #2563eb; }
.badge-status    { background: #eff6ff; color: #2563eb; }
.badge-warning   { background: #fefce8; color: #ca8a04; }
.badge-history   { background: #f3f4f6; color: #6b7280; }
.badge-forward   { background: #fff7ed; color: #c2410c; }
.badge-broadcast { background: #fdf4ff; color: #9333ea; }
.badge-danger    { background: #fef2f2; color: #dc2626; }
.badge-success   { background: #f0fdf4; color: #16a34a; }
.badge-secondary { background: #f3f4f6; color: #6b7280; }

/* ── PARTICIPANTE ── */
.mc-participant {
  display: flex; align-items: center; gap: 8px;
  padding: 5px 10px;
  background: #eef6ff; border: 1px solid #bfdbfe;
  border-radius: 8px; flex-shrink: 0;
}
.mc-participant-avatar {
  width: 32px; height: 32px; border-radius: 50%;
  object-fit: cover; flex-shrink: 0;
}
.mc-participant-avatar--initial {
  display: flex; align-items: center; justify-content: center;
  color: white; font-weight: 700; font-size: 11px;
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
}
.mc-participant-info { display: flex; flex-direction: column; gap: 1px; }
.mc-participant-name  { font-size: 12px; font-weight: 700; color: #2563eb; }
.mc-participant-phone { font-size: 11px; color: #4b5563; font-family: monospace; }
.mc-participant-lid   { font-size: 10px; color: #9ca3af; font-family: monospace; }

/* ── BODY ── */
.mc-body { display: flex; flex-direction: column; gap: 6px; margin-bottom: 8px; }
.mc-inreply { font-size: 11px; color: #6b7280; }
.mc-inreply a { color: #7C3AED; text-decoration: none; }
.mc-text { font-size: 13px; color: #1f2937; line-height: 1.5; white-space: pre-wrap; word-break: break-word; }

/* anúncio */
.mc-ads {
  display: flex; gap: 8px; align-items: flex-start;
  padding: 8px; background: #fff5f5;
  border: 1px solid #fecaca; border-radius: 8px;
}
.mc-ads-thumb { width: 56px; height: 56px; object-fit: cover; border-radius: 6px; flex-shrink: 0; }
.mc-ads-info { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.mc-ads-info strong { font-size: 12px; color: #be123c; }
.mc-ads-meta { font-size: 11px; color: #6b7280; }
.mc-ads-link { font-size: 11px; color: #dc2626; text-decoration: none; }
.mc-ads-link:hover { text-decoration: underline; }

/* url */
.mc-url {
  display: flex; gap: 8px; align-items: flex-start;
  padding: 8px; background: #f9fafb;
  border: 1px solid #e5e7eb; border-radius: 8px;
}
.mc-url-thumb { width: 56px; height: 56px; object-fit: cover; border-radius: 6px; flex-shrink: 0; }
.mc-url-info { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.mc-url-info strong { font-size: 12px; color: #111827; }
.mc-url-desc { font-size: 11px; color: #6b7280; margin: 0; }
.mc-url-link { font-size: 11px; color: #7C3AED; text-decoration: none; word-break: break-all; }
.mc-url-link:hover { text-decoration: underline; }

/* mídia */
.mc-img  { max-width: 240px; max-height: 240px; border-radius: 8px; display: block; cursor: pointer; }
.mc-audio { width: 100%; max-width: 320px; }
.mc-video { max-width: 100%; max-height: 240px; border-radius: 8px; display: block; }
.mc-media-actions { display: flex; gap: 4px; margin-top: 4px; flex-wrap: wrap; }

.mc-file {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 10px; background: #f3f4f6; border-radius: 8px;
}
.mc-file-icon { color: #6b7280; flex-shrink: 0; }
.mc-file-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.mc-file-name { font-weight: 600; font-size: 12px; color: #111827; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mc-file-meta { font-size: 10px; color: #9ca3af; font-family: monospace; }

/* erros */
.mc-exceptions {
  font-size: 12px; color: #dc2626; padding: 6px 8px;
  background: #fef2f2; border: 1px solid #fecaca; border-radius: 6px;
}
.mc-exceptions ul { margin: 4px 0 0; padding-left: 14px; color: #991b1b; }
.mc-exceptions li { margin-bottom: 2px; }

/* ── FOOTER ── */
.mc-footer {
  display: flex; align-items: center; gap: 8px; flex-wrap: wrap;
  padding-top: 8px; border-top: 1px solid #f3f4f6;
}
.mc-id      { font-size: 10px; color: #d1d5db; font-family: monospace; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mc-trackid { font-size: 10px; color: #d1d5db; font-family: monospace; }
.mc-actions { display: flex; align-items: center; gap: 4px; flex-wrap: wrap; margin-left: auto; }
.mc-type    { font-size: 10px; padding: 1px 5px; background: #f3f4f6; color: #9ca3af; border-radius: 3px; font-family: monospace; }
.btn-revoke { background: #fef2f2 !important; border-color: #fecaca !important; color: #dc2626 !important; }
.btn-revoke:hover:not(:disabled) { background: #fee2e2 !important; }

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

html[data-theme='dark'] .back-link {
  background: rgba(15, 23, 42, 0.92);
  border-color: rgba(71, 85, 105, 0.3);
  color: #e2e8f0;
}

html[data-theme='dark'] .back-link:hover {
  background: rgba(30, 41, 59, 0.94);
  border-color: rgba(124, 58, 237, 0.3);
  color: #f8fafc;
}

html[data-theme='dark'] .btn-refresh {
  background: rgba(15, 23, 42, 0.92);
  border-color: rgba(71, 85, 105, 0.32);
  color: #e2e8f0;
}

html[data-theme='dark'] .btn-refresh:hover:not(:disabled) {
  border-color: rgba(124, 58, 237, 0.46);
  color: #c4b5fd;
}

html[data-theme='dark'] .filters-bar {
  background: rgba(15, 23, 42, 0.88);
  border: 1px solid rgba(71, 85, 105, 0.24);
}

html[data-theme='dark'] .messages-count {
  background: rgba(30, 41, 59, 0.94);
  color: #cbd5e1;
  border: 1px solid rgba(71, 85, 105, 0.24);
}

html[data-theme='dark'] .pagination-controls {
  background: rgba(15, 23, 42, 0.92);
  border: 1px solid rgba(71, 85, 105, 0.24);
  box-shadow: 0 16px 30px rgba(2, 6, 23, 0.26);
}

html[data-theme='dark'] .btn-page {
  background: rgba(30, 41, 59, 0.94);
  color: #e2e8f0;
  border: 1px solid rgba(71, 85, 105, 0.26);
}

html[data-theme='dark'] .btn-page:hover:not(:disabled) {
  background: rgba(76, 29, 149, 0.88);
  border-color: rgba(124, 58, 237, 0.58);
}

html[data-theme='dark'] .btn-page:disabled {
  background: rgba(30, 41, 59, 0.58);
  color: #64748b;
  border-color: rgba(71, 85, 105, 0.16);
}

html[data-theme='dark'] .page-info {
  background: rgba(30, 41, 59, 0.94);
  color: #cbd5e1;
}

html[data-theme='dark'] .page-size-select {
  background: rgba(11, 22, 40, 0.96);
  border-color: #334155;
  color: #e2e8f0;
}

html[data-theme='dark'] .page-size-select:hover:not(:disabled) {
  border-color: rgba(124, 58, 237, 0.52);
}

html[data-theme='dark'] .page-size-select:disabled {
  background: rgba(30, 41, 59, 0.58);
}

html[data-theme='dark'] .btn-small {
  background: rgba(30, 41, 59, 0.94);
  border-color: rgba(71, 85, 105, 0.24);
  color: #e2e8f0;
}

html[data-theme='dark'] .btn-small:hover {
  background: rgba(51, 65, 85, 0.96);
}

html[data-theme='dark'] .mc {
  background: rgba(15, 23, 42, 0.88);
  border-color: rgba(71, 85, 105, 0.3);
}
html[data-theme='dark'] .mc-name    { color: #f1f5f9; }
html[data-theme='dark'] .mc-phone   { color: #94a3b8; }
html[data-theme='dark'] .mc-chat    { color: #a78bfa; }
html[data-theme='dark'] .mc-time    { color: #64748b; }
html[data-theme='dark'] .mc-text    { color: #e2e8f0; }
html[data-theme='dark'] .mc-footer  { border-top-color: rgba(71, 85, 105, 0.2); }
html[data-theme='dark'] .mc-id,
html[data-theme='dark'] .mc-trackid { color: #475569; }
html[data-theme='dark'] .mc-type    { background: rgba(30, 41, 59, 0.9); color: #94a3b8; }
html[data-theme='dark'] .mc-file    { background: rgba(30, 41, 59, 0.7); border: 1px solid rgba(71,85,105,0.2); }
html[data-theme='dark'] .mc-file-name { color: #e2e8f0; }
html[data-theme='dark'] .mc-participant { background: rgba(12, 74, 110, 0.3); border-color: rgba(56, 189, 248, 0.25); }
html[data-theme='dark'] .mc-participant-name  { color: #38bdf8; }
html[data-theme='dark'] .mc-participant-phone { color: #94a3b8; }
html[data-theme='dark'] .mc-participant-lid   { color: #64748b; }
html[data-theme='dark'] .mc-ads  { background: rgba(127, 29, 29, 0.18); border-color: rgba(239,68,68,0.2); }
html[data-theme='dark'] .mc-url  { background: rgba(30, 41, 59, 0.6);  border-color: rgba(71,85,105,0.2); }
html[data-theme='dark'] .mc-url-info strong { color: #f1f5f9; }
html[data-theme='dark'] .mc-exceptions { background: rgba(127, 29, 29, 0.18); border-color: rgba(239,68,68,0.2); }
</style>
