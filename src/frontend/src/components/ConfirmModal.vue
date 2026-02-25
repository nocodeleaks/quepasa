<template>
  <div v-if="show">
    <div class="modal-backdrop fade show"></div>
    <div class="modal d-block" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ title }}</h5>
            <button type="button" class="btn-close" aria-label="Close" @click="$emit('cancel')"></button>
          </div>
          <div class="modal-body">
            <slot>{{ message }}</slot>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="$emit('cancel')">{{ cancelLabel }}</button>
            <button type="button" class="btn btn-danger" @click="$emit('confirm')">{{ confirmLabel }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  props: {
    show: { type: Boolean, required: true },
    title: { type: String, default: 'Confirm' },
    message: { type: String, default: '' },
    confirmLabel: { type: String, default: 'Confirm' },
    cancelLabel: { type: String, default: 'Cancel' }
  },
  emits: ['confirm', 'cancel']
})
</script>
