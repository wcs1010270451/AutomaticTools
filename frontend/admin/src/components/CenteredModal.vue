<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  open: boolean
  labelledby: string
  width?: string
}>(), {
  width: '480px',
})

const emit = defineEmits<{
  close: []
}>()

const panel = ref<HTMLElement | null>(null)
let previousBodyOverflow = ''

watch(
  () => props.open,
  async (open) => {
    if (open) {
      previousBodyOverflow = document.body.style.overflow
      document.body.style.overflow = 'hidden'
      await nextTick()
      panel.value?.focus()
      return
    }
    restoreBodyOverflow()
  },
)

function restoreBodyOverflow() {
  document.body.style.overflow = previousBodyOverflow
}

function handleKeydown(event: KeyboardEvent) {
  if (props.open && event.key === 'Escape') {
    emit('close')
  }
}

window.addEventListener('keydown', handleKeydown)

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
  restoreBodyOverflow()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="centered-modal">
      <div v-if="open" class="modal-overlay" @mousedown.self="emit('close')">
        <div
          ref="panel"
          class="modal-panel"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="labelledby"
          :style="{ '--modal-width': width }"
          tabindex="-1"
        >
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  z-index: 1000;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 24px;
  background: oklch(18% 0.018 165 / 0.46);
}

.modal-panel {
  width: min(var(--modal-width), calc(100vw - 48px));
  max-height: calc(100vh - 48px);
  overflow-y: auto;
  background: oklch(99% 0.003 150);
  border: 1px solid oklch(83% 0.022 158);
  border-radius: 8px;
  box-shadow: 0 22px 60px oklch(16% 0.025 165 / 0.24);
  outline: none;
}

.centered-modal-enter-active,
.centered-modal-leave-active {
  transition: opacity 180ms ease-out;
}

.centered-modal-enter-active .modal-panel,
.centered-modal-leave-active .modal-panel {
  transition: transform 180ms cubic-bezier(0.22, 1, 0.36, 1);
}

.centered-modal-enter-from,
.centered-modal-leave-to {
  opacity: 0;
}

.centered-modal-enter-from .modal-panel,
.centered-modal-leave-to .modal-panel {
  transform: translateY(10px);
}

@media (prefers-reduced-motion: reduce) {
  .centered-modal-enter-active,
  .centered-modal-leave-active,
  .centered-modal-enter-active .modal-panel,
  .centered-modal-leave-active .modal-panel {
    transition: none;
  }
}
</style>
