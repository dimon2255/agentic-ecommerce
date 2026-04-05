<template>
  <form class="space-y-6" @submit.prevent="$emit('submit')">
    <SlugInput
      :name="form.name"
      :slug="form.slug"
      @update:name="form.name = $event"
      @update:slug="form.slug = $event"
    />

    <div>
      <label class="block text-sm font-medium text-secondary mb-1.5">Parent Category</label>
      <select v-model="form.parent_id" class="input-dark">
        <option :value="null">None (root category)</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">
          {{ cat.name }}
        </option>
      </select>
    </div>

    <div class="flex items-center gap-3 pt-4 border-t border-[var(--border-default)]">
      <button type="submit" class="btn-accent px-6 py-2.5 rounded-lg" :disabled="saving">
        {{ saving ? 'Saving...' : submitLabel }}
      </button>
      <NuxtLink to="/admin/categories" class="text-sm text-secondary hover:text-[var(--text-primary)] transition-colors">
        Cancel
      </NuxtLink>
    </div>
  </form>
</template>

<script setup lang="ts">
export interface CategoryFormData {
  name: string
  slug: string
  parent_id: string | null
}

defineProps<{
  form: CategoryFormData
  categories: { id: string; name: string }[]
  saving?: boolean
  submitLabel?: string
}>()

defineEmits<{
  submit: []
}>()
</script>
