<template>
  <div class="user-create-page">
    <!-- Header -->
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile">
        <i class="fa fa-arrow-left"></i>
        Voltar
      </button>
      <div class="header-content">
        <h1>
          <i class="fa fa-user-plus"></i>
          Criar Usuário
        </h1>
        <p>Adicione um novo usuário ao sistema</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="create-card">
      <form @submit.prevent="createUser" class="create-form">
        <div v-if="error" class="error-box">
          <i class="fa fa-exclamation-triangle"></i>
          <span>{{ error }}</span>
        </div>

        <div v-if="success" class="success-box">
          <i class="fa fa-check-circle"></i>
          <span>{{ success }}</span>
        </div>

        <div class="form-group">
          <label for="email">
            <i class="fa fa-envelope"></i>
            Email
          </label>
          <input 
            id="email"
            v-model="email" 
            type="email" 
            class="form-input" 
            placeholder="usuario@exemplo.com"
            required
          />
        </div>

        <div class="form-group">
          <label for="password">
            <i class="fa fa-lock"></i>
            Senha
          </label>
          <div class="password-wrapper">
            <input 
              id="password"
              v-model="password" 
              :type="showPassword ? 'text' : 'password'" 
              class="form-input" 
              placeholder="••••••••"
              required
            />
            <button 
              type="button" 
              class="toggle-password" 
              @click="showPassword = !showPassword"
            >
              <i :class="showPassword ? 'fa fa-eye-slash' : 'fa fa-eye'"></i>
            </button>
          </div>
          <div class="password-strength" v-if="password">
            <div class="strength-bar" :class="passwordStrengthClass" :style="{ width: passwordStrength + '%' }"></div>
          </div>
        </div>

        <div class="form-group">
          <label for="confirmPassword">
            <i class="fa fa-lock"></i>
            Confirmar Senha
          </label>
          <input 
            id="confirmPassword"
            v-model="confirmPassword" 
            type="password" 
            class="form-input" 
            placeholder="••••••••"
            required
          />
          <small v-if="confirmPassword && password !== confirmPassword" class="error-hint">
            As senhas não coincidem
          </small>
        </div>

        <div class="form-actions">
          <router-link to="/account" class="btn-secondary">
            Cancelar
          </router-link>
          <button 
            type="submit" 
            class="btn-primary" 
            :disabled="loading || !isFormValid"
          >
            <i v-if="loading" class="fa fa-spinner fa-spin"></i>
            <i v-else class="fa fa-user-plus"></i>
            {{ loading ? 'Criando...' : 'Criar Usuário' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from 'vue'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

export default defineComponent({
  setup() {
    const email = ref('')
    const password = ref('')
    const confirmPassword = ref('')
    const showPassword = ref(false)
    const loading = ref(false)
    const error = ref('')
    const success = ref('')

    const passwordStrength = computed(() => {
      const pwd = password.value
      if (!pwd) return 0
      let strength = 0
      if (pwd.length >= 6) strength += 25
      if (pwd.length >= 8) strength += 25
      if (/[A-Z]/.test(pwd)) strength += 25
      if (/[0-9]/.test(pwd) || /[^A-Za-z0-9]/.test(pwd)) strength += 25
      return strength
    })

    const passwordStrengthClass = computed(() => {
      if (passwordStrength.value < 50) return 'weak'
      if (passwordStrength.value < 75) return 'medium'
      return 'strong'
    })

    const isFormValid = computed(() => {
      return email.value && 
             password.value && 
             password.value.length >= 4 && 
             password.value === confirmPassword.value
    })

    async function createUser() {
      if (!isFormValid.value) return

      loading.value = true
      error.value = ''
      success.value = ''

      try {
        await api.post('/api/user', {
          email: email.value,
          password: password.value
        })
        success.value = 'Usuário criado com sucesso!'
        pushToast('Usuário criado com sucesso!', 'success')
        
        // Clear form
        email.value = ''
        password.value = ''
        confirmPassword.value = ''
      } catch (err: any) {
        const msg = err?.response?.data?.result || err.message || 'Erro ao criar usuário'
        error.value = msg
        pushToast(msg, 'error')
      } finally {
        loading.value = false
      }
    }

    return { 
      email, password, confirmPassword, showPassword, loading, error, success,
      passwordStrength, passwordStrengthClass, isFormValid,
      createUser
    }
  }
})
</script>

<style scoped>
.user-create-page {
  max-width: 500px;
  margin: 0 auto;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  text-decoration: none;
  font-size: 14px;
  margin-bottom: 16px;
}

.back-link:hover {
  color: #374151;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px;
}

.page-header h1 i {
  color: var(--branding-primary, #7C3AED);
}

.page-header p {
  color: #6b7280;
  margin: 0;
}

.create-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.error-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 10px;
  color: #dc2626;
  font-size: 14px;
}

.success-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #dcfce7;
  border: 1px solid #bbf7d0;
  border-radius: 10px;
  color: #16a34a;
  font-size: 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.form-group label i {
  color: #9ca3af;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 16px;
  transition: all 0.2s;
  background: #f9fafb;
}

.form-input:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
  background: white;
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.1);
}

.password-wrapper {
  position: relative;
}

.password-wrapper .form-input {
  padding-right: 48px;
}

.toggle-password {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  padding: 8px;
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
}

.toggle-password:hover {
  color: #6b7280;
}

.password-strength {
  height: 4px;
  background: #e5e7eb;
  border-radius: 2px;
  overflow: hidden;
}

.strength-bar {
  height: 100%;
  transition: all 0.3s;
}

.strength-bar.weak { background: #ef4444; }
.strength-bar.medium { background: #f59e0b; }
.strength-bar.strong { background: #10b981; }

.error-hint {
  font-size: 12px;
  color: #dc2626;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 8px;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.25);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e5e7eb;
}
</style>
