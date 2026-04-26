<template>
  <div class="position-fixed top-0 end-0 p-3" style="z-index: 1080;">
    <div v-for="t in toasts" :key="t.id" class="toast show mb-2" role="alert" aria-live="assertive" aria-atomic="true">
      <div :class="['toast-header', headerClass(t.type)]">
        <strong class="me-auto text-white">{{ t.type }}</strong>
        <small class="text-white-50">agora</small>
        <button type="button" class="btn-close btn-close-white ms-2 mb-1" @click="dismiss(t.id)"></button>
      </div>
      <div class="toast-body">
        {{ t.message }}
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { useToasts } from '@/services/toast'

export default defineComponent({
  setup() {
    const toasts = useToasts()

    function dismiss(id: number) {
      const idx = toasts.findIndex((t) => t.id === id)
      if (idx !== -1) toasts.splice(idx, 1)
    }

    function headerClass(type: string) {
      if (type === 'success') return 'bg-success'
      if (type === 'error') return 'bg-danger'
      return 'bg-info'
    }

    return { toasts, dismiss, headerClass }
  }
})
</script>

<style scoped>
.toast {
  min-width: 220px;
}
</style>
