<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60" @click="$emit('cancel')" />
        <div class="relative glass-strong rounded-xl p-6 w-full max-w-sm animate-scale-in space-y-4">
          <h3 class="text-lg font-display font-semibold text-[var(--text-primary)]">{{ title }}</h3>
          <p class="text-sm text-secondary">{{ message }}</p>
          <div class="flex justify-end gap-3 pt-2">
            <button
              class="px-4 py-2 text-sm rounded-lg border border-[var(--border-default)] text-secondary hover:text-[var(--text-primary)] hover:border-[var(--border-strong)] transition-colors"
              @click="$emit('cancel')"
            >
              Cancel
            </button>
            <button
              class="px-4 py-2 text-sm rounded-lg font-medium transition-colors"
              :class="variant === 'danger'
                ? 'bg-[var(--color-error)] text-white hover:bg-[var(--color-error)]/80'
                : 'btn-accent'"
              @click="$emit('confirm')"
            >
              {{ confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  open: boolean
  title: string
  message: string
  confirmText?: string
  variant?: 'danger' | 'default'
}>(), {
  confirmText: 'Confirm',
  variant: 'default',
})

defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<style scoped>
.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.2s ease;
}
.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}
</style>
