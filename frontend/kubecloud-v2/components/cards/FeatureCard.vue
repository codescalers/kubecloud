<template>
  <div
    v-tilt-effect
    class="h-100"
  >
    <v-card class="text-center h-100">
      <v-icon color="primary" :icon="icon" size="50" />

      <p class="text-h5 font-weight-bold mt-4 mb-1 text-no-wrap" v-text="title" />
      <p class="text-body-1 text-accent" v-text="description" />
    </v-card>
  </div>
</template>

<script setup lang="ts">
defineProps<{ icon: string, title: string, description: string }>()
</script>
