<template>
  <v-card color="surface-variant" class="pa-4 d-flex flex-column" height="100%">
    <div class="d-flex justify-space-between align-center mb-3">
      <h3 class="text-h6 font-weight-medium">
        Node {{ node.nodeId }}
      </h3>
      <div class="text-h6 text-success font-weight-bold">${{ node.price_usd ?? 'N/A' }}/month</div>
    </div>

    <div v-if="node.country" class="d-flex justify-space-between align-center mb-3">
      <div class="d-flex align-center text-medium-emphasis">
        <v-icon size="16" class="me-1">mdi-map-marker</v-icon>
        {{ node.country }}
      </div>
      <v-chip v-if="node.gpu" color="deep-purple-accent-2" size="small" variant="elevated">
        <v-icon size="16" class="me-1">mdi-nvidia</v-icon>
        GPU
      </v-chip>
    </div>

    <v-divider class="mb-3" />

    <div class="mb-4">
      <div class="d-flex align-center mb-2">
        <v-icon size="18" class="me-2" color="primary">mdi-cpu-64-bit</v-icon>
        <span class="text-body-2 text-medium-emphasis me-2">CPU:</span>
        <span class="text-body-2">{{ node.cpu }} vCPU</span>
      </div>
      <div class="d-flex align-center mb-2">
        <v-icon size="18" class="me-2" color="success">mdi-memory</v-icon>
        <span class="text-body-2 text-medium-emphasis me-2">RAM:</span>
        <span class="text-body-2">{{ node.ram }} GB</span>
      </div>
      <div class="d-flex align-center mb-2">
        <v-icon size="18" class="me-2" color="info">mdi-harddisk</v-icon>
        <span class="text-body-2 text-medium-emphasis me-2">Storage:</span>
        <span class="text-body-2">{{ node.storage }} GB</span>
      </div>
    </div>

    <v-spacer />

    <v-btn
      v-if="isAuthenticated"
      color="primary"
      variant="elevated"
      block
      @click="$emit('reserve', node.nodeId)"
      aria-label="Reserve Node"
      :loading="loading"
      :disabled="disabled || loading"
    >
      Reserve Node
    </v-btn>
    <v-btn
      v-else
      color="primary"
      variant="outlined"
      block
      @click="$emit('signin')"
      aria-label="Sign In to Reserve"
    >
      Sign In to Reserve
    </v-btn>
  </v-card>
</template>

<script setup lang="ts">
import type { NormalizedNode } from '../types/normalizedNode';
import { defineProps, defineEmits } from 'vue';

const props = defineProps<{ node: NormalizedNode; isAuthenticated: boolean; loading?: boolean; disabled?: boolean }>();
const emit = defineEmits(['reserve', 'signin']);

function formatStorage(val: number) {
  if (val >= 1024) {
    return (val / 1024).toLocaleString(undefined, { maximumFractionDigits: 1, minimumFractionDigits: 1 }) + ' TB';
  }
  return Math.round(val).toLocaleString() + ' GB';
}
</script>


