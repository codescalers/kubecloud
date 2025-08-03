<template>
  <v-card color="surface-variant" style="border: 2px solid white !important;">
    <v-card-title class="text-h5 font-weight-semibold text-white pa-6 pb-4">
      <slot name="title" />
    </v-card-title>
    <v-card-text class="pa-6 pt-0 text-white">
      <slot name="content" />
    </v-card-text>
    <v-card-actions v-if="$slots.actions" class="pa-6 pt-0 justify-end">
      <slot name="actions" />
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
</script>
