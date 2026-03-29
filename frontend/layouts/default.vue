<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <NuxtLink to="/" class="text-xl font-bold text-gray-900 tracking-tight">
            FlexShop
          </NuxtLink>
          <nav class="flex items-center gap-6">
            <NuxtLink to="/catalog" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors">
              Catalog
            </NuxtLink>
            <NuxtLink to="/cart" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors">
              Cart
            </NuxtLink>
          </nav>
        </div>
      </div>
    </header>
    <main>
      <slot />
    </main>
    <footer class="bg-white border-t border-gray-200 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <p class="text-sm text-gray-400 text-center">
          &copy; {{ new Date().getFullYear() }} FlexShop. All rights reserved.
        </p>
      </div>
    </footer>
  </div>
</template>
