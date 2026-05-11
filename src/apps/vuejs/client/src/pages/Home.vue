<template>
  <div class="home-page">
    <!-- Compact page header -->
    <div class="page-header">
      <div class="page-title-row">
        <h1 class="page-title">{{ t('home_title') }}</h1>
        <div v-if="!loading && hasServers" class="metric-chips">
          <span class="mchip mchip-total">{{ servers.length }} {{ t('total') }}</span>
          <span class="mchip mchip-online">{{ connectedCount }} {{ t('connected') }}</span>
          <span class="mchip mchip-offline">{{ disconnectedCount }} {{ t('disconnected') }}</span>
        </div>
      </div>
      <button @click="createNewServer" class="btn-new" :disabled="creating">
        <div v-if="creating" class="spin-xs"></div>
        <i v-else class="fa fa-plus"></i>
        <span class="btn-new-label">{{ t('new_session') }}</span>
      </button>
    </div>

    <!-- Search + view toggle bar -->
    <div v-if="hasServers" class="search-row">
      <div class="search-field">
        <i class="fa fa-search sf-icon"></i>
        <input
          v-model="searchQuery"
          @keyup.enter="applySearch"
          class="sf-input"
          type="search"
          :placeholder="t('search_placeholder')"
        />
        <button v-if="searchQuery" class="sf-clear" @click="clearSearch"><i class="fa fa-times"></i></button>
      </div>
      <div class="view-toggle">
        <button class="vbtn" :class="{ active: viewMode === 'card' }" @click="viewMode = 'card'" :title="t('card_view')"><i class="fa fa-th-large"></i></button>
        <button class="vbtn" :class="{ active: viewMode === 'table' }" @click="viewMode = 'table'" :title="t('table_view')"><i class="fa fa-list"></i></button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="state-center">
      <div class="spin-md"></div>
    </div>

    <!-- Error -->
    <div v-if="error" class="alert-error">
      <i class="fa fa-exclamation-circle"></i>
      <span>{{ error }}</span>
      <button @click="load" class="link-btn">{{ t('error_retry') }}</button>
    </div>

    <!-- Empty state -->
    <div v-else-if="!loading && servers.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" class="empty-icon">
        <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2z"/>
      </svg>
      <p>{{ t('no_sessions_description') }}</p>
      <button @click="createNewServer" class="btn-add-lg" :disabled="creating">
        <i class="fa fa-plus"></i> {{ t('connect_whatsapp') }}
      </button>
    </div>

    <!-- No results -->
    <div v-else-if="!loading && servers.length > 0 && displayServers.length === 0" class="empty-state">
      <p>{{ t('no_results') }}</p>
      <button class="link-btn" @click="clearSearch">{{ t('no_results_hint') }}</button>
    </div>

    <!-- Card View: rectangular cards grid (default) -->
    <div v-else-if="viewMode === 'card'" class="sessions-grid">
      <div v-for="srv in displayServers" :key="srv.token" class="scard" :class="getStatusClass(srv)">
        <!-- Card top: avatar + identity -->
        <div class="scard-head">
          <div class="scard-avatar" :class="getStatusClass(srv)">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2z"/>
            </svg>
          </div>
          <div class="scard-identity">
            <div class="scard-phone">{{ formatWid(srv.wid) || t('not_connected') }}</div>
            <span class="scard-badge" :class="getStatusClass(srv)">{{ srv.state || t('unknown') }}</span>
          </div>
          <div v-if="srv.dispatch_count > 0" class="scard-dispatch" :title="t('col_dispatch')">
            <i class="fa fa-bell"></i> {{ srv.dispatch_count }}
          </div>
        </div>

        <!-- Token -->
        <div class="scard-token-row" @click="copyToken(srv.token)" :title="srv.token">
          <code class="scard-token">{{ srv.token }}</code>
          <i v-if="copiedToken === srv.token" class="fa fa-check text-success"></i>
          <i v-else class="fa fa-copy scard-copy-icon"></i>
        </div>

        <!-- Feature flags (connected only) -->
        <div v-if="isConnected(srv)" class="scard-flags">
          <button class="fltbtn" :class="{ on: srv.groups }" @click.stop="toggleGroups(srv)" :title="t('groups')" :disabled="toggling === srv.token"><i class="fa fa-users"></i></button>
          <button class="fltbtn" :class="{ on: srv.broadcasts }" @click.stop="toggleBroadcasts(srv)" :title="t('broadcasts')" :disabled="toggling === srv.token"><i class="fa fa-bullhorn"></i></button>
          <button class="fltbtn" :class="{ on: srv.read_receipts }" @click.stop="toggleReadReceipts(srv)" :title="t('read_receipts')" :disabled="toggling === srv.token"><i class="fa fa-check-double"></i></button>
          <button class="fltbtn" :class="{ on: srv.calls }" @click.stop="toggleCalls(srv)" :title="t('calls')" :disabled="toggling === srv.token"><i class="fa fa-phone"></i></button>
        </div>

        <!-- Card actions -->
        <div class="scard-actions">
          <router-link v-if="!isConnected(srv)" :to="`/server/${srv.token}/qrcode`" class="scard-btn scard-btn-connect" :title="t('connect')"><i class="fa fa-qrcode"></i> {{ t('connect') }}</router-link>
          <router-link v-if="isConnected(srv)" :to="`/server/${srv.token}/send`" class="scard-btn" :title="t('send')"><i class="fa fa-paper-plane"></i></router-link>
          <router-link :to="`/server/${srv.token}`" class="scard-btn" :title="t('open')"><i class="fa fa-eye"></i></router-link>
          <div class="dropdown">
            <button class="scard-btn scard-btn-more" type="button" data-bs-toggle="dropdown" aria-expanded="false" :disabled="toggling === srv.token">
              <i class="fa fa-ellipsis-v"></i>
            </button>
            <ul class="dropdown-menu dropdown-menu-end">
              <li><router-link :to="`/server/${srv.token}`" class="dropdown-item"><i class="fa fa-eye me-2"></i> {{ t('open') }}</router-link></li>
              <template v-if="isConnected(srv)">
                <li><router-link :to="`/server/${srv.token}/send`" class="dropdown-item"><i class="fa fa-paper-plane me-2"></i> {{ t('send_message') }}</router-link></li>
                <li><router-link :to="`/server/${srv.token}/messages`" class="dropdown-item"><i class="fa fa-inbox me-2"></i> {{ t('messages') }}</router-link></li>
                <li><router-link :to="`/server/${srv.token}/lid/send`" class="dropdown-item"><i class="fa fa-paper-plane me-2"></i> {{ t('menu_lid_send') }}</router-link></li>
                <li><router-link :to="`/server/${srv.token}/lid/mappings`" class="dropdown-item"><i class="fa fa-random me-2"></i> {{ t('menu_lid_mappings') }}</router-link></li>
                <li><hr class="dropdown-divider"></li>
                <li><button class="dropdown-item" :class="{ active: srv.devel }" @click="toggleDebug(srv)"><i class="fa fa-bug me-2"></i> {{ t('debug') }} {{ srv.devel ? t('state_on_short') : t('state_off_short') }}</button></li>
                <li><button class="dropdown-item text-warning" @click="disconnectServer(srv)"><i class="fa fa-unlink me-2"></i> {{ t('disconnect') }}</button></li>
              </template>
              <li><hr class="dropdown-divider"></li>
              <li><button class="dropdown-item" :class="{ active: srv.groups }" @click="toggleGroups(srv)"><i class="fa fa-users me-2"></i> {{ t('groups') }}</button></li>
              <li><button class="dropdown-item" :class="{ active: srv.broadcasts }" @click="toggleBroadcasts(srv)"><i class="fa fa-bullhorn me-2"></i> {{ t('broadcasts') }}</button></li>
              <li><button class="dropdown-item" :class="{ active: srv.read_receipts }" @click="toggleReadReceipts(srv)"><i class="fa fa-check-double me-2"></i> {{ t('read_receipts') }}</button></li>
              <li><button class="dropdown-item" :class="{ active: srv.calls }" @click="toggleCalls(srv)"><i class="fa fa-phone me-2"></i> {{ t('calls') }}</button></li>
              <li><hr class="dropdown-divider"></li>
              <li><router-link :to="`/dispatching?token=${srv.token}`" class="dropdown-item"><i class="fa fa-link me-2"></i> {{ t('dispatching') }}</router-link></li>
              <li><router-link :to="`/rabbitmq?token=${srv.token}`" class="dropdown-item"><i class="fa fa-database me-2"></i> {{ t('rabbitmq') }}</router-link></li>
              <li><hr class="dropdown-divider"></li>
              <li v-if="!isConnected(srv)"><router-link :to="`/server/${srv.token}/qrcode`" class="dropdown-item text-success"><i class="fa fa-qrcode me-2"></i> {{ t('connect') }}</router-link></li>
              <li v-if="srv.wid"><button class="dropdown-item" @click="toggleServer(srv)"><i :class="srv.verified ? 'fa fa-power-off me-2' : 'fa fa-play me-2'"></i> {{ srv.verified ? t('disable') : t('enable') }}</button></li>
              <li><button class="dropdown-item text-danger" @click="deleteServer(srv)"><i class="fa fa-trash me-2"></i> {{ t('remove') }}</button></li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Table View: compact session rows -->
    <div v-else class="sessions-list">
      <div v-for="srv in displayServers" :key="srv.token" class="srow" :class="getStatusClass(srv)">
        <!-- Avatar -->
        <div class="srow-avatar" :class="getStatusClass(srv)">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2z"/>
          </svg>
        </div>
        <!-- Main info -->
        <div class="srow-info">
          <div class="srow-phone">{{ formatWid(srv.wid) || t('not_connected') }}</div>
          <div class="srow-sub">
            <span class="sbadge" :class="getStatusClass(srv)">{{ srv.state || t('unknown') }}</span>
            <code class="srow-token" @click="copyToken(srv.token)" :title="srv.token">
              {{ srv.token }}<i v-if="copiedToken === srv.token" class="fa fa-check text-success ms-1"></i>
            </code>
          </div>
        </div>
        <!-- Feature flags (connected only) -->
        <div class="srow-flags" v-if="isConnected(srv)">
          <button class="fltbtn" :class="{ on: srv.groups }" @click="toggleGroups(srv)" :title="t('groups')" :disabled="toggling === srv.token"><i class="fa fa-users"></i></button>
          <button class="fltbtn" :class="{ on: srv.broadcasts }" @click="toggleBroadcasts(srv)" :title="t('broadcasts')" :disabled="toggling === srv.token"><i class="fa fa-bullhorn"></i></button>
          <button class="fltbtn" :class="{ on: srv.read_receipts }" @click="toggleReadReceipts(srv)" :title="t('read_receipts')" :disabled="toggling === srv.token"><i class="fa fa-check-double"></i></button>
          <button class="fltbtn" :class="{ on: srv.calls }" @click="toggleCalls(srv)" :title="t('calls')" :disabled="toggling === srv.token"><i class="fa fa-phone"></i></button>
        </div>
        <!-- Dispatch badge -->
        <div class="srow-dispatch" v-if="srv.dispatch_count > 0" :title="t('col_dispatch')">
          <i class="fa fa-bell"></i> {{ srv.dispatch_count }}
        </div>
        <!-- Action shortcuts -->
        <div class="srow-actions">
          <router-link v-if="!isConnected(srv)" :to="`/server/${srv.token}/qrcode`" class="srow-btn srow-btn-connect" :title="t('connect')"><i class="fa fa-qrcode"></i></router-link>
          <router-link v-if="isConnected(srv)" :to="`/server/${srv.token}/send`" class="srow-btn" :title="t('send')"><i class="fa fa-paper-plane"></i></router-link>
          <router-link :to="`/server/${srv.token}`" class="srow-btn" :title="t('open')"><i class="fa fa-eye"></i></router-link>
          <div class="dropdown">
            <button class="srow-btn srow-btn-more" type="button" data-bs-toggle="dropdown" aria-expanded="false" :disabled="toggling === srv.token">
              <i class="fa fa-ellipsis-v"></i>
            </button>
            <ul class="dropdown-menu dropdown-menu-end">
              <li><router-link :to="`/server/${srv.token}`" class="dropdown-item"><i class="fa fa-eye me-2"></i> {{ t('open') }}</router-link></li>
              <template v-if="isConnected(srv)">
                <li><router-link :to="`/server/${srv.token}/send`" class="dropdown-item"><i class="fa fa-paper-plane me-2"></i> {{ t('send_message') }}</router-link></li>
                <li><router-link :to="`/server/${srv.token}/messages`" class="dropdown-item"><i class="fa fa-inbox me-2"></i> {{ t('messages') }}</router-link></li>
                <li><router-link :to="`/server/${srv.token}/lid/send`" class="dropdown-item"><i class="fa fa-paper-plane me-2"></i> {{ t('menu_lid_send') }}</router-link></li>
                <li><router-link :to="`/server/${srv.token}/lid/mappings`" class="dropdown-item"><i class="fa fa-random me-2"></i> {{ t('menu_lid_mappings') }}</router-link></li>
                <li><hr class="dropdown-divider"></li>
                <li><button class="dropdown-item" :class="{ active: srv.devel }" @click="toggleDebug(srv)"><i class="fa fa-bug me-2"></i> {{ t('debug') }} {{ srv.devel ? t('state_on_short') : t('state_off_short') }}</button></li>
                <li><button class="dropdown-item text-warning" @click="disconnectServer(srv)"><i class="fa fa-unlink me-2"></i> {{ t('disconnect') }}</button></li>
              </template>
              <li><hr class="dropdown-divider"></li>
              <li><button class="dropdown-item" :class="{ active: srv.groups }" @click="toggleGroups(srv)"><i class="fa fa-users me-2"></i> {{ t('groups') }} {{ srv.groups ? t('state_on_short') : t('state_off_short') }}</button></li>
              <li><button class="dropdown-item" :class="{ active: srv.broadcasts }" @click="toggleBroadcasts(srv)"><i class="fa fa-bullhorn me-2"></i> {{ t('broadcasts') }} {{ srv.broadcasts ? t('state_on_short') : t('state_off_short') }}</button></li>
              <li><button class="dropdown-item" :class="{ active: srv.read_receipts }" @click="toggleReadReceipts(srv)"><i class="fa fa-check-double me-2"></i> {{ t('read_receipts') }} {{ srv.read_receipts ? t('state_on_short') : t('state_off_short') }}</button></li>
              <li><button class="dropdown-item" :class="{ active: srv.calls }" @click="toggleCalls(srv)"><i class="fa fa-phone me-2"></i> {{ t('calls') }} {{ srv.calls ? t('state_on_short') : t('state_off_short') }}</button></li>
              <li><hr class="dropdown-divider"></li>
              <li><router-link :to="`/dispatching?token=${srv.token}`" class="dropdown-item"><i class="fa fa-link me-2"></i> {{ t('dispatching') }}</router-link></li>
              <li><router-link :to="`/rabbitmq?token=${srv.token}`" class="dropdown-item"><i class="fa fa-database me-2"></i> {{ t('rabbitmq') }}</router-link></li>
              <li><hr class="dropdown-divider"></li>
              <li v-if="!isConnected(srv)"><router-link :to="`/server/${srv.token}/qrcode`" class="dropdown-item text-success"><i class="fa fa-qrcode me-2"></i> {{ t('connect') }}</router-link></li>
              <li v-if="srv.wid"><button class="dropdown-item" @click="toggleServer(srv)"><i :class="srv.verified ? 'fa fa-power-off me-2' : 'fa fa-play me-2'"></i> {{ srv.verified ? t('disable') : t('enable') }}</button></li>
              <li><button class="dropdown-item text-danger" @click="deleteServer(srv)"><i class="fa fa-trash me-2"></i> {{ t('remove') }}</button></li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div class="pager" v-if="!loading && hasServers && filteredServers.length > 0">
      <span class="pager-info">{{ ((currentPage - 1) * pageSize) + 1 }}–{{ Math.min(currentPage * pageSize, filteredServers.length) }} / {{ filteredServers.length }}</span>
      <div class="pager-nav">
        <button class="pager-btn" @click="prevPage" :disabled="currentPage <= 1"><i class="fa fa-chevron-left"></i></button>
        <span class="pager-indicator">{{ currentPage }} / {{ totalPages }}</span>
        <button class="pager-btn" @click="nextPage" :disabled="currentPage >= totalPages"><i class="fa fa-chevron-right"></i></button>
      </div>
      <select v-model="pageSize" class="pager-size">
        <option v-for="opt in pageSizeOptions" :key="opt" :value="opt">{{ opt }}</option>
      </select>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/services/api'
import { useServerLifecycleRefresh } from '@/composables/useServerLifecycleRefresh'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

export default defineComponent({
  setup() {
    const router = useRouter()
    const servers = ref<any[]>([])
    const loading = ref(true)
    const error = ref('')
    const viewMode = ref('card') // 'card' or 'table'
    const copiedToken = ref('')
    const toggling = ref('')
    const creating = ref(false)

  const { t, locale, setLocale } = useLocale()

    const connectedCount = computed(() => 
      servers.value.filter(s => s.state?.toLowerCase() === 'ready').length
    )

    const disconnectedCount = computed(() => 
      servers.value.filter(s => s.state?.toLowerCase() !== 'ready').length
    )

    // Track if we ever had servers (to keep search bar visible during empty search results)
    const allServersCount = ref(0)
    const hasServers = computed(() => allServersCount.value > 0)

    // Pagination state
    const currentPage = ref(1)
    const pageSize = ref(10)
    const pageSizeOptions = [10, 25, 50, 100]

    // Search state (debounced)
    const searchQuery = ref('')
    const debouncedQuery = ref('')
    let searchTimeout: any = null

    // Filtered servers (without pagination)
    const filteredServers = computed(() => {
      const q = debouncedQuery.value.trim().toLowerCase()
      if (!q) return servers.value
      return servers.value.filter(s => {
        const token = (s.token || '').toLowerCase()
        const wid = (s.wid || '').toLowerCase()
        const state = (s.state || '').toLowerCase()
        return token.includes(q) || wid.includes(q) || state.includes(q)
      })
    })

    // Total pages
    const totalPages = computed(() => {
      return Math.ceil(filteredServers.value.length / pageSize.value) || 1
    })

    // Paginated servers
    const displayServers = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value
      const end = start + pageSize.value
      return filteredServers.value.slice(start, end)
    })

    // Reset to page 1 when search or page size changes
    watch([debouncedQuery, pageSize], () => {
      currentPage.value = 1
    })

    // Navigation functions
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

    function applySearch() {
      // immediately apply (useful for enter key)
      debouncedQuery.value = searchQuery.value
      if (searchTimeout) {
        clearTimeout(searchTimeout)
        searchTimeout = null
      }
    }

    function clearSearch() {
      searchQuery.value = ''
      debouncedQuery.value = ''
    }

    // Watch input to debounce search
    watch(searchQuery, (val) => {
      if (searchTimeout) clearTimeout(searchTimeout)
      searchTimeout = setTimeout(() => {
        debouncedQuery.value = val
      }, 300)
    })

    // When the debounced query changes, call the server-side search endpoint
    watch(debouncedQuery, async (val) => {
      const q = (val || '').trim()
      if (q.length === 0) {
        // Empty query => reload all servers
        await load()
        return
      }

      await searchServers(q)
    })

    async function load() {
      try {
        loading.value = true
        error.value = ''
        const res = await api.get('/api/sessions')
        servers.value = res.data.servers || []
        // Update allServersCount only on full load (not search)
        allServersCount.value = servers.value.length
        // Use server-configured view mode as default
        if (res.data.serversViewMode) {
          viewMode.value = res.data.serversViewMode
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || t('error_load_sessions')
      } finally {
        loading.value = false
      }
    }

    async function searchServers(query: string) {
      try {
        loading.value = true
        error.value = ''
        const body = { query: query, page: 1, limit: 50 }
        const res = await api.post('/api/sessions/search', body)
        // Replace servers with response
        servers.value = res.data.servers || []
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || t('error_search_sessions')
      } finally {
        loading.value = false
      }
    }

    async function refreshCurrentView() {
      const query = debouncedQuery.value.trim()
      if (query) {
        await searchServers(query)
        return
      }

      await load()
    }

    function getStatusClass(srv: any) {
      const state = srv.state?.toLowerCase() || ''
      if (state === 'ready') return 'connected'
      if (state === 'connecting' || state === 'starting') return 'connecting'
      return 'disconnected'
    }

    function getConnectionClass(srv: any) {
      const conn = (srv.connection || srv.state || '').toLowerCase()
      if (conn === 'ready') return 'ready'
      if (conn === 'connecting') return 'connecting'
      if (conn === 'unverified') return 'unverified'
      return 'disconnected'
    }

    function isConnected(srv: any) {
      return srv.state?.toLowerCase() === 'ready'
    }

    // Tri-state helper: returns CSS class based on value (-1, 0, 1)
    // -1 = off (red), 0 = unset (no color), 1 = on (green)
    function getTriStateClass(val: number | boolean | null | undefined): string {
      if (val === 1 || val === true) return 'state-on'
      if (val === -1 || val === false) return 'state-off'
      return 'state-unset'
    }

    // Tri-state title helper
    function getTriStateTitle(name: string, val: number | boolean | null | undefined): string {
      if (val === 1 || val === true) return t('tristate_on', name)
      if (val === -1 || val === false) return t('tristate_off', name)
      return t('tristate_default', name)
    }

    // Token truncation is handled via CSS (class .truncated) to allow copy and selectable text
    // function truncateToken(token: string) { /* removed - use CSS */ }

    function formatUptime(seconds: number) {
      if (!seconds || seconds < 0) return '-'
      const d = Math.floor(seconds / 86400)
      const h = Math.floor((seconds % 86400) / 3600)
      const m = Math.floor((seconds % 3600) / 60)
      if (d > 0) return `${d}d ${h}h ${m}m`
      if (h > 0) return `${h}h ${m}m`
      return `${m}m`
    }

    // Format WID to show only phone number (remove session/server and @s.whatsapp.net)
    // Example: "554333749900:44@s.whatsapp.net" -> "554333749900"
    function formatWid(wid: string | null | undefined): string {
      if (!wid) return ''
      // Remove @s.whatsapp.net, @lid, @g.us, etc.
      let phone = wid.split('@')[0]
      // Remove session/server part (after colon)
      phone = phone.split(':')[0]
      return phone
    }

    async function copyToken(token: string) {
      const onSuccess = () => {
        copiedToken.value = token
        pushToast(t('token_copied'), 'success')
        setTimeout(() => copiedToken.value = '', 2000)
      }

      try {
        await navigator.clipboard.writeText(token)
        onSuccess()
      } catch {
        // Fallback for older browsers
        const textArea = document.createElement('textarea')
        textArea.value = token
        document.body.appendChild(textArea)
        textArea.select()
        document.execCommand('copy')
        document.body.removeChild(textArea)
        onSuccess()
      }
    }

    async function toggleServer(srv: any) {
      try {
        toggling.value = srv.token
        // Use srv.verified (not connection state) to decide the action:
        // verified=true means it's active → disable; verified=false means disabled → enable.
        const endpoint = srv.verified ? 'disable' : 'enable'
        await api.post(`/api/session/${endpoint}`, { token: srv.token })
        await load()
        pushToast(t('session_updated'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_update_session'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function disconnectServer(srv: any) {
      if (!confirm(t('confirm_disconnect'))) return
      try {
        toggling.value = srv.token
        await api.post('/api/session/disable', { token: srv.token })
        await load()
        pushToast(t('session_disconnected'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_disconnect'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleDebug(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/session/debug', { token: srv.token })
        await load()
        pushToast(t('debug_updated'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_update_debug'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleGroups(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/session/option', { token: srv.token, option: 'groups' })
        await load()
        pushToast(t('groups_updated'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_update_groups'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleBroadcasts(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/session/option', { token: srv.token, option: 'broadcasts' })
        await load()
        pushToast(t('broadcasts_updated'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_update_broadcasts'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleReadReceipts(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/session/option', { token: srv.token, option: 'readreceipts' })
        await load()
        pushToast(t('read_receipts_updated'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_update_read_receipts'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleCalls(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/session/option', { token: srv.token, option: 'calls' })
        await load()
        pushToast(t('calls_updated'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_update_calls'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function deleteServer(srv: any) {
      if (!confirm(t('confirm_remove'))) return
      try {
        toggling.value = srv.token
        await api.delete('/api/sessions', { data: { token: srv.token } })
        await load()
        pushToast(t('session_removed'), 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('error_remove_session'), 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function createNewServer() {
      if (creating.value) return
      creating.value = true
      try {
        const response = await api.post('/api/sessions', {})
        const createdServer = response.data?.server || response.data
        if (createdServer?.token) {
          const token = createdServer.token
          pushToast(t('session_created'), 'success')
          router.push(`/server/${encodeURIComponent(token)}`)
        } else {
          throw new Error(t('error_token_not_received'))
        }
      } catch (err: any) {
        console.error('Error creating server:', err)
        const errorMsg = err.response?.data?.message || err.message || t('error_create_session')
        pushToast(errorMsg, 'error')
      } finally {
        creating.value = false
      }
    }

    useServerLifecycleRefresh({
      onRefresh: refreshCurrentView,
      onConnectError: () => {
        // The page still works with manual refresh if websocket auth is unavailable.
      },
    })

    onMounted(() => {
      load()
    })

    return { 
      servers, loading, error, connectedCount, disconnectedCount, viewMode,
      searchQuery, displayServers, copiedToken, toggling, hasServers, creating,
      filteredServers, currentPage, pageSize, pageSizeOptions, totalPages,
      load, getStatusClass, getConnectionClass, isConnected, formatUptime, formatWid,
      copyToken, toggleServer, toggleDebug, toggleGroups, toggleBroadcasts, 
      toggleReadReceipts, toggleCalls, disconnectServer, deleteServer,
      applySearch, clearSearch, goToPage, nextPage, prevPage, createNewServer,
      getTriStateClass, getTriStateTitle,
      t, locale, setLocale
    }
  }
})
</script>

<style scoped>
/* ===== Layout ===== */
.home-page { max-width: 1100px; margin: 0 auto; }

/* ===== Page Header ===== */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.9rem;
  padding: 0.65rem 0;
}
.page-title-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
  min-width: 0;
}
.page-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
  white-space: nowrap;
}
.metric-chips {
  display: flex;
  gap: 0.35rem;
  flex-wrap: wrap;
}
.mchip {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 0.55rem;
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 600;
}
.mchip-total  { background: rgba(226, 232, 240, 0.7); color: #475569; }
.mchip-online { background: rgba(187, 247, 208, 0.7); color: #15803d; }
.mchip-offline{ background: rgba(226, 232, 240, 0.7); color: #6b7280; }

/* ===== New Session Button ===== */
.btn-new {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  height: 34px;
  padding: 0 0.8rem;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.18s;
  white-space: nowrap;
}
.btn-new:hover { opacity: 0.88; }
.btn-new:disabled { opacity: 0.5; cursor: not-allowed; }
@media (max-width: 480px) { .btn-new-label { display: none; } }

/* ===== Spinners ===== */
@keyframes spin { to { transform: rotate(360deg); } }
.spin-xs {
  width: 13px; height: 13px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
.spin-md {
  width: 32px; height: 32px;
  border: 3px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
  margin: 40px auto;
  display: block;
}
.state-center { text-align: center; }

/* ===== Alert / Error ===== */
.alert-error {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 0.9rem;
  background: rgba(254, 242, 242, 0.82);
  border: 1px solid rgba(254, 202, 202, 0.6);
  border-radius: 10px;
  color: #dc2626;
  font-size: 0.875rem;
  margin-bottom: 0.75rem;
}
.link-btn {
  background: none; border: none;
  color: var(--branding-primary, #7C3AED);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  text-decoration: underline;
  padding: 0;
  margin-left: auto;
}

/* ===== Empty State ===== */
.empty-state {
  text-align: center;
  padding: 2.5rem 1.5rem;
  background: rgba(255,255,255,0.6);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 18px;
}
.empty-icon { color: #d1d5db; margin-bottom: 0.75rem; }
.empty-state p { color: #6b7280; margin: 0 0 1rem; font-size: 0.92rem; }
.btn-add-lg {
  display: inline-flex; align-items: center; gap: 0.4rem;
  height: 38px; padding: 0 1.1rem;
  border: none; border-radius: 10px;
  background: var(--branding-primary, #7C3AED);
  color: white; font-size: 0.85rem; font-weight: 600; cursor: pointer;
}
.btn-add-lg:hover { opacity: 0.88; }
.btn-add-lg:disabled { opacity: 0.5; cursor: not-allowed; }

/* ===== Search Row ===== */
.search-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.search-field {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  height: 38px;
  padding: 0 0.6rem;
  background: rgba(255,255,255,0.72);
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 12px;
  backdrop-filter: blur(8px);
}
.sf-icon { color: #9ca3af; font-size: 0.8rem; flex-shrink: 0; }
.sf-input {
  flex: 1; min-width: 0; border: none; outline: none;
  background: transparent; font-size: 0.875rem; color: #111827;
}
.sf-clear {
  background: none; border: none; color: #9ca3af; cursor: pointer; padding: 0;
}
.view-toggle {
  display: flex;
  background: rgba(238, 242, 255, 0.72);
  border-radius: 10px;
  padding: 2px;
  gap: 2px;
}
.vbtn {
  width: 34px; height: 34px;
  border: none; background: transparent; color: #9ca3af;
  border-radius: 8px; cursor: pointer; font-size: 0.8rem;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.vbtn.active {
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
}

/* ===== Table View ===== */
.servers-table-wrapper {
  overflow: hidden;
  background: rgba(255,255,255,0.72);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.12);
  backdrop-filter: blur(10px);
}
.servers-table { width: 100%; border-collapse: collapse; }
.servers-table th {
  background: rgba(248, 250, 252, 0.7);
  padding: 9px 12px;
  text-align: left;
  font-size: 0.7rem;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-bottom: 1px solid rgba(229, 231, 235, 0.6);
}
.servers-table td {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(243, 244, 246, 0.7);
  vertical-align: middle;
  font-size: 0.875rem;
}
.servers-table tbody tr:hover { background: rgba(250, 247, 255, 0.5); }
.status-cell { text-align: center; width: 50px; }
.phone-cell { font-weight: 600; color: #111827; }
.token-cell { font-family: monospace; }
.token-code {
  cursor: pointer; padding: 3px 6px;
  background: rgba(243, 244, 246, 0.7);
  border-radius: 5px; font-size: 0.75rem;
}
.token-code:hover { background: rgba(229, 231, 235, 0.85); }
.token-code.truncated {
  display: inline-block;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}
.connection-cell { width: 110px; }
.connection-badge {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 10px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
}
.connection-badge.ready       { background: rgba(245,239,255,0.8); color: var(--branding-secondary, #5B21B6); }
.connection-badge.connecting  { background: rgba(254,243,199,0.8); color: #92400e; }
.connection-badge.unverified  { background: rgba(254,242,242,0.8); color: #dc2626; }
.connection-badge.disconnected{ background: rgba(243,244,246,0.8); color: #6b7280; }
.dispatch-cell { display: flex; align-items: center; justify-content: center; gap: 4px; }
.dispatch-count { font-weight: 600; }
.actions-cell { width: 60px; text-align: center; }
.action-dropdown-btn {
  width: 32px; height: 32px;
  display: flex; align-items: center; justify-content: center;
  border: none; background: rgba(243,244,246,0.8);
  border-radius: 8px; color: #6b7280; cursor: pointer; transition: all 0.15s;
}
.action-dropdown-btn:hover { background: var(--branding-primary, #7C3AED); color: white; }
.action-dropdown-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* ===== Dropdown Menu ===== */
.dropdown-menu {
  min-width: 190px; padding: 5px 0;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.09);
  border: 1px solid rgba(229,231,235,0.7);
  background: rgba(255,255,255,0.92);
  backdrop-filter: blur(12px);
}
.dropdown-item {
  padding: 8px 14px;
  font-size: 0.84rem;
  display: flex; align-items: center;
}
.dropdown-item:hover { background: rgba(243,244,246,0.7); }
.dropdown-item.active { background: rgba(124,58,237,0.08); color: var(--branding-primary, #7C3AED); }
.dropdown-item i { width: 18px; font-size: 0.8rem; }

/* ===== Cards Grid (default card view) ===== */
.sessions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 0.75rem;
  overflow: visible;
}
.scard {
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 12px;
  overflow: visible;
  display: flex;
  flex-direction: column;
  position: relative;
  transition: box-shadow 0.18s, transform 0.18s;
}
.scard:hover {
  box-shadow: 0 4px 18px rgba(0,0,0,0.08);
  transform: translateY(-2px);
}
.scard.connected  { border-top: 3px solid #22c55e; }
.scard.connecting { border-top: 3px solid #f59e0b; }
.scard.disconnected { border-top: 3px solid #e5e7eb; }

.scard-head {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.7rem 0.8rem 0.5rem;
}
.scard-avatar {
  flex-shrink: 0;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.scard-avatar.connected    { background: linear-gradient(135deg, #22c55e, #16a34a); }
.scard-avatar.connecting   { background: linear-gradient(135deg, #f59e0b, #d97706); }
.scard-avatar.disconnected { background: linear-gradient(135deg, #94a3b8, #64748b); }
.scard-identity {
  flex: 1;
  min-width: 0;
}
.scard-phone {
  font-size: 0.9rem;
  font-weight: 700;
  color: #1f2937;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.scard-badge {
  display: inline-block;
  font-size: 0.65rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.scard-badge.connected    { background: #dcfce7; color: #15803d; }
.scard-badge.connecting   { background: #fef3c7; color: #92400e; }
.scard-badge.disconnected { background: #f1f5f9; color: #475569; }
.scard-dispatch {
  flex-shrink: 0;
  font-size: 0.72rem;
  font-weight: 700;
  color: #7C3AED;
  background: #ede9fe;
  padding: 2px 7px;
  border-radius: 8px;
}

.scard-token-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0.8rem 0.4rem;
  cursor: pointer;
  background: #f8fafc;
  border-top: 1px solid rgba(148,163,184,0.1);
  border-bottom: 1px solid rgba(148,163,184,0.1);
}
.scard-token {
  flex: 1;
  font-size: 0.68rem;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
}
.scard-copy-icon { font-size: 0.7rem; color: #94a3b8; flex-shrink: 0; }
.scard-token-row:hover .scard-copy-icon { color: #7C3AED; }

.scard-flags {
  display: flex;
  gap: 0.3rem;
  padding: 0.4rem 0.8rem;
}

.scard-actions {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.45rem 0.7rem;
  border-top: 1px solid rgba(148,163,184,0.08);
  margin-top: auto;
  position: relative;
  overflow: visible;
}

.scard-actions .dropdown-menu {
  z-index: 1085;
}
.scard-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  height: 30px;
  min-width: 30px;
  padding: 0 0.5rem;
  border-radius: 7px;
  border: 1px solid rgba(148,163,184,0.2);
  background: transparent;
  color: #475569;
  font-size: 0.75rem;
  cursor: pointer;
  text-decoration: none;
  transition: background 0.15s, color 0.15s;
}
.scard-btn:hover { background: #f1f5f9; color: #1f2937; }
.scard-btn-connect {
  background: #dcfce7;
  border-color: #bbf7d0;
  color: #15803d;
  font-weight: 600;
}
.scard-btn-connect:hover { background: #bbf7d0; color: #166534; }
.scard-btn-more { margin-left: auto; }

/* ===== Session Rows (Table View) ===== */
.sessions-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
.srow {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.55rem 0.75rem;
  background: rgba(255,255,255,0.72);
  border: 1px solid rgba(148,163,184,0.12);
  border-left: 3px solid #e5e7eb;
  border-radius: 14px;
  backdrop-filter: blur(10px);
  transition: box-shadow 0.15s, transform 0.15s;
}
.srow:hover { box-shadow: 0 4px 14px rgba(15,23,42,0.06); transform: translateY(-1px); }
.srow.connected    { border-left-color: var(--branding-primary, #7C3AED); }
.srow.connecting   { border-left-color: #f59e0b; }
.srow.disconnected { border-left-color: #d1d5db; }

/* Avatar */
.srow-avatar {
  width: 36px; height: 36px; flex-shrink: 0;
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  color: white;
}
.srow-avatar.connected    { background: linear-gradient(135deg, var(--branding-primary,#7C3AED), var(--branding-secondary,#5B21B6)); }
.srow-avatar.connecting   { background: linear-gradient(135deg, #f59e0b, #d97706); }
.srow-avatar.disconnected { background: linear-gradient(135deg, #9ca3af, #6b7280); }

/* Info */
.srow-info { flex: 1; min-width: 0; }
.srow-phone {
  font-size: 0.9rem; font-weight: 600; color: #111827;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.srow-sub {
  display: flex; align-items: center; gap: 0.4rem; margin-top: 2px;
  flex-wrap: nowrap; overflow: hidden;
}
.sbadge {
  display: inline-block;
  padding: 1px 7px;
  border-radius: 999px;
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
  flex-shrink: 0;
}
.sbadge.connected    { background: rgba(245,239,255,0.9); color: var(--branding-secondary,#5B21B6); }
.sbadge.connecting   { background: rgba(254,243,199,0.9); color: #92400e; }
.sbadge.disconnected { background: rgba(243,244,246,0.9); color: #6b7280; }

.srow-token {
  font-family: monospace; font-size: 0.7rem; color: #9ca3af;
  cursor: pointer; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  max-width: 160px;
}
.srow-token:hover { color: #6b7280; }

/* Feature flag toggles */
.srow-flags {
  display: flex; gap: 3px; flex-shrink: 0;
}
.fltbtn {
  width: 28px; height: 28px;
  border: 1px solid rgba(229,231,235,0.8);
  border-radius: 8px;
  background: rgba(249,250,251,0.8);
  color: #d1d5db; cursor: pointer; font-size: 0.68rem;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.fltbtn:hover { background: rgba(243,244,246,0.9); color: #9ca3af; }
.fltbtn.on {
  background: rgba(220,252,231,0.85); color: #16a34a;
  border-color: rgba(134,239,172,0.5);
}
.fltbtn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Dispatch badge */
.srow-dispatch {
  display: flex; align-items: center; gap: 3px;
  font-size: 0.72rem; font-weight: 600; color: #059669;
  background: rgba(209,250,229,0.7);
  padding: 2px 7px; border-radius: 999px;
  flex-shrink: 0; white-space: nowrap;
}

/* Action buttons */
.srow-actions {
  display: flex; align-items: center; gap: 3px; flex-shrink: 0;
  position: relative;
  overflow: visible;
}

.srow-actions .dropdown-menu {
  z-index: 1085;
}
.srow-btn {
  width: 32px; height: 32px;
  border: none;
  background: rgba(243,244,246,0.8);
  border-radius: 9px;
  color: #6b7280; cursor: pointer; font-size: 0.8rem;
  display: flex; align-items: center; justify-content: center;
  text-decoration: none;
  transition: all 0.15s;
}
.srow-btn:hover { background: rgba(229,231,235,0.9); color: #374151; }
.srow-btn-connect { background: rgba(220,252,231,0.8); color: #16a34a; }
.srow-btn-connect:hover { background: rgba(187,247,208,0.9); }
.srow-btn-more:hover { background: var(--branding-primary,#7C3AED); color: white; }
.srow-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* ===== Pagination ===== */
.pager {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.75rem;
  padding: 0.5rem 0.75rem;
  background: rgba(255,255,255,0.65);
  border: 1px solid rgba(148,163,184,0.12);
  border-radius: 12px;
  font-size: 0.8rem;
  color: #6b7280;
  flex-wrap: wrap;
}
.pager-info { flex: 1; min-width: 0; }
.pager-nav { display: flex; align-items: center; gap: 0.4rem; }
.pager-btn {
  width: 30px; height: 30px;
  border: 1px solid rgba(148,163,184,0.2);
  border-radius: 8px; background: white;
  color: #374151; cursor: pointer; display: flex;
  align-items: center; justify-content: center; font-size: 0.75rem;
}
.pager-btn:hover:not(:disabled) { background: #f5f3ff; border-color: var(--branding-primary,#7C3AED); }
.pager-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.pager-indicator { min-width: 52px; text-align: center; font-size: 0.78rem; color: #6b7280; }
.pager-size {
  padding: 4px 8px;
  border: 1px solid rgba(148,163,184,0.2);
  border-radius: 7px; background: white;
  color: #374151; font-size: 0.8rem; cursor: pointer;
}

/* ===== Utilities ===== */
.text-success { color: #16a34a; }
.text-warning { color: #f59e0b; }
.text-danger  { color: #dc2626; }
.ms-1 { margin-left: 4px; }
.me-2 { margin-right: 6px; }
.dropdown-divider { margin: 3px 0; border-color: rgba(229,231,235,0.6); }

/* ===== Responsive ===== */
@media (max-width: 680px) {
  .srow-flags { display: none; }
  .srow-token { max-width: 100px; }
  .pager { flex-direction: column; align-items: stretch; }
  .pager-nav { justify-content: center; }
}
@media (max-width: 480px) {
  .srow-dispatch { display: none; }
  .page-title { font-size: 0.95rem; }
  .servers-table-wrapper { overflow-x: auto; }
  .servers-table { min-width: 600px; }
}
</style>