<script setup lang="ts">
import {
  AlertCircle,
  ArrowDown,
  ArrowUp,
  CheckCircle2,
  KeyRound,
  ListOrdered,
  LockKeyhole,
  Plus,
  RefreshCw,
  Search,
  Server,
  ShieldCheck,
  Trash2,
  Wifi
} from '@lucide/vue'
import { computed, onMounted, ref } from 'vue'

interface SpamStatus {
  configured: boolean
  unlocked: boolean
}

interface SpamSection {
  token: string
  wid?: string
  user?: string
  contextid?: string
  verified: boolean
  status: string
  ready: boolean
  inSpam: boolean
  enabled: boolean
  position?: number
  label?: string
}

const masterKey = ref(sessionStorage.getItem('quepasa.spam.masterkey') || '')
const status = ref<SpamStatus>({ configured: false, unlocked: false })
const sections = ref<SpamSection[]>([])
const searchResults = ref<SpamSection[]>([])
const search = ref('')
const loading = ref(true)
const validating = ref(false)
const searching = ref(false)
const savingToken = ref('')
const error = ref('')
const notice = ref('')

const unlocked = computed(() => status.value.configured && status.value.unlocked)
const activeSections = computed(() => sections.value.filter((item) => item.enabled).length)
const readySections = computed(() => sections.value.filter((item) => item.enabled && item.ready).length)

onMounted(async () => {
  await refreshStatus()
  if (status.value.configured && masterKey.value) {
    await unlock()
  }
  loading.value = false
})

async function refreshStatus() {
  error.value = ''
  try {
    status.value = await fetchJson<SpamStatus>('/spam/status', { allowAnonymous: true })
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function unlock() {
  validating.value = true
  error.value = ''
  try {
    await loadSections()
    sessionStorage.setItem('quepasa.spam.masterkey', masterKey.value)
    status.value.unlocked = true
  } catch (err) {
    status.value.unlocked = false
    error.value = errorMessage(err)
  } finally {
    validating.value = false
  }
}

function lock() {
  masterKey.value = ''
  status.value.unlocked = false
  sections.value = []
  searchResults.value = []
  sessionStorage.removeItem('quepasa.spam.masterkey')
}

async function loadSections() {
  const response = await fetchJson<{ items: SpamSection[] }>('/spam/sections')
  sections.value = response.items ?? []
}

async function runSearch() {
  searching.value = true
  error.value = ''
  try {
    const response = await fetchJson<{ items: SpamSection[] }>('/spam/sections/search', {
      method: 'POST',
      body: { search: search.value, limit: 100 }
    })
    searchResults.value = response.items ?? []
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    searching.value = false
  }
}

async function addSection(item: SpamSection) {
  savingToken.value = item.token
  error.value = ''
  notice.value = ''
  try {
    await fetchJson('/spam/sections', {
      method: 'POST',
      body: {
        token: item.token,
        enabled: true,
        label: item.label || item.wid || item.user || ''
      }
    })
    notice.value = 'Seção adicionada ao serviço de spam.'
    await loadSections()
    await runSearch()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    savingToken.value = ''
  }
}

async function removeSection(item: SpamSection) {
  savingToken.value = item.token
  error.value = ''
  notice.value = ''
  try {
    await fetchJson(`/spam/sections?token=${encodeURIComponent(item.token)}`, { method: 'DELETE' })
    notice.value = 'Seção removida do serviço de spam.'
    await loadSections()
    if (searchResults.value.length) await runSearch()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    savingToken.value = ''
  }
}

async function toggleSection(item: SpamSection) {
  savingToken.value = item.token
  error.value = ''
  notice.value = ''
  try {
    await fetchJson('/spam/sections', {
      method: 'PATCH',
      body: {
        token: item.token,
        enabled: !item.enabled,
        position: item.position || 0,
        label: item.label || ''
      }
    })
    notice.value = item.enabled ? 'Seção pausada.' : 'Seção ativada.'
    await loadSections()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    savingToken.value = ''
  }
}

async function moveSection(index: number, delta: number) {
  const nextIndex = index + delta
  if (nextIndex < 0 || nextIndex >= sections.value.length) return

  const ordered = [...sections.value]
  const [item] = ordered.splice(index, 1)
  ordered.splice(nextIndex, 0, item)
  sections.value = ordered

  try {
    await fetchJson('/spam/sections/reorder', {
      method: 'POST',
      body: { tokens: ordered.map((section) => section.token) }
    })
    await loadSections()
  } catch (err) {
    error.value = errorMessage(err)
    await loadSections()
  }
}

async function refreshAll() {
  await refreshStatus()
  if (unlocked.value) {
    await loadSections()
    if (search.value || searchResults.value.length) await runSearch()
  }
}

function sectionTitle(item: SpamSection) {
  return item.wid || item.token
}

function sectionSubtitle(item: SpamSection) {
  return [item.user, item.contextid].filter(Boolean).join(' · ') || 'Sem proprietário registrado'
}

function apiPath(path: string) {
  const runtimeConfig = (window as unknown as { quepasa?: { apiBase?: string } }).quepasa
  const base = (runtimeConfig?.apiBase || '/api').replace(/\/+$/, '')
  return `${base}${path}`
}

async function fetchJson<T = unknown>(
  path: string,
  options: { method?: string; body?: unknown; allowAnonymous?: boolean } = {}
): Promise<T> {
  const headers: Record<string, string> = {}
  if (!options.allowAnonymous) {
    headers['X-QUEPASA-MASTERKEY'] = masterKey.value
  }
  if (options.body !== undefined) {
    headers['Content-Type'] = 'application/json'
  }

  const response = await fetch(apiPath(path), {
    method: options.method || 'GET',
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body)
  })

  const contentType = response.headers.get('content-type') || ''
  const payload = contentType.includes('application/json') ? await response.json() : await response.text()
  if (!response.ok) {
    throw new Error(typeof payload === 'string' ? payload : payload.result || payload.status || 'Falha na requisição')
  }
  return payload as T
}

function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : 'Falha inesperada'
}
</script>

<template>
  <main class="spam-shell">
    <aside class="rail" aria-label="Estado do serviço">
      <div class="brand-lock">
        <div class="brand-mark"><ShieldCheck :size="26" /></div>
        <div>
          <strong>Spam Control</strong>
          <span>QuePasa master</span>
        </div>
      </div>

      <div class="rail-stat">
        <span>Seções na fila</span>
        <strong>{{ sections.length }}</strong>
      </div>
      <div class="rail-stat">
        <span>Ativas</span>
        <strong>{{ activeSections }}</strong>
      </div>
      <div class="rail-stat">
        <span>Prontas</span>
        <strong>{{ readySections }}</strong>
      </div>

      <button class="ghost-action" type="button" @click="refreshAll">
        <RefreshCw :size="18" />
        Atualizar
      </button>
      <button v-if="unlocked" class="ghost-action danger" type="button" @click="lock">
        <LockKeyhole :size="18" />
        Bloquear
      </button>
    </aside>

    <section class="workspace" aria-live="polite">
      <header class="topbar">
        <div>
          <p class="eyebrow">WHATSAPP</p>
          <h1>Serviço de spam</h1>
          <span>Defina quais seções o endpoint <code>/spam</code> pode usar e em qual ordem.</span>
        </div>
        <div class="state-pill" :class="{ ok: unlocked, blocked: !status.configured }">
          <CheckCircle2 v-if="unlocked" :size="18" />
          <AlertCircle v-else :size="18" />
          {{ unlocked ? 'Master liberado' : status.configured ? 'Master key necessária' : 'Master key ausente' }}
        </div>
      </header>

      <div v-if="loading" class="empty-panel">
        <RefreshCw class="spin" :size="28" />
        <strong>Carregando serviço</strong>
      </div>

      <div v-else-if="!status.configured" class="empty-panel blocked-panel">
        <LockKeyhole :size="36" />
        <strong>Master key não configurada</strong>
        <span>Configure `MASTERKEY` no ambiente do QuePasa para habilitar este console.</span>
      </div>

      <form v-else-if="!unlocked" class="unlock-panel" @submit.prevent="unlock">
        <KeyRound :size="34" />
        <div>
          <h2>Acesso master</h2>
          <p>Informe a master key para gerenciar as seções autorizadas no `/spam`.</p>
        </div>
        <label>
          <span>Master key</span>
          <input v-model="masterKey" type="password" autocomplete="current-password" required />
        </label>
        <button class="primary-action" type="submit" :disabled="validating">
          <LockKeyhole :size="18" />
          {{ validating ? 'Validando' : 'Entrar' }}
        </button>
      </form>

      <template v-else>
        <div v-if="error" class="notice error"><AlertCircle :size="18" />{{ error }}</div>
        <div v-if="notice" class="notice ok"><CheckCircle2 :size="18" />{{ notice }}</div>

        <div class="control-grid">
          <section class="panel search-panel">
            <div class="panel-heading">
              <div>
                <p class="eyebrow">PESQUISA</p>
                <h2>Encontrar seções</h2>
              </div>
              <Search :size="22" />
            </div>

            <form class="search-box" @submit.prevent="runSearch">
              <input v-model="search" type="search" placeholder="Token, telefone, usuário ou contexto" />
              <button type="submit" :disabled="searching">
                <Search :size="18" />
                {{ searching ? 'Buscando' : 'Buscar' }}
              </button>
            </form>

            <div class="result-list">
              <article v-for="item in searchResults" :key="item.token" class="section-card">
                <div class="status-dot" :class="{ ready: item.ready }" aria-hidden="true"></div>
                <div class="section-copy">
                  <strong>{{ sectionTitle(item) }}</strong>
                  <span>{{ sectionSubtitle(item) }}</span>
                  <code>{{ item.token }}</code>
                </div>
                <button
                  v-if="!item.inSpam"
                  class="icon-command add"
                  type="button"
                  :disabled="savingToken === item.token"
                  :aria-label="`Adicionar ${sectionTitle(item)}`"
                  @click="addSection(item)"
                >
                  <Plus :size="20" />
                </button>
                <span v-else class="queued-badge"><ListOrdered :size="16" />Na fila</span>
              </article>

              <div v-if="!searchResults.length" class="hint-box">
                <Search :size="24" />
                <strong>Pesquise para localizar seções</strong>
                <span>A busca percorre todos os usuários e contextos cadastrados.</span>
              </div>
            </div>
          </section>

          <section class="panel queue-panel">
            <div class="panel-heading">
              <div>
                <p class="eyebrow">ORDEM</p>
                <h2>Fila do `/spam`</h2>
              </div>
              <ListOrdered :size="22" />
            </div>

            <div class="queue-table" role="table" aria-label="Seções do serviço de spam">
              <div class="queue-row queue-head" role="row">
                <span>Ordem</span>
                <span>Seção</span>
                <span>Status</span>
                <span>Ações</span>
              </div>

              <article v-for="(item, index) in sections" :key="item.token" class="queue-row" role="row">
                <div class="order-cell">
                  <strong>{{ index + 1 }}</strong>
                  <div class="stack-actions">
                    <button type="button" :disabled="index === 0" :aria-label="`Subir ${sectionTitle(item)}`" @click.stop="moveSection(index, -1)">
                      <ArrowUp :size="16" />
                    </button>
                    <button type="button" :disabled="index === sections.length - 1" :aria-label="`Descer ${sectionTitle(item)}`" @click.stop="moveSection(index, 1)">
                      <ArrowDown :size="16" />
                    </button>
                  </div>
                </div>

                <div class="section-copy">
                  <strong>{{ sectionTitle(item) }}</strong>
                  <span>{{ sectionSubtitle(item) }}</span>
                  <code>{{ item.token }}</code>
                </div>

                <div class="status-cell">
                  <span class="connection-state" :class="{ ready: item.ready, disabled: !item.enabled }">
                    <Wifi v-if="item.ready" :size="16" />
                    <Server v-else :size="16" />
                    {{ item.enabled ? item.status : 'Pausada' }}
                  </span>
                </div>

                <div class="row-actions">
                  <button class="soft-command" type="button" :disabled="savingToken === item.token" @click="toggleSection(item)">
                    {{ item.enabled ? 'Pausar' : 'Ativar' }}
                  </button>
                  <button
                    class="icon-command danger"
                    type="button"
                    :disabled="savingToken === item.token"
                    :aria-label="`Remover ${sectionTitle(item)}`"
                    @click="removeSection(item)"
                  >
                    <Trash2 :size="18" />
                  </button>
                </div>
              </article>

              <div v-if="!sections.length" class="hint-box wide">
                <ListOrdered :size="26" />
                <strong>Nenhuma seção configurada</strong>
                <span>Enquanto esta fila estiver vazia, o `/spam` usa o comportamento legado.</span>
              </div>
            </div>
          </section>
        </div>
      </template>
    </section>
  </main>
</template>
