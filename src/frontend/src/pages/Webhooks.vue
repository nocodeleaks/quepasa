<template>
  <div class="webhooks-page">
    <!-- Header -->
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile">
        <i class="fa fa-arrow-left"></i>
        Voltar
      </button>
      <div class="header-content">
        <h1>
          <i class="fa fa-globe"></i>
          Webhooks
        </h1>
        <p v-if="data?.server?.wid">Servidor: {{ data.server.wid }}</p>
        <p v-else-if="data?.server?.token">Token: {{ truncateToken(data.server.token) }}</p>
        <p v-else>Gerencie as integrações de webhook</p>
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
      <p>Para gerenciar webhooks, acesse através de um servidor específico.</p>
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
    <div v-else-if="data" class="webhooks-content">
      <!-- Add New Form -->
      <div class="add-card">
        <div class="card-header">
          <i class="fa fa-plus-circle"></i>
          <h2>Adicionar Webhook</h2>
        </div>
        <div class="card-body">
          <form @submit.prevent="createWebhook" class="add-form">
            <div class="form-row">
              <div class="form-group flex-grow">
                <label>URL do Webhook</label>
                <input 
                  v-model="newUrl" 
                  type="url" 
                  class="form-input" 
                  placeholder="https://example.com/webhook"
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
                <input type="checkbox" v-model="newForwardInternal" />
                <span>Forward Internal</span>
              </label>
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
              <label class="checkbox-label">
                <input type="checkbox" v-model="newCalls" />
                <span>Chamadas</span>
              </label>
            </div>

            <button type="submit" class="btn-primary" :disabled="!newUrl || creating">
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
          <h2>Webhooks Ativos</h2>
          <span class="count-badge">{{ data.webhooks?.length || 0 }}</span>
        </div>
        <div class="card-body">
          <div v-if="!data.webhooks || data.webhooks.length === 0" class="empty-state">
            <i class="fa fa-inbox"></i>
            <p>Nenhum webhook configurado</p>
          </div>

          <div v-else class="webhook-list">
            <div v-for="w in data.webhooks" :key="w.url" class="webhook-item">
              <div class="item-header">
                <div class="item-info">
                  <div class="item-main">
                    <strong class="webhook-url">{{ w.url }}</strong>
                    <span v-if="w.trackid" class="track-id">Track: {{ w.trackid }}</span>
                  </div>
                  <div class="item-status">
                    <span class="status-indicator" :class="{ success: !w.failure, error: w.failure }">
                      <i :class="w.failure ? 'fa fa-times-circle' : 'fa fa-check-circle'"></i>
                      {{ w.failure ? 'Falha' : 'OK' }}
                    </span>
                  </div>
                </div>

                <div class="item-actions">
                  <button class="btn-edit" @click="startEdit(w)" title="Editar">
                    <i class="fa fa-edit"></i>
                  </button>
                  <button class="btn-delete" @click="confirmDelete(w.url)" title="Remover">
                    <i class="fa fa-trash"></i>
                  </button>
                </div>
              </div>

              <!-- Extra data display -->
              <div v-if="w.extra" class="item-extra">
                <span class="extra-label">Extra:</span>
                <code class="extra-value">{{ formatExtra(w.extra) }}</code>
              </div>

              <div class="item-flags">
                <button 
                  class="flag-btn" 
                  :class="getTriStateClass(w.forwardinternal ? 1 : -1)" 
                  @click="toggleWebhookFlag(w.url, 'webhook-forwardinternal')"
                  title="Forward Internal"
                >
                  <i class="fa fa-share"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="getTriStateClass(w.broadcasts)" 
                  @click="toggleWebhookFlag(w.url, 'webhook-broadcasts')"
                  title="Broadcasts"
                >
                  <i class="fa fa-bullhorn"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="getTriStateClass(w.groups)" 
                  @click="toggleWebhookFlag(w.url, 'webhook-groups')"
                  title="Grupos"
                >
                  <i class="fa fa-users"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="getTriStateClass(w.readreceipts)" 
                  @click="toggleWebhookFlag(w.url, 'webhook-readreceipts')"
                  title="Confirmações de Leitura"
                >
                  <i class="fa fa-check-double"></i>
                </button>
                <button 
                  class="flag-btn" 
                  :class="getTriStateClass(w.calls)" 
                  @click="toggleWebhookFlag(w.url, 'webhook-calls')"
                  title="Chamadas"
                >
                  <i class="fa fa-phone"></i>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Confirm Modal -->
    <ConfirmModal 
      :show="showConfirm" 
      title="Remover Webhook" 
      message="Tem certeza que deseja remover este webhook?" 
      @confirm="doDeleteWebhook" 
      @cancel="showConfirm = false" 
      confirmLabel="Remover" 
      cancelLabel="Cancelar" 
    />

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal-content edit-modal">
        <div class="modal-header">
          <h3><i class="fa fa-edit"></i> Editar Webhook</h3>
          <button class="modal-close" @click="showEditModal = false">
            <i class="fa fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="saveEdit">
            <div class="form-group">
              <label>URL do Webhook</label>
              <input 
                v-model="editData.url" 
                type="url" 
                class="form-input" 
                placeholder="https://example.com/webhook"
                required
              />
            </div>

            <div class="form-group">
              <label>Track ID</label>
              <input 
                v-model="editData.trackId" 
                type="text" 
                class="form-input" 
                placeholder="Identificador para evitar loop"
              />
            </div>

            <div class="form-group">
              <label>Extra (JSON)</label>
              <textarea 
                v-model="editData.extraStr" 
                class="form-input extra-textarea" 
                placeholder='{"chave": "valor"}'
                rows="4"
              ></textarea>
              <small class="form-hint">Dados JSON extras que serão enviados junto com o payload</small>
            </div>

            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="editData.forwardInternal" />
                <span>Forward Internal</span>
              </label>
            </div>

            <div class="form-group options-grid">
              <div class="option-item">
                <label>Broadcasts</label>
                <TriStateToggle v-model="editData.broadcasts" />
              </div>
              <div class="option-item">
                <label>Grupos</label>
                <TriStateToggle v-model="editData.groups" />
              </div>
              <div class="option-item">
                <label>Confirmações</label>
                <TriStateToggle v-model="editData.readReceipts" />
              </div>
              <div class="option-item">
                <label>Chamadas</label>
                <TriStateToggle v-model="editData.calls" />
              </div>
            </div>

            <div class="modal-actions">
              <button type="button" class="btn-secondary" @click="showEditModal = false">
                Cancelar
              </button>
              <button type="submit" class="btn-primary" :disabled="saving">
                <i v-if="saving" class="fa fa-spinner fa-spin"></i>
                <i v-else class="fa fa-save"></i>
                {{ saving ? 'Salvando...' : 'Salvar' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, reactive } from 'vue'
import api from '@/services/api'
import { useRoute } from 'vue-router'
import ConfirmModal from '@/components/ConfirmModal.vue'
import TriStateToggle from '@/components/TriStateToggle.vue'
import { pushToast } from '@/services/toast'

export default defineComponent({
  components: { ConfirmModal, TriStateToggle },
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

    // Tri-state helper: returns CSS class based on value (-1, 0, 1)
    function getTriStateClass(val: number | boolean | null | undefined): string {
      if (val === 1 || val === true) return 'state-on'
      if (val === -1 || val === false) return 'state-off'
      return 'state-unset'
    }

    // Format extra field for display
    function formatExtra(extra: any): string {
      if (!extra) return ''
      if (typeof extra === 'string') return extra
      try {
        return JSON.stringify(extra, null, 2)
      } catch {
        return String(extra)
      }
    }

    // Convert API value to tri-state number
    function toTriState(val: any): number {
      if (val === 1 || val === true) return 1
      if (val === -1 || val === false) return -1
      return 0
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
        const res = await api.get('/api/webhooks', { params: { token } })
        data.value = res.data
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao carregar webhooks'
      } finally {
        loading.value = false
      }
    }

    // create webhook
    const newUrl = ref('')
    const newTrackId = ref('')
    const newForwardInternal = ref(false)
    const newBroadcasts = ref(true)
    const newGroups = ref(true)
    const newReadReceipts = ref(false)
    const newCalls = ref(false)
    const newExtra = ref('')

    async function createWebhook() {
      if (!newUrl.value) return

      creating.value = true
      try {
        const token = (route.query.token as string) || (data.value?.server?.token || '')
        if (!token) throw new Error('Token não encontrado')
        
        let extraParsed = null
        if (newExtra.value) extraParsed = JSON.parse(newExtra.value)
        
        await api.post('/api/webhooks', {
          token,
          url: newUrl.value,
          trackId: newTrackId.value,
          forwardInternal: newForwardInternal.value,
          broadcasts: newBroadcasts.value,
          groups: newGroups.value,
          readReceipts: newReadReceipts.value,
          calls: newCalls.value,
          extra: extraParsed,
        })
        
        await load()
        
        // Clear fields
        newUrl.value = ''
        newTrackId.value = ''
        newForwardInternal.value = false
        newBroadcasts.value = true
        newGroups.value = true
        newReadReceipts.value = false
        newCalls.value = false
        newExtra.value = ''
        
        pushToast('Webhook criado com sucesso', 'success')
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao criar webhook'
        error.value = msg
        pushToast(msg, 'error')
      } finally {
        creating.value = false
      }
    }

    // confirmation flow
    const showConfirm = ref(false)
    const confirmUrl = ref('')

    function confirmDelete(url: string) {
      confirmUrl.value = url
      showConfirm.value = true
    }

    async function doDeleteWebhook() {
      try {
        const token = (route.query.token as string) || (data.value?.server?.token || '')
        if (!token) throw new Error('Token não encontrado')
        
        await api.delete('/api/webhooks', { 
          params: { token }, 
          data: { url: confirmUrl.value } 
        })
        
        showConfirm.value = false
        confirmUrl.value = ''
        
        await load()
        pushToast('Webhook removido', 'success')
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao remover'
        error.value = msg
        pushToast(msg, 'error')
      }
    }

    async function toggleWebhookFlag(url: string, key: string) {
      try {
        const token = (route.query.token as string) || (data.value?.server?.token || '')
        if (!token) throw new Error('Token não encontrado')
        
        await api.post('/api/toggle', { token, key, url })
        await load()
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao alternar'
        pushToast(msg, 'error')
      }
    }

    // Edit webhook
    const showEditModal = ref(false)
    const saving = ref(false)
    const editData = reactive({
      originalUrl: '',
      url: '',
      trackId: '',
      extraStr: '',
      forwardInternal: false,
      broadcasts: 0,
      groups: 0,
      readReceipts: 0,
      calls: 0
    })

    function startEdit(webhook: any) {
      editData.originalUrl = webhook.url || ''
      editData.url = webhook.url || ''
      editData.trackId = webhook.trackid || ''
      editData.forwardInternal = webhook.forwardinternal || false
      editData.broadcasts = toTriState(webhook.broadcasts)
      editData.groups = toTriState(webhook.groups)
      editData.readReceipts = toTriState(webhook.readreceipts)
      editData.calls = toTriState(webhook.calls)
      
      // Format extra as JSON string for editing
      if (webhook.extra) {
        try {
          editData.extraStr = typeof webhook.extra === 'string' 
            ? webhook.extra 
            : JSON.stringify(webhook.extra, null, 2)
        } catch {
          editData.extraStr = ''
        }
      } else {
        editData.extraStr = ''
      }
      
      showEditModal.value = true
    }

    async function saveEdit() {
      saving.value = true
      try {
        const token = (route.query.token as string) || (data.value?.server?.token || '')
        if (!token) throw new Error('Token não encontrado')

        // Parse extra JSON
        let extraParsed = null
        if (editData.extraStr.trim()) {
          try {
            extraParsed = JSON.parse(editData.extraStr)
          } catch (e) {
            pushToast('Extra deve ser um JSON válido', 'error')
            saving.value = false
            return
          }
        }

        await api.put('/api/webhooks', {
          token,
          originalUrl: editData.originalUrl,
          url: editData.url,
          trackId: editData.trackId,
          forwardInternal: editData.forwardInternal,
          broadcasts: editData.broadcasts,
          groups: editData.groups,
          readReceipts: editData.readReceipts,
          calls: editData.calls,
          extra: extraParsed
        })

        showEditModal.value = false
        await load()
        pushToast('Webhook atualizado com sucesso', 'success')
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao atualizar webhook'
        pushToast(msg, 'error')
      } finally {
        saving.value = false
      }
    }

    onMounted(() => {
      load()
    })

    return { 
      data, error, loading, creating,
      newUrl, newTrackId, newForwardInternal, newBroadcasts, newGroups, newReadReceipts, newCalls, newExtra, 
      createWebhook, showConfirm, confirmDelete, doDeleteWebhook, toggleWebhookFlag,
      truncateToken, hasToken, getTriStateClass, formatExtra,
      showEditModal, saving, editData, startEdit, saveEdit
    }
  }
})
</script>

<style scoped>
.webhooks-page {
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

.webhooks-content {
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
  flex-wrap: wrap;
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

.webhook-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.webhook-item {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.item-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
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

.webhook-url {
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

.item-status {
  margin-top: 6px;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 8px;
}

.status-indicator.success {
  background: #dcfce7;
  color: #16a34a;
}

.status-indicator.error {
  background: #fef2f2;
  color: #dc2626;
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

/* Tri-state classes for flag buttons */
.flag-btn.state-unset {
  background: #f3f4f6;
  color: #9ca3af;
}

.flag-btn.state-off {
  background: #fee2e2;
  color: #dc2626;
}

.flag-btn.state-on {
  background: #dcfce7;
  color: #16a34a;
}

/* Item actions */
.item-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.btn-edit {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: #e0e7ff;
  border-radius: 8px;
  color: #4f46e5;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-edit:hover {
  background: #c7d2fe;
}

/* Extra data display */
.item-extra {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  padding: 8px 12px;
  background: #f3f4f6;
  border-radius: 8px;
  font-size: 12px;
}

.extra-label {
  color: #6b7280;
  font-weight: 600;
  flex-shrink: 0;
}

.extra-value {
  color: #374151;
  font-family: 'Fira Code', 'Consolas', monospace;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 100px;
  overflow-y: auto;
}

/* Modal styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content.edit-modal {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 18px;
  color: #111827;
}

.modal-header h3 i {
  color: var(--branding-primary, #7C3AED);
}

.modal-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: #f3f4f6;
  border-radius: 8px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.modal-close:hover {
  background: #e5e7eb;
  color: #374151;
}

.modal-body {
  padding: 24px;
}

.modal-body .form-group {
  margin-bottom: 16px;
}

.modal-body .form-input {
  width: 100%;
}

.extra-textarea {
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  resize: vertical;
}

.form-hint {
  color: #6b7280;
  font-size: 12px;
  margin-top: 4px;
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.option-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.option-item label {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  background: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

@media (max-width: 768px) {
  .form-row {
    flex-direction: column;
  }

  .webhook-item {
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

  .webhook-item {
    position: relative;
    padding-right: 56px;
  }
}
</style>
