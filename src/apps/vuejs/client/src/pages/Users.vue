<template>
  <div class="users-page">
    <div class="page-header">
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
          </svg>
          {{ t('users_title') }}
        </h1>
        <p class="hide-mobile">{{ t('users_subtitle') }}</p>
      </div>
      <div class="header-actions">
        <router-link to="/users/create" class="btn-primary">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
          </svg>
          {{ t('users_new') }}
        </router-link>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner-large"></div>
      <p>{{ t('users_loading') }}</p>
    </div>

    <div v-if="error" class="error-box">
      <i class="fa fa-exclamation-triangle"></i>
      <span>{{ error }}</span>
      <button @click="load" class="retry-btn">{{ t('error_retry') }}</button>
    </div>

    <div v-else-if="!loading && users.length === 0" class="empty-state">
      <div class="empty-icon">
        <i class="fa fa-users fa-4x"></i>
      </div>
      <h2>{{ t('users_empty_title') }}</h2>
      <p>{{ t('users_empty_desc') }}</p>
      <router-link to="/users/create" class="btn-primary-large">
        <i class="fa fa-user-plus"></i>
        {{ t('users_create_cta') }}
      </router-link>
    </div>

    <div v-else class="users-list">
      <div v-for="user in users" :key="user.username" class="user-card">
        <div class="user-info">
          <div class="user-avatar">
            <i class="fa fa-user"></i>
          </div>
          <div class="user-details">
            <h3>{{ user.username }}</h3>
            <small v-if="user.timestamp">{{ t('users_created_at', [formatDate(user.timestamp)]) }}</small>
          </div>
        </div>
        <div class="user-actions">
          <button
            class="btn-danger-small"
            @click="confirmDelete(user)"
            :disabled="deleting === user.username"
            :title="t('delete')"
          >
            <i v-if="deleting === user.username" class="fa fa-spinner fa-spin"></i>
            <i v-else class="fa fa-trash"></i>
          </button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-icon danger">
          <i class="fa fa-exclamation-triangle fa-3x"></i>
        </div>
        <h3>{{ t('users_delete_title') }}</h3>
        <p>{{ t('users_delete_message', [userToDelete?.username || '']) }}</p>
        <div class="modal-actions">
          <button @click="showDeleteModal = false" class="btn-secondary">{{ t('cancel') }}</button>
          <button @click="deleteUser" class="btn-danger">{{ t('delete') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useMasterKey } from '@/composables/useMasterKey'
import { useLocale } from '@/i18n'

interface User {
  username: string
  timestamp?: string
}

export default defineComponent({
  setup() {
    const { masterKeyHeaders } = useMasterKey()
    const { t, locale } = useLocale()
    const users = ref<User[]>([])
    const loading = ref(true)
    const error = ref('')
    const deleting = ref('')
    const showDeleteModal = ref(false)
    const userToDelete = ref<User | null>(null)

    async function load() {
      loading.value = true
      error.value = ''
      try {
        const res = await api.get('/api/users', { headers: masterKeyHeaders() })
        users.value = res.data.users || []
      } catch (err: any) {
        error.value = err?.response?.data?.result || t('users_error_load')
      } finally {
        loading.value = false
      }
    }

    function formatDate(dateStr: string) {
      try {
        return new Date(dateStr).toLocaleDateString(locale.value)
      } catch {
        return dateStr
      }
    }

    function confirmDelete(user: User) {
      userToDelete.value = user
      showDeleteModal.value = true
    }

    async function deleteUser() {
      if (!userToDelete.value) return

      const username = userToDelete.value.username
      deleting.value = username
      showDeleteModal.value = false

      try {
        await api.delete(`/api/user/${encodeURIComponent(username)}`, { headers: masterKeyHeaders() })
        pushToast(t('users_deleted'), 'success')
        await load()
      } catch (err: any) {
        pushToast(err?.response?.data?.result || t('users_error_delete'), 'error')
      } finally {
        deleting.value = ''
        userToDelete.value = null
      }
    }

    onMounted(() => {
      load()
    })

    return {
      t,
      users,
      loading,
      error,
      deleting,
      showDeleteModal,
      userToDelete,
      load,
      formatDate,
      confirmDelete,
      deleteUser,
    }
  }
})
</script>

<style scoped>
.users-page {
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-content h1 {
  font-size: 24px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-content h1 svg {
  color: var(--branding-primary, #7C3AED);
}

.btn-primary svg {
  flex-shrink: 0;
}

.header-content p {
  color: #6b7280;
  margin: 0;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: var(--branding-primary, #7C3AED);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
}

.btn-primary:hover {
  opacity: 0.9;
}

.loading-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
}

.spinner-large {
  width: 40px;
  height: 40px;
  border: 3px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-box {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
}

.retry-btn {
  margin-left: auto;
  padding: 6px 12px;
  background: #dc2626;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  background: #f9fafb;
  border-radius: 12px;
}

.empty-icon {
  color: #9ca3af;
  margin-bottom: 16px;
}

.empty-state h2 {
  font-size: 20px;
  color: #374151;
  margin: 0 0 8px;
}

.empty-state p {
  color: #6b7280;
  margin: 0 0 24px;
}

.btn-primary-large {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 14px 28px;
  background: var(--branding-primary, #7C3AED);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  text-decoration: none;
}

.users-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
}

.user-card.user-self {
  border: 2px solid var(--branding-primary, #7C3AED);
  background: #f5f3ff;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  width: 44px;
  height: 44px;
  background: var(--branding-primary, #7C3AED);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-avatar.avatar-self {
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.4);
}

.user-details h3 {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
  margin: 0 0 2px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.self-badge {
  font-size: 10px;
  padding: 2px 6px;
  background: var(--branding-primary, #7C3AED);
  color: white;
  border-radius: 4px;
  font-weight: 600;
  text-transform: uppercase;
}

.user-details small {
  color: #6b7280;
}

.btn-danger-small {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  border-radius: 8px;
  cursor: pointer;
}

.btn-danger-small:hover {
  background: #fee2e2;
}

.btn-danger-small:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  background: white;
  border-radius: 16px;
  padding: 32px;
  max-width: 400px;
  width: 100%;
  text-align: center;
}

.modal-icon {
  width: 72px;
  height: 72px;
  margin: 0 auto 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-icon.danger {
  background: #fef2f2;
  color: #dc2626;
}

.modal-content h3 {
  font-size: 20px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 12px;
}

.modal-content p {
  color: #6b7280;
  margin: 0 0 24px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.btn-secondary {
  padding: 12px 24px;
  background: #f3f4f6;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  color: #374151;
  cursor: pointer;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.btn-danger {
  padding: 12px 24px;
  background: #dc2626;
  color: white;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
}

.btn-danger:hover {
  background: #b91c1c;
}

@media (max-width: 768px) {
  .hide-mobile {
    display: none !important;
  }

  .users-page {
    padding: 0 16px;
  }

  .page-header {
    position: sticky;
    top: 0;
    background: white;
    padding: 16px 0;
    margin: 0 0 16px;
    z-index: 100;
  }
}
</style>
