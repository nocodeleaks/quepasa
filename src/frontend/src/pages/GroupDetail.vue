<template>
  <div class="group-detail-page">
    <!-- Loading -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Carregando grupo...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="error-container">
      <i class="fa fa-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <router-link :to="`/server/${token}/groups`" class="btn-back">Voltar aos grupos</router-link>
    </div>

    <!-- Content -->
    <div v-else class="group-layout">
      <!-- Messages column (left) -->
      <div class="messages-column">
        <div class="messages-header">
          <div>
            <p class="eyebrow">Histórico recente</p>
            <h2>Mensagens</h2>
          </div>
          <small v-if="totalMessages">{{ visibleMessages.length }} de {{ totalMessages }} mensagens</small>
        </div>
        <div
          class="messages-list"
          v-if="visibleMessages.length"
          ref="messagesListRef"
          @scroll.passive="onMessagesScroll"
        >
          <div v-for="m in visibleMessages" :key="m.id" class="message-item">
            <div class="message-avatar">
              <img v-if="participantPictures[m.participant?.phone || m.participant?.id]" :src="participantPictures[m.participant?.phone || m.participant?.id]" />
              <div v-else class="avatar-placeholder-small"><i class="fa fa-user"></i></div>
            </div>
            <div class="message-content">
              <div class="message-meta">
                <span class="message-sender">{{ m.participant?.title || formatPhone(m.participant?.phone || m.participant?.id) }}</span>
                <span class="message-time">{{ formatTime(m.timestamp) }}</span>
              </div>
              <div class="message-text">{{ messagePreview(m) }}</div>
            </div>
          </div>
          <div class="load-more-state" v-if="isLoadingMore">Carregando mais mensagens...</div>
          <div class="load-more-state" v-else-if="hasMoreMessages">
            <button class="load-more-btn" type="button" @click="loadMoreMessages">Carregar mais 50</button>
          </div>
          <div class="load-more-state muted" v-else>
            Você chegou ao início do histórico
          </div>
        </div>
        <div v-else class="empty-messages">Nenhuma mensagem recente</div>
      </div>

      <!-- Details column (right) -->
      <div class="details-column">
        <!-- Header with group photo -->
        <div class="group-header">
          <router-link :to="`/server/${token}/groups`" class="back-btn">
            <i class="fa fa-arrow-left"></i>
          </router-link>
          
          <div class="group-photo-container">
            <img v-if="groupPicture" :src="groupPicture" :alt="group.Name" class="group-photo" />
            <div v-else class="group-photo-placeholder">
              <i class="fa fa-users"></i>
            </div>
            <button v-if="isAdmin" @click="setGroupPhoto" class="edit-photo-btn" title="Alterar foto">
              <i class="fa fa-camera"></i>
            </button>
          </div>

          <h1 class="group-name">
            {{ group.Name || 'Grupo sem nome' }}
            <button v-if="isAdmin" @click="setGroupName" class="edit-btn" title="Alterar nome">
              <i class="fa fa-pencil"></i>
            </button>
          </h1>
          
          <p class="group-meta">
            Grupo · {{ group.Participants?.length || 0 }} membros
          </p>
        </div>

        <!-- Description -->
        <div class="section" v-if="group.Topic || isAdmin">
          <div class="section-header">
            <i class="fa fa-info-circle"></i>
            <span>Descrição</span>
            <button v-if="isAdmin" @click="setGroupTopic" class="edit-btn" title="Alterar descrição">
              <i class="fa fa-pencil"></i>
            </button>
          </div>
          <p class="description-text" v-if="group.Topic">{{ group.Topic }}</p>
          <p class="description-empty" v-else>Nenhuma descrição definida</p>
          <p class="description-meta" v-if="group.TopicSetAt">
            Criada em {{ formatDate(group.TopicSetAt) }}
          </p>
        </div>

        <!-- Quick Actions -->
        <div class="section actions-section">
          <button class="action-btn" @click="getInvite">
            <i class="fa fa-link"></i>
            <span>Link de convite</span>
          </button>
          <button v-if="isAdmin" class="action-btn" @click="addParticipant">
            <i class="fa fa-user-plus"></i>
            <span>Adicionar</span>
          </button>
          <button class="action-btn" @click="searchParticipants">
            <i class="fa fa-search"></i>
            <span>Pesquisar</span>
          </button>
        </div>

        <!-- Participants -->
        <div class="section">
          <div class="section-header">
            <i class="fa fa-users"></i>
            <span>{{ group.Participants?.length || 0 }} membros</span>
          </div>
          
          <!-- Search box -->
          <div class="search-box" v-if="showSearch">
            <input 
              v-model="participantSearch" 
              type="text" 
              placeholder="Pesquisar membros..."
              class="search-input"
            />
          </div>

          <div class="participants-list">
            <div 
              v-for="p in filteredParticipants" 
              :key="p.JID" 
              class="participant-item"
            >
              <div class="participant-avatar">
                <img 
                  v-if="participantPictures[p.PhoneNumber || p.JID]" 
                  :src="participantPictures[p.PhoneNumber || p.JID]" 
                  :alt="p.DisplayName"
                />
                <div v-else class="avatar-placeholder">
                  <i class="fa fa-user"></i>
                </div>
              </div>
              <div class="participant-info">
                <div class="participant-name">
                  <span v-if="p.DisplayName">~ {{ p.DisplayName }}</span>
                  <span v-else>{{ formatPhone(p.PhoneNumber || p.JID) }}</span>
                </div>
                <div class="participant-phone" v-if="p.DisplayName && p.PhoneNumber">
                  {{ formatPhone(p.PhoneNumber) }}
                </div>
              </div>
              <div class="participant-badges">
                <span v-if="p.IsSuperAdmin" class="badge badge-owner">Criador</span>
                <span v-else-if="p.IsAdmin" class="badge badge-admin">Admin</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Leave Group -->
        <div class="section danger-section">
          <button class="danger-btn" @click="leaveGroup">
            <i class="fa fa-sign-out"></i>
            <span>Sair do grupo</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

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

    const group = ref<any>({})
    const groupPicture = ref<string>('')
    const participantPictures = ref<Record<string, string>>({})
    const messages = ref<any[]>([])
    const visibleMessages = ref<any[]>([])
    const loading = ref(false)
    const error = ref('')
    const showSearch = ref(false)
    const participantSearch = ref('')
    const myPhone = ref<string>('')
    const messagesListRef = ref<HTMLElement | null>(null)
    const isLoadingMore = ref(false)

    const PAGE_SIZE = 50

    const totalMessages = computed(() => messages.value.length)
    const hasMoreMessages = computed(() => visibleMessages.value.length < messages.value.length)

    // Check if current user is admin
    const isAdmin = computed(() => {
      if (!group.value.Participants || !myPhone.value) return false
      const me = group.value.Participants.find((p: Participant) => 
        p.PhoneNumber?.includes(myPhone.value) || p.JID?.includes(myPhone.value)
      )
      return me?.IsAdmin || me?.IsSuperAdmin || false
    })

    // Filter participants by search
    const filteredParticipants = computed(() => {
      const participants = group.value.Participants || []
      if (!participantSearch.value) return participants
      
      const query = participantSearch.value.toLowerCase()
      return participants.filter((p: Participant) => 
        p.DisplayName?.toLowerCase().includes(query) ||
        p.PhoneNumber?.includes(query) ||
        p.JID?.includes(query)
      )
    })

    async function load() {
      loading.value = true
      error.value = ''
      messages.value = []
      visibleMessages.value = []
      try {
        // Get group info
        const res = await api.get(`/api/groups/get?token=${encodeURIComponent(token)}&groupid=${encodeURIComponent(groupid)}`)
        group.value = res.data?.groupinfo || {}

        // Get group picture
        loadGroupPicture()

        // Get server info to find my phone
        const serverRes = await api.get(`/api/server/${token}/info`)
        const wid = serverRes.data?.wid || ''
        myPhone.value = wid.replace('@s.whatsapp.net', '').replace('@lid', '')

        // Load some participant pictures (first 10 for performance)
        await loadParticipantPictures()

        // Load recent messages for this group
        await loadMessages()
      } catch (e: any) {
        error.value = e?.response?.data?.result || e?.message || 'Erro ao carregar grupo'
      } finally {
        loading.value = false
      }
    }

    async function loadGroupPicture() {
      try {
        const res = await api.get(`/api/picinfo/${encodeURIComponent(groupid)}?token=${encodeURIComponent(token)}`)
        if (res.data?.url) {
          groupPicture.value = res.data.url
        }
      } catch {
        // Silently ignore
      }
    }

    async function loadParticipantPictures() {
      const participants = (group.value.Participants || []).slice(0, 10)
      for (const p of participants) {
        const id = p.PhoneNumber || p.JID
        if (!id) continue
        try {
          const res = await api.get(`/api/picinfo/${encodeURIComponent(id)}?token=${encodeURIComponent(token)}`)
          if (res.data?.url) {
            participantPictures.value[id] = res.data.url
          }
        } catch {
          // Silently ignore
        }
      }
    }

    function formatPhone(phone: string): string {
      if (!phone) return ''
      // Remove @s.whatsapp.net or @lid
      let clean = phone.replace('@s.whatsapp.net', '').replace('@lid', '')
      // Format as +55 11 99999-9999
      if (clean.length >= 12) {
        const country = clean.slice(0, 2)
        const ddd = clean.slice(2, 4)
        const part1 = clean.slice(4, 9)
        const part2 = clean.slice(9)
        return `+${country} ${ddd} ${part1}-${part2}`
      }
      return clean
    }

    function formatDate(dateStr: string): string {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return date.toLocaleDateString('pt-BR', { 
        day: '2-digit', 
        month: '2-digit', 
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
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

    async function loadMessages() {
      try {
        const res = await api.get(`/api/server/${token}/messages`)
        const msgs = res.data?.messages || []
        const groupMsgs: any[] = []
        for (const msg of msgs) {
          const chatId = msg.chat?.id
          if (!chatId || chatId !== groupid) continue
          if (msg.type === 'unhandled' || msg.type === 'revoked' || msg.type === 'system') continue
          if (msg.debug?.reason === 'discard') continue
          // Accept if has text, attachment, or inreply
          if (!msg.text && !msg.attachment && !msg.inreply) continue

          let preview = msg.text || ''
          if (msg.attachment) {
            const mime = msg.attachment.mimetype || ''
            if (mime.startsWith('image/')) preview = preview ? `📷 ${preview}` : '📷 Imagem'
            else if (mime.startsWith('video/')) preview = preview ? `🎥 ${preview}` : '🎥 Vídeo'
            else if (mime.startsWith('audio/') || msg.type === 'ptt') preview = preview ? `🎵 ${preview}` : '🎵 Áudio'
            else if (mime.includes('pdf')) preview = preview ? `📄 ${preview}` : '📄 PDF'
            else preview = preview ? `📎 ${preview}` : '📎 Arquivo'
          }

          if (!preview && msg.inreply) preview = '↩️ Resposta'
          if (!preview) continue

          groupMsgs.push({ ...msg, text: preview })
        }
        // Sort by timestamp desc
        groupMsgs.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
        messages.value = groupMsgs
        visibleMessages.value = groupMsgs.slice(0, PAGE_SIZE)
      } catch (e) {
        // ignore
      }
    }

    function loadMoreMessages() {
      if (!hasMoreMessages.value || isLoadingMore.value) return
      isLoadingMore.value = true
      const nextSlice = messages.value.slice(visibleMessages.value.length, visibleMessages.value.length + PAGE_SIZE)
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

    function messagePreview(m: any, maxLength = 120) {
      const t = m.text || ''
      if (t.length <= maxLength) return t
      return t.substring(0, maxLength) + '...'
    }

    function searchParticipants() {
      showSearch.value = !showSearch.value
      if (!showSearch.value) {
        participantSearch.value = ''
      }
    }

    async function leaveGroup() {
      if (!confirm('Deseja realmente sair do grupo?')) return
      try {
        await api.post('/api/groups/leave', { token, group_jid: groupid })
        pushToast('Saída do grupo solicitada', 'success')
        router.push(`/server/${token}/groups`)
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao sair do grupo', 'error')
      }
    }

    async function setGroupName() {
      const name = prompt('Novo nome do grupo (<=25 caracteres):', group.value.Name || '')
      if (!name) return
      try {
        await api.put('/api/groups/name', { token, group_jid: groupid, name })
        pushToast('Nome do grupo atualizado', 'success')
        await load()
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao atualizar nome', 'error')
      }
    }

    async function setGroupTopic() {
      const topic = prompt('Nova descrição do grupo:', group.value.Topic || '')
      if (topic == null) return
      try {
        await api.put('/api/groups/description', { token, group_jid: groupid, topic })
        pushToast('Descrição atualizada', 'success')
        await load()
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao atualizar descrição', 'error')
      }
    }

    async function addParticipant() {
      const phones = prompt('Telefone(s) para adicionar (separados por vírgula):')
      if (!phones) return
      const participants = phones.split(',').map((s: string) => s.trim()).filter(Boolean)
      try {
        await api.put('/api/groups/participants', { token, group_jid: groupid, participants })
        pushToast('Participantes adicionados', 'success')
        await load()
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao adicionar participante', 'error')
      }
    }

    async function setGroupPhoto() {
      const url = prompt('URL da imagem do grupo (ou vazio para cancelar):')
      if (!url) return
      try {
        await api.put('/api/groups/photo', { token, group_jid: groupid, image_url: url })
        pushToast('Foto do grupo atualizada', 'success')
        await load()
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao alterar foto', 'error')
      }
    }

    async function getInvite() {
      try {
        const res = await api.get(`/api/invite?chatid=${encodeURIComponent(groupid)}&token=${encodeURIComponent(token)}`)
        const url = res.data?.url
        if (url) {
          // Copy to clipboard
          await navigator.clipboard.writeText(url)
          pushToast('Link copiado para a área de transferência!', 'success')
        } else {
          pushToast('Nenhum link de convite disponível', 'error')
        }
      } catch (e: any) {
        pushToast(e?.response?.data?.result || e?.message || 'Erro ao obter link', 'error')
      }
    }

    onMounted(() => { load() })

    return {
      token,
      group,
      groupPicture,
      participantPictures,
      messages,
      visibleMessages,
      loading,
      error,
      isAdmin,
      showSearch,
      participantSearch,
      totalMessages,
      hasMoreMessages,
      isLoadingMore,
      messagesListRef,
      filteredParticipants,
      formatPhone,
      formatDate,
      formatTime,
      messagePreview,
      loadMoreMessages,
      onMessagesScroll,
      searchParticipants,
      leaveGroup,
      setGroupName,
      setGroupTopic,
      setGroupPhoto,
      addParticipant,
      getInvite
    }
  }
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

/* Loading */
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
  border-top-color: #00034B;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 15px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Error */
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
  background: #00034B;
  color: white;
  border-radius: 8px;
  text-decoration: none;
}

/* Header */
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
  color: #00034B;
  text-decoration: none;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #00034B;
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
  background: linear-gradient(135deg, #00034B, #000266);
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
  background: #00034B;
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
  color: #00034B;
}

.group-meta {
  margin: 0;
  color: #6b7280;
  font-size: 0.9rem;
}

/* Sections */
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
  color: #00034B;
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

/* Actions */
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
  color: #00034B;
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

/* Search */
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
  border-color: #00034B;
}

/* Participants */
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
  color: #00034B;
}

.badge-admin {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}

/* Danger Section */
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

/* Responsive */
@media (max-width: 640px) {
  .group-detail-page {
    padding: 15px;
  }

  .group-photo, .group-photo-placeholder {
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
