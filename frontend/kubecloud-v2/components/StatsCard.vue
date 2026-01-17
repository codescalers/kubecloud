<template>
  <v-card class="text-center">
    <v-avatar
      :icon="icon"
      :color="color"
      variant="tonal"
      :size="50"
      rounded="lg"
    />

    <p class="text-h5 font-weight-bold mt-4 mb-1 text-no-wrap" v-text="value" />
    <p class="text-subtitle-2 text-uppercase text-accent" v-text="title" />
  </v-card>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  icon: string
  color: string
  value: string | number
}>()
</script>
