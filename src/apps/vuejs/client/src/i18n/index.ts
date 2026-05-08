import { reactive, computed } from 'vue'
import en from './en'
import pt from './pt'
import type { Messages } from './en'

const STORAGE_KEY = 'quepasa_locale'
const DEFAULT_LOCALE: Locale = 'en-US'

export type Locale = 'en-US' | 'pt-BR'

export const localeOptions: Array<{ locale: Locale; label: string; title: string }> = [
  { locale: 'en-US', label: 'EN', title: 'English' },
  { locale: 'pt-BR', label: 'PT', title: 'Português' },
]

const translations: Record<Locale, Messages> = {
  'en-US': en,
  'pt-BR': pt,
}

function normalizeLocale(input?: string | null): Locale | null {
  if (!input) return null

  const normalized = input.trim().replace('_', '-').toLowerCase()

  if (normalized === 'pt-br' || normalized === 'pt' || normalized.startsWith('pt-')) {
    return 'pt-BR'
  }

  if (normalized === 'en-us' || normalized === 'en' || normalized.startsWith('en-')) {
    return 'en-US'
  }

  return null
}

function detectBrowserLocale(): Locale {
  const candidates = [
    ...(navigator.languages || []),
    navigator.language,
  ]

  for (const candidate of candidates) {
    const locale = normalizeLocale(candidate)
    if (locale) return locale
  }

  return DEFAULT_LOCALE
}

function loadLocale(): Locale {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    const normalized = normalizeLocale(stored)
    if (normalized) return normalized
  } catch { /* ignore */ }
  return detectBrowserLocale()
}

// Module-level reactive state shared across all composable calls.
const state = reactive({ locale: loadLocale() })

export function setLocale(locale: Locale) {
  const normalized = normalizeLocale(locale) || DEFAULT_LOCALE
  state.locale = normalized
  try {
    localStorage.setItem(STORAGE_KEY, normalized)
  } catch { /* ignore */ }
}

export function useLocale() {
  function t(key: keyof Messages, ...args: string[]): string {
    const currentLocale = normalizeLocale(state.locale) || DEFAULT_LOCALE
    const dict: any = translations[currentLocale] ?? translations[DEFAULT_LOCALE]
    let msg: string = dict[key] ?? (translations[DEFAULT_LOCALE] as any)[key] ?? String(key)
    args.forEach((arg, i) => {
      msg = msg.replace(`{${i}}`, arg)
    })
    return msg
  }

  return {
    t,
    locale: computed(() => state.locale as Locale),
    setLocale,
  }
}
