<template>
  <DialogCardLayout
    title="Drain Balance"
    icon="mdi-water-remove"
    icon-color="warning"
    @cancel="$emit('cancel')"
  >
    <template #description>
      <div>
        <span class="text-subtitle-2 opacity-50">Drain balance from user:</span>&nbsp;
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
        Are you sure you want to drain all balance from user {{ user?.username }}? This will
        transfer their funds to the system account.
      </p>
    </v-card-text>

    <template #actions>
      <v-btn
        variant="text"
        color="warning"
        prepend-icon="mdi-water-remove"
        @click="$emit('confirm')"
      >
        Drain Balance
      </v-btn>
    </template>
  </DialogCardLayout>
</template>

<script setup lang="ts">
import type { ServicesUserWithUSDBalance } from "../generated/api"

defineProps<{ user?: ServicesUserWithUSDBalance }>()
defineEmits<{ (e: "confirm" | "cancel"): void }>()
</script>
