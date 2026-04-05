<template>
  <div class="space-y-2">
    <div
      v-for="bar in normalizedBars"
      :key="bar.label"
      class="flex items-center gap-3 text-sm"
    >
      <span class="w-20 text-xs text-muted truncate text-right shrink-0">{{ bar.label }}</span>
      <div class="flex-1 h-6 bg-[var(--bg-surface)] rounded overflow-hidden">
        <div
          class="h-full rounded transition-all duration-300"
          :class="barColor"
          :style="{ width: `${bar.percent}%` }"
        />
      </div>
      <span class="w-20 text-xs text-secondary shrink-0">{{ bar.formatted }}</span>
    </div>
    <p v-if="bars.length === 0" class="text-sm text-muted text-center py-4">No data</p>
  </div>
</template>

<script setup lang="ts">
export interface BarData {
  label: string
  value: number
  formatted?: string
}

const props = withDefaults(defineProps<{
  bars: BarData[]
  color?: 'accent' | 'success' | 'info'
}>(), {
  color: 'accent',
})

const colorMap: Record<string, string> = {
  accent: 'bg-accent/70',
  success: 'bg-[var(--color-success)]/70',
  info: 'bg-[var(--color-info)]/70',
}

const barColor = computed(() => colorMap[props.color] || colorMap.accent)

const normalizedBars = computed(() => {
  const max = Math.max(...props.bars.map(b => b.value), 1)
  return props.bars.map(b => ({
    ...b,
    percent: (b.value / max) * 100,
    formatted: b.formatted ?? String(b.value),
  }))
})
</script>
