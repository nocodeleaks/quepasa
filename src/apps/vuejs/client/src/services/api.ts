import axios from 'axios'

// Read the API base path injected by the Go backend into index.html at serve
// time via window.__QUEPASA_CONFIG__. The value reflects whatever API_PREFIX
// the server owner configured (default: "/api"). Do NOT hardcode a fallback
// here — if the config is missing the SPA is being served incorrectly.
export const apiBase: string =
  (window as any).quepasa?.apiBase ?? ''

function trimSlashes(input: string): string {
  return input.replace(/^\/+|\/+$/g, '')
}

function resolveApiUrl(url: string): string {
  // Keep absolute URLs untouched.
  if (/^https?:\/\//i.test(url)) return url

  // Normalize only canonical API-prefixed requests.
  if (!url.startsWith('/api/')) return url

  const configuredBase = trimSlashes(apiBase)
  if (!configuredBase) return url

  const legacySuffix = url.substring('/api/'.length)
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
