<template>
  <div v-if="show">
    <div class="modal-backdrop fade show"></div>
    <div class="modal d-block" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ resolvedTitle }}</h5>
            <button type="button" class="btn-close" :aria-label="t('close')" @click="$emit('cancel')"></button>
          </div>
          <div class="modal-body">
            <slot>{{ message }}</slot>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="$emit('cancel')">{{ resolvedCancelLabel }}</button>
            <button type="button" class="btn btn-danger" @click="$emit('confirm')">{{ resolvedConfirmLabel }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from 'vue'
import { useLocale } from '@/i18n'

export default defineComponent({
  props: {
    show: { type: Boolean, required: true },
    title: { type: String, default: '' },
    message: { type: String, default: '' },
    confirmLabel: { type: String, default: '' },
    cancelLabel: { type: String, default: '' }
  },
  emits: ['confirm', 'cancel'],
  setup(props) {
    const { t } = useLocale()

    const resolvedTitle = computed(() => props.title || t('confirm_title'))
    const resolvedConfirmLabel = computed(() => props.confirmLabel || t('confirm'))
    const resolvedCancelLabel = computed(() => props.cancelLabel || t('cancel'))

    return { t, resolvedTitle, resolvedConfirmLabel, resolvedCancelLabel }
  }
})
</script>
