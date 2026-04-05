<template>
  <div class="card-dark p-6 space-y-2">
    <p class="text-xs font-semibold uppercase tracking-wider text-muted">{{ label }}</p>
    <div v-if="loading" class="h-8 w-24 bg-[var(--bg-hover)] rounded animate-pulse" />
    <p v-else class="text-2xl font-display font-bold text-[var(--text-primary)]">
      <span v-if="prefix" class="text-accent">{{ prefix }}</span>{{ formattedValue }}
    </p>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  label: string
  value: number | string
  prefix?: string
  loading?: boolean
  format?: 'number' | 'currency' | 'none'
}>()

const formattedValue = computed(() => {
  if (typeof props.value === 'string') return props.value
  switch (props.format) {
    case 'currency':
      return props.value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
    case 'none':
      return String(props.value)
    default:
      return props.value.toLocaleString('en-US')
  }
})
</script>
