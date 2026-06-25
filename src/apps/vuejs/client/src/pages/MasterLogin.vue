<template>
  <div class="master-login-page">
    <div class="master-login-card">
      <div class="master-icon">
        <i class="fa fa-shield-alt"></i>
      </div>
      <h1>{{ t('master_login_title') }}</h1>
      <p class="master-subtitle">{{ t('master_login_subtitle') }}</p>

      <form @submit.prevent="submit" class="master-form">
        <div v-if="error" class="error-box">
          <i class="fa fa-exclamation-triangle"></i>
          <span>{{ error }}</span>
        </div>

        <div class="form-group">
          <label for="master-key">
            <i class="fa fa-key"></i>
            {{ t('master_key_label') }}
          </label>
          <div class="password-wrapper">
            <input
              id="master-key"
              v-model="keyInput"
              :type="showKey ? 'text' : 'password'"
              class="form-input"
              :placeholder="t('master_key_placeholder')"
              autocomplete="off"
              required
            />
            <button type="button" class="toggle-password" @click="showKey = !showKey" tabindex="-1">
              <i :class="showKey ? 'fa fa-eye-slash' : 'fa fa-eye'"></i>
            </button>
          </div>
        </div>

        <button type="submit" class="btn-master" :disabled="verifying">
          <i v-if="verifying" class="fa fa-spinner fa-spin"></i>
          <i v-else class="fa fa-unlock-alt"></i>
          {{ verifying ? t('master_verifying') : t('master_enter') }}
        </button>

        <button type="button" class="btn-cancel" @click="$router.back()">
          {{ t('cancel') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMasterKey } from '@/composables/useMasterKey'
import { useLocale } from '@/i18n'

export default defineComponent({
  setup() {
    const router = useRouter()
    const route = useRoute()
    const { verifying, verifyMasterKey } = useMasterKey()
    const { t } = useLocale()

    const keyInput = ref('')
    const showKey = ref(false)
    const error = ref('')

    async function submit() {
      error.value = ''
      if (!keyInput.value.trim()) return

      const ok = await verifyMasterKey(keyInput.value.trim())
      if (ok) {
        const redirect = (route.query.redirect as string) || '/users'
        router.push(redirect)
      } else {
        error.value = t('master_invalid_key')
        keyInput.value = ''
      }
    }

    return { keyInput, showKey, verifying, error, submit, t }
  }
})
</script>

<style scoped>
.master-login-page {
  min-height: 80vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
}

.master-login-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  padding: 2.5rem 2rem;
  width: 100%;
  max-width: 420px;
  text-align: center;
}

.master-icon {
  width: 72px;
  height: 72px;
  background: linear-gradient(135deg, #7C3AED22, #7C3AED44);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.25rem;
  font-size: 2rem;
  color: var(--branding-primary, #7C3AED);
}

h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #111827;
  margin: 0 0 0.4rem;
}

.master-subtitle {
  color: #6b7280;
  font-size: 0.9rem;
  margin: 0 0 1.75rem;
}

.master-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  text-align: left;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 6px;
}

.password-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.form-input {
  width: 100%;
  padding: 10px 40px 10px 12px;
  border: 1.5px solid #e5e7eb;
  border-radius: 8px;
  font-size: 1rem;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.form-input:focus {
  border-color: var(--branding-primary, #7C3AED);
}

.toggle-password {
  position: absolute;
  right: 10px;
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
  padding: 0;
  font-size: 0.95rem;
}

.btn-master {
  width: 100%;
  padding: 11px;
  background: var(--branding-primary, #7C3AED);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: opacity 0.2s;
}

.btn-master:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-master:not(:disabled):hover {
  opacity: 0.9;
}

.btn-cancel {
  width: 100%;
  padding: 10px;
  background: transparent;
  color: #6b7280;
  border: 1.5px solid #e5e7eb;
  border-radius: 8px;
  font-size: 0.95rem;
  cursor: pointer;
  transition: border-color 0.2s;
}

.btn-cancel:hover {
  border-color: #9ca3af;
  color: #374151;
}

.error-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 0.875rem;
}
</style>
