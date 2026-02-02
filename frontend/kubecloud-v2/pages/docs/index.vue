<template>
  <!-- <div>
    {{ page }}
  </div> -->

  <div v-html="page?.md?.html" />
</template>

<script setup lang="ts">
defineProps<{ page: Doc }>()
</script>
