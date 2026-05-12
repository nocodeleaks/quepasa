<template>
  <div class="tri-state-toggle" :class="stateClass" :disabled="disabled">
    <button 
      class="tri-btn off" 
      :class="{ active: modelValue === -1 }" 
      @click="setValue(-1)"
      :disabled="disabled"
      :title="t('tristate_off_short')"
    >
      <i class="fa fa-times"></i>
    </button>
    <button 
      class="tri-btn default" 
      :class="{ active: modelValue === 0 }" 
      @click="setValue(0)"
      :disabled="disabled"
      :title="t('tristate_default_short')"
    >
      <i class="fa fa-minus"></i>
    </button>
    <button 
      class="tri-btn on" 
      :class="{ active: modelValue === 1 }" 
      @click="setValue(1)"
      :disabled="disabled"
      :title="t('tristate_on_short')"
    >
      <i class="fa fa-check"></i>
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'
import { useLocale } from '@/i18n'

export default defineComponent({
  name: 'TriStateToggle',
  props: {
    modelValue: {
      type: Number,
      required: true,
      validator: (v: number) => [-1, 0, 1].includes(v)
    },
    disabled: {
      type: Boolean,
      default: false
    }
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
    const { t } = useLocale()

    const stateClass = computed(() => {
      if (props.disabled) return 'disabled'
      if (props.modelValue === -1) return 'state-off'
      if (props.modelValue === 1) return 'state-on'
      return 'state-unset'
    })

    function setValue(val: number) {
      if (props.disabled) return
      emit('update:modelValue', val)
      emit('change', val)
    }
    return { setValue, stateClass, t }
  }
})
</script>

<style scoped>
.tri-state-toggle {
  display: inline-flex;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e5e7eb;
  background: #f9fafb;
}

.tri-state-toggle.disabled {
  opacity: 0.6;
  pointer-events: none;
}

/* State-based border colors */
.tri-state-toggle.state-off {
  border-color: #fca5a5;
}

.tri-state-toggle.state-on {
  border-color: #86efac;
}

.tri-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 32px;
  border: none;
  background: transparent;
  color: #9ca3af;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tri-btn:hover:not(:disabled) {
  background: #e5e7eb;
}

.tri-btn:not(:last-child) {
  border-right: 1px solid #e5e7eb;
}

/* Active state: OFF (red) */
.tri-btn.off.active {
  background: #fee2e2;
  color: #dc2626;
}

/* Active state: Default/Unset (no color - neutral gray) */
.tri-btn.default.active {
  background: #f3f4f6;
  color: #6b7280;
}

/* Active state: ON (green) */
.tri-btn.on.active {
  background: #dcfce7;
  color: #16a34a;
}

.tri-btn i {
  font-size: 12px;
}

html[data-theme='dark'] .tri-state-toggle {
  border-color: rgba(71, 85, 105, 0.42);
  background: rgba(15, 23, 42, 0.92);
  box-shadow: inset 0 0 0 1px rgba(2, 6, 23, 0.22);
}

html[data-theme='dark'] .tri-state-toggle.state-off {
  border-color: rgba(248, 113, 113, 0.55);
}

html[data-theme='dark'] .tri-state-toggle.state-on {
  border-color: rgba(74, 222, 128, 0.48);
}

html[data-theme='dark'] .tri-btn {
  color: #64748b;
}

html[data-theme='dark'] .tri-btn:hover:not(:disabled) {
  background: rgba(30, 41, 59, 0.96);
  color: #cbd5e1;
}

html[data-theme='dark'] .tri-btn:not(:last-child) {
  border-right-color: rgba(71, 85, 105, 0.34);
}

html[data-theme='dark'] .tri-btn.off.active {
  background: rgba(127, 29, 29, 0.42);
  color: #fca5a5;
}

html[data-theme='dark'] .tri-btn.default.active {
  background: rgba(51, 65, 85, 0.84);
  color: #e2e8f0;
}

html[data-theme='dark'] .tri-btn.on.active {
  background: rgba(20, 83, 45, 0.46);
  color: #86efac;
}
</style>
