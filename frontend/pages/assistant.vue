<template>
  <div />
</template>

<script setup lang="ts">
const { open } = useAssistantPanel()
const router = useRouter()

onMounted(() => {
  router.replace('/')
  nextTick(() => open())
})
</script>
