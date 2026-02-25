<template>
  <div class="environment-page">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2>Environment Variables</h2>
      <button @click="$router.back()" class="back-link hide-mobile">
        <i class="fa fa-arrow-left me-2"></i> Back
      </button>
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <div v-else-if="error" class="alert alert-danger">
      {{ error }}
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
                    <th style="width: 30%">Variable</th>
                    <th style="width: 35%">Value</th>
                    <th style="width: 35%">Description</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="v in cat.variables" :key="v.name">
                    <td class="font-monospace text-primary">{{ v.name }}</td>
                    <td>
                      <code v-if="v.value" class="env-value">{{ v.value }}</code>
                      <span v-else class="text-muted">Not set</span>
                    </td>
                    <td class="text-muted small">{{ v.description }}</td>
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
import { RouterLink } from 'vue-router'
import api from '@/services/api'

interface EnvVar {
  name: string
  value: string
  description: string
}

interface Category {
  name: string
  variables: EnvVar[]
}

export default defineComponent({
  name: 'Environment',
  components: { RouterLink },
  setup() {
    const loading = ref(true)
    const error = ref('')
    const categories = ref<Category[]>([])

    const loadEnvironment = async () => {
      try {
        const res = await api.get('/api/environment')
        if (res.data?.categories) {
          categories.value = res.data.categories
        }
      } catch (e: any) {
        error.value = e.message || 'Failed to load environment variables'
      } finally {
        loading.value = false
      }
    }

    onMounted(loadEnvironment)

    return {
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
}

.table td {
  vertical-align: middle;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .hide-mobile {
    display: none !important;
  }

  .table {
    font-size: 0.85rem;
  }

  .table th:nth-child(3),
  .table td:nth-child(3) {
    display: none;
  }

  .table th:nth-child(1),
  .table td:nth-child(1) {
    width: 40%;
  }

  .table th:nth-child(2),
  .table td:nth-child(2) {
    width: 60%;
  }
}
</style>
