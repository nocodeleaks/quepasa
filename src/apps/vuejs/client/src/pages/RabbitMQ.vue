<template>
  <div class="rabbitmq-page">
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
            <path d="M4 4h16v4H4V4zm0 6h10v4H4v-4zm0 6h16v4H4v-4zm12-6h4v4h-4v-4z"/>
          </svg>
          {{ t('rabbitmq_title') }}
        </h1>
        <p v-if="currentToken">{{ t('rabbitmq_server_label', [truncateToken(currentToken)]) }}</p>
        <p v-else>{{ t('rabbitmq_manage') }}</p>
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
      <h2>{{ t('rabbitmq_no_token_title') }}</h2>
      <p>{{ t('rabbitmq_no_token_desc') }}</p>
      <router-link to="/" class="btn-primary">
        <i class="fa fa-arrow-left"></i>
        {{ t('rabbitmq_back_to_servers') }}
      </router-link>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('rabbitmq_loading') }}</p>
    </div>

    <div v-else-if="data" class="rabbitmq-content">
      <div class="add-card">
        <div class="card-header">
          <i class="fa fa-plus-circle"></i>
          <h2>{{ t('rabbitmq_add_title') }}</h2>
        </div>
        <div class="card-body">
          <form @submit.prevent="createRabbit" class="add-form">
            <div class="form-row">
              <div class="form-group flex-grow">
                <label>{{ t('rabbitmq_connection_string_label') }}</label>
                <input
                  v-model="newConnectionString"
                  type="text"
                  class="form-input"
                  :placeholder="t('rabbitmq_connection_placeholder')"
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
                  :placeholder="t('rabbitmq_extra_placeholder')"
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
              <label class="checkbox-label">
                <input type="checkbox" v-model="newDirect" />
                <span>{{ t('direct') }}</span>
              </label>
            </div>

            <button type="submit" class="btn-primary" :disabled="!newConnectionString || creating">
              <i v-if="creating" class="fa fa-spinner fa-spin"></i>
              <i v-else class="fa fa-plus"></i>
              {{ creating ? t('rabbitmq_adding') : t('rabbitmq_add_btn') }}
            </button>
          </form>
        </div>
      </div>

      <div class="list-card">
        <div class="card-header">
          <i class="fa fa-list"></i>
          <h2>{{ t('rabbitmq_active_title') }}</h2>
          <span class="count-badge">{{ data.rabbitmq.length }}</span>
        </div>
        <div class="card-body">
          <div v-if="data.rabbitmq.length === 0" class="empty-state">
            <i class="fa fa-inbox"></i>
            <p>{{ t('rabbitmq_empty') }}</p>
          </div>

          <div v-else class="rabbitmq-list">
            <div v-for="item in data.rabbitmq" :key="item.connection_string" class="rabbitmq-item">
              <div class="item-info">
                <div class="item-main">
                  <strong class="connection-string">{{ item.connection_string }}</strong>
                  <span v-if="item.trackid" class="track-id">{{ t('webhooks_track_prefix', [item.trackid]) }}</span>
                </div>
                <div class="item-meta">
                  <span v-if="item.exchange_name"><i class="fa fa-exchange-alt"></i> {{ item.exchange_name }}</span>
                  <span v-if="item.routing_key"><i class="fa fa-stream"></i> {{ item.routing_key }}</span>
                </div>
                <div v-if="item.extra" class="item-extra">
                  <span class="extra-label">{{ t('webhooks_extra_prefix') }}</span>
                  <code class="extra-value">{{ formatExtra(item.extra) }}</code>
                </div>
              </div>

              <div class="item-flags">
                <button
                  class="flag-btn"
                  :class="getTriStateClass(item.forwardinternal ? 1 : -1)"
                  @click="toggleRabbitFlag(item, 'forwardinternal')"
                  :title="t('webhooks_forward_internal')"
                >
                  <i class="fa fa-share"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(item.broadcasts)"
                  @click="toggleRabbitFlag(item, 'broadcasts')"
                  :title="t('broadcasts')"
                >
                  <i class="fa fa-bullhorn"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(item.groups)"
                  @click="toggleRabbitFlag(item, 'groups')"
                  :title="t('webhooks_groups')"
                >
                  <i class="fa fa-users"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(item.readreceipts)"
                  @click="toggleRabbitFlag(item, 'readreceipts')"
                  :title="t('webhooks_confirmations')"
                >
                  <i class="fa fa-check-double"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(item.calls)"
                  @click="toggleRabbitFlag(item, 'calls')"
                  :title="t('webhooks_calls')"
                >
                  <i class="fa fa-phone"></i>
                </button>
                <button
                  class="flag-btn"
                  :class="getTriStateClass(item.direct)"
                  @click="toggleRabbitFlag(item, 'direct')"
                  :title="t('direct')"
                >
                  <i class="fa fa-comment"></i>
                </button>
              </div>

              <button class="btn-delete" @click="confirmDelete(item.connection_string)" :title="t('remove')">
                <i class="fa fa-trash"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ConfirmModal
      :show="showConfirm"
      :title="t('rabbitmq_confirm_delete_title')"
      :message="t('rabbitmq_confirm_delete_msg')"
      @confirm="doDelete"
      @cancel="showConfirm = false"
      :confirmLabel="t('remove')"
      :cancelLabel="t('cancel')"
    />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import ConfirmModal from '@/components/ConfirmModal.vue'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

type RabbitItem = {
  connection_string: string
  exchange_name?: string
  routing_key?: string
  trackid?: string
  forwardinternal?: boolean
  broadcasts?: number | boolean | null
  groups?: number | boolean | null
  readreceipts?: number | boolean | null
  calls?: number | boolean | null
  direct?: number | boolean | null
  extra?: unknown
}

export default defineComponent({
  components: { ConfirmModal },
  setup() {
    const { t } = useLocale()
    const route = useRoute()
    const data = ref<{ server: { token: string }; rabbitmq: RabbitItem[] } | null>(null)
    const error = ref('')
    const loading = ref(true)
    const creating = ref(false)
    const hasToken = ref(true)
    const showConfirm = ref(false)
    const confirmConnectionString = ref('')

    const newConnectionString = ref('')
    const newTrackId = ref('')
    const newForwardInternal = ref(false)
    const newBroadcasts = ref(true)
    const newGroups = ref(true)
    const newReadReceipts = ref(false)
    const newCalls = ref(false)
    const newDirect = ref(true)
    const newExtra = ref('')

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

    function parseExtra(extraText: string) {
      if (!extraText.trim()) return null
      return JSON.parse(extraText)
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

    function buildRabbitPayload(item: Partial<RabbitItem>) {
      return {
        connection_string: item.connection_string || '',
        trackid: item.trackid || '',
        forwardinternal: item.forwardinternal === true,
        broadcasts: toTriState(item.broadcasts),
        groups: toTriState(item.groups),
        readreceipts: toTriState(item.readreceipts),
        calls: toTriState(item.calls),
        direct: toTriState(item.direct),
        extra: item.extra ?? null,
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
        const res = await api.get('/api/dispatches/rabbitmq', { params: { token: currentToken.value } })
        data.value = {
          server: { token: currentToken.value },
          rabbitmq: res.data?.rabbitmq || [],
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || t('rabbitmq_error_load')
      } finally {
        loading.value = false
      }
    }

    async function upsertRabbit(payload: ReturnType<typeof buildRabbitPayload>) {
      await api.post('/api/dispatches/rabbitmq', { token: currentToken.value, ...payload })
    }

    async function createRabbit() {
      if (!newConnectionString.value || !currentToken.value) return

      creating.value = true
      error.value = ''

      try {
        await upsertRabbit({
          connection_string: newConnectionString.value,
          trackid: newTrackId.value,
          forwardinternal: newForwardInternal.value,
          broadcasts: toTriState(newBroadcasts.value),
          groups: toTriState(newGroups.value),
          readreceipts: toTriState(newReadReceipts.value),
          calls: toTriState(newCalls.value),
          extra: parseExtra(newExtra.value),
        })

        await load()

        newConnectionString.value = ''
        newTrackId.value = ''
        newForwardInternal.value = false
        newBroadcasts.value = true
        newGroups.value = true
        newReadReceipts.value = false
        newCalls.value = false
        newExtra.value = ''

        pushToast(t('rabbitmq_created'), 'success')
      } catch (err: any) {
        const message = err?.response?.data?.result || err.message || t('rabbitmq_error_create')
        error.value = message
        pushToast(message, 'error')
      } finally {
        creating.value = false
      }
    }

    function confirmDelete(connectionString: string) {
      confirmConnectionString.value = connectionString
      showConfirm.value = true
    }

    async function doDelete() {
      if (!currentToken.value) return

      try {
        await api.delete('/api/dispatches/rabbitmq', {
          data: { token: currentToken.value, connection_string: confirmConnectionString.value },
        })

        showConfirm.value = false
        confirmConnectionString.value = ''
        await load()
        pushToast(t('rabbitmq_deleted'), 'success')
      } catch (err: any) {
        const message = err?.response?.data?.result || err.message || t('rabbitmq_error_delete')
        error.value = message
        pushToast(message, 'error')
      }
    }

    async function toggleRabbitFlag(item: RabbitItem, key: 'forwardinternal' | 'broadcasts' | 'groups' | 'readreceipts' | 'calls' | 'direct') {
      try {
        const payload = buildRabbitPayload(item)

        if (key === 'forwardinternal') {
          payload.forwardinternal = !payload.forwardinternal
        } else {
          payload[key] = nextTriState(payload[key])
        }

        await upsertRabbit(payload)
        await load()
      } catch (err: any) {
        const message = err?.response?.data?.result || err.message || t('rabbitmq_error_toggle')
        pushToast(message, 'error')
      }
    }

    onMounted(() => {
      load()
    })

    return {
      t,
      confirmDelete,
      createRabbit,
      creating,
      currentToken,
      data,
      doDelete,
      error,
      formatExtra,
      getTriStateClass,
      hasToken,
      loading,
      newBroadcasts,
      newCalls,
      newConnectionString,
      newDirect,
      newExtra,
      newForwardInternal,
      newGroups,
      newReadReceipts,
      newTrackId,
      showConfirm,
      toggleRabbitFlag,
      truncateToken,
    }
  },
})
</script>

<style scoped>
.rabbitmq-page {
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

.rabbitmq-content {
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

.rabbitmq-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rabbitmq-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
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

.connection-string {
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

.item-meta {
  display: flex;
  gap: 12px;
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
}

.item-meta i {
  margin-right: 4px;
}

.item-extra {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  padding: 8px 12px;
  background: #f3f4f6;
  border-radius: 8px;
  font-size: 12px;
  margin-top: 10px;
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

@media (max-width: 768px) {
  .form-row {
    flex-direction: column;
  }

  .rabbitmq-item {
    position: relative;
    flex-direction: column;
    align-items: stretch;
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
