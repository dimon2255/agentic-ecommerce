<template>
  <span
    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
    :class="colorClass"
  >
    {{ label || status }}
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  status: string
  label?: string
}>()

const colorMap: Record<string, string> = {
  // Product statuses
  draft: 'bg-[var(--bg-hover)] text-muted',
  active: 'bg-[var(--color-success-bg)] text-[var(--color-success)] border border-[var(--color-success-border)]',
  archived: 'bg-[var(--color-warning-bg)] text-[var(--color-warning)] border border-[var(--color-warning-border)]',
  inactive: 'bg-[var(--bg-hover)] text-muted',
  // Order statuses
  pending: 'bg-[var(--color-info-bg)] text-[var(--color-info)] border border-[var(--color-info-border)]',
  paid: 'bg-[var(--color-success-bg)] text-[var(--color-success)] border border-[var(--color-success-border)]',
  shipped: 'bg-accent/10 text-accent border border-accent/20',
  completed: 'bg-[var(--color-success-bg)] text-[var(--color-success)] border border-[var(--color-success-border)]',
  cancelled: 'bg-[var(--color-error-bg)] text-[var(--color-error)] border border-[var(--color-error-border)]',
}

const colorClass = computed(() => colorMap[props.status] || 'bg-[var(--bg-hover)] text-muted')
</script>
