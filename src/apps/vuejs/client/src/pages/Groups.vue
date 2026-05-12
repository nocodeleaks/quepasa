<template>
  <div class="groups-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile" type="button">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
        </svg>
        {{ t('back') }}
      </button>
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
          </svg>
          {{ t('groups_title') }}
        </h1>
        <p>{{ token }}</p>
      </div>
    </div>

    <div class="toolbar">
      <div class="search-wrap">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" class="search-icon">
          <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
        </svg>
        <input v-model="searchQuery" type="text" class="search-input" :placeholder="t('groups_search_placeholder')" />
        <button v-if="searchQuery" class="clear-btn" type="button" @click="searchQuery = ''">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
            <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
          </svg>
        </button>
      </div>
      <div class="toolbar-right">
        <div class="view-toggle">
          <button type="button" :class="['toggle-btn', { active: viewMode === 'card' }]" @click="viewMode = 'card'" :title="t('card_view')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M3 3h8v8H3zm10 0h8v8h-8zM3 13h8v8H3zm10 0h8v8h-8z"/>
            </svg>
          </button>
          <button type="button" :class="['toggle-btn', { active: viewMode === 'list' }]" @click="viewMode = 'list'" :title="t('table_view')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M3 13h2v-2H3v2zm0 4h2v-2H3v2zm0-8h2V7H3v2zm4 4h14v-2H7v2zm0 4h14v-2H7v2zM7 7v2h14V7H7z"/>
            </svg>
          </button>
        </div>
        <button class="btn-icon" type="button" @click="load" :title="t('groups_refresh')">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor">
            <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
          </svg>
        </button>
        <button class="btn-primary" type="button" @click="openCreateModal">
          <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
          </svg>
          {{ t('groups_create') }}
        </button>
      </div>
    </div>

    <div v-if="error" class="error-box">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('groups_loading') }}</p>
    </div>

    <template v-else>
      <div v-if="groups.length" class="stats-bar">
        <span class="stat-chip">
          {{ filteredGroups.length }} / {{ groups.length }} {{ t('groups_title').toLowerCase() }}
        </span>
        <select v-model.number="pageSize" class="page-size-select">
          <option :value="12">12</option>
          <option :value="24">24</option>
          <option :value="50">50</option>
          <option :value="100">100</option>
        </select>
      </div>

      <div v-if="filteredGroups.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" width="52" height="52" fill="currentColor">
          <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
        </svg>
        <p>{{ searchQuery ? t('groups_empty_for_query').replace('{0}', searchQuery) : t('groups_empty') }}</p>
      </div>

      <!-- Card view -->
      <div v-else-if="viewMode === 'card'" class="groups-grid">
        <div v-for="group in displayGroups" :key="group.JID" class="group-card" @click="goToGroup(group.JID)">
          <div class="card-avatar">
            <img v-if="groupPictures[group.JID]" :src="groupPictures[group.JID]" :alt="group.Name" @error="handleImageError(group.JID)" />
            <div v-else class="avatar-initials" :style="avatarStyle(group)">{{ initials(group.Name) }}</div>
          </div>
          <div class="card-body">
            <div class="card-top">
              <span class="group-name">{{ group.Name || t('groups_unnamed') }}</span>
              <span v-if="group.lastMessage" class="msg-time">{{ formatTime(group.lastMessage.timestamp) }}</span>
            </div>
            <div class="card-preview">
              <span v-if="group.lastMessage" class="msg-preview">
                <b>{{ group.lastMessage.senderName }}:</b> {{ truncate(group.lastMessage.text) }}
              </span>
              <span v-else class="msg-empty">{{ t('groups_no_recent_messages') }}</span>
            </div>
            <div class="card-footer">
              <span class="participants-pill">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor">
                  <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
                </svg>
                {{ group.Participants?.length || 0 }}
              </span>
              <span v-if="group.IsAnnounce" class="badge badge-announce" :title="t('groups_only_admins_send')">
                <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor">
                  <path d="M18 11v2h4v-2h-4zm-2 6.61c.96.71 2.21 1.65 3.2 2.39.4-.53.8-1.07 1.2-1.6-.99-.74-2.24-1.68-3.2-2.4-.4.54-.8 1.08-1.2 1.61zM20.4 5.6c-.4-.53-.8-1.07-1.2-1.6-.99.74-2.24 1.68-3.2 2.4.4.53.8 1.07 1.2 1.6.96-.72 2.21-1.65 3.2-2.4zM4 9c-1.1 0-2 .9-2 2v2c0 1.1.9 2 2 2h1v4h2v-4h1l5 3V6L8 9H4zm11.5 3c0-1.33-.58-2.53-1.5-3.35v6.69c.92-.81 1.5-2.01 1.5-3.34z"/>
                </svg>
              </span>
              <span v-if="group.IsParent" class="badge badge-community" :title="t('groups_community')">
                <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor">
                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14H9V8h2v8zm4 0h-2V8h2v8z"/>
                </svg>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- List view -->
      <div v-else class="groups-list">
        <div v-for="group in displayGroups" :key="group.JID" class="group-row" @click="goToGroup(group.JID)">
          <div class="row-avatar">
            <img v-if="groupPictures[group.JID]" :src="groupPictures[group.JID]" :alt="group.Name" @error="handleImageError(group.JID)" />
            <div v-else class="avatar-initials avatar-sm" :style="avatarStyle(group)">{{ initials(group.Name) }}</div>
          </div>
          <div class="row-info">
            <span class="group-name">{{ group.Name || t('groups_unnamed') }}</span>
            <span class="row-meta">
              {{ t('groups_participants_count').replace('{0}', String(group.Participants?.length || 0)) }}
              <span v-if="group.Topic" class="topic-inline"> · {{ group.Topic }}</span>
            </span>
          </div>
          <div class="row-badges">
            <span v-if="group.IsAnnounce" class="badge badge-announce">{{ t('groups_only_admins_send') }}</span>
            <span v-if="group.IsParent" class="badge badge-community">{{ t('groups_community') }}</span>
          </div>
          <svg class="row-chevron" viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/>
          </svg>
        </div>
      </div>

      <div v-if="totalPages > 1" class="pagination">
        <button class="page-btn" type="button" :disabled="currentPage === 1" @click="goToPage(1)">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M18.41 16.59L13.82 12l4.59-4.59L17 6l-6 6 6 6zM6 6h2v12H6z"/></svg>
        </button>
        <button class="page-btn" type="button" :disabled="currentPage === 1" @click="prevPage">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z"/></svg>
        </button>
        <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
        <button class="page-btn" type="button" :disabled="currentPage >= totalPages" @click="nextPage">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
        </button>
        <button class="page-btn" type="button" :disabled="currentPage >= totalPages" @click="goToPage(totalPages)">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M5.59 7.41L10.18 12l-4.59 4.59L7 18l6-6-6-6zM16 6h2v12h-2z"/></svg>
        </button>
      </div>
    </template>

    <!-- Create Group Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreateModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
              <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
            </svg>
          </div>
          <h3>{{ t('groups_create') }}</h3>
          <button type="button" class="modal-close" @click="closeCreateModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div class="field">
            <label class="field-label">{{ t('groups_field_name') }}</label>
            <input
              v-model="createForm.title"
              ref="titleInputRef"
              type="text"
              class="field-input"
              :placeholder="t('groups_field_name_placeholder')"
              maxlength="25"
              @keydown.enter="submitCreate"
            />
            <span class="field-hint">{{ createForm.title.length }} / 25</span>
          </div>

          <div class="field">
            <label class="field-label">{{ t('groups_field_participants') }}</label>
            <textarea
              v-model="createForm.participantsRaw"
              class="field-input field-textarea"
              :placeholder="t('groups_field_participants_placeholder')"
              rows="3"
            />
            <span class="field-hint">{{ t('groups_field_participants_hint') }}</span>
          </div>

          <div v-if="createError" class="modal-error">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
            {{ createError }}
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn-cancel" @click="closeCreateModal" :disabled="creating">
            {{ t('cancel') }}
          </button>
          <button type="button" class="btn-confirm" @click="submitCreate" :disabled="creating || !createForm.title.trim()">
            <svg v-if="creating" viewBox="0 0 24 24" width="14" height="14" fill="currentColor" class="spin-icon">
              <path d="M12 4V1L8 5l4 4V6c3.31 0 6 2.69 6 6s-2.69 6-6 6-6-2.69-6-6H4c0 4.42 3.58 8 8 8s8-3.58 8-8-3.58-8-8-8z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
            </svg>
            {{ creating ? t('groups_creating') : t('groups_create') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

interface LastMessage { timestamp: string; senderName: string; text: string }
interface Group { JID: string; Name: string; Participants: any[]; IsAnnounce: boolean; IsParent: boolean; Topic?: string; lastMessage?: LastMessage }

const AVATAR_COLORS = ['#7C3AED','#2563EB','#059669','#D97706','#DC2626','#0891B2','#65A30D','#9333EA']

export default defineComponent({
  setup() {
    const route = useRoute()
    const router = useRouter()
    const token = route.params.token as string
    const { t, locale } = useLocale()

    const groups = ref<Group[]>([])
    const groupPictures = ref<Record<string, string>>({})
    const loading = ref(false)
    const error = ref('')
    const searchQuery = ref('')
    const viewMode = ref<'card' | 'list'>('card')
    const currentPage = ref(1)
    const pageSize = ref(24)

    const showCreateModal = ref(false)
    const creating = ref(false)
    const createError = ref('')
    const titleInputRef = ref<HTMLInputElement | null>(null)
    const createForm = ref({ title: '', participantsRaw: '' })

    const filteredGroups = computed(() => {
      if (!searchQuery.value) return groups.value
      const q = searchQuery.value.toLowerCase()
      return groups.value.filter(g =>
        (g.Name || '').toLowerCase().includes(q) ||
        g.JID.toLowerCase().includes(q) ||
        (g.Topic || '').toLowerCase().includes(q)
      )
    })

    const totalPages = computed(() => Math.max(1, Math.ceil(filteredGroups.value.length / pageSize.value)))

    const displayGroups = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value
      return filteredGroups.value.slice(start, start + pageSize.value)
    })

    watch([searchQuery, pageSize], () => { currentPage.value = 1 })
    watch(totalPages, () => { if (currentPage.value > totalPages.value) currentPage.value = totalPages.value })

    function goToPage(p: number) { currentPage.value = Math.max(1, Math.min(p, totalPages.value)) }
    function nextPage() { if (currentPage.value < totalPages.value) currentPage.value++ }
    function prevPage() { if (currentPage.value > 1) currentPage.value-- }

    function initials(name: string) {
      if (!name) return '?'
      return name.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase()
    }

    function avatarStyle(g: Group) {
      const idx = g.JID.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % AVATAR_COLORS.length
      return { background: AVATAR_COLORS[idx] }
    }

    function truncate(text: string, max = 55) {
      if (!text || text.length <= max) return text
      return text.substring(0, max) + '…'
    }

    function formatTime(timestamp: string) {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const oneDay = 86400000
      if (diff < oneDay && date.getDate() === now.getDate())
        return date.toLocaleTimeString(locale.value, { hour: '2-digit', minute: '2-digit' })
      if (diff < 2 * oneDay) return t('time_yesterday')
      if (diff < 7 * oneDay) return date.toLocaleDateString(locale.value, { weekday: 'short' })
      return date.toLocaleDateString(locale.value, { day: '2-digit', month: '2-digit' })
    }

    function handleImageError(jid: string) { delete groupPictures.value[jid] }

    function goToGroup(jid: string) { router.push(`/server/${token}/groups/${encodeURIComponent(jid)}`) }

    function buildPreview(msg: any): string {
      let text = msg.text || ''
      if (msg.attachment) {
        const mime = msg.attachment.mimetype || ''
        if (mime.startsWith('image/')) text = text ? `[IMG] ${text}` : `[IMG] ${t('media_image')}`
        else if (mime.startsWith('video/')) text = text ? `[VID] ${text}` : `[VID] ${t('media_video')}`
        else if (mime.startsWith('audio/') || msg.type === 'ptt') text = text ? `[AUD] ${text}` : `[AUD] ${t('media_audio')}`
        else text = text ? `[ARQ] ${text}` : `[ARQ] ${t('media_file')}`
      }
      if (!text && msg.inreply) text = `[RPL] ${t('media_reply')}`
      return text
    }

    async function load() {
      loading.value = true
      error.value = ''
      try {
        const res = await api.get('/api/groups', { params: { token } })
        const raw = (res.data?.groups || []) as Group[]
        raw.sort((a, b) => (a.Name || '').localeCompare(b.Name || ''))
        groups.value = raw
        void loadGroupPictures(raw.slice(0, 20))
        await loadLastMessages()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.message || t('groups_error_load')
      } finally {
        loading.value = false
      }
    }

    async function loadGroupPictures(list: Group[]) {
      for (const g of list) {
        try {
          const res = await api.post('/api/media/pictures/info', { token, chatId: g.JID })
          if (res.data?.info?.url) groupPictures.value[g.JID] = res.data.info.url
        } catch { /* ignore */ }
      }
    }

    async function loadLastMessages() {
      try {
        const res = await api.get('/api/messages', { params: { token } })
        const messages = res.data?.messages || []
        const last: Record<string, LastMessage> = {}
        for (const msg of messages) {
          const chatId = msg.chat?.id
          if (!chatId?.endsWith('@g.us')) continue
          if (['unhandled','revoked','system'].includes(msg.type)) continue
          if (msg.debug?.reason === 'discard') continue
          const preview = buildPreview(msg)
          if (!preview) continue
          if (!last[chatId] || new Date(msg.timestamp) > new Date(last[chatId].timestamp)) {
            last[chatId] = {
              timestamp: msg.timestamp,
              senderName: msg.participant?.title || msg.participant?.phone || t('unknown'),
              text: preview,
            }
          }
        }
        groups.value = groups.value
          .map(g => ({ ...g, lastMessage: last[g.JID] }))
          .sort((a, b) => {
            if (a.lastMessage && b.lastMessage)
              return new Date(b.lastMessage.timestamp).getTime() - new Date(a.lastMessage.timestamp).getTime()
            if (a.lastMessage) return -1
            if (b.lastMessage) return 1
            return (a.Name || '').localeCompare(b.Name || '')
          })
      } catch { /* optional */ }
    }

    function openCreateModal() {
      createForm.value = { title: '', participantsRaw: '' }
      createError.value = ''
      showCreateModal.value = true
      nextTick(() => titleInputRef.value?.focus())
    }

    function closeCreateModal() {
      if (creating.value) return
      showCreateModal.value = false
    }

    async function submitCreate() {
      if (!createForm.value.title.trim() || creating.value) return
      creating.value = true
      createError.value = ''
      const participants = createForm.value.participantsRaw
        .split(',').map(p => p.trim()).filter(Boolean)
      try {
        await api.post('/api/groups', { token, title: createForm.value.title.trim(), participants })
        pushToast(t('groups_created'), 'success')
        showCreateModal.value = false
        await load()
      } catch (err: any) {
        createError.value = err?.response?.data?.result || err?.message || t('groups_error_create')
      } finally {
        creating.value = false
      }
    }

    onMounted(load)

    return {
      t, token, groups, groupPictures, loading, error, searchQuery, viewMode,
      currentPage, pageSize, filteredGroups, totalPages, displayGroups,
      initials, avatarStyle, truncate, formatTime, handleImageError,
      goToGroup, goToPage, nextPage, prevPage, load,
      showCreateModal, creating, createError, createForm, titleInputRef,
      openCreateModal, closeCreateModal, submitCreate,
    }
  },
})
</script>

<style scoped>
.groups-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
}

.header-content h1 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px;
}

.header-content h1 svg { color: var(--branding-primary, #7C3AED); }

.header-content p {
  color: #6b7280;
  font-size: 13px;
  font-family: monospace;
  margin: 0;
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

.back-link:hover { background: #eef2ff; border-color: #c7d2fe; color: #312e81; }

.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.search-wrap {
  flex: 1;
  min-width: 200px;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon { position: absolute; left: 12px; color: #9ca3af; pointer-events: none; }

.search-input {
  width: 100%;
  padding: 10px 34px 10px 36px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  background: #f9fafb;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
  background: white;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.08);
}

.clear-btn {
  position: absolute;
  right: 10px;
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
  padding: 4px;
  display: flex;
}

.clear-btn:hover { color: #374151; }

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.view-toggle {
  display: flex;
  background: #f3f4f6;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e5e7eb;
}

.toggle-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  background: transparent;
  border: none;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn.active {
  background: var(--branding-primary, #7C3AED);
  color: white;
}

.toggle-btn:hover:not(.active) { color: #374151; }

.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover { background: #e5e7eb; }

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 16px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(124,58,237,0.25); }

.error-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  color: #dc2626;
  margin-bottom: 20px;
  font-size: 14px;
}

.loading-state { text-align: center; padding: 60px 0; color: #6b7280; }

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 14px;
}

@keyframes spin { to { transform: rotate(360deg); } }

.stats-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
}

.stat-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px;
  background: #f5f3ff;
  color: var(--branding-primary, #7C3AED);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.page-size-select {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 5px 10px;
  background: white;
  color: #374151;
  font-size: 13px;
  cursor: pointer;
}

.empty-state {
  text-align: center;
  padding: 60px 0;
  color: #9ca3af;
}

.empty-state svg { color: #d1d5db; margin-bottom: 14px; }
.empty-state p { font-size: 15px; font-weight: 500; margin: 0; }

/* Card grid */
.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 14px;
}

.group-card {
  display: flex;
  gap: 14px;
  padding: 16px;
  background: white;
  border-radius: 14px;
  border: 1px solid #e5e7eb;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  cursor: pointer;
  transition: all 0.18s;
}

.group-card:hover {
  border-color: var(--branding-primary, #7C3AED);
  box-shadow: 0 4px 14px rgba(124,58,237,0.12);
  transform: translateY(-2px);
}

.card-avatar { flex-shrink: 0; }

.card-avatar img {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  object-fit: cover;
}

.avatar-initials {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
  font-weight: 700;
}

.avatar-sm { width: 44px; height: 44px; border-radius: 10px; font-size: 14px; }

.card-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 5px; }

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}

.group-name {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.msg-time { font-size: 11px; color: #9ca3af; white-space: nowrap; flex-shrink: 0; }

.card-preview { font-size: 12px; color: #6b7280; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.msg-empty { font-style: italic; opacity: 0.7; }

.card-footer {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
}

.participants-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #6b7280;
}

/* List view */
.groups-list {
  background: white;
  border-radius: 14px;
  border: 1px solid #e5e7eb;
  overflow: hidden;
}

.group-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.15s;
  border-bottom: 1px solid #f3f4f6;
}

.group-row:last-child { border-bottom: none; }
.group-row:hover { background: #f9fafb; }

.row-avatar { flex-shrink: 0; }

.row-avatar img {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  object-fit: cover;
}

.row-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.row-meta {
  font-size: 12px;
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topic-inline { opacity: 0.8; }

.row-badges { display: flex; gap: 6px; flex-shrink: 0; }
.row-chevron { color: #d1d5db; flex-shrink: 0; }

/* Shared badges */
.badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 7px;
  border-radius: 20px;
  font-size: 10px;
  font-weight: 600;
}

.badge-announce { background: #fef3c7; color: #92400e; }
.badge-community { background: #dbeafe; color: #1e40af; }

/* Pagination */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 24px;
}

.page-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: white;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) { border-color: var(--branding-primary, #7C3AED); color: var(--branding-primary, #7C3AED); }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.page-info { font-size: 14px; font-weight: 600; color: #374151; min-width: 60px; text-align: center; }

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-card {
  background: white;
  border-radius: 18px;
  width: 100%;
  max-width: 460px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 24px 16px;
  border-bottom: 1px solid #f3f4f6;
}

.modal-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #f5f3ff;
  color: var(--branding-primary, #7C3AED);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.modal-header h3 {
  flex: 1;
  font-size: 17px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: none;
  border: none;
  border-radius: 8px;
  color: #9ca3af;
  cursor: pointer;
  transition: all 0.15s;
}

.modal-close:hover { background: #f3f4f6; color: #374151; }

.modal-body {
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field { display: flex; flex-direction: column; gap: 6px; }

.field-label {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}

.field-input {
  padding: 10px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  background: #f9fafb;
  transition: all 0.2s;
  font-family: inherit;
  resize: vertical;
}

.field-input:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
  background: white;
  box-shadow: 0 0 0 4px rgba(124,58,237,0.08);
}

.field-textarea { min-height: 80px; }

.field-hint {
  font-size: 11px;
  color: #9ca3af;
}

.modal-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 13px;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px 20px;
  border-top: 1px solid #f3f4f6;
}

.btn-cancel {
  padding: 9px 20px;
  background: #f3f4f6;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-cancel:hover:not(:disabled) { background: #e5e7eb; }
.btn-cancel:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-confirm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 20px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-confirm:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(124,58,237,0.3); }
.btn-confirm:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

.spin-icon { animation: spin 0.8s linear infinite; }

@media (max-width: 600px) {
  .hide-mobile { display: none; }
  .groups-grid { grid-template-columns: 1fr; }
  .row-badges { display: none; }
  .modal-card { border-radius: 14px; }
}
</style>
