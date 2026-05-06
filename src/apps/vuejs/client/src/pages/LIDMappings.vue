<template>
  <div class="lid-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link">Voltar</button>
      <h1>Consulta de mapeamentos @lid</h1>
      <p>Consulta bidirecional @lid ↔ telefone usando os resolvers do sistema.</p>
    </div>

    <div class="grid">
      <section class="card">
        <h2>@lid -> telefone</h2>
        <form @submit.prevent="lookupByLid" class="form-grid">
          <label>
            Token da sessao
            <input v-model="token" type="text" class="form-input" required />
          </label>
          <label>
            LID
            <input v-model="lid" type="text" class="form-input" placeholder="121281638842371@lid" required />
          </label>
          <button class="btn-primary" type="submit" :disabled="loadingLid">
            {{ loadingLid ? 'Consultando...' : 'Consultar telefone' }}
          </button>
        </form>
        <div v-if="lidResult" class="result-box">
          <div><strong>LID:</strong> {{ lidResult.lid || lid }}</div>
          <div><strong>Telefone:</strong> {{ lidResult.phone || '-' }}</div>
        </div>
      </section>

      <section class="card">
        <h2>telefone -> @lid</h2>
        <form @submit.prevent="lookupByPhone" class="form-grid">
          <label>
            Token da sessao
            <input v-model="token" type="text" class="form-input" required />
          </label>
          <label>
            Telefone
            <input v-model="phone" type="text" class="form-input" placeholder="+5511999999999" required />
          </label>
          <button class="btn-primary" type="submit" :disabled="loadingPhone">
            {{ loadingPhone ? 'Consultando...' : 'Consultar @lid' }}
          </button>
        </form>
        <div v-if="phoneResult" class="result-box">
          <div><strong>Telefone:</strong> {{ phoneResult.phone || phone }}</div>
          <div><strong>LID:</strong> {{ phoneResult.lid || '-' }}</div>
        </div>
      </section>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>
    <div v-if="rawResponse" class="raw-box">
      <strong>Resposta bruta</strong>
      <pre>{{ rawResponse }}</pre>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/services/api'

export default defineComponent({
  setup() {
    const route = useRoute()
    const token = ref(String(route.params.token || ''))

    const lid = ref('')
    const phone = ref('')
    const loadingLid = ref(false)
    const loadingPhone = ref(false)

    const lidResult = ref<any | null>(null)
    const phoneResult = ref<any | null>(null)
    const error = ref('')
    const rawResponse = ref('')

    const parsePayload = (data: any) => ({
      lid: data?.lid || data?.LID || data?.result?.lid || data?.result?.LID || '',
      phone: data?.phone || data?.result?.phone || '',
    })

    const lookupByLid = async () => {
      loadingLid.value = true
      error.value = ''
      try {
        const res = await api.get('/api/contacts/identifier', {
          params: {
            token: token.value,
            lid: lid.value.trim(),
          },
        })

        rawResponse.value = JSON.stringify(res.data, null, 2)
        lidResult.value = parsePayload(res.data)
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.response?.data?.message || err?.message || 'Falha ao consultar mapeamento por LID'
      } finally {
        loadingLid.value = false
      }
    }

    const lookupByPhone = async () => {
      loadingPhone.value = true
      error.value = ''
      try {
        const res = await api.get('/api/contacts/identifier', {
          params: {
            token: token.value,
            phone: phone.value.trim(),
          },
        })

        rawResponse.value = JSON.stringify(res.data, null, 2)
        phoneResult.value = parsePayload(res.data)
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.response?.data?.message || err?.message || 'Falha ao consultar mapeamento por telefone'
      } finally {
        loadingPhone.value = false
      }
    }

    return {
      token,
      lid,
      phone,
      loadingLid,
      loadingPhone,
      lidResult,
      phoneResult,
      error,
      rawResponse,
      lookupByLid,
      lookupByPhone,
    }
  },
})
</script>

<style scoped>
.lid-page { max-width: 1100px; margin: 0 auto; }
.page-header h1 { margin: 0.5rem 0; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 1rem; }
.card { background: #fff; border: 1px solid #dbe3ef; border-radius: 14px; padding: 1rem; }
.form-grid { display: grid; gap: 0.8rem; }
.form-input { width: 100%; border: 1px solid #c7d2e3; border-radius: 10px; padding: 0.65rem 0.75rem; }
.btn-primary, .back-link { border: 0; border-radius: 10px; padding: 0.6rem 0.9rem; }
.btn-primary { background: #0f766e; color: #fff; }
.back-link { background: #f3f4f6; color: #111827; margin-bottom: 0.5rem; }
.result-box { margin-top: 0.8rem; background: #ecfeff; border: 1px solid #a5f3fc; color: #164e63; padding: 0.75rem; border-radius: 10px; }
.error-box { margin-top: 1rem; background: #fee2e2; color: #991b1b; padding: 0.75rem; border-radius: 10px; }
.raw-box { margin-top: 1rem; background: #0b1220; color: #e5e7eb; padding: 0.75rem; border-radius: 10px; }
.raw-box pre { margin: 0.5rem 0 0; white-space: pre-wrap; word-break: break-word; }
</style>
