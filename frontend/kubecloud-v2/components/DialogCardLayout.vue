<template>
  <v-card :style="{ padding: '0 !important' }">
    <v-card-title class="pa-8 d-flex align-center justify-space-between border-b">
      <div>
        <div class="d-flex ga-2 align-baseline">
          <v-icon v-if="icon" :icon="icon" size="small" :color="iconColor ?? 'success'" />
          <slot v-else name="icon" />

          <span v-if="title" class="text-h5 font-weight-bold" v-text="title" />
          <slot v-else name="title" />
        </div>
        <p v-if="description" class="text-subtitle-2 opacity-50" v-text="description" />
        <slot v-else name="description" />
      </div>

      <v-btn
        icon
        size="small"
        variant="plain"
        :style="{ borderRadius: '50% !important' }"
        @click="$emit('cancel')"
      >
        <v-icon icon="mdi-close" size="large" />
      </v-btn>
    </v-card-title>

    <slot v-if="$slots.outer" name="outer" />

    <v-card-text v-if="$slots.default">
      <slot />
    </v-card-text>

    <v-divider v-if="$slots.actions" />
    <v-card-actions v-if="$slots.actions" class="px-6 py-4 flex-row-reverse justify-start">
      <slot name="actions" />
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ title?: string, description?: string, icon?: string, iconColor?: string }>()
defineEmits<{ (e: "cancel"): void }>()
</script>
