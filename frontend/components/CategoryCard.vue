<template>
  <NuxtLink
    :to="`/catalog/${category.slug}`"
    class="group block bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
  >
    <div class="aspect-[4/3] bg-gradient-to-br from-primary-50 to-primary-100 flex items-center justify-center">
      <span class="text-4xl text-primary-300 group-hover:scale-110 transition-transform">
        {{ icon }}
      </span>
    </div>
    <div class="p-4">
      <h3 class="font-semibold text-gray-900 group-hover:text-primary-600 transition-colors">
        {{ category.name }}
      </h3>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
const props = defineProps<{
  category: { id: string; name: string; slug: string }
  icon?: string
}>()

const icon = props.icon || '+'
</script>
