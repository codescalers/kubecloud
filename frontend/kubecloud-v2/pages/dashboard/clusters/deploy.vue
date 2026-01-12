<template>
  <div>Deploy cluster</div>
</template>

<script setup lang="ts">
const { drawer, container } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onUnmounted(drawer.open)

onMounted(container.fluidize)
onUnmounted(container.containerize)
</script>
