<template>
  <div class="app-shell">
    <header v-if="!isLoginPage" class="app-shell-header">
      <div class="app-topbar-wrap" :style="shellStyle">
        <div class="app-topbar">
        <div class="topbar-left">
          <RouterLink class="brand-pill" to="/">
            <img v-if="branding.logo" :src="branding.logo" alt="Logo" class="brand-logo" />
            <div class="brand-copy">
              <span class="brand-title">{{ branding.title || 'QuePasa' }}</span>
              <span class="brand-caption">{{ branding.companyName || t('home_subtitle') }}</span>
            </div>
          </RouterLink>
        </div>

        <nav class="desktop-nav d-none d-lg-flex">
          <RouterLink class="nav-pill" to="/">{{ t('nav_home') }}</RouterLink>
          <RouterLink class="nav-pill" to="/connect">{{ t('nav_connect') }}</RouterLink>
          <RouterLink v-if="session.user.value" class="nav-pill" to="/account">{{ t('nav_account') }}</RouterLink>
          <RouterLink v-if="hasMasterKey" class="nav-pill nav-pill-master" :class="{ 'nav-pill-master-active': isMasterAuthenticated() }" to="/master">{{ t('nav_master') }}</RouterLink>
          <a class="nav-pill nav-pill-secondary" href="/swagger/" target="_blank">{{ t('nav_api_docs') }}</a>
        </nav>

        <div class="desktop-actions d-none d-lg-flex">
          <div v-if="session.user.value" class="user-avatar-shell">
            <div class="user-hover-card">{{ session.user.value.email || session.user.value.username }}</div>
            <button class="user-avatar-btn" type="button" :aria-label="session.user.value.email || session.user.value.username">
              <i class="fa fa-user-circle"></i>
            </button>
          </div>
          <div class="lang-switch">
            <button class="lang-btn" :class="{ active: locale === 'en-US' }" @click="setLocale('en-US')" title="English">EN</button>
            <button class="lang-btn" :class="{ active: locale === 'pt-BR' }" @click="setLocale('pt-BR')" title="Português">PT</button>
          </div>
          <button v-if="session.user.value" class="logout-btn" @click="logout">{{ t('logout') }}</button>
        </div>

        <div class="mobile-actions d-lg-none">
          <div v-if="session.user.value" class="user-avatar-shell user-avatar-shell-mobile">
            <button class="user-avatar-btn" type="button" :aria-label="session.user.value.email || session.user.value.username">
              <i class="fa fa-user-circle"></i>
            </button>
          </div>
          <div class="lang-switch lang-switch-compact">
            <button class="lang-btn" :class="{ active: locale === 'en-US' }" @click="setLocale('en-US')" title="English">EN</button>
            <button class="lang-btn" :class="{ active: locale === 'pt-BR' }" @click="setLocale('pt-BR')" title="Português">PT</button>
          </div>
          <button
            class="menu-btn"
            type="button"
            data-bs-toggle="offcanvas"
            data-bs-target="#mobileMenuSheet"
            aria-controls="mobileMenuSheet"
            :aria-label="t('nav_more')"
          >
            <i class="fa fa-bars"></i>
          </button>
        </div>
        </div>
      </div>

      <nav class="mobile-dock d-lg-none">
        <RouterLink class="dock-link" to="/">
          <i class="fa fa-home"></i>
          <span>{{ t('nav_home') }}</span>
        </RouterLink>
        <RouterLink class="dock-link" to="/connect">
          <i class="fa fa-link"></i>
          <span>{{ t('nav_connect') }}</span>
        </RouterLink>
        <RouterLink v-if="session.user.value" class="dock-link" to="/account">
          <i class="fa fa-user-circle"></i>
          <span>{{ t('nav_account') }}</span>
        </RouterLink>
        <button
          class="dock-link dock-link-button"
          type="button"
          data-bs-toggle="offcanvas"
          data-bs-target="#mobileMenuSheet"
          aria-controls="mobileMenuSheet"
        >
          <i class="fa fa-ellipsis-h"></i>
          <span>{{ t('nav_more') }}</span>
        </button>
      </nav>

      <div class="offcanvas offcanvas-bottom mobile-menu-sheet" tabindex="-1" id="mobileMenuSheet" aria-labelledby="mobileMenuSheetLabel">
        <div class="offcanvas-header">
          <div>
            <div class="sheet-title" id="mobileMenuSheetLabel">{{ branding.title || 'QuePasa' }}</div>
            <div class="sheet-subtitle">{{ t('nav_more') }}</div>
          </div>
          <button type="button" class="btn-close" data-bs-dismiss="offcanvas" :aria-label="t('close')"></button>
        </div>
        <div class="offcanvas-body">
          <div v-if="session.user.value" class="sheet-user-pill">
            <i class="fa fa-user-circle"></i>
            <span>{{ session.user.value.email || session.user.value.username }}</span>
          </div>

          <div class="sheet-grid">
            <button class="sheet-tile" type="button" @click="navigateTo('/')">
              <i class="fa fa-home"></i>
              <span>{{ t('nav_home') }}</span>
            </button>
            <button class="sheet-tile" type="button" @click="navigateTo('/connect')">
              <i class="fa fa-link"></i>
              <span>{{ t('nav_connect') }}</span>
            </button>
            <button v-if="session.user.value" class="sheet-tile" type="button" @click="navigateTo('/account')">
              <i class="fa fa-user-circle"></i>
              <span>{{ t('nav_account') }}</span>
            </button>
            <button v-if="hasMasterKey" class="sheet-tile sheet-tile-master" type="button" @click="navigateTo('/master')">
              <i class="fa fa-shield-alt"></i>
              <span>{{ t('nav_master') }}</span>
            </button>
            <a class="sheet-tile" href="/swagger/" target="_blank" @click="closeOffcanvas">
              <i class="fa fa-book"></i>
              <span>{{ t('nav_api_docs') }}</span>
            </a>
            <button v-if="session.user.value" class="sheet-tile sheet-tile-danger" type="button" @click="logout">
              <i class="fa fa-sign-out-alt"></i>
              <span>{{ t('logout') }}</span>
            </button>
          </div>

          <div class="sheet-language">
            <span>{{ t('language_label') }}</span>
            <div class="lang-switch lang-switch-sheet">
              <button class="lang-btn" :class="{ active: locale === 'en-US' }" @click="setLocale('en-US')" title="English">EN</button>
              <button class="lang-btn" :class="{ active: locale === 'pt-BR' }" @click="setLocale('pt-BR')" title="Português">PT</button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <main :class="mainClass">
      <div v-if="session.loading.value" class="shell-loading">{{ t('loading_session') }}</div>
      <RouterView v-else />
    </main>

    <Toaster />

    <footer v-if="!isLoginPage" class="app-footer">
      <div class="footer-content">
        <div class="footer-left">
          <span>© {{ year }}</span>
          <a v-if="branding.companyUrl" :href="branding.companyUrl" target="_blank" class="company-link">
            {{ branding.companyName || branding.title || 'QuePasa' }}
          </a>
          <span v-else>{{ branding.companyName || branding.title || 'QuePasa' }}</span>
        </div>
        <div class="footer-right">
          <span class="version-badge">v{{ appVersion }}</span>
        </div>
      </div>
    </footer>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, computed, ref } from 'vue'
import { RouterLink, RouterView, useRouter, useRoute } from 'vue-router'
import { useSessionStore } from '@/stores/session'
import { useMasterKey } from '@/composables/useMasterKey'
import api from './services/api'
import Toaster from '@/components/Toaster.vue'
import { useLocale } from '@/i18n'

export default defineComponent({
  components: { RouterLink, RouterView, Toaster },
  setup() {
    const year = new Date().getFullYear()
    const session = useSessionStore()
    const { isMasterAuthenticated } = useMasterKey()
    const router = useRouter()
    const route = useRoute()
    const appVersion = ref('0.0.0')
    const hasMasterKey = ref(false)
    const relaxedSessions = ref(true)

    const { t, locale, setLocale } = useLocale()

    const branding = ref({
      title: 'QuePasa',
      logo: '',
      favicon: '',
      primaryColor: '#7C3AED',
      secondaryColor: '#5B21B6',
      accentColor: '#8B5CF6',
      companyName: '',
      companyUrl: ''
    })

    const isLoginPage = computed(() => route.path === '/login' || route.path === '/setup')

    const mainClass = computed(() => {
      if (isLoginPage.value) return 'app-main app-main-auth'
      return 'app-main container'
    })

    const shellStyle = computed(() => ({
      background: `linear-gradient(135deg, color-mix(in srgb, ${branding.value.primaryColor} 92%, black), color-mix(in srgb, ${branding.value.secondaryColor} 88%, black))`,
      boxShadow: `0 4px 14px color-mix(in srgb, ${branding.value.primaryColor} 20%, transparent)`
    }))

    const loadBranding = async () => {
      try {
        const res = await api.get('/api/auth/config')
        if (res.data?.branding) {          branding.value = { ...branding.value, ...res.data.branding }

          const root = document.documentElement
          root.style.setProperty('--branding-primary', branding.value.primaryColor)
          root.style.setProperty('--branding-secondary', branding.value.secondaryColor)
          root.style.setProperty('--branding-accent', branding.value.accentColor)

          document.title = branding.value.title

          if (branding.value.favicon) {
            let favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement
            if (!favicon) {
              favicon = document.createElement('link')
              favicon.rel = 'icon'
              document.head.appendChild(favicon)
            }
            favicon.href = branding.value.favicon
          }
        }

        if (res.data?.version) {
          appVersion.value = res.data.version
        }

        // Load master key availability after branding
        if (session.user.value) {
          try {
            const accountRes = await api.get('/api/account')
            hasMasterKey.value = Boolean(accountRes.data?.hasMasterKey)
            relaxedSessions.value = accountRes.data?.relaxedSessions !== false
          } catch {
            // ignore
          }
        }
      } catch {
        // ignore
      }
    }

    const logout = async () => {
      closeOffcanvas()
      try {
        await api.get('/logout')
      } catch {
        // ignore
      }
      session.clearSession()
      router.push('/login')
    }

    const closeOffcanvas = () => {
      const offcanvasEl = document.getElementById('mobileMenuSheet')
      if (offcanvasEl) {
        const bsOffcanvas = (window as any).bootstrap?.Offcanvas?.getInstance(offcanvasEl)
        if (bsOffcanvas) {
          bsOffcanvas.hide()
        }
      }
    }

    const navigateTo = (path: string) => {
      closeOffcanvas()
      router.push(path)
    }

    onMounted(() => {
      loadBranding()
    })

    return { year, session, logout, isLoginPage, branding, shellStyle, mainClass, appVersion, navigateTo, closeOffcanvas, t, locale, setLocale, hasMasterKey, isMasterAuthenticated, relaxedSessions }
  }
})
</script>

<style>
.app-shell {
  min-height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(124, 58, 237, 0.08), transparent 26%),
    radial-gradient(circle at top right, rgba(14, 165, 233, 0.06), transparent 22%),
    linear-gradient(180deg, #f7f7fb 0%, #f4f6fb 100%);
}

.app-shell-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1020;
  background: rgba(247, 247, 251, 0.98);
  backdrop-filter: blur(6px);
  border-bottom: 1px solid rgba(148, 163, 184, 0.15);
}

.app-topbar,
.mobile-dock {
  position: relative;
}

.app-topbar-wrap {
  width: 100%;
}

.app-topbar {
  width: min(1240px, 100%);
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.7rem 1rem;
  color: #fff;
}

.topbar-left,
.desktop-actions,
.mobile-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.mobile-actions {
  margin-left: auto;
}

.brand-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
  color: inherit;
  text-decoration: none;
}

.brand-logo {
  width: 40px;
  height: 40px;
  border-radius: 14px;
  object-fit: cover;
  background: rgba(255, 255, 255, 0.14);
  padding: 0.2rem;
}

.brand-copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.brand-title {
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.1;
}

.brand-caption {
  max-width: 180px;
  font-size: 0.74rem;
  color: rgba(255, 255, 255, 0.78);
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.desktop-nav {
  flex: 1;
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.nav-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 32px;
  padding: 0.3rem 0.8rem;
  border-radius: 999px;
  font-size: 0.82rem;
  color: rgba(255, 255, 255, 0.84);
  text-decoration: none;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.10);
  font-weight: 600;
  transition: transform 0.18s ease, background 0.18s ease, color 0.18s ease;
}

.nav-pill:hover,
.nav-pill.router-link-active,
.nav-pill.router-link-exact-active {
  color: #fff;
  background: rgba(255, 255, 255, 0.12);
  transform: translateY(-1px);
}

.nav-pill-secondary {
  background: rgba(17, 24, 39, 0.10);
}

.nav-pill-master {
  background: rgba(220, 38, 38, 0.12);
  color: #dc2626;
}

.nav-pill-master-active {
  background: rgba(220, 38, 38, 0.22);
  font-weight: 700;
}

.sheet-tile-master {
  background: rgba(220, 38, 38, 0.08);
  color: #dc2626;
}

.user-avatar-shell {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.user-avatar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.10);
  color: #fff;
  font-size: 1rem;
  cursor: default;
}

.user-hover-card {
  position: absolute;
  right: calc(100% + 0.75rem);
  top: 50%;
  transform: translateY(-50%) translateX(8px);
  opacity: 0;
  pointer-events: none;
  display: inline-flex;
  align-items: center;
  min-height: 42px;
  max-width: 260px;
  padding: 0.7rem 0.9rem;
  border-radius: 14px;
  background: rgba(17, 24, 39, 0.78);
  color: #fff;
  font-size: 0.88rem;
  font-weight: 600;
  white-space: nowrap;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.14);
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.user-avatar-shell:hover .user-hover-card {
  opacity: 1;
  transform: translateY(-50%) translateX(0);
}

.lang-switch {
  display: inline-flex;
  padding: 0.18rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.10);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.lang-btn,
.logout-btn,
.menu-btn,
.dock-link-button {
  border: 0;
}

.lang-btn {
  min-width: 42px;
  padding: 0.45rem 0.7rem;
  border-radius: 999px;
  background: transparent;
  color: rgba(255, 255, 255, 0.74);
  font-size: 0.78rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}

.lang-btn.active,
.lang-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.14);
}

.logout-btn,
.menu-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 44px;
  padding: 0.65rem 1rem;
  border-radius: 16px;
  color: #fff;
  background: rgba(17, 24, 39, 0.12);
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.18s ease, background 0.18s ease;
}

.logout-btn:hover,
.menu-btn:hover {
  transform: translateY(-1px);
  background: rgba(17, 24, 39, 0.18);
}

.mobile-actions {
  gap: 0.55rem;
}

.lang-switch-compact .lang-btn {
  min-width: 38px;
  padding-inline: 0.55rem;
}

.mobile-dock {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.5rem;
  max-width: 1240px;
  margin: 0 auto;
  padding: 0.45rem;
  background: rgba(255, 255, 255, 0.62);
  border-top: 1px solid rgba(148, 163, 184, 0.12);
  backdrop-filter: blur(18px);
}

.dock-link {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  min-height: 58px;
  padding: 0.55rem 0.35rem;
  border-radius: 16px;
  color: #64748b;
  text-decoration: none;
  background: transparent;
  font-size: 0.72rem;
  font-weight: 600;
}

.dock-link i {
  font-size: 1rem;
}

.dock-link.router-link-active,
.dock-link.router-link-exact-active,
.dock-link:hover,
.dock-link-button:hover {
  color: var(--branding-primary, #7C3AED);
  background: rgba(124, 58, 237, 0.08);
}

.dock-link-button {
  font: inherit;
  cursor: pointer;
}

.mobile-menu-sheet {
  height: auto !important;
  max-height: 78vh;
  border-top-left-radius: 28px;
  border-top-right-radius: 28px;
  background: linear-gradient(180deg, rgba(255,255,255,0.94), rgba(248,247,255,0.88));
}

.mobile-menu-sheet .offcanvas-header {
  align-items: flex-start;
  padding: 1.2rem 1.2rem 0.75rem;
}

.sheet-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: #111827;
}

.sheet-subtitle {
  color: #64748b;
  font-size: 0.88rem;
}

.sheet-user-pill {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.9rem 1rem;
  border-radius: 18px;
  background: rgba(244, 240, 255, 0.72);
  color: #4c1d95;
  font-weight: 600;
}

.sheet-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 1rem;
}

.sheet-tile {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.55rem;
  min-height: 96px;
  padding: 1rem;
  border-radius: 20px;
  text-decoration: none;
  color: #0f172a;
  background: rgba(255, 255, 255, 0.74);
  border: 1px solid rgba(148, 163, 184, 0.18);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.04);
}

.sheet-tile i {
  font-size: 1.05rem;
  color: var(--branding-primary, #7C3AED);
}

.sheet-tile-danger i,
.sheet-tile-danger {
  color: #b91c1c;
}

.sheet-language {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1rem;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  background: rgba(238, 242, 255, 0.72);
  color: #334155;
  font-weight: 600;
}

.lang-switch-sheet {
  background: rgba(124, 58, 237, 0.10);
  border: 0;
}

.lang-switch-sheet .lang-btn {
  color: #6b21a8;
}

.lang-switch-sheet .lang-btn.active,
.lang-switch-sheet .lang-btn:hover {
  color: #fff;
  background: var(--branding-primary, #7C3AED);
}

.app-main {
  position: relative;
  z-index: 2;
  padding-top: 7rem;
  padding-bottom: 4.5rem;
}

.app-main-auth {
  padding-bottom: 0;
}

.shell-loading {
  text-align: center;
  padding: 4rem 1rem;
  color: #64748b;
  font-weight: 600;
}

.app-footer {
  position: relative;
  z-index: 0;
  margin-top: 1rem;
  padding: 1.2rem 0 1.8rem;
}

.footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.footer-left {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
  color: #64748b;
  font-size: 0.88rem;
}

.company-link {
  color: var(--branding-primary, #7C3AED);
  text-decoration: none;
}

.company-link:hover {
  text-decoration: underline;
}

.version-badge {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0.35rem 0.85rem;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: #fff;
  font-size: 0.76rem;
  font-weight: 700;
  pointer-events: none;
}

@media (min-width: 992px) {
  .app-shell-header {
    background: linear-gradient(180deg, rgba(247, 247, 251, 0.98), rgba(247, 247, 251, 0.92));
  }

  .app-main {
    padding-top: 5.1rem;
    padding-bottom: 2rem;
  }

  .mobile-dock {
    display: none;
  }
}

@media (max-width: 991.98px) {
  .desktop-nav,
  .desktop-actions {
    display: none !important;
  }

  .app-topbar {
    padding: 0.68rem 0.85rem;
  }

  .brand-caption {
    max-width: 132px;
  }

  .user-hover-card {
    display: none;
  }
}

@media (max-width: 767.98px) {
  .app-shell-header {
    background: linear-gradient(180deg, rgba(247, 247, 251, 0.98), rgba(247, 247, 251, 0.94));
  }

  .app-topbar {
    gap: 0.65rem;
    padding: 0.7rem 0.75rem;
  }

  .brand-logo {
    width: 36px;
    height: 36px;
    border-radius: 12px;
  }

  .brand-title {
    font-size: 0.9rem;
  }

  .brand-caption {
    font-size: 0.64rem;
  }

  .menu-btn {
    min-width: 44px;
    padding-inline: 0.85rem;
  }

  .user-avatar-btn {
    width: 42px;
    height: 42px;
  }

  .sheet-grid {
    grid-template-columns: 1fr 1fr;
  }

  .footer-content {
    flex-direction: column;
    text-align: center;
  }
}
</style>
