<template>
  <div class="users-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1>
          <i class="fa fa-users"></i>
          Users
        </h1>
        <p class="hide-mobile">Manage users in your account</p>
      </div>
      <div class="header-actions">
        <router-link to="/users/create" class="btn-primary">
          <i class="fa fa-user-plus"></i>
          New User
        </router-link>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner-large"></div>
      <p>Loading users...</p>
    </div>

    <!-- Error -->
    <div v-if="error" class="error-box">
      <i class="fa fa-exclamation-triangle"></i>
      <span>{{ error }}</span>
      <button @click="load" class="retry-btn">Retry</button>
    </div>

    <!-- Empty state -->
    <div v-else-if="!loading && users.length === 0" class="empty-state">
      <div class="empty-icon">
        <i class="fa fa-users fa-4x"></i>
      </div>
      <h2>No users yet</h2>
      <p>Create your first user to get started</p>
      <router-link to="/users/create" class="btn-primary-large">
        <i class="fa fa-user-plus"></i>
        Create User
      </router-link>
    </div>

    <!-- Users List -->
    <div v-else class="users-list">
      <div v-for="user in users" :key="user.username" class="user-card" :class="{ 'user-self': user.is_self }">
        <div class="user-info">
          <div class="user-avatar" :class="{ 'avatar-self': user.is_self }">
            <i class="fa fa-user"></i>
          </div>
          <div class="user-details">
            <h3>
              {{ user.username }}
              <span v-if="user.is_self" class="self-badge">You</span>
            </h3>
            <small v-if="user.timestamp">Created: {{ formatDate(user.timestamp) }}</small>
          </div>
        </div>
        <div class="user-actions">
          <button 
            v-if="!user.is_self"
            class="btn-danger-small" 
            @click="confirmDelete(user)"
            :disabled="deleting === user.username"
          >
            <i v-if="deleting === user.username" class="fa fa-spinner fa-spin"></i>
            <i v-else class="fa fa-trash"></i>
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-icon danger">
          <i class="fa fa-exclamation-triangle fa-3x"></i>
        </div>
        <h3>Delete User?</h3>
        <p>Are you sure you want to delete <strong>{{ userToDelete?.username }}</strong>? This action cannot be undone.</p>
        <div class="modal-actions">
          <button @click="showDeleteModal = false" class="btn-secondary">Cancel</button>
          <button @click="deleteUser" class="btn-danger">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

interface User {
  username: string
  created_by?: string
  timestamp?: string
  is_self?: boolean
}

export default defineComponent({
  setup() {
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
        const res = await api.get('/api/users')
        users.value = res.data.users || []
      } catch (err: any) {
        error.value = err?.response?.data?.result || 'Failed to load users'
      } finally {
        loading.value = false
      }
    }

    function formatDate(dateStr: string) {
      try {
        return new Date(dateStr).toLocaleDateString()
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
      
      deleting.value = userToDelete.value.username
      showDeleteModal.value = false
      
      try {
        await api.delete('/api/user', { 
          data: { username: userToDelete.value.username } 
        })
        pushToast('User deleted successfully', 'success')
        await load()
      } catch (err: any) {
        pushToast(err?.response?.data?.result || 'Failed to delete user', 'error')
      } finally {
        deleting.value = ''
        userToDelete.value = null
      }
    }

    onMounted(() => {
      load()
    })

    return {
      users,
      loading,
      error,
      deleting,
      showDeleteModal,
      userToDelete,
      load,
      formatDate,
      confirmDelete,
      deleteUser
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

/* Modal */
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

/* Responsive */
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
