<template>
  <form class="space-y-6" @submit.prevent="$emit('submit')">
    <SlugInput
      :name="form.name"
      :slug="form.slug"
      @update:name="form.name = $event"
      @update:slug="form.slug = $event"
    />

    <div>
      <label class="block text-sm font-medium text-secondary mb-1.5">Category</label>
      <select v-model="form.category_id" class="input-dark">
        <option value="" disabled>Select a category</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">
          {{ cat.parent_id ? '\u00A0\u00A0\u2514\u00A0' : '' }}{{ cat.name }}
        </option>
      </select>
    </div>

    <div>
      <label class="block text-sm font-medium text-secondary mb-1.5">Description</label>
      <textarea
        v-model="form.description"
        class="input-dark min-h-[100px] resize-y"
        placeholder="Product description"
      />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-medium text-secondary mb-1.5">Base Price</label>
        <div class="relative">
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-muted text-sm">$</span>
          <input
            v-model.number="form.base_price"
            type="number"
            step="0.01"
            min="0"
            class="input-dark pl-7"
            placeholder="0.00"
          />
        </div>
      </div>
      <div>
        <label class="block text-sm font-medium text-secondary mb-1.5">Status</label>
        <select v-model="form.status" class="input-dark">
          <option value="draft">Draft</option>
          <option value="active">Active</option>
          <option value="archived">Archived</option>
        </select>
      </div>
    </div>

    <ImageUploader v-model="form.images" />

    <div class="flex items-center gap-3 pt-4 border-t border-[var(--border-default)]">
      <button type="submit" class="btn-accent px-6 py-2.5 rounded-lg" :disabled="saving">
        {{ saving ? 'Saving...' : submitLabel }}
      </button>
      <NuxtLink to="/admin/products" class="text-sm text-secondary hover:text-[var(--text-primary)] transition-colors">
        Cancel
      </NuxtLink>
    </div>
  </form>
</template>

<script setup lang="ts">
export interface ProductFormData {
  name: string
  slug: string
  category_id: string
  description: string
  base_price: number
  status: string
  images: string[]
}

const props = defineProps<{
  form: ProductFormData
  categories: { id: string; name: string; parent_id: string | null }[]
  saving?: boolean
  submitLabel?: string
}>()

defineEmits<{
  submit: []
}>()
</script>
