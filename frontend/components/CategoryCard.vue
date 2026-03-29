<template>
  <NuxtLink
    :to="`/catalog/${category.slug}`"
    class="group block bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
  >
    <div class="aspect-[4/3] flex items-center justify-content-center relative overflow-hidden" :class="gradientClass">
      <div class="absolute inset-0 flex items-center justify-center">
        <span class="text-6xl font-bold text-white/30 group-hover:scale-110 transition-transform select-none">
          {{ category.name.charAt(0) }}
        </span>
      </div>
      <div class="absolute bottom-3 left-3 right-3">
        <span class="inline-block px-2 py-1 bg-white/20 backdrop-blur-sm rounded text-xs text-white/80 font-medium">
          {{ label }}
        </span>
      </div>
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
}>()

const gradients = [
  'bg-gradient-to-br from-blue-500 to-indigo-600',
  'bg-gradient-to-br from-emerald-500 to-teal-600',
  'bg-gradient-to-br from-orange-500 to-rose-600',
  'bg-gradient-to-br from-violet-500 to-purple-600',
  'bg-gradient-to-br from-cyan-500 to-blue-600',
  'bg-gradient-to-br from-pink-500 to-fuchsia-600',
]

const hash = props.category.name.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0)
const gradientClass = gradients[hash % gradients.length]

const labels: Record<string, string> = {
  electronics: 'Gadgets & Tech',
  clothing: 'Fashion & Apparel',
  laptops: 'Portable Computing',
  't-shirts': 'Casual Wear',
}
const label = labels[props.category.slug] || 'Browse Collection'
</script>
