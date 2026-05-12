<template>
  <div class="environment-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link hide-mobile">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
        </svg>
        {{ t('back') }}
      </button>
      <div class="header-content">
        <h1>
          <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
            <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/>
          </svg>
          {{ t('environment_title') }}
        </h1>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('loading_generic') }}</p>
    </div>

    <div v-else-if="error" class="error-box">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <div v-else>
      <div class="accordion" id="envAccordion">
        <div v-for="(cat, idx) in categories" :key="cat.name" class="accordion-item">
          <h2 class="accordion-header" :id="'heading' + idx">
            <button
              class="accordion-button"
              :class="{ collapsed: idx !== 0 }"
              type="button"
              data-bs-toggle="collapse"
              :data-bs-target="'#collapse' + idx"
              :aria-expanded="idx === 0 ? 'true' : 'false'"
              :aria-controls="'collapse' + idx"
            >
              <span class="category-badge me-2">{{ cat.variables.length }}</span>
              {{ cat.name }}
            </button>
          </h2>
          <div
            :id="'collapse' + idx"
            class="accordion-collapse collapse"
            :class="{ show: idx === 0 }"
            :aria-labelledby="'heading' + idx"
            data-bs-parent="#envAccordion"
          >
            <div class="accordion-body p-0">
              <table class="table table-hover mb-0">
                <thead class="table-light">
                  <tr>
                    <th style="width: 30%">{{ t('environment_variable') }}</th>
                    <th style="width: 70%">{{ t('environment_value') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="v in cat.variables" :key="v.name">
                    <td class="font-monospace text-primary">{{ v.name }}</td>
                    <td>
                      <code v-if="v.value" class="env-value">{{ v.value }}</code>
                      <span v-else class="text-muted">{{ t('environment_not_set') }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue'
import api from '@/services/api'
import { useMasterKey } from '@/composables/useMasterKey'
import { useLocale } from '@/i18n'

interface EnvVar {
  name: string
  value: string
}

interface Category {
  name: string
  variables: EnvVar[]
}

export default defineComponent({
  name: 'Environment',
  setup() {
    const loading = ref(true)
    const error = ref('')
    const { masterKeyHeaders } = useMasterKey()
    const { t } = useLocale()
    const categories = ref<Category[]>([])

    const formatValue = (value: unknown): string => {
      if (value === null || value === undefined || value === '') return ''
      if (typeof value === 'object') return JSON.stringify(value)
      return String(value)
    }

    const buildCategories = (source: Record<string, any>): Category[] => {
      return Object.entries(source)
        .filter(([, value]) => value && typeof value === 'object' && !Array.isArray(value))
        .map(([name, value]) => ({
          name,
          variables: Object.entries(value as Record<string, any>)
            .map(([key, entry]) => ({
              name: key,
              value: formatValue(entry)
            }))
            .sort((a, b) => a.name.localeCompare(b.name))
        }))
        .filter((category) => category.variables.length > 0)
        .sort((a, b) => a.name.localeCompare(b.name))
    }

    const loadEnvironment = async () => {
      try {
        const res = await api.get('/api/system/environment', { headers: masterKeyHeaders() })
        categories.value = buildCategories(res.data?.settings || res.data?.preview || {})
      } catch (e: any) {
        error.value = e?.response?.data?.result || e.message || t('environment_error_load')
      } finally {
        loading.value = false
      }
    }

    onMounted(loadEnvironment)

    return {
      t,
      loading,
      error,
      categories
    }
  }
})
</script>

<style scoped>
.environment-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
}

.header-content h1 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.header-content h1 svg {
  color: var(--branding-primary, #7C3AED);
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #334155;
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 10px;
  padding: 6px 12px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.back-link:hover {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #312e81;
}

.loading-state {
  text-align: center;
  padding: 60px 0;
  color: #6b7280;
}

.spinner {
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
  gap: 10px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  color: #dc2626;
  margin-bottom: 20px;
}

.category-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--branding-primary, #7C3AED);
  color: white;
  font-size: 0.75rem;
  font-weight: 600;
}

.accordion-button:not(.collapsed) {
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
}

.accordion-button:not(.collapsed) .category-badge {
  background: rgba(255, 255, 255, 0.3);
}

.env-value {
  background: #f8f9fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.85rem;
  word-break: break-all;
  white-space: pre-wrap;
}

.table td {
  vertical-align: middle;
}

@media (max-width: 768px) {
  .hide-mobile {
    display: none !important;
  }

  .table {
    font-size: 0.85rem;
  }
}
</style>
