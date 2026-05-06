<template>
  <div class="lid-page">
    <div class="page-header">
      <button @click="$router.back()" class="back-link">{{ t('back') }}</button>
      <h1>{{ t('lid_direct_title') }}</h1>
      <p>{{ t('lid_direct_subtitle') }}</p>
    </div>

    <div class="card">
      <form @submit.prevent="sendDirectLid" class="form-grid">
        <label>
          {{ t('lid_direct_recipient_label') }}
          <input v-model="chatid" type="text" class="form-input" :placeholder="t('lid_direct_recipient_placeholder')" required />
        </label>

        <label>
          {{ t('lid_direct_text_label') }}
          <textarea v-model="text" class="form-textarea" rows="5" required></textarea>
        </label>

        <label>
          {{ t('lid_direct_inreply_label') }}
          <input v-model="inreply" type="text" class="form-input" />
        </label>

        <label>
          {{ t('lid_direct_trackid_label') }}
          <input v-model="trackid" type="text" class="form-input" />
        </label>

        <div class="actions">
          <button class="btn-primary" type="submit" :disabled="sending">
            {{ sending ? t('lid_direct_sending') : t('lid_direct_send_button') }}
          </button>
          <RouterLink :to="`/server/${encodeURIComponent(token)}/lid/mappings`" class="btn-secondary">
            {{ t('lid_direct_open_mappings') }}
          </RouterLink>
        </div>
      </form>

      <div v-if="success" class="success-box">
        <strong>{{ t('lid_direct_success_title') }}</strong>
        <div>{{ t('lid_direct_message_id_label') }}: {{ success.id }}</div>
        <div>{{ t('lid_direct_chatid_label') }}: {{ success.chatid }}</div>
        <div>{{ t('lid_direct_trackid_label_short') }}: {{ success.trackid || '-' }}</div>
      </div>

      <div v-if="error" class="error-box">{{ error }}</div>

      <div v-if="rawResponse" class="raw-box">
        <strong>{{ t('raw_response') }}</strong>
        <pre>{{ rawResponse }}</pre>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useLocale } from '@/i18n'
import api from '@/services/api'

export default defineComponent({
  setup() {
    const route = useRoute()
    const { t } = useLocale()
    const token = ref(String(route.params.token || ''))
    const chatid = ref('')
    const text = ref('')
    const inreply = ref('')
    const trackid = ref('')
    const sending = ref(false)
    const error = ref('')
    const rawResponse = ref('')
    const success = ref<any | null>(null)

    const sendDirectLid = async () => {
      sending.value = true
      error.value = ''
      success.value = null
      rawResponse.value = ''

      try {
        const payload: Record<string, string> = {
          token: token.value,
          chatid: chatid.value.trim(),
          text: text.value,
        }

        if (inreply.value.trim()) payload.inreply = inreply.value.trim()
        if (trackid.value.trim()) payload.trackid = trackid.value.trim()

        const res = await api.post('/api/messages/lid/direct', payload)
        rawResponse.value = JSON.stringify(res.data, null, 2)

        const result = res.data?.result || res.data?.data || {}
        success.value = {
          id: result?.id || '',
          chatid: result?.chatid || chatid.value,
          trackid: result?.trackid || trackid.value,
        }
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.response?.data?.message || err?.message || t('lid_direct_error_send')
      } finally {
        sending.value = false
      }
    }

    return {
      t,
      token,
      chatid,
      text,
      inreply,
      trackid,
      sending,
      error,
      rawResponse,
      success,
      sendDirectLid,
    }
  },
})
</script>

<style scoped>
.lid-page { max-width: 980px; margin: 0 auto; }
.page-header h1 { margin: 0.5rem 0; }
.card { background: #fff; border: 1px solid #dbe3ef; border-radius: 14px; padding: 1rem; }
.form-grid { display: grid; gap: 0.85rem; }
.form-input, .form-textarea { width: 100%; border: 1px solid #c7d2e3; border-radius: 10px; padding: 0.65rem 0.75rem; }
.actions { display: flex; gap: 0.7rem; flex-wrap: wrap; }
.btn-primary, .btn-secondary, .back-link { border: 0; border-radius: 10px; padding: 0.6rem 0.9rem; text-decoration: none; }
.btn-primary { background: #0f766e; color: #fff; }
.btn-secondary { background: #e2e8f0; color: #111827; }
.back-link { background: #f3f4f6; color: #111827; }
.error-box { margin-top: 0.8rem; background: #fee2e2; color: #991b1b; padding: 0.75rem; border-radius: 10px; }
.success-box { margin-top: 0.8rem; background: #dcfce7; color: #14532d; padding: 0.75rem; border-radius: 10px; }
.raw-box { margin-top: 0.8rem; background: #0b1220; color: #e5e7eb; padding: 0.75rem; border-radius: 10px; }
.raw-box pre { margin: 0.5rem 0 0; white-space: pre-wrap; word-break: break-word; }
</style>
