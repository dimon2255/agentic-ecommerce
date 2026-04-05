<template>
  <div class="space-y-4">
    <div>
      <label class="block text-sm font-medium text-secondary mb-1.5">Name</label>
      <input
        :value="name"
        class="input-dark"
        placeholder="Product name"
        @input="onNameInput(($event.target as HTMLInputElement).value)"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-secondary mb-1.5">Slug</label>
      <div class="flex gap-2">
        <input
          :value="slug"
          class="input-dark"
          placeholder="product-slug"
          @input="$emit('update:slug', ($event.target as HTMLInputElement).value); autoSlug = false"
        />
        <button
          type="button"
          class="shrink-0 px-3 py-2 text-xs rounded-lg border border-[var(--border-default)] text-secondary hover:text-accent hover:border-accent/30 transition-colors"
          title="Auto-generate from name"
          @click="regenerateSlug"
        >
          Auto
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  name: string
  slug: string
}>()

const emit = defineEmits<{
  'update:name': [value: string]
  'update:slug': [value: string]
}>()

const autoSlug = ref(true)

function slugify(text: string): string {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

function onNameInput(value: string) {
  emit('update:name', value)
  if (autoSlug.value) {
    emit('update:slug', slugify(value))
  }
}

function regenerateSlug() {
  autoSlug.value = true
  emit('update:slug', slugify(props.name))
}
</script>
