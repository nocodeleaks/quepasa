import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '@/pages/Home.vue'
import AccountPage from '@/pages/Account.vue'
import LoginPage from '@/pages/Login.vue'
import WebhooksPage from '@/pages/Webhooks.vue'
import RabbitMQPage from '@/pages/RabbitMQ.vue'
import ServerPage from '@/pages/Server.vue'
import SetupPage from '@/pages/Setup.vue'
import ConnectPage from '@/pages/Connect.vue'
import UsersPage from '@/pages/Users.vue'
import UserCreatePage from '@/pages/UserCreate.vue'
import EnvironmentPage from '@/pages/Environment.vue'
import QRCodePage from '@/pages/QRCode.vue'
import PairCodePage from '@/pages/PairCode.vue'
import MessagesPage from '@/pages/Messages.vue'
import SendMessagePage from '@/pages/SendMessage.vue'
import { useSessionStore } from '@/stores/session'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomePage, meta: { requiresAuth: true } },
    { path: '/account', name: 'account', component: AccountPage, meta: { requiresAuth: true } },
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
    { path: '/setup', name: 'setup', component: SetupPage, meta: { requiresAuth: true } },
    { path: '/users', name: 'users', component: UsersPage, meta: { requiresAuth: true } },
    { path: '/users/create', name: 'users.create', component: UserCreatePage, meta: { requiresAuth: true } },
    { path: '/environment', name: 'environment', component: EnvironmentPage, meta: { requiresAuth: true } },
    { path: '/login', name: 'login', component: LoginPage },
  ],
})

router.beforeEach(async (to) => {
  const session = useSessionStore()
  if (session.loading.value) {
    await session.loadSession()
  }

  if (to.meta.requiresAuth && !session.user.value) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.name === 'login' && session.user.value) {
    return { name: 'home' }
  }

  return true
})

export default router
