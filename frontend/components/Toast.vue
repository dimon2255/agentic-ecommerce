<template>
  <Teleport to="body">
    <div class="fixed bottom-4 right-4 z-[200] flex flex-col gap-2">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          role="alert"
          class="px-4 py-3 rounded-lg shadow-lg text-sm font-medium max-w-sm border backdrop-blur-sm"
          :class="variantClasses[toast.variant]"
        >
          {{ toast.message }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useToast } from '~/composables/useToast'

const { toasts } = useToast()

const variantClasses: Record<string, string> = {
  success: 'bg-[var(--color-success-bg)] border-[var(--color-success-border)] text-[var(--color-success)]',
  error: 'bg-[var(--color-error-bg)] border-[var(--color-error-border)] text-[var(--color-error)]',
  warning: 'bg-[var(--color-warning-bg)] border-[var(--color-warning-border)] text-[var(--color-warning)]',
  info: 'bg-[var(--color-info-bg)] border-[var(--color-info-border)] text-[var(--color-info)]',
}
</script>

<style scoped>
.toast-enter-active { transition: all 0.3s ease-out; }
.toast-leave-active { transition: all 0.2s ease-in; }
.toast-enter-from { opacity: 0; transform: translateX(100%); }
.toast-leave-to { opacity: 0; transform: translateX(100%); }
</style>
