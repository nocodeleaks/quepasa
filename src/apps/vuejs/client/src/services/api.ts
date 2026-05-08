import axios from 'axios'

// Read the API base path injected by the Go backend into index.html at serve
// time via window.__QUEPASA_CONFIG__. The value reflects whatever API_PREFIX
// the server owner configured (default: "/api"). Do NOT hardcode a fallback
// here — if the config is missing the SPA is being served incorrectly.
export const apiBase: string =
  (window as any).quepasa?.apiBase ?? ''

const canonicalV5Prefixes = [
  '/api/auth/',
  '/api/system/',
  '/api/users',
  '/api/sessions',
  '/api/session/',
  '/api/dispatches/',
  '/api/contacts',
  '/api/messages',
  '/api/media/',
  '/api/chats/',
  '/api/groups',
  '/api/status/',
  '/api/labels'
]

function trimSlashes(input: string): string {
  return input.replace(/^\/+|\/+$/g, '')
}

function isCanonicalV5Path(url: string): boolean {
  return canonicalV5Prefixes.some((prefix) => url === prefix || url.startsWith(prefix))
}

function normalizeVueApiVersion(url: string): string {
  if (!url.startsWith('/api/')) return url
  if (url.startsWith('/api/v')) return url
  if (!isCanonicalV5Path(url)) return url

  const suffix = url.substring('/api/'.length)
  return `/api/v5/${suffix}`
}

function resolveApiUrl(url: string): string {
  // Keep absolute URLs untouched.
  if (/^https?:\/\//i.test(url)) return url

  const normalizedUrl = normalizeVueApiVersion(url)

  // Normalize only canonical API-prefixed requests.
  if (!normalizedUrl.startsWith('/api/')) return normalizedUrl

  const configuredBase = trimSlashes(apiBase)
  if (!configuredBase) return normalizedUrl

  const legacySuffix = normalizedUrl.substring('/api/'.length)
  return `/${configuredBase}/${legacySuffix}`
}

const api = axios.create({
  // baseURL is set to the origin root so that absolute paths (e.g. /api/*)
  // are passed through as-is. Components that need to build URLs manually
  // (e.g. WebSocket endpoints) should use the exported `apiBase` constant.
  baseURL: '/',
  withCredentials: true,
  headers: { 'Accept': 'application/json' }
})

api.interceptors.request.use((config) => {
  const rawUrl = config.url ?? ''
  config.url = resolveApiUrl(rawUrl)
  return config
})

export default api
