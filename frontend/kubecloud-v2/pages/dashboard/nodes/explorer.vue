<template>
  <div>
    <v-btn @click="$router.back()">Back</v-btn>
    <div>Dashboard Nodes Explorer</div>
  </div>
</template>

<script setup lang="ts">
const { drawer } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onUnmounted(drawer.open)
</script>
