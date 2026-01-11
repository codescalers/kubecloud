<template>
  <v-card>
    <div class="d-flex align-center justify-center">
      <v-avatar
        :icon="icon"
        :color="color"
        variant="tonal"
        :size="50"
        rounded="lg"
      />
    </div>

    <p class="text-center text-h4 font-weight-bold mt-4 mb-1 text-no-wrap" v-text="value" />
    <p class="text-center text-subtitle-2 text-uppercase opacity-50" v-text="title" />
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
