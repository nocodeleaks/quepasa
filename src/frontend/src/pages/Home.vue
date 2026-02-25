<template>
  <div class="home-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1>Seus Servidores</h1>
        <p class="hide-mobile">Gerencie suas conexões WhatsApp</p>
      </div>
      <div class="header-actions">
        <!-- Desktop Controls -->
        <div class="desktop-controls" v-if="hasServers">
          <!-- Search box -->
          <div class="search-box">
            <input
              v-model="searchQuery"
              @keyup.enter="applySearch"
              class="search-input"
              type="search"
              placeholder="Search..."
              aria-label="Search servers"
              title="Search servers"
            />
            <button v-if="searchQuery" class="search-clear" @click="clearSearch" title="Clear search">
              <i class="fa fa-times"></i>
            </button>
          </div>

          <!-- View Toggle -->
          <div class="view-toggle">
            <button 
              class="view-btn" 
              :class="{ active: viewMode === 'card' }" 
              @click="viewMode = 'card'"
              title="Card view"
            >
              <i class="fa fa-th-large"></i>
            </button>
            <button 
              class="view-btn" 
              :class="{ active: viewMode === 'table' }" 
              @click="viewMode = 'table'"
              title="Table view"
            >
              <i class="fa fa-list"></i>
            </button>
          </div>
        </div>

        <!-- New Server button -->
        <button @click="createNewServer" class="btn-add" :disabled="creating">
          <template v-if="creating">
            <div class="spinner-small"></div>
          </template>
          <template v-else>
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
            </svg>
            <span class="hide-mobile">New Server</span>
          </template>
        </button>
      </div>
    </div>

    <!-- Mobile Controls Row -->
    <div class="mobile-controls" v-if="hasServers">
      <!-- Search box -->
      <div class="search-box">
        <input
          v-model="searchQuery"
          @keyup.enter="applySearch"
          class="search-input"
          type="search"
          placeholder="Search..."
          aria-label="Search servers"
          title="Search servers"
        />
        <button v-if="searchQuery" class="search-clear" @click="clearSearch" title="Clear search">
          <i class="fa fa-times"></i>
        </button>
      </div>

      <!-- View Toggle -->
      <div class="view-toggle">
        <button 
          class="view-btn" 
          :class="{ active: viewMode === 'card' }" 
          @click="viewMode = 'card'"
          title="Card view"
        >
          <i class="fa fa-th-large"></i>
        </button>
        <button 
          class="view-btn" 
          :class="{ active: viewMode === 'table' }" 
          @click="viewMode = 'table'"
          title="Table view"
        >
          <i class="fa fa-list"></i>
        </button>
      </div>

      <!-- New Server button (mobile) -->
      <button @click="createNewServer" class="btn-add-mobile" :disabled="creating">
        <template v-if="creating">
          <div class="spinner-small"></div>
        </template>
        <template v-else>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
          </svg>
        </template>
      </button>
    </div>

    <!-- Stats -->
    <div class="stats-bar" v-if="!loading && hasServers">
      <div class="stat-item">
        <span class="stat-value">{{ servers.length }}</span>
        <span class="stat-label">Total</span>
      </div>
      <div class="stat-item connected">
        <span class="stat-value">{{ connectedCount }}</span>
        <span class="stat-label">Conectados</span>
      </div>
      <div class="stat-item disconnected">
        <span class="stat-value">{{ disconnectedCount }}</span>
        <span class="stat-label">Desconectados</span>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner-large"></div>
      <p>Carregando servidores...</p>
    </div>

    <!-- Error -->
    <div v-if="error" class="error-box">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>{{ error }}</span>
      <button @click="load" class="retry-btn">Tentar novamente</button>
    </div>

    <!-- Empty state -->
    <div v-else-if="!loading && servers.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" width="80" height="80" fill="currentColor">
          <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2z"/>
        </svg>
      </div>
      <h2>Nenhum servidor configurado</h2>
      <p>Adicione seu primeiro servidor WhatsApp para começar</p>
      <button @click="createNewServer" class="btn-primary-large" :disabled="creating">
        <template v-if="creating">
          <div class="spinner-small"></div>
          <span>Criando...</span>
        </template>
        <template v-else>
          <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
          </svg>
          Conectar WhatsApp
        </template>
      </button>
    </div>

    <!-- No results (after search) -->
    <div v-else-if="!loading && servers.length > 0 && displayServers.length === 0" class="no-results">
      <h3>No servers match your search</h3>
      <p>Try a different query or clear the search.</p>
    </div>

    <!-- Table View -->
    <div v-else-if="viewMode === 'table'" class="servers-table-wrapper">
      <table class="servers-table">
        <thead>
          <tr>
            <th>Active</th>
            <th>Phone</th>
            <th>Token</th>
            <th>Dispatch</th>
            <th>Connection</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="srv in displayServers" :key="srv.token">
            <td class="status-cell">
              <i 
                :class="srv.verified ? 'fa fa-check-square text-success' : 'fa fa-exclamation-triangle text-warning'"
                :title="srv.verified ? 'Server active' : 'Server not verified'"
              ></i>
            </td>
            <td class="phone-cell">{{ formatWid(srv.wid) || '—' }}</td>
            <td class="token-cell">
              <code 
                class="token-code truncated" 
                @click="copyToken(srv.token)" 
                :title="srv.token"
              >
                {{ srv.token }}
              </code>
              <i v-if="copiedToken === srv.token" class="fa fa-check text-success ms-1"></i>
            </td>
            <td class="status-cell">
              <div class="dispatch-cell" title="Total dispatch configurations">
                <span v-if="srv.dispatch_count > 0" class="dispatch-count">{{ srv.dispatch_count }}</span>
                <span v-else class="dispatch-count">0</span>
              </div>
            </td>
            <td class="connection-cell">
              <span class="connection-badge" :class="getConnectionClass(srv)">
                {{ srv.connection || srv.state || 'Unknown' }}
              </span>
            </td>
            <td class="actions-cell">
              <div class="dropdown">
                <button 
                  class="action-dropdown-btn" 
                  type="button" 
                  data-bs-toggle="dropdown" 
                  aria-expanded="false"
                  :disabled="toggling === srv.token"
                >
                  <i class="fa fa-ellipsis-v"></i>
                </button>
                <ul class="dropdown-menu dropdown-menu-end">
                  <!-- View/Open -->
                  <li>
                    <router-link :to="`/server/${srv.token}`" class="dropdown-item">
                      <i class="fa fa-eye me-2"></i> Open
                    </router-link>
                  </li>
                  
                  <!-- Connected actions -->
                  <template v-if="isConnected(srv)">
                    <li>
                      <router-link :to="`/server/${srv.token}/send`" class="dropdown-item">
                        <i class="fa fa-paper-plane me-2"></i> Send Message
                      </router-link>
                    </li>
                    <li>
                      <router-link :to="`/server/${srv.token}/messages`" class="dropdown-item">
                        <i class="fa fa-inbox me-2"></i> Messages
                      </router-link>
                    </li>
                    <li><hr class="dropdown-divider"></li>
                    <li>
                      <button class="dropdown-item" :class="{ active: srv.devel }" @click="toggleDebug(srv)">
                        <i class="fa fa-bug me-2"></i> Debug {{ srv.devel ? '(ON)' : '(OFF)' }}
                      </button>
                    </li>
                    <li>
                      <button class="dropdown-item text-warning" @click="disconnectServer(srv)">
                        <i class="fa fa-unlink me-2"></i> Disconnect
                      </button>
                    </li>
                  </template>
                  
                  <li><hr class="dropdown-divider"></li>
                  
                  <!-- Toggle buttons -->
                  <li>
                    <button class="dropdown-item" :class="{ active: srv.groups }" @click="toggleGroups(srv)">
                      <i class="fa fa-users me-2"></i> Groups {{ srv.groups ? '(ON)' : '(OFF)' }}
                    </button>
                  </li>
                  <li>
                    <button class="dropdown-item" :class="{ active: srv.broadcasts }" @click="toggleBroadcasts(srv)">
                      <i class="fa fa-bullhorn me-2"></i> Broadcasts {{ srv.broadcasts ? '(ON)' : '(OFF)' }}
                    </button>
                  </li>
                  <li>
                    <button class="dropdown-item" :class="{ active: srv.read_receipts }" @click="toggleReadReceipts(srv)">
                      <i class="fa fa-check-double me-2"></i> Read Receipts {{ srv.read_receipts ? '(ON)' : '(OFF)' }}
                    </button>
                  </li>
                  <li>
                    <button class="dropdown-item" :class="{ active: srv.calls }" @click="toggleCalls(srv)">
                      <i class="fa fa-phone me-2"></i> Calls {{ srv.calls ? '(ON)' : '(OFF)' }}
                    </button>
                  </li>
                  
                  <li><hr class="dropdown-divider"></li>
                  
                  <!-- Dispatch -->
                  <li>
                    <router-link :to="`/dispatching?token=${srv.token}`" class="dropdown-item">
                      <i class="fa fa-link me-2"></i> Dispatching
                    </router-link>
                  </li>
                  <li>
                    <router-link :to="`/rabbitmq?token=${srv.token}`" class="dropdown-item">
                      <i class="fa fa-database me-2"></i> RabbitMQ
                    </router-link>
                  </li>
                  
                  <li><hr class="dropdown-divider"></li>
                  
                  <!-- Connect (if not connected) -->
                  <li v-if="!isConnected(srv)">
                    <router-link :to="`/server/${srv.token}/qrcode`" class="dropdown-item text-success">
                      <i class="fa fa-qrcode me-2"></i> Connect
                    </router-link>
                  </li>
                  
                  <!-- Enable/Disable -->
                  <li>
                    <button class="dropdown-item" @click="toggleServer(srv)">
                      <i :class="srv.verified ? 'fa fa-power-off me-2' : 'fa fa-play me-2'"></i>
                      {{ srv.verified ? 'Disable' : 'Enable' }}
                    </button>
                  </li>
                  
                  <!-- Delete -->
                  <li>
                    <button class="dropdown-item text-danger" @click="deleteServer(srv)">
                      <i class="fa fa-trash me-2"></i> Remove
                    </button>
                  </li>
                </ul>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Card View (default) -->
    <div v-else class="servers-grid">
      <div v-for="srv in displayServers" :key="srv.token" class="server-card" :class="getStatusClass(srv)">
        <div class="server-header">
          <div class="server-avatar" :class="getStatusClass(srv)">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
              <path d="M12.04 2c-5.46 0-9.91 4.45-9.91 9.91 0 1.75.46 3.45 1.32 4.95L2.05 22l5.25-1.38c1.45.79 3.08 1.21 4.74 1.21 5.46 0 9.91-4.45 9.91-9.91 0-2.65-1.03-5.14-2.9-7.01A9.816 9.816 0 0012.04 2z"/>
            </svg>
          </div>
          <div class="server-info">
            <h3>{{ formatWid(srv.wid) || 'Not connected' }}</h3>
            <span class="status-badge" :class="getStatusClass(srv)">{{ srv.state || 'Unknown' }}</span>
          </div>
        </div>

        <div class="server-details">
          <div class="detail-row">
            <span class="detail-label">Token:</span>
            <code class="detail-value token-code truncated" @click="copyToken(srv.token)" :title="srv.token">
              {{ srv.token }}
              <i v-if="copiedToken === srv.token" class="fa fa-check text-success ms-1"></i>
            </code>
          </div>
          <div class="detail-row" v-if="srv.uptime_seconds >= 0">
            <span class="detail-label">Uptime:</span>
            <span class="detail-value">{{ formatUptime(srv.uptime_seconds) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Dispatch:</span>
            <span class="detail-value">
              <i v-if="srv.dispatch_count > 0" class="fa fa-bell text-success" title="Dispatch count"></i>
              <span class="ms-1">{{ srv.dispatch_count || 0 }}</span>
              <i v-if="srv.webhook_count > 0" class="fa fa-link text-success ms-2" title="Dispatching: {{ srv.webhook_count }}"></i>
              <i v-if="srv.rabbitmq_count > 0" class="fa fa-database text-success ms-1" title="RabbitMQ: {{ srv.rabbitmq_count }}"></i>
            </span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Status:</span>
            <div class="flags-row">
              <span class="flag-badge" :class="getTriStateClass(srv.groups)" :title="getTriStateTitle('Groups', srv.groups)">
                <i class="fa fa-users"></i>
              </span>
              <span class="flag-badge" :class="getTriStateClass(srv.broadcasts)" :title="getTriStateTitle('Broadcasts', srv.broadcasts)">
                <i class="fa fa-bullhorn"></i>
              </span>
              <span class="flag-badge" :class="getTriStateClass(srv.readreceipts)" :title="getTriStateTitle('Read Receipts', srv.readreceipts)">
                <i class="fa fa-check-double"></i>
              </span>
              <span class="flag-badge" :class="getTriStateClass(srv.calls)" :title="getTriStateTitle('Calls', srv.calls)">
                <i class="fa fa-phone"></i>
              </span>
            </div>
          </div>
        </div>
        <!-- Quick Toggle Actions (only when connected) -->
        <div class="quick-toggles" v-if="isConnected(srv)">
          <button 
            class="toggle-btn" 
            :class="{ active: srv.groups }" 
            @click="toggleGroups(srv)" 
            title="Groups"
            :disabled="toggling === srv.token"
          >
            <i class="fa fa-users"></i>
          </button>
          <button 
            class="toggle-btn" 
            :class="{ active: srv.broadcasts }" 
            @click="toggleBroadcasts(srv)" 
            title="Broadcasts"
            :disabled="toggling === srv.token"
          >
            <i class="fa fa-bullhorn"></i>
          </button>
          <button 
            class="toggle-btn" 
            :class="{ active: srv.read_receipts }" 
            @click="toggleReadReceipts(srv)" 
            title="Read Receipts"
            :disabled="toggling === srv.token"
          >
            <i class="fa fa-check-double"></i>
          </button>
          <button 
            class="toggle-btn" 
            :class="{ active: srv.calls }" 
            @click="toggleCalls(srv)" 
            title="Calls"
            :disabled="toggling === srv.token"
          >
            <i class="fa fa-phone"></i>
          </button>
        </div>

        <div class="server-actions">
          <!-- Open (always visible) -->
          <router-link :to="`/server/${srv.token}`" class="btn-action">
            <i class="fa fa-eye"></i>
            Open
          </router-link>

          <!-- Not connected: only Connect + Remove -->
          <router-link 
            v-if="!isConnected(srv)" 
            :to="`/server/${srv.token}/qrcode`" 
            class="btn-action success"
          >
            <i class="fa fa-qrcode"></i>
            Connect
          </router-link>

          <button 
            class="btn-action danger" 
            @click="deleteServer(srv)" 
            title="Remove Server"
            :disabled="toggling === srv.token"
          >
            <i class="fa fa-trash"></i>
            Remove
          </button>

          <!-- Connected: show full actions -->
          <template v-if="isConnected(srv)">
            <router-link :to="`/server/${srv.token}/send`" class="btn-action">
              <i class="fa fa-paper-plane"></i>
              Send
            </router-link>
            <router-link :to="`/server/${srv.token}/messages`" class="btn-action">
              <i class="fa fa-inbox"></i>
              Messages
            </router-link>
            <router-link :to="`/dispatching?token=${srv.token}`" class="btn-action">
              <i class="fa fa-link"></i>
              Dispatching
            </router-link>
            <button 
              class="btn-action warning" 
              @click="disconnectServer(srv)" 
              title="Disconnect"
              :disabled="toggling === srv.token"
            >
              <i class="fa fa-unlink"></i>
              Disconnect
            </button>
            <button 
              class="btn-action" 
              :class="{ active: srv.devel }" 
              @click="toggleDebug(srv)" 
              title="Toggle Debug"
              :disabled="toggling === srv.token"
            >
              <i class="fa fa-bug"></i>
              Debug
            </button>
          </template>

          <!-- Disable Server (always visible) -->
          <button 
            class="btn-action" 
            :class="{ warning: srv.verified }" 
            @click="toggleServer(srv)" 
            :title="srv.verified ? 'Disable Server' : 'Enable Server'"
            :disabled="toggling === srv.token"
          >
            <i :class="srv.verified ? 'fa fa-power-off' : 'fa fa-play'"></i>
            {{ srv.verified ? 'Disable' : 'Enable' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div class="pagination-bar" v-if="!loading && hasServers && filteredServers.length > 0">
      <div class="pagination-info">
        Showing {{ ((currentPage - 1) * pageSize) + 1 }}-{{ Math.min(currentPage * pageSize, filteredServers.length) }} of {{ filteredServers.length }} servers
      </div>
      <div class="pagination-controls">
        <div class="page-size-selector">
          <label for="pageSize">Per page:</label>
          <select id="pageSize" v-model="pageSize">
            <option v-for="opt in pageSizeOptions" :key="opt" :value="opt">{{ opt }}</option>
          </select>
        </div>
        <div class="page-nav">
          <button class="page-btn" @click="prevPage" :disabled="currentPage <= 1" title="Previous page">
            <i class="fa fa-chevron-left"></i>
          </button>
          <span class="page-indicator">Page {{ currentPage }} of {{ totalPages }}</span>
          <button class="page-btn" @click="nextPage" :disabled="currentPage >= totalPages" title="Next page">
            <i class="fa fa-chevron-right"></i>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

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
        const res = await api.get('/api/servers')
        servers.value = res.data.servers || []
        // Update allServersCount only on full load (not search)
        allServersCount.value = servers.value.length
        // Use server-configured view mode as default
        if (res.data.serversViewMode) {
          viewMode.value = res.data.serversViewMode
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao carregar servidores'
      } finally {
        loading.value = false
      }
    }

    async function searchServers(query: string) {
      try {
        loading.value = true
        error.value = ''
        const body = { query: query, page: 1, limit: 50 }
        const res = await api.post('/api/servers/search', body)
        // Replace servers with response
        servers.value = res.data.servers || []
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Error searching servers'
      } finally {
        loading.value = false
      }
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
      if (val === 1 || val === true) return `${name}: ON (forçado)`
      if (val === -1 || val === false) return `${name}: OFF (forçado)`
      return `${name}: Padrão do sistema`
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
        pushToast('Token copiado!', 'success')
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
        const action = isConnected(srv) ? 'stop' : 'start'
        await api.post('/api/command', { token: srv.token, action })
        await load()
        pushToast('Servidor atualizado', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao alterar servidor', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function disconnectServer(srv: any) {
      if (!confirm('Deseja realmente desconectar este servidor?')) return
      try {
        toggling.value = srv.token
        await api.post('/api/command', { token: srv.token, action: 'stop' })
        await load()
        pushToast('Servidor desconectado', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao desconectar', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleDebug(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/debug', { token: srv.token })
        await load()
        pushToast('Debug atualizado', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao alterar debug', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleGroups(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/command', { token: srv.token, action: 'groups' })
        await load()
        pushToast('Grupos atualizado', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao alterar grupos', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleBroadcasts(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/command', { token: srv.token, action: 'broadcasts' })
        await load()
        pushToast('Broadcasts atualizado', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao alterar broadcasts', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleReadReceipts(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/command', { token: srv.token, action: 'readreceipts' })
        await load()
        pushToast('Confirmações de leitura atualizado', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao alterar confirmações', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function toggleCalls(srv: any) {
      try {
        toggling.value = srv.token
        await api.post('/api/command', { token: srv.token, action: 'calls' })
        await load()
        pushToast('Ligações atualizado', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao alterar ligações', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function deleteServer(srv: any) {
      if (!confirm('Deseja realmente REMOVER este servidor? Esta ação não pode ser desfeita.')) return
      try {
        toggling.value = srv.token
        await api.post('/api/delete', { token: srv.token, key: 'server' })
        await load()
        pushToast('Servidor removido', 'success')
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Erro ao remover servidor', 'error')
      } finally {
        toggling.value = ''
      }
    }

    async function createNewServer() {
      if (creating.value) return
      creating.value = true
      try {
        const response = await api.post('/api/server/create', {})
        if (response.data && response.data.token) {
          const token = response.data.token
          pushToast('Servidor criado com sucesso!', 'success')
          router.push(`/server/${token}`)
        } else {
          throw new Error('Token não recebido do servidor')
        }
      } catch (err: any) {
        console.error('Error creating server:', err)
        const errorMsg = err.response?.data?.message || err.message || 'Erro ao criar servidor'
        pushToast(errorMsg, 'error')
      } finally {
        creating.value = false
      }
    }

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
      getTriStateClass, getTriStateTitle
    }
  }
})
</script>

<style scoped>
.home-page {
  max-width: 1400px;
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

.header-content h1 {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px;
}

.header-content p {
  color: #6b7280;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.desktop-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  border: 1px solid #e5e7eb;
  padding: 6px 8px;
  border-radius: 8px;
}

.search-input {
  border: none;
  outline: none;
  min-width: 220px;
}

.search-clear {
  background: transparent;
  border: none;
  color: #9ca3af;
  cursor: pointer;
}

.no-results {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.view-toggle {
  display: flex;
  background: #f3f4f6;
  border-radius: 8px;
  overflow: hidden;
}

.view-btn {
  padding: 8px 12px;
  border: none;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.view-btn:hover {
  background: #e5e7eb;
}

.view-btn.active {
  background: var(--branding-primary, #7C3AED);
  color: white;
}

.btn-add {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border-radius: 12px;
  text-decoration: none;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-add:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(124, 58, 237, 0.25);
}

.stats-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
}

.stat-label {
  font-size: 13px;
  color: #6b7280;
}

.stat-item.connected .stat-value { color: var(--branding-primary, #7C3AED); }
.stat-item.disconnected .stat-value { color: #6b7280; }

.loading-state {
  text-align: center;
  padding: 60px 0;
}

.spinner-large {
  width: 50px;
  height: 50px;
  border: 4px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

.spinner-small {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-state p {
  color: #6b7280;
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
}

.retry-btn {
  margin-left: auto;
  padding: 8px 16px;
  background: white;
  border: 1px solid #dc2626;
  border-radius: 8px;
  color: #dc2626;
  font-weight: 600;
  cursor: pointer;
}

.retry-btn:hover {
  background: #fef2f2;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  background: white;
  border-radius: 20px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
}

.empty-icon {
  color: #d1d5db;
  margin-bottom: 20px;
}

.empty-state h2 {
  font-size: 24px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 8px;
}

.empty-state p {
  color: #6b7280;
  margin: 0 0 24px;
}

.btn-primary-large {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 16px 32px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border-radius: 14px;
  text-decoration: none;
  font-size: 18px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-primary-large:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(124, 58, 237, 0.25);
}

/* Table View Styles */
.servers-table-wrapper {
  background: white;
  border-radius: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.servers-table {
  width: 100%;
  border-collapse: collapse;
}

.servers-table th {
  background: #f9fafb;
  padding: 14px 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid #e5e7eb;
}

.servers-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f3f4f6;
  vertical-align: middle;
}

.servers-table tbody tr:hover {
  background: #f9fafb;
}

.status-cell {
  text-align: center;
  width: 60px;
}

.phone-cell {
  font-weight: 500;
  color: #111827;
}

.token-cell {
  font-family: monospace;
}

.token-code {
  cursor: pointer;
  padding: 4px 8px;
  background: #f3f4f6;
  border-radius: 6px;
  font-size: 12px;
  transition: all 0.2s;
}

.token-code:hover {
  background: #e5e7eb;
}

/* CSS-based truncation for long tokens */
.token-code.truncated {
  display: inline-block;
  max-width: 140px; /* adjust as needed */
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.connection-cell {
  width: 120px;
}

.connection-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.connection-badge.ready { background: #f5efff; color: var(--branding-secondary, #5B21B6); }
.connection-badge.connecting { background: #fef3c7; color: #92400e; }
.connection-badge.unverified { background: #fef2f2; color: #dc2626; }
.connection-badge.disconnected { background: #f3f4f6; color: #6b7280; }

.dispatch-cell { display:flex; align-items:center; justify-content:center; gap:6px; }
.dispatch-count { font-weight:600; }

.actions-cell {
  width: 80px;
  text-align: center;
}

/* Actions Dropdown */
.action-dropdown-btn {
  width: 36px;
  height: 36px;
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

.action-dropdown-btn:hover {
  background: var(--branding-primary, #7C3AED);
  color: white;
}

.action-dropdown-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dropdown-menu {
  min-width: 200px;
  padding: 8px 0;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  border: 1px solid #e5e7eb;
}

.dropdown-item {
  padding: 10px 16px;
  font-size: 14px;
  display: flex;
  align-items: center;
}

.dropdown-item:hover {
  background: #f3f4f6;
}

.dropdown-item.active {
  background: rgba(124, 58, 237, 0.1);
  color: var(--branding-primary, #7C3AED);
}

.dropdown-item i {
  width: 20px;
}

.action-buttons {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.action-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: #f3f4f6;
  border-radius: 6px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
}

.action-btn:hover {
  background: #e5e7eb;
  color: #374151;
}

.action-btn.active {
  background: var(--branding-primary, #7C3AED);
  color: white;
}

.action-btn.success {
  background: #dcfce7;
  color: #16a34a;
}

.action-btn.success:hover {
  background: #bbf7d0;
}

.action-btn.danger {
  background: #fef2f2;
  color: #dc2626;
}

.action-btn.danger:hover {
  background: #fecaca;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Card View Styles */
.servers-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.server-card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  border-left: 4px solid #e5e7eb;
  transition: all 0.2s;
}

.server-card:hover {
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
}

.server-card.connected { border-left-color: var(--branding-primary, #7C3AED); }
.server-card.connecting { border-left-color: #f59e0b; }
.server-card.disconnected { border-left-color: #9ca3af; }

.server-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
}

.server-avatar {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.server-avatar.connected { background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6)); }
.server-avatar.connecting { background: linear-gradient(135deg, #f59e0b, #d97706); }
.server-avatar.disconnected { background: linear-gradient(135deg, #9ca3af, #6b7280); }

.server-info h3 {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
  margin: 0 0 6px;
}

.status-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.connected { background: #f5efff; color: var(--branding-secondary, #5B21B6); }
.status-badge.connecting { background: #fef3c7; color: #92400e; }
.status-badge.disconnected { background: #f3f4f6; color: #6b7280; }

.server-details {
  margin-bottom: 16px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 10px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}

.detail-label {
  font-size: 13px;
  color: #6b7280;
}

.detail-value {
  font-size: 13px;
  color: #111827;
  font-weight: 500;
}

.flags-row {
  display: flex;
  gap: 4px;
}

.flag-badge {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  background: #e5e7eb;
  color: #9ca3af;
  font-size: 11px;
}

/* Tri-state: unset (0) - no color, neutral gray */
.flag-badge.state-unset {
  background: #f3f4f6;
  color: #9ca3af;
}

/* Tri-state: off (-1) - red */
.flag-badge.state-off {
  background: #fee2e2;
  color: #dc2626;
}

/* Tri-state: on (1) - green */
.flag-badge.state-on {
  background: #dcfce7;
  color: #16a34a;
}

/* Legacy active class for compatibility */
.flag-badge.active {
  background: var(--branding-primary, #7C3AED);
  color: white;
}

.server-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.btn-action {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #f3f4f6;
  border: none;
  border-radius: 8px;
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  transition: all 0.2s;
  cursor: pointer;
}

.btn-action:hover {
  background: #e5e7eb;
}

.btn-action.primary {
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
}

.btn-action.primary:hover {
  box-shadow: 0 4px 8px rgba(124, 58, 237, 0.25);
}

.btn-action.success {
  background: #dcfce7;
  color: #16a34a;
}

.btn-action.success:hover {
  background: #bbf7d0;
}

.btn-action.danger {
  background: #fef2f2;
  color: #dc2626;
}

.btn-action.danger:hover {
  background: #fecaca;
}

.btn-action.active {
  background: var(--branding-primary, #7C3AED);
  color: white;
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Quick Toggle Buttons */
.quick-toggles {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px;
  background: #f8fafc;
  border-radius: 8px;
  justify-content: center;
}

.toggle-btn {
  width: 36px;
  height: 36px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: white;
  color: #9ca3af;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.toggle-btn:hover {
  background: #f3f4f6;
  color: #6b7280;
}

.toggle-btn.active {
  background: #10b981;
  color: white;
  border-color: #10b981;
}

.toggle-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Warning button style */
.btn-action.warning {
  background: #f59e0b;
  color: white;
}

.btn-action.warning:hover {
  background: #d97706;
}

/* Utility classes */
.text-success { color: #16a34a; }
.text-warning { color: #f59e0b; }
.text-muted { color: #9ca3af; }
.ms-1 { margin-left: 4px; }

/* Pagination */
.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: var(--card-bg);
  border-radius: 8px;
  margin-top: 20px;
  gap: 16px;
  flex-wrap: wrap;
}

.pagination-info {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 24px;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-size-selector label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.page-size-selector select {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--input-bg);
  color: var(--text-primary);
  font-size: 0.9rem;
  cursor: pointer;
}

.page-nav {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--input-bg);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background: var(--hover-bg);
  border-color: var(--primary-color);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-indicator {
  color: var(--text-secondary);
  font-size: 0.9rem;
  min-width: 100px;
  text-align: center;
}

/* Responsive */
@media (max-width: 768px) {
  .hide-mobile {
    display: none !important;
  }

  .home-page {
    padding: 0;
    margin: 0;
  }

  .page-header {
    position: sticky;
    top: 0;
    z-index: 100;
    background: white;
    margin: 0;
    padding: 12px 16px;
    border-radius: 0;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .header-content h1 {
    font-size: 20px;
  }

  .mobile-controls {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: #f9fafb;
    position: sticky;
    top: 60px;
    z-index: 99;
  }

  .mobile-controls .search-box {
    flex: 1;
  }

  .mobile-controls .search-input {
    min-width: 0;
    width: 100%;
  }

  .btn-add-mobile {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    background: var(--branding-primary, #7C3AED);
    color: white;
    border-radius: 8px;
    text-decoration: none;
    flex-shrink: 0;
  }

  .servers-table-wrapper {
    overflow-x: auto;
  }
  
  .servers-table {
    min-width: 800px;
  }
  
  .servers-grid {
    grid-template-columns: 1fr;
    padding: 0 16px;
  }

  .stats-bar {
    margin: 0 16px 16px;
  }

  .pagination-bar {
    flex-direction: column;
    align-items: stretch;
    text-align: center;
    margin: 16px;
  }

  .pagination-controls {
    justify-content: center;
    flex-wrap: wrap;
  }

  .desktop-controls {
    display: none;
  }

  .btn-add span {
    display: none;
  }

  .btn-add {
    padding: 10px;
    border-radius: 8px;
  }
}

/* Desktop: hide mobile-only elements */
@media (min-width: 769px) {
  .mobile-controls {
    display: none;
  }

  .btn-add-mobile {
    display: none;
  }
}
</style>
