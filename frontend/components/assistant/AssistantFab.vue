<template>
  <button
    @click="toggle"
    :aria-label="isOpen ? 'Close shopping assistant' : 'Open shopping assistant'"
    :aria-expanded="isOpen"
    class="fixed bottom-6 right-6 z-[55] w-14 h-14 rounded-full btn-accent flex items-center justify-center shadow-lg transition-all duration-300 hover:scale-110 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-surface-base"
  >
    <!-- Chat icon when closed -->
    <svg v-if="!isOpen" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
      <path stroke-linecap="round" stroke-linejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" />
    </svg>
    <!-- Close icon when open -->
    <svg v-else class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
      <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
    <!-- Unread indicator -->
    <span
      v-if="hasUnread && !isOpen"
      class="absolute -top-0.5 -right-0.5 w-3.5 h-3.5 bg-[var(--color-error-border)] rounded-full border-2 border-surface-base animate-pulse"
    />
  </button>
</template>

<script setup lang="ts">
const { isOpen, hasUnread, toggle } = useAssistantPanel()
</script>
