<template>
  <NuxtLink
    :to="`/catalog/${category.slug}`"
    class="group block card-dark glow-hover overflow-hidden"
  >
    <div class="aspect-[4/3] relative overflow-hidden" :class="gradientClass">
      <div class="absolute inset-0 bg-black/20 group-hover:bg-black/10 transition-colors duration-500"></div>
      <div class="absolute inset-0 flex items-center justify-center">
        <span class="text-7xl font-display font-bold text-white/10 group-hover:text-white/20 group-hover:scale-110 transition-all duration-500 select-none">
          {{ category.name.charAt(0) }}
        </span>
      </div>
      <div class="absolute bottom-3 left-3 right-3">
        <span class="inline-block px-2.5 py-1 bg-white/10 backdrop-blur-md rounded-md text-xs text-white/70 font-medium">
          {{ label }}
        </span>
      </div>
    </div>
    <div class="p-4">
      <h3 class="font-display font-semibold text-[var(--text-primary)] group-hover:text-accent transition-colors">
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
  'bg-gradient-to-br from-amber-700/60 to-orange-900/60',
  'bg-gradient-to-br from-emerald-700/60 to-teal-900/60',
  'bg-gradient-to-br from-rose-700/60 to-pink-900/60',
  'bg-gradient-to-br from-violet-700/60 to-indigo-900/60',
  'bg-gradient-to-br from-cyan-700/60 to-blue-900/60',
  'bg-gradient-to-br from-fuchsia-700/60 to-purple-900/60',
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
