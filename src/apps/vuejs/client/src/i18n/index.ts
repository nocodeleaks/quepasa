import { reactive, computed } from 'vue'
import en from './en'
import pt from './pt'
import type { Messages } from './en'

const STORAGE_KEY = 'quepasa_locale'

export type Locale = 'en-US' | 'pt-BR'

const translations: Record<Locale, Messages> = {
  'en-US': en,
  'pt-BR': pt,
}

function detectBrowserLocale(): Locale {
  const lang = navigator.language || ''
  return lang.toLowerCase().startsWith('pt') ? 'pt-BR' : 'en-US'
}

function loadLocale(): Locale {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'en-US' || stored === 'pt-BR') return stored as Locale
  } catch { /* ignore */ }
  return detectBrowserLocale()
}

// Module-level reactive state shared across all composable calls.
const state = reactive({ locale: loadLocale() })

export function setLocale(locale: Locale) {
  state.locale = locale
  try {
    localStorage.setItem(STORAGE_KEY, locale)
  } catch { /* ignore */ }
}

export function useLocale() {
  function t(key: keyof Messages, ...args: string[]): string {
    const dict: any = translations[state.locale as Locale] ?? translations['en-US']
    let msg: string = dict[key] ?? (translations['en-US'] as any)[key] ?? String(key)
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
