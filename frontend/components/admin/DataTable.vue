<template>
  <div class="card-dark overflow-hidden">
    <!-- Toolbar slot -->
    <div v-if="$slots.toolbar" class="px-6 py-4 border-b border-[var(--border-default)]">
      <slot name="toolbar" />
    </div>

    <!-- Table -->
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[var(--border-default)]">
            <th
              v-for="col in columns"
              :key="col.key"
              class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted"
              :class="{ 'cursor-pointer hover:text-secondary': col.sortable }"
              @click="col.sortable && toggleSort(col.key)"
            >
              <span class="flex items-center gap-1">
                {{ col.label }}
                <span v-if="col.sortable && sortBy === col.key" class="text-accent">
                  {{ sortDir === 'asc' ? '\u2191' : '\u2193' }}
                </span>
              </span>
            </th>
            <th v-if="$slots.actions" class="px-6 py-3 w-20" />
          </tr>
        </thead>
        <tbody v-if="loading">
          <tr v-for="i in perPage" :key="i">
            <td v-for="col in columns" :key="col.key" class="px-6 py-4">
              <div class="h-4 bg-[var(--bg-hover)] rounded animate-pulse" />
            </td>
            <td v-if="$slots.actions" class="px-6 py-4">
              <div class="h-4 w-8 bg-[var(--bg-hover)] rounded animate-pulse" />
            </td>
          </tr>
        </tbody>
        <tbody v-else-if="rows.length === 0">
          <tr>
            <td :colspan="columns.length + ($slots.actions ? 1 : 0)" class="px-6 py-12 text-center text-muted">
              {{ emptyText }}
            </td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr
            v-for="(row, idx) in rows"
            :key="row.id || idx"
            class="border-b border-[var(--border-subtle)] hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
            @click="$emit('row-click', row)"
          >
            <td v-for="col in columns" :key="col.key" class="px-6 py-4">
              <slot :name="`cell-${col.key}`" :row="row" :value="row[col.key]">
                {{ row[col.key] }}
              </slot>
            </td>
            <td v-if="$slots.actions" class="px-6 py-4">
              <slot name="actions" :row="row" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="px-6 py-3 flex items-center justify-between border-t border-[var(--border-default)]">
      <span class="text-xs text-muted">
        {{ total }} result{{ total !== 1 ? 's' : '' }}
      </span>
      <div class="flex items-center gap-2">
        <button
          :disabled="page <= 1"
          class="px-3 py-1 text-xs rounded border border-[var(--border-default)] text-secondary hover:text-[var(--text-primary)] hover:border-[var(--border-strong)] disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          @click="$emit('update:page', page - 1)"
        >
          Prev
        </button>
        <span class="text-xs text-muted">{{ page }} / {{ totalPages }}</span>
        <button
          :disabled="page >= totalPages"
          class="px-3 py-1 text-xs rounded border border-[var(--border-default)] text-secondary hover:text-[var(--text-primary)] hover:border-[var(--border-strong)] disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          @click="$emit('update:page', page + 1)"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface Column {
  key: string
  label: string
  sortable?: boolean
}

const props = withDefaults(defineProps<{
  columns: Column[]
  rows: any[]
  loading?: boolean
  total?: number
  page?: number
  perPage?: number
  totalPages?: number
  emptyText?: string
}>(), {
  loading: false,
  total: 0,
  page: 1,
  perPage: 20,
  totalPages: 0,
  emptyText: 'No data found',
})

const emit = defineEmits<{
  'update:page': [page: number]
  'sort': [key: string, dir: string]
  'row-click': [row: any]
}>()

const sortBy = ref('')
const sortDir = ref<'asc' | 'desc'>('asc')

function toggleSort(key: string) {
  if (sortBy.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = key
    sortDir.value = 'asc'
  }
  emit('sort', sortBy.value, sortDir.value)
}
</script>
