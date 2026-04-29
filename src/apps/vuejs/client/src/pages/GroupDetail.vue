<template>
  <div class="group-detail-page">
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>{{ t('group_detail_loading') }}</p>
    </div>

    <div v-else-if="error" class="error-container">
      <i class="fa fa-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <router-link :to="`/server/${token}/groups`" class="btn-back">{{ t('group_detail_back_to_groups') }}</router-link>
    </div>

    <div v-else class="group-layout">
      <div class="messages-column">
        <div class="messages-header">
          <div>
            <p class="eyebrow">{{ t('group_detail_recent_history') }}</p>
            <h2>{{ t('messages') }}</h2>
          </div>
          <small v-if="totalMessages">{{ t('group_detail_messages_count', String(visibleMessages.length), String(totalMessages)) }}</small>
        </div>

        <div
          class="messages-list"
          v-if="visibleMessages.length"
          ref="messagesListRef"
          @scroll.passive="onMessagesScroll"
        >
          <div v-for="message in visibleMessages" :key="message.id" class="message-item">
            <div class="message-avatar">
              <img
                v-if="participantPictures[message.participant?.phone || message.participant?.id]"
                :src="participantPictures[message.participant?.phone || message.participant?.id]"
              />
              <div v-else class="avatar-placeholder-small"><i class="fa fa-user"></i></div>
            </div>
            <div class="message-content">
              <div class="message-meta">
                <span class="message-sender">
                  {{ message.participant?.title || formatPhone(message.participant?.phone || message.participant?.id) }}
                </span>
                <span class="message-time">{{ formatTime(message.timestamp) }}</span>
              </div>
              <div class="message-text">{{ messagePreview(message) }}</div>
            </div>
          </div>

          <div class="load-more-state" v-if="isLoadingMore">{{ t('group_detail_loading_more') }}</div>
          <div class="load-more-state" v-else-if="hasMoreMessages">
            <button class="load-more-btn" type="button" @click="loadMoreMessages">{{ t('group_detail_load_more_50') }}</button>
          </div>
          <div class="load-more-state muted" v-else>{{ t('group_detail_history_start') }}</div>
        </div>

        <div v-else class="empty-messages">{{ t('groups_no_recent_messages') }}</div>
      </div>

      <div class="details-column">
        <div class="group-header">
          <router-link :to="`/server/${token}/groups`" class="back-btn">
            <i class="fa fa-arrow-left"></i>
          </router-link>

          <div class="group-photo-container">
            <img v-if="groupPicture" :src="groupPicture" :alt="group.Name" class="group-photo" />
            <div v-else class="group-photo-placeholder">
              <i class="fa fa-users"></i>
            </div>
            <button v-if="isAdmin" @click="setGroupPhoto" class="edit-photo-btn" :title="t('group_detail_change_photo')">
              <i class="fa fa-camera"></i>
            </button>
          </div>

          <h1 class="group-name">
            {{ group.Name || t('groups_unnamed') }}
            <button v-if="isAdmin" @click="setGroupName" class="edit-btn" :title="t('group_detail_change_name')">
              <i class="fa fa-pencil"></i>
            </button>
          </h1>

          <p class="group-meta">{{ t('group_detail_meta', String(group.Participants?.length || 0)) }}</p>
        </div>

        <div class="section" v-if="group.Topic || isAdmin">
          <div class="section-header">
            <i class="fa fa-info-circle"></i>
            <span>{{ t('group_detail_description') }}</span>
            <button v-if="isAdmin" @click="setGroupTopic" class="edit-btn" :title="t('group_detail_change_description')">
              <i class="fa fa-pencil"></i>
            </button>
          </div>
          <p class="description-text" v-if="group.Topic">{{ group.Topic }}</p>
          <p class="description-empty" v-else>{{ t('group_detail_no_description') }}</p>
          <p class="description-meta" v-if="group.TopicSetAt">{{ t('group_detail_created_at', formatDate(group.TopicSetAt)) }}</p>
        </div>

        <div class="section actions-section">
          <button class="action-btn" @click="getInvite">
            <i class="fa fa-link"></i>
            <span>{{ t('group_detail_invite_link') }}</span>
          </button>
          <button v-if="isAdmin" class="action-btn" @click="addParticipant">
            <i class="fa fa-user-plus"></i>
            <span>{{ t('group_detail_add') }}</span>
          </button>
          <button class="action-btn" @click="searchParticipants">
            <i class="fa fa-search"></i>
            <span>{{ t('group_detail_search') }}</span>
          </button>
        </div>

        <div class="section">
          <div class="section-header">
            <i class="fa fa-users"></i>
            <span>{{ t('group_detail_members_count', String(group.Participants?.length || 0)) }}</span>
          </div>

          <div class="search-box" v-if="showSearch">
            <input
              v-model="participantSearch"
              type="text"
              :placeholder="t('group_detail_search_members')"
              class="search-input"
            />
          </div>

          <div class="participants-list">
            <div v-for="participant in filteredParticipants" :key="participant.JID" class="participant-item">
              <div class="participant-avatar">
                <img
                  v-if="participantPictures[participant.PhoneNumber || participant.JID]"
                  :src="participantPictures[participant.PhoneNumber || participant.JID]"
                  :alt="participant.DisplayName"
                />
                <div v-else class="avatar-placeholder">
                  <i class="fa fa-user"></i>
                </div>
              </div>

              <div class="participant-info">
                <div class="participant-name">
                  <span v-if="participant.DisplayName">~ {{ participant.DisplayName }}</span>
                  <span v-else>{{ formatPhone(participant.PhoneNumber || participant.JID) }}</span>
                </div>
                <div class="participant-phone" v-if="participant.DisplayName && participant.PhoneNumber">
                  {{ formatPhone(participant.PhoneNumber) }}
                </div>
              </div>

              <div class="participant-badges">
                <span v-if="participant.IsSuperAdmin" class="badge badge-owner">{{ t('group_detail_owner') }}</span>
                <span v-else-if="participant.IsAdmin" class="badge badge-admin">{{ t('group_detail_admin') }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="section danger-section">
          <button class="danger-btn" @click="leaveGroup">
            <i class="fa fa-sign-out"></i>
            <span>{{ t('group_detail_leave_group') }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

interface Participant {
  JID: string
  PhoneNumber: string
  LID: string
  IsAdmin: boolean
  IsSuperAdmin: boolean
  DisplayName: string
}

export default defineComponent({
  setup() {
    const route = useRoute()
    const router = useRouter()
    const token = route.params.token as string
    const groupid = route.params.id as string
    const encodedGroupId = encodeURIComponent(groupid)

    const group = ref<any>({})
    const groupPicture = ref('')
    const participantPictures = ref<Record<string, string>>({})
    const messages = ref<any[]>([])
    const visibleMessages = ref<any[]>([])
    const loading = ref(false)
    const error = ref('')
    const showSearch = ref(false)
    const participantSearch = ref('')
    const { t, locale } = useLocale()
    const myPhone = ref('')
    const messagesListRef = ref<HTMLElement | null>(null)
    const isLoadingMore = ref(false)

    const pageSize = 50

    const totalMessages = computed(() => messages.value.length)
    const hasMoreMessages = computed(() => visibleMessages.value.length < messages.value.length)

    const isAdmin = computed(() => {
      if (!group.value.Participants || !myPhone.value) return false
      const me = group.value.Participants.find((participant: Participant) => {
        return participant.PhoneNumber?.includes(myPhone.value) || participant.JID?.includes(myPhone.value)
      })
      return me?.IsAdmin || me?.IsSuperAdmin || false
    })

    const filteredParticipants = computed(() => {
      const participants = group.value.Participants || []
      if (!participantSearch.value) return participants

      const query = participantSearch.value.toLowerCase()
      return participants.filter((participant: Participant) => {
        return (
          participant.DisplayName?.toLowerCase().includes(query) ||
          participant.PhoneNumber?.includes(query) ||
          participant.JID?.includes(query)
        )
      })
    })

    function buildMessagePreview(message: any) {
      let preview = message.text || ''

      if (message.attachment) {
        const mime = message.attachment.mimetype || ''

        if (mime.startsWith('image/')) preview = preview ? `[IMG] ${preview}` : '[IMG] Imagem'
        else if (mime.startsWith('video/')) preview = preview ? `[VID] ${preview}` : '[VID] Video'
        else if (mime.startsWith('audio/') || message.type === 'ptt') preview = preview ? `[AUD] ${preview}` : '[AUD] Audio'
        else if (mime.includes('pdf')) preview = preview ? `[PDF] ${preview}` : '[PDF] PDF'
        else preview = preview ? `[ARQ] ${preview}` : '[ARQ] Arquivo'
      }

      if (!preview && message.inreply) preview = '[RPL] Resposta'
      return preview
    }

    async function load() {
      loading.value = true
      error.value = ''
      messages.value = []
      visibleMessages.value = []

      try {
        const res = await api.post('/api/groups/get', { token, groupId: groupid })
        group.value = res.data?.groupinfo || {}

        await loadGroupPicture()

        const serverRes = await api.post('/api/sessions/get', { token })
        const wid = serverRes.data?.server?.wid || ''
        myPhone.value = wid.replace('@s.whatsapp.net', '').replace('@lid', '')

        await loadParticipantPictures()
        await loadMessages()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.message || t('group_detail_error_load')
      } finally {
        loading.value = false
      }
    }

    async function loadGroupPicture() {
      try {
        const res = await api.post('/api/media/pictures/info', { token, chatId: groupid })
        if (res.data?.info?.url) {
          groupPicture.value = res.data.info.url
        }
      } catch {
        // ignore picture failures
      }
    }

    async function loadParticipantPictures() {
      const participants = (group.value.Participants || []).slice(0, 10)
      for (const participant of participants) {
        const id = participant.PhoneNumber || participant.JID
        if (!id) continue

        try {
          const res = await api.post('/api/media/pictures/info', { token, chatId: id })
          if (res.data?.info?.url) {
            participantPictures.value[id] = res.data.info.url
          }
        } catch {
          // ignore picture failures
        }
      }
    }

    function formatPhone(phone: string) {
      if (!phone) return ''
      let clean = phone.replace('@s.whatsapp.net', '').replace('@lid', '')
      if (clean.length >= 12) {
        const country = clean.slice(0, 2)
        const ddd = clean.slice(2, 4)
        const part1 = clean.slice(4, 9)
        const part2 = clean.slice(9)
        return `+${country} ${ddd} ${part1}-${part2}`
      }
      return clean
    }

    function formatDate(dateStr: string) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return date.toLocaleDateString(locale.value, {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    }

    function formatTime(timestamp: string) {
      if (!timestamp) return ''

      const date = new Date(timestamp)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const oneDay = 24 * 60 * 60 * 1000

      if (diff < oneDay && date.getDate() === now.getDate()) {
        return date.toLocaleTimeString(locale.value, { hour: '2-digit', minute: '2-digit' })
      }
      if (diff < 2 * oneDay) return t('time_yesterday')
      if (diff < 7 * oneDay) return date.toLocaleDateString(locale.value, { weekday: 'short' })
      return date.toLocaleDateString(locale.value, { day: '2-digit', month: '2-digit' })
    }

    async function loadMessages() {
      try {
        const res = await api.get('/api/messages', { params: { token } })
        const rawMessages = res.data?.messages || []
        const groupMessages: any[] = []

        for (const message of rawMessages) {
          const chatId = message.chat?.id
          if (!chatId || chatId !== groupid) continue
          if (message.type === 'unhandled' || message.type === 'revoked' || message.type === 'system') continue
          if (message.debug?.reason === 'discard') continue
          if (!message.text && !message.attachment && !message.inreply) continue

          const preview = buildMessagePreview(message)
          if (!preview) continue

          groupMessages.push({ ...message, text: preview })
        }

        groupMessages.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
        messages.value = groupMessages
        visibleMessages.value = groupMessages.slice(0, pageSize)
      } catch {
        // messages are optional enrichment
      }
    }

    function loadMoreMessages() {
      if (!hasMoreMessages.value || isLoadingMore.value) return
      isLoadingMore.value = true
      const nextSlice = messages.value.slice(visibleMessages.value.length, visibleMessages.value.length + pageSize)
      visibleMessages.value = visibleMessages.value.concat(nextSlice)
      requestAnimationFrame(() => {
        isLoadingMore.value = false
      })
    }

    function onMessagesScroll() {
      const el = messagesListRef.value
      if (!el || isLoadingMore.value || !hasMoreMessages.value) return

      const distanceToBottom = el.scrollHeight - (el.scrollTop + el.clientHeight)
      if (distanceToBottom < 160) {
        loadMoreMessages()
      }
    }

    function messagePreview(message: any, maxLength = 120) {
      const text = message.text || ''
      if (text.length <= maxLength) return text
      return `${text.substring(0, maxLength)}...`
    }

    function searchParticipants() {
      showSearch.value = !showSearch.value
      if (!showSearch.value) participantSearch.value = ''
    }

    async function leaveGroup() {
      if (!confirm(t('group_detail_confirm_leave'))) return

      try {
        await api.post('/api/groups/leave', { token, groupId: groupid })
        pushToast(t('group_detail_leave_requested'), 'success')
        router.push(`/server/${token}/groups`)
      } catch (err: any) {
        pushToast(err?.response?.data?.result || err?.message || t('group_detail_error_leave'), 'error')
      }
    }

    async function setGroupName() {
      const name = prompt(t('group_detail_prompt_name'), group.value.Name || '')
      if (!name) return

      try {
        await api.patch('/api/groups', { token, groupId: groupid, name })
        pushToast(t('group_detail_name_updated'), 'success')
        await load()
      } catch (err: any) {
        pushToast(err?.response?.data?.result || err?.message || t('group_detail_error_name'), 'error')
      }
    }

    async function setGroupTopic() {
      const topic = prompt(t('group_detail_prompt_description'), group.value.Topic || '')
      if (topic == null) return

      try {
        await api.patch('/api/groups', { token, groupId: groupid, topic })
        pushToast(t('group_detail_description_updated'), 'success')
        await load()
      } catch (err: any) {
        pushToast(err?.response?.data?.result || err?.message || t('group_detail_error_description'), 'error')
      }
    }

    async function addParticipant() {
      const phones = prompt(t('group_detail_prompt_add_participants'))
      if (!phones) return

      const participants = phones.split(',').map((value: string) => value.trim()).filter(Boolean)

      try {
        await api.put('/api/groups/participants', {
          token,
          groupId: groupid,
          action: 'add',
          participants,
        })
        pushToast(t('group_detail_participants_added'), 'success')
        await load()
      } catch (err: any) {
        pushToast(err?.response?.data?.result || err?.message || t('group_detail_error_add_participant'), 'error')
      }
    }

    async function setGroupPhoto() {
      const url = prompt(t('group_detail_prompt_photo_url'))
      if (!url) return

      try {
        await api.put('/api/groups/photo', { token, groupId: groupid, image_url: url })
        pushToast(t('group_detail_photo_updated'), 'success')
        await load()
      } catch (err: any) {
        pushToast(err?.response?.data?.result || err?.message || t('group_detail_error_photo'), 'error')
      }
    }

    async function getInvite() {
      try {
        const res = await api.post('/api/groups/invite', { token, groupId: groupid })
        const url = res.data?.url

        if (!url) {
          pushToast(t('group_detail_no_invite_link'), 'error')
          return
        }

        await navigator.clipboard.writeText(url)
        pushToast(t('group_detail_invite_copied'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || err?.message || t('group_detail_error_invite'), 'error')
      }
    }

    onMounted(() => {
      load()
    })

    return {
      addParticipant,
      error,
      filteredParticipants,
      formatDate,
      formatPhone,
      formatTime,
      getInvite,
      group,
      groupPicture,
      hasMoreMessages,
      isAdmin,
      isLoadingMore,
      leaveGroup,
      loadMoreMessages,
      loading,
      messagePreview,
      messages,
      messagesListRef,
      onMessagesScroll,
      participantPictures,
      participantSearch,
      searchParticipants,
      setGroupName,
      setGroupPhoto,
      setGroupTopic,
      showSearch,
      token,
      totalMessages,
      t,
      visibleMessages,
    }
  },
})
</script>

<style scoped>
.group-detail-page {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 24px;
}

.group-layout {
  display: grid;
  grid-template-columns: 1.35fr 1fr;
  gap: 20px;
  align-items: start;
}

.messages-column {
  background: #0f172a;
  color: #e5e7eb;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
  min-height: 520px;
}

.messages-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 12px;
}

.messages-header h2 {
  margin: 0;
  font-size: 1.4rem;
  letter-spacing: 0.01em;
}

.messages-header small {
  color: #cbd5e1;
}

.eyebrow {
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-size: 0.7rem;
  color: #94a3b8;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: calc(100vh - 220px);
  overflow-y: auto;
  padding-right: 6px;
}

.message-item {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 10px;
  padding: 12px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.06), rgba(255, 255, 255, 0.02));
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.message-avatar img,
.avatar-placeholder-small {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  object-fit: cover;
  background: #1f2937;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow: hidden;
}

.message-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: #cbd5e1;
}

.message-sender {
  font-weight: 600;
  color: #fff;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.message-time {
  font-size: 0.78rem;
  color: #94a3b8;
}

.message-text {
  color: #e2e8f0;
  line-height: 1.5;
  word-break: break-word;
}

.load-more-state {
  text-align: center;
  padding: 8px;
  color: #cbd5e1;
}

.load-more-state.muted {
  color: #94a3b8;
}

.load-more-btn {
  background: #e5e7eb;
  color: #0f172a;
  border: none;
  padding: 8px 14px;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 600;
}

.load-more-btn:hover {
  background: #cbd5e1;
}

.empty-messages {
  color: #94a3b8;
  text-align: center;
  padding: 20px 0;
}

@media (max-width: 1024px) {
  .group-layout {
    grid-template-columns: 1fr;
  }

  .messages-column {
    order: 2;
  }

  .details-column {
    order: 1;
  }
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 20px;
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

.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 20px;
  color: #dc3545;
  text-align: center;
}

.error-container i {
  font-size: 3rem;
  margin-bottom: 15px;
}

.btn-back {
  margin-top: 20px;
  padding: 10px 20px;
  background: #00034b;
  color: white;
  border-radius: 8px;
  text-decoration: none;
}

.group-header {
  text-align: center;
  padding: 20px 0 30px;
  position: relative;
}

.back-btn {
  position: absolute;
  left: 0;
  top: 20px;
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

.group-photo-container {
  position: relative;
  display: inline-block;
  margin-bottom: 15px;
}

.group-photo {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid #e5e7eb;
}

.group-photo-placeholder {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: linear-gradient(135deg, #00034b, #000266);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 3rem;
}

.edit-photo-btn {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #00034b;
  color: white;
  border: 3px solid white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.edit-photo-btn:hover {
  background: #000266;
}

.group-name {
  margin: 0 0 5px;
  font-size: 1.5rem;
  color: #111827;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.edit-btn {
  padding: 5px 8px;
  background: transparent;
  border: none;
  color: #6b7280;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.edit-btn:hover {
  background: #e5e7eb;
  color: #00034b;
}

.group-meta {
  margin: 0;
  color: #6b7280;
  font-size: 0.9rem;
}

.section {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  color: #111827;
  margin-bottom: 12px;
}

.section-header i {
  color: #00034b;
}

.description-text {
  margin: 0;
  color: #374151;
  white-space: pre-wrap;
  line-height: 1.5;
}

.description-empty {
  margin: 0;
  color: #9ca3af;
  font-style: italic;
}

.description-meta {
  margin: 10px 0 0;
  font-size: 0.75rem;
  color: #9ca3af;
}

.actions-section {
  display: flex;
  justify-content: space-around;
  padding: 12px;
}

.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 20px;
  background: transparent;
  border: none;
  color: #00034b;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #e5e7eb;
}

.action-btn i {
  font-size: 1.2rem;
}

.action-btn span {
  font-size: 0.8rem;
}

.search-box {
  margin-bottom: 12px;
}

.search-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 0.9rem;
  outline: none;
  transition: border-color 0.2s;
}

.search-input:focus {
  border-color: #00034b;
}

.participants-list {
  max-height: 400px;
  overflow-y: auto;
}

.participant-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #f3f4f6;
}

.participant-item:last-child {
  border-bottom: none;
}

.participant-avatar img {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  object-fit: cover;
}

.participant-avatar .avatar-placeholder {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  background: #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
}

.participant-info {
  flex: 1;
  min-width: 0;
}

.participant-name {
  font-weight: 500;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.participant-phone {
  font-size: 0.8rem;
  color: #6b7280;
}

.participant-badges {
  display: flex;
  gap: 5px;
}

.badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 500;
}

.badge-owner {
  background: rgba(0, 3, 75, 0.1);
  color: #00034b;
}

.badge-admin {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}

.danger-section {
  background: #fef2f2;
  border-color: #fecaca;
}

.danger-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 12px;
  background: transparent;
  border: none;
  color: #dc2626;
  font-weight: 500;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
}

.danger-btn:hover {
  background: #fee2e2;
}

@media (max-width: 640px) {
  .group-detail-page {
    padding: 15px;
  }

  .group-photo,
  .group-photo-placeholder {
    width: 100px;
    height: 100px;
  }

  .group-photo-placeholder {
    font-size: 2.5rem;
  }

  .action-btn {
    padding: 10px 15px;
  }
}
</style>
