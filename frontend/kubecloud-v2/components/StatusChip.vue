<template>
  <div class="w-100 text-center">
    <div
      :class="`d-inline-flex px-2 py-1 align-center border rounded-xl ga-2 border-${color}`"
      :style="{
        backgroundColor: `rgba(var(--v-theme-${color}), 0.1)`,
      }"
    >
      <v-avatar :color="color" size="8" />
      <span
        :class="`text-caption  text-${color} text-capitalize font-weight-medium`"
        v-text="status ?? 'N/A'"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ status: string }>()

const color = useStatusColor(() => props.status)
</script>
