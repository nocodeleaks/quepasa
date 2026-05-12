<template>
  <div class="webhooks-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
        </svg>
        {{ t('back') }}
      </button>
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M12 2a10 10 0 100 20 10 10 0 000-20zm0 18a8 8 0 110-16 8 8 0 010 16zm-1-5h2v2h-2zm0-8h2v6h-2z"/>
          </svg>
          {{ t('webhooks_title') }}
        </h1>
        <p v-if="currentToken">{{ t('webhooks_server_label', [truncateToken(currentToken)]) }}</p>
        <p v-else>{{ t('webhooks_manage') }}</p>
      </div>
    </div>

    <div v-if="error" class="error-box">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <div v-else-if="!hasToken && !loading" class="no-token-warning">
      <div class="warning-icon">
        <i class="fa fa-exclamation-circle"></i>
      </div>
      <h2>{{ t('webhooks_no_token_title') }}</h2>
      <p>{{ t('webhooks_no_token_desc') }}</p>
      <router-link to="/" class="btn-primary">
        <i class="fa fa-arrow-left"></i>
        {{ t('webhooks_back_to_servers') }}
      </router-link>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('webhooks_loading') }}</p>
    </div>

    <div v-else-if="data" class="webhooks-content">
      <div class="add-card">
        <div class="card-header">
          <i class="fa fa-plus-circle"></i>
          <h2>{{ t('webhooks_add_title') }}</h2>
        </div>
        <div class="card-body">
          <form @submit.prevent="createWebhook" class="add-form">
            <div class="form-row">
              <div class="form-group flex-grow">
                <label>{{ t('webhooks_url_label') }}</label>
                <input
                  v-model="newUrl"
                  type="url"
                  class="form-input"
                  :placeholder="t('webhooks_url_placeholder')"
                  required
                />
              </div>
              <div class="form-group">
                <label>{{ t('webhooks_track_id_label') }}</label>
                <input
                  v-model="newTrackId"
                  type="text"
                  class="form-input"
                  :placeholder="t('optional')"
                />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group flex-grow">
                <label>{{ t('webhooks_extra_label') }}</label>
                <textarea
                  v-model="newExtra"
                  class="form-input extra-textarea"
                  rows="3"
                  :placeholder="t('webhooks_extra_placeholder')"
                ></textarea>
              </div>
            </div>

            <div class="form-row options-row">
              <label class="checkbox-label">
                <input type="checkbox" v-model="newForwardInternal" />
                <span>{{ t('webhooks_forward_internal') }}</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="newBroadcasts" />
                <span>{{ t('broadcasts') }}</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="newGroups" />
                <span>{{ t('webhooks_groups') }}</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="newReadReceipts" />
                <span>{{ t('webhooks_confirmations') }}</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="newCalls" />
                <span>{{ t('webhooks_calls') }}</span>
              </label>
            </div>

            <button type="submit" class="btn-primary" :disabled="!newUrl || creating">
              <i v-if="creating" class="fa fa-spinner fa-spin"></i>
              <i v-else class="fa fa-plus"></i>
              {{ creating ? t('webhooks_adding') : t('webhooks_add_btn') }}
            </button>
          </form>
        </div>
      </div>

      <div class="list-card">
        <div class="card-header">
          <i class="fa fa-list"></i>
          <h2>{{ t('webhooks_active_title') }}</h2>
          <span class="count-badge">{{ data.webhooks.length }}</span>
        </div>
        <div class="card-body">
          <div v-if="data.webhooks.length === 0" class="empty-state">
            <i class="fa fa-inbox"></i>
            <p>{{ t('webhooks_empty') }}</p>
          </div>

          <div v-else class="webhook-list">
            <div v-for="webhook in data.webhooks" :key="webhook.url" class="webhook-item">
              <div class="item-header">
                <div class="item-info">
                  <div class="item-main">
                    <strong class="webhook-url">{{ webhook.url }}</strong>
                    <span v-if="webhook.trackid" class="track-id">{{ t('webhooks_track_prefix', [webhook.trackid]) }}</span>
                  </div>
                  <div class="item-status">
                    <span class="status-indicator" :class="{ success: !webhook.failure, error: !!webhook.failure }">
                      <i :class="webhook.failure ? 'fa fa-times-circle' : 'fa fa-check-circle'"></i>
                      {{ webhook.failure ? t('webhooks_failure') : t('webhooks_ok') }}
                    </span>
                  </div>
                </div>

                <div class="item-actions">
                  <button class="btn-edit" @click="startEdit(webhook)" :title="t('edit')">
                    <i class="fa fa-edit"></i>
                  </button>
                  <button class="btn-delete" @click="confirmDelete(webhook.url)" :title="t('remove')">
                    <i class="fa fa-trash"></i>
                  </button>
                </div>
              </div>

              <div v-if="webhook.extra" class="item-extra">
                <span class="extra-label">{{ t('webhooks_extra_prefix') }}</span>
                <code class="extra-value">{{ formatExtra(webhook.extra) }}</code>
              </div>

              <div class="item-flags">
                <button
                  class="flag-btn"
                  :class="getTriStateClass(webhook.forwardinternal ? 1 : -1)"
                  @click="toggleWebhookFlag(webhook, 'forwardinternal')"
                  :title="t('webhooks_forward_internal')"
                >
                  <i class="fa fa-share"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(webhook.broadcasts)"
                  @click="toggleWebhookFlag(webhook, 'broadcasts')"
                  :title="t('broadcasts')"
                >
                  <i class="fa fa-bullhorn"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(webhook.groups)"
                  @click="toggleWebhookFlag(webhook, 'groups')"
                  :title="t('webhooks_groups')"
                >
                  <i class="fa fa-users"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(webhook.readreceipts)"
                  @click="toggleWebhookFlag(webhook, 'readreceipts')"
                  :title="t('webhooks_confirmations')"
                >
                  <i class="fa fa-check-double"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(webhook.calls)"
                  @click="toggleWebhookFlag(webhook, 'calls')"
                  :title="t('webhooks_calls')"
                >
                  <i class="fa fa-phone"></i>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ConfirmModal
      :show="showConfirm"
      :title="t('webhooks_confirm_delete_title')"
      :message="t('webhooks_confirm_delete_msg')"
      @confirm="doDeleteWebhook"
      @cancel="showConfirm = false"
      :confirmLabel="t('remove')"
      :cancelLabel="t('cancel')"
    />

    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal-content edit-modal">
        <div class="modal-header">
          <h3><i class="fa fa-edit"></i> {{ t('webhooks_edit_title') }}</h3>
          <button class="modal-close" @click="showEditModal = false">
            <i class="fa fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="saveEdit">
            <div class="form-group">
              <label>{{ t('webhooks_url_label') }}</label>
              <input
                v-model="editData.url"
                type="url"
                class="form-input"
                :placeholder="t('webhooks_url_placeholder')"
                required
              />
            </div>

            <div class="form-group">
              <label>{{ t('webhooks_track_id_label') }}</label>
              <input
                v-model="editData.trackid"
                type="text"
                class="form-input"
                :placeholder="t('webhooks_track_id_placeholder')"
              />
            </div>

            <div class="form-group">
              <label>{{ t('webhooks_extra_label') }}</label>
              <textarea
                v-model="editData.extraStr"
                class="form-input extra-textarea"
                rows="4"
                :placeholder="t('webhooks_extra_placeholder')"
              ></textarea>
              <small class="form-hint">{{ t('webhooks_edit_hint') }}</small>
            </div>

            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="editData.forwardinternal" />
                <span>{{ t('webhooks_forward_internal') }}</span>
              </label>
            </div>

            <div class="form-group options-grid">
              <div class="option-item">
                <label>{{ t('broadcasts') }}</label>
                <TriStateToggle v-model="editData.broadcasts" />
              </div>
              <div class="option-item">
                <label>{{ t('webhooks_groups') }}</label>
                <TriStateToggle v-model="editData.groups" />
              </div>
              <div class="option-item">
                <label>{{ t('webhooks_confirmations') }}</label>
                <TriStateToggle v-model="editData.readreceipts" />
              </div>
              <div class="option-item">
                <label>{{ t('webhooks_calls') }}</label>
                <TriStateToggle v-model="editData.calls" />
              </div>
            </div>

            <div class="modal-actions">
              <button type="button" class="btn-secondary" @click="showEditModal = false">
                {{ t('cancel') }}
              </button>
              <button type="submit" class="btn-primary" :disabled="saving">
                <i v-if="saving" class="fa fa-spinner fa-spin"></i>
                <i v-else class="fa fa-save"></i>
                {{ saving ? t('webhooks_saving') : t('webhooks_save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import ConfirmModal from '@/components/ConfirmModal.vue'
import TriStateToggle from '@/components/TriStateToggle.vue'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

type WebhookItem = {
  url: string
  trackid?: string
  forwardinternal?: boolean
  broadcasts?: number | boolean | null
  groups?: number | boolean | null
  readreceipts?: number | boolean | null
  calls?: number | boolean | null
  extra?: unknown
  failure?: string | null
}

export default defineComponent({
  components: { ConfirmModal, TriStateToggle },
  setup() {
    const { t } = useLocale()
    const route = useRoute()
    const data = ref<{ server: { token: string }; webhooks: WebhookItem[] } | null>(null)
    const error = ref('')
    const loading = ref(true)
    const creating = ref(false)
    const hasToken = ref(true)
    const showConfirm = ref(false)
    const confirmUrl = ref('')
    const showEditModal = ref(false)
    const saving = ref(false)

    const newUrl = ref('')
    const newTrackId = ref('')
    const newForwardInternal = ref(false)
    const newBroadcasts = ref(true)
    const newGroups = ref(true)
    const newReadReceipts = ref(false)
    const newCalls = ref(false)
    const newExtra = ref('')

    const editData = reactive({
      originalUrl: '',
      url: '',
      trackid: '',
      extraStr: '',
      forwardinternal: false,
      broadcasts: 0,
      groups: 0,
      readreceipts: 0,
      calls: 0,
    })

    const currentToken = computed(() => String(route.query.token || ''))

    function truncateToken(token: string) {
      if (!token) return ''
      if (token.length <= 16) return token
      return `${token.substring(0, 8)}...${token.substring(token.length - 4)}`
    }

    function getTriStateClass(value: number | boolean | null | undefined) {
      if (value === 1 || value === true) return 'state-on'
      if (value === -1 || value === false) return 'state-off'
      return 'state-unset'
    }

    function toTriState(value: any): number {
      if (value === 1 || value === true) return 1
      if (value === -1 || value === false) return -1
      return 0
    }

    function nextTriState(value: number | boolean | null | undefined): number {
      const normalized = toTriState(value)
      if (normalized === 0) return 1
      if (normalized === 1) return -1
      return 0
    }

    function formatExtra(extra: unknown): string {
      if (!extra) return ''
      if (typeof extra === 'string') return extra
      try {
        return JSON.stringify(extra, null, 2)
      } catch {
        return String(extra)
      }
    }

    function parseExtra(extraText: string) {
      if (!extraText.trim()) return null
      return JSON.parse(extraText)
    }

    function buildWebhookPayload(webhook: Partial<WebhookItem>) {
      return {
        url: webhook.url || '',
        trackid: webhook.trackid || '',
        forwardinternal: webhook.forwardinternal === true,
        broadcasts: toTriState(webhook.broadcasts),
        groups: toTriState(webhook.groups),
        readreceipts: toTriState(webhook.readreceipts),
        calls: toTriState(webhook.calls),
        extra: webhook.extra ?? null,
      }
    }

    async function load() {
      loading.value = true
      error.value = ''

      try {
        if (!currentToken.value) {
          hasToken.value = false
          data.value = null
          return
        }

        hasToken.value = true
        const res = await api.get('/api/dispatches/webhooks', { params: { token: currentToken.value } })
        data.value = {
          server: { token: currentToken.value },
          webhooks: res.data?.webhooks || [],
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || t('webhooks_error_load')
      } finally {
        loading.value = false
      }
    }

    async function upsertWebhook(payload: ReturnType<typeof buildWebhookPayload>) {
      await api.post('/api/dispatches/webhooks', { token: currentToken.value, ...payload })
    }

    async function createWebhook() {
      if (!newUrl.value || !currentToken.value) return

      creating.value = true
      error.value = ''

      try {
        await upsertWebhook({
          url: newUrl.value,
          trackid: newTrackId.value,
          forwardinternal: newForwardInternal.value,
          broadcasts: toTriState(newBroadcasts.value),
          groups: toTriState(newGroups.value),
          readreceipts: toTriState(newReadReceipts.value),
          calls: toTriState(newCalls.value),
          extra: parseExtra(newExtra.value),
        })

        await load()

        newUrl.value = ''
        newTrackId.value = ''
        newForwardInternal.value = false
        newBroadcasts.value = true
        newGroups.value = true
        newReadReceipts.value = false
        newCalls.value = false
        newExtra.value = ''

        pushToast(t('webhooks_created'), 'success')
      } catch (err: any) {
        const message = err?.response?.data?.result || err.message || t('webhooks_error_create')
        error.value = message
        pushToast(message, 'error')
      } finally {
        creating.value = false
      }
    }

    function confirmDelete(url: string) {
      confirmUrl.value = url
      showConfirm.value = true
    }

    async function doDeleteWebhook() {
      if (!currentToken.value) return

      try {
        await api.delete('/api/dispatches/webhooks', {
          data: { token: currentToken.value, url: confirmUrl.value },
        })

        showConfirm.value = false
        confirmUrl.value = ''
        await load()
        pushToast(t('webhooks_deleted'), 'success')
      } catch (err: any) {
        const message = err?.response?.data?.result || err.message || t('webhooks_error_delete')
        error.value = message
        pushToast(message, 'error')
      }
    }

    async function toggleWebhookFlag(webhook: WebhookItem, key: 'forwardinternal' | 'broadcasts' | 'groups' | 'readreceipts' | 'calls') {
      try {
        const payload = buildWebhookPayload(webhook)

        if (key === 'forwardinternal') {
          payload.forwardinternal = !payload.forwardinternal
        } else {
          payload[key] = nextTriState(payload[key])
        }

        await upsertWebhook(payload)
        await load()
      } catch (err: any) {
        const message = err?.response?.data?.result || err.message || t('webhooks_error_toggle')
        pushToast(message, 'error')
      }
    }

    function startEdit(webhook: WebhookItem) {
      editData.originalUrl = webhook.url || ''
      editData.url = webhook.url || ''
      editData.trackid = webhook.trackid || ''
      editData.forwardinternal = webhook.forwardinternal === true
      editData.broadcasts = toTriState(webhook.broadcasts)
      editData.groups = toTriState(webhook.groups)
      editData.readreceipts = toTriState(webhook.readreceipts)
      editData.calls = toTriState(webhook.calls)
      editData.extraStr = webhook.extra ? formatExtra(webhook.extra) : ''
      showEditModal.value = true
    }

    async function saveEdit() {
      if (!currentToken.value) return

      saving.value = true
      try {
        if (editData.originalUrl && editData.originalUrl !== editData.url) {
          await api.delete('/api/dispatches/webhooks', {
            data: { token: currentToken.value, url: editData.originalUrl },
          })
        }

        await upsertWebhook({
          url: editData.url,
          trackid: editData.trackid,
          forwardinternal: editData.forwardinternal,
          broadcasts: editData.broadcasts,
          groups: editData.groups,
          readreceipts: editData.readreceipts,
          calls: editData.calls,
          extra: parseExtra(editData.extraStr),
        })

        showEditModal.value = false
        await load()
        pushToast(t('webhooks_updated'), 'success')
      } catch (err: any) {
        const message = err?.response?.data?.result || err.message || t('webhooks_error_update')
        pushToast(message, 'error')
      } finally {
        saving.value = false
      }
    }

    onMounted(() => {
      load()
    })

    return {
      t,
      confirmDelete,
      createWebhook,
      creating,
      currentToken,
      data,
      doDeleteWebhook,
      editData,
      error,
      formatExtra,
      getTriStateClass,
      hasToken,
      loading,
      newBroadcasts,
      newCalls,
      newExtra,
      newForwardInternal,
      newGroups,
      newReadReceipts,
      newTrackId,
      newUrl,
      saveEdit,
      saving,
      showConfirm,
      showEditModal,
      startEdit,
      toggleWebhookFlag,
      truncateToken,
    }
  },
})
</script>

<style scoped>
.webhooks-page {
  max-width: 1000px;
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

.header-content h1 svg {
  color: var(--branding-primary, #7c3aed);
}

.header-content p {
  color: #6b7280;
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

.back-link:hover {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #312e81;
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
  background: linear-gradient(135deg, var(--branding-primary, #7c3aed), var(--branding-secondary, #5b21b6));
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
  border-top-color: var(--branding-primary, #7c3aed);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.webhooks-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.add-card,
.list-card {
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
  color: var(--branding-primary, #7c3aed);
}

.card-header h2 {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
  margin: 0;
}

.count-badge {
  margin-left: auto;
  background: var(--branding-primary, #7c3aed);
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
  border-color: var(--branding-primary, #7c3aed);
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.extra-textarea {
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  resize: vertical;
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

.checkbox-label input[type='checkbox'] {
  width: 16px;
  height: 16px;
  accent-color: var(--branding-primary, #7c3aed);
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--branding-primary, #7c3aed), var(--branding-secondary, #5b21b6));
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

.item-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.btn-delete,
.btn-edit {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-delete {
  background: #fef2f2;
  color: #dc2626;
}

.btn-delete:hover {
  background: #fecaca;
}

.btn-edit {
  background: #e0e7ff;
  color: #4f46e5;
}

.btn-edit:hover {
  background: #c7d2fe;
}

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
  color: var(--branding-primary, #7c3aed);
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
    position: relative;
    padding-right: 56px;
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
}
</style>
