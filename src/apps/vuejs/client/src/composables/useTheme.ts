import { computed, reactive } from 'vue'

const STORAGE_KEY = 'quepasa_theme'
const DARK_MEDIA_QUERY = '(prefers-color-scheme: dark)'

export type ThemeMode = 'light' | 'dark'

function normalizeTheme(input?: string | null): ThemeMode | null {
  if (!input) return null
  const normalized = input.trim().toLowerCase()
  if (normalized === 'light' || normalized === 'dark') return normalized
  return null
}

function detectSystemTheme(): ThemeMode {
  if (typeof window !== 'undefined' && 'matchMedia' in window) {
    return window.matchMedia(DARK_MEDIA_QUERY).matches ? 'dark' : 'light'
  }
  return 'light'
}

function getStoredTheme(): ThemeMode | null {
  if (typeof window === 'undefined') return null
  try {
    return normalizeTheme(window.localStorage.getItem(STORAGE_KEY))
  } catch {
    return null
  }
}

function loadThemeState() {
  const storedTheme = getStoredTheme()
  return {
    theme: storedTheme || detectSystemTheme(),
    followsSystem: storedTheme === null,
  }
}

const state = reactive(loadThemeState())

let initialized = false

function ensureThemeColorMeta() {
  if (typeof document === 'undefined') return null

  let meta = document.querySelector('meta[name="theme-color"]') as HTMLMetaElement | null
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = 'theme-color'
    document.head.appendChild(meta)
  }

  return meta
}

function applyTheme(theme: ThemeMode) {
  if (typeof document === 'undefined') return

  const root = document.documentElement
  root.setAttribute('data-theme', theme)
  root.setAttribute('data-bs-theme', theme)
  root.style.colorScheme = theme

  const meta = ensureThemeColorMeta()
  if (meta) {
    meta.content = theme === 'dark' ? '#07111f' : '#f6f8fc'
  }
}

function persistTheme(theme: ThemeMode | null) {
  if (typeof window === 'undefined') return
  try {
    if (theme) {
      window.localStorage.setItem(STORAGE_KEY, theme)
    } else {
      window.localStorage.removeItem(STORAGE_KEY)
    }
  } catch {
    // ignore storage failures
  }
}

export function setTheme(theme: ThemeMode) {
  state.theme = theme
  state.followsSystem = false
  persistTheme(theme)
  applyTheme(theme)
}

export function toggleTheme() {
  setTheme(state.theme === 'dark' ? 'light' : 'dark')
}

export function initializeTheme() {
  applyTheme(state.theme)
  if (initialized || typeof window === 'undefined' || !('matchMedia' in window)) return

  initialized = true
  const mediaQuery = window.matchMedia(DARK_MEDIA_QUERY)
  const handleChange = (event: MediaQueryListEvent) => {
    if (!state.followsSystem) return
    state.theme = event.matches ? 'dark' : 'light'
    applyTheme(state.theme)
  }

  if ('addEventListener' in mediaQuery) {
    mediaQuery.addEventListener('change', handleChange)
  } else if ('addListener' in mediaQuery) {
    mediaQuery.addListener(handleChange)
  }
}

export function useTheme() {
  return {
    theme: computed(() => state.theme),
    isDark: computed(() => state.theme === 'dark'),
    toggleTheme,
    setTheme,
  }
}
