<template>
  <div class="flex-grow-1 pa-4 d-flex flex-column align-center">
    <v-icon :icon="icon" :color="color" />
    <span class="text-caption font-weight-bold opacity-50 mt-2 mb-0" v-text="title" />
    <span class="text-subtitle-2 font-weight-bold opacity-90"> {{ value }} {{ unit }} </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{ icon: string; color: string; title: string; value: number; unit: string }>()
</script>
