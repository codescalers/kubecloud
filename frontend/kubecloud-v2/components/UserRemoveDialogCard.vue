<template>
  <DialogCardLayout
    title="Remove User"
    icon="mdi-trash-can-outline"
    icon-color="error"
    @cancel="$emit('cancel')"
  >
    <template #description>
      <div>
        <span class="text-subtitle-2 opacity-50">Remove user:</span>&nbsp;
        <span
          class="text-subtitle-2 text-primary px-2 py-1 rounded-lg"
          :style="{
            backgroundColor: 'rgba(var(--v-theme-primary), var(--v-border-opacity))',
          }"
        >{{ user?.email ?? "N/A" }}</span>
      </div>
    </template>

    <v-card-text>
      <p class="text-body-1 opacity-70">
        Are you sure you want to remove user {{ user?.username }}? This will delete their account
        and all their data.
      </p>
    </v-card-text>

    <template #actions>
      <v-btn
        variant="text"
        color="error"
        prepend-icon="mdi-trash-can-outline"
        @click="$emit('confirm')"
      >
        Remove User
      </v-btn>
    </template>
  </DialogCardLayout>
</template>

<script setup lang="ts">
import type { ServicesUserWithUSDBalance } from "../generated/api"

defineProps<{ user?: ServicesUserWithUSDBalance }>()
defineEmits<{ (e: "confirm" | "cancel"): void }>()
</script>
