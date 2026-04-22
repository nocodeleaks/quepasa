<template>
  <div :class="containerClass">
    <header v-if="!isLoginPage" class="mb-4">
      <nav class="navbar navbar-expand-lg navbar-dark rounded" :style="navbarStyle">
        <div class="container-fluid">
          <RouterLink class="navbar-brand d-flex align-items-center" to="/">
            <img v-if="branding.logo" :src="branding.logo" alt="Logo" class="navbar-logo me-2" />
            {{ branding.title || 'QuePasa' }}
          </RouterLink>
          
          <!-- Mobile: Offcanvas Toggle -->
          <button
            class="navbar-toggler"
            type="button"
            data-bs-toggle="offcanvas"
            data-bs-target="#navbarOffcanvas"
            aria-controls="navbarOffcanvas"
            aria-label="Toggle navigation"
          >
            <span class="navbar-toggler-icon"></span>
          </button>

          <!-- Desktop Navbar Collapse -->
          <div class="collapse navbar-collapse d-none d-lg-flex" id="navbarNav">
            <ul class="navbar-nav me-auto">
              <li class="nav-item">
                <RouterLink class="nav-link" to="/">Home</RouterLink>
              </li>
              <li class="nav-item">
                <RouterLink class="nav-link" to="/account">Account</RouterLink>
              </li>
              <li class="nav-item dropdown">
                <a class="nav-link dropdown-toggle" href="#" id="manageMenu" role="button" data-bs-toggle="dropdown" aria-expanded="false">Manage</a>
                <ul class="dropdown-menu" aria-labelledby="manageMenu">
                  <li><RouterLink class="dropdown-item" to="/users">Users</RouterLink></li>
                  <li><RouterLink class="dropdown-item" to="/users/create">Create User</RouterLink></li>
                  <li><hr class="dropdown-divider"></li>
                  <li><RouterLink class="dropdown-item" to="/environment">Environment</RouterLink></li>
                  <li><hr class="dropdown-divider"></li>
                  <li><a class="dropdown-item" href="/swagger/" target="_blank">API Docs (Swagger)</a></li>
                </ul>
              </li>
            </ul>
            <div class="d-flex align-items-center text-white" v-if="session.user.value">
              <small class="me-3">{{ session.user.value.username }}</small>
              <button class="btn btn-outline-light btn-sm" @click="logout">Logout</button>
            </div>
          </div>

          <!-- Mobile Offcanvas Menu (hidden on desktop via d-lg-none) -->
          <div class="offcanvas offcanvas-end d-lg-none" tabindex="-1" id="navbarOffcanvas" aria-labelledby="navbarOffcanvasLabel">
            <div class="offcanvas-header" :style="navbarStyle">
              <h5 class="offcanvas-title text-white" id="navbarOffcanvasLabel">
                <img v-if="branding.logo" :src="branding.logo" alt="Logo" class="navbar-logo me-2" />
                {{ branding.title || 'QuePasa' }}
              </h5>
              <button type="button" class="btn-close btn-close-white" data-bs-dismiss="offcanvas" aria-label="Close"></button>
            </div>
            <div class="offcanvas-body">
              <ul class="navbar-nav">
                <li class="nav-item">
                  <a class="nav-link" href="#" @click.prevent="navigateTo('/')">
                    <i class="fa fa-home me-2"></i> Home
                  </a>
                </li>
                <li class="nav-item">
                  <a class="nav-link" href="#" @click.prevent="navigateTo('/account')">
                    <i class="fa fa-user me-2"></i> Account
                  </a>
                </li>
                <li><hr class="my-2"></li>
                <li class="nav-item">
                  <a class="nav-link" href="#" @click.prevent="navigateTo('/setup')">
                    <i class="fa fa-cog me-2"></i> Setup
                  </a>
                </li>
                <li class="nav-item">
                  <a class="nav-link" href="#" @click.prevent="navigateTo('/users')">
                    <i class="fa fa-users me-2"></i> Users
                  </a>
                </li>
                <li class="nav-item">
                  <a class="nav-link" href="#" @click.prevent="navigateTo('/users/create')">
                    <i class="fa fa-user-plus me-2"></i> Create User
                  </a>
                </li>
                <li><hr class="my-2"></li>
                <li class="nav-item">
                  <a class="nav-link" href="#" @click.prevent="navigateTo('/environment')">
                    <i class="fa fa-cog me-2"></i> Environment
                  </a>
                </li>
                <li class="nav-item">
                  <a class="nav-link" href="/swagger/" target="_blank">
                    <i class="fa fa-book me-2"></i> API Docs
                  </a>
                </li>
                <li><hr class="my-2"></li>
                <li class="nav-item" v-if="session.user.value">
                  <div class="nav-link text-muted small">
                    <i class="fa fa-user-circle me-2"></i> {{ session.user.value.username }}
                  </div>
                </li>
                <li class="nav-item">
                  <a class="nav-link text-danger" href="#" @click.prevent="logout">
                    <i class="fa fa-sign-out-alt me-2"></i> Logout
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </nav>
    </header>

    <main>
      <div v-if="session.loading.value" class="text-center py-5">Carregando sessão...</div>
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
import api from './services/api'
import Toaster from '@/components/Toaster.vue'

export default defineComponent({
  components: { RouterLink, RouterView, Toaster },
  setup() {
    const year = new Date().getFullYear()
    const session = useSessionStore()
    const router = useRouter()
    const route = useRoute()
    const appVersion = ref('0.0.0')
    
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
    
    const isLoginPage = computed(() => route.path === '/login')
    
    const containerClass = computed(() => {
      if (isLoginPage.value) return ''
      return 'container py-4'
    })
    
    const navbarStyle = computed(() => ({
      background: `linear-gradient(135deg, ${branding.value.primaryColor}, ${branding.value.secondaryColor})`
    }))

    const loadBranding = async () => {
      try {
        const res = await api.get('/api/login/config')
        if (res.data?.branding) {
          branding.value = { ...branding.value, ...res.data.branding }
          
          // Apply CSS variables globally
          const root = document.documentElement
          root.style.setProperty('--branding-primary', branding.value.primaryColor)
          root.style.setProperty('--branding-secondary', branding.value.secondaryColor)
          root.style.setProperty('--branding-accent', branding.value.accentColor)
          
          // Update document title
          document.title = branding.value.title
          
          // Update favicon
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
        
        // Get version from response
        if (res.data?.version) {
          appVersion.value = res.data.version
        }
      } catch {
        // ignore
      }
    }

    const checkSession = async () => {
      await session.loadSession()
    }

    const logout = async () => {
      closeOffcanvas()
      try {
        await api.get('/logout')
      } catch (_) {
        /* ignore */
      }
      session.clearSession()
      router.push('/login')
    }
    
    const closeOffcanvas = () => {
      const offcanvasEl = document.getElementById('navbarOffcanvas')
      if (offcanvasEl) {
        // Use Bootstrap's Offcanvas API to close
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
      checkSession()
    })

    return { year, session, logout, isLoginPage, branding, navbarStyle, containerClass, appVersion, navigateTo }
  }
})
</script>

<style>
.navbar-logo {
  height: 32px;
  width: auto;
}

/* Footer styles */
.app-footer {
  margin-top: 2rem;
  padding: 1rem 0;
  border-top: 1px solid #e9ecef;
  background: #f8f9fa;
}

.footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #6c757d;
  font-size: 0.875rem;
}

.company-link {
  color: var(--branding-primary, #7C3AED);
  text-decoration: none;
}

.company-link:hover {
  text-decoration: underline;
}

.footer-right {
  display: flex;
  align-items: center;
}

.version-badge {
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  padding: 0.25rem 0.75rem;
  border-radius: 1rem;
  font-size: 0.75rem;
  font-weight: 600;
}

/* Offcanvas styles */
.offcanvas .nav-link {
  padding: 0.75rem 1rem;
  color: #333;
  font-size: 1rem;
}

.offcanvas .nav-link:hover {
  background: #f8f9fa;
}

/* Mobile offcanvas - full width */
@media (max-width: 991.98px) {
  .offcanvas {
    width: 100% !important;
  }
}

/* Mobile responsive */
@media (max-width: 768px) {
  #app.container {
    max-width: 100%;
    padding-left: 0;
    padding-right: 0;
    padding-top: 0;
  }

  header.mb-4 {
    margin-bottom: 0 !important;
    position: sticky;
    top: 0;
    z-index: 1000;
  }

  .navbar.rounded {
    border-radius: 0 !important;
  }

  .footer-content {
    flex-direction: column;
    gap: 0.5rem;
    text-align: center;
  }
}
</style>
