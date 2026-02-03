<template>
  <!-- <div>
    {{ page }}
  </div> -->

  <v-card>
    <!-- {{ page?.md?.attributes }} -->
    <div v-html="page?.md?.html" />
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ page: Doc }>()
</script>
