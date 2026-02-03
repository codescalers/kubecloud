<template>
  <v-card>
    <div v-html="page?.md?.html" />
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ page: Doc }>()
</script>
