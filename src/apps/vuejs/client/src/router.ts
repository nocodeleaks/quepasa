import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '@/pages/Home.vue'
import LoginPage from '@/pages/Login.vue'
import WebhooksPage from '@/pages/Webhooks.vue'
import RabbitMQPage from '@/pages/RabbitMQ.vue'
import ServerPage from '@/pages/Server.vue'
import ConnectPage from '@/pages/Connect.vue'
import QRCodePage from '@/pages/QRCode.vue'
import PairCodePage from '@/pages/PairCode.vue'
import MessagesPage from '@/pages/Messages.vue'
import SendMessagePage from '@/pages/SendMessage.vue'
import AccountPage from '@/pages/Account.vue'
import EnvironmentPage from '@/pages/Environment.vue'
import UsersPage from '@/pages/Users.vue'
import UserCreatePage from '@/pages/UserCreate.vue'
import SetupPage from '@/pages/Setup.vue'
import { useSessionStore } from '@/stores/session'

function hasSessionCookie() {
  return document.cookie.split(';').some((entry) => entry.trim().startsWith('jwt='))
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: HomePage, meta: { requiresAuth: true } },
    { path: '/webhooks', name: 'webhooks', component: WebhooksPage, meta: { requiresAuth: true } },
    { path: '/dispatching', name: 'dispatching', component: WebhooksPage, meta: { requiresAuth: true } },
    { path: '/rabbitmq', name: 'rabbitmq', component: RabbitMQPage, meta: { requiresAuth: true } },
    { path: '/connect', name: 'connect', component: ConnectPage, meta: { requiresAuth: true } },
    { path: '/server/:token', name: 'server', component: ServerPage, meta: { requiresAuth: true } },
    { path: '/server/:token/qrcode', name: 'server.qrcode', component: QRCodePage, meta: { requiresAuth: true } },
    { path: '/server/:token/paircode', name: 'server.paircode', component: PairCodePage, meta: { requiresAuth: true } },
    { path: '/server/:token/messages', name: 'server.messages', component: MessagesPage, meta: { requiresAuth: true } },
    { path: '/server/:token/send', name: 'server.send', component: SendMessagePage, meta: { requiresAuth: true } },
    { path: '/server/:token/groups', name: 'server.groups', component: () => import('@/pages/Groups.vue'), meta: { requiresAuth: true } },
    { path: '/server/:token/groups/:id', name: 'server.groups.detail', component: () => import('@/pages/GroupDetail.vue'), meta: { requiresAuth: true } },
    { path: '/account', name: 'account', component: AccountPage, meta: { requiresAuth: true } },
    { path: '/environment', name: 'environment', component: EnvironmentPage, meta: { requiresAuth: true } },
    { path: '/users', name: 'users', component: UsersPage, meta: { requiresAuth: true } },
    { path: '/users/create', name: 'users.create', component: UserCreatePage, meta: { requiresAuth: true } },
    { path: '/setup', name: 'setup', component: SetupPage },
    { path: '/login', name: 'login', component: LoginPage },
  ],
})

router.beforeEach(async (to) => {
  const session = useSessionStore()
  const requiresAuth = Boolean(to.meta.requiresAuth)
  // Always probe the session when navigating to a protected page or the login/setup
  // pages. The jwt cookie is HttpOnly so document.cookie cannot detect it — probing
  // the backend directly is the only reliable way to know whether the user is already
  // authenticated.
  const shouldProbeSession = requiresAuth || to.name === 'login' || to.name === 'setup'

  if (session.loading.value && shouldProbeSession) {
    await session.loadSession({ allowUnauthorized: true })
  } else if (session.loading.value) {
    session.resolveUnauthenticated()
  }

  if (requiresAuth && !session.user.value) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.name === 'login' && session.user.value) {
    return { name: 'home' }
  }

  return true
})

export default router
