<template>
  <v-card class="node-card mb-4" elevation="0">
    <div class="price-area px-4 pt-4 pb-2 mb-3">
      <div class="d-flex align-center mb-1">
        <span :style="`color:${priceColor}; font-size:1.5rem; font-weight:700; letter-spacing:0.01em;`">${{ monthlyPrice }}</span>
        <span class="text-caption ml-2" :style="`color:${priceLabelColor}; font-size:1.1rem; font-weight:500;`">/month</span>
      </div>
      <div class="d-flex align-center">
        <span :style="`color:${priceColor}; font-size:1.1rem; font-weight:600;`">${{ hourlyPrice }}</span>
        <span class="text-caption ml-1" :style="`color:${priceLabelColor}; font-size:1.05rem; font-weight:500;`">/hr</span>
      </div>
    </div>
    <div class="d-flex align-center justify-space-between px-4 pb-1 mb-3">
      <span class="text-h6 font-weight-bold text-white">Node {{ node.nodeId }}</span>
      <v-chip v-if="node.gpu" color="#0ea5e9" variant="outlined" size="small" class="ml-2">GPU</v-chip>
    </div>
    <div v-if="node.country" class="d-flex align-center px-4 pb-1">
      <v-icon size="16" class="mr-1" :color="priceLabelColor">mdi-map-marker</v-icon>
      <span class="text-body-2" :style="`color:${priceLabelColor};`">{{ node.country }}</span>
    </div>
    <v-card-text class="py-0 px-4">
      <div v-for="r in resources" :key="r.label" class="resource-row">
        <span class="resource-icon"><v-icon size="18" :color="r.color">{{ r.icon }}</v-icon></span>
        <span class="font-weight-medium" :style="`color:${priceLabelColor}; min-width:40px;`">{{ r.label }}</span>
        <span class="text-white ml-1">{{ r.value() }}</span>
      </div>
    </v-card-text>
    <v-card-actions class="pt-3 px-4 pb-4">
      <v-btn
        :color="buttonColor"
        variant="outlined"
        block
        class="font-weight-bold"
        style="border-radius:8px;"
        @click="handleAction"
        :aria-label="buttonLabel"
        :loading="loading"
        :disabled="disabled || loading"
      >
        {{ buttonLabel }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import type { NormalizedNode } from '../types/normalizedNode';
import { defineProps, defineEmits, computed } from 'vue';


const props = defineProps<{ node: NormalizedNode; loading?: boolean; disabled?: boolean; buttonLabel?: string }>();
const buttonLabel = computed(() => props.buttonLabel || 'Reserve Node');
const buttonColor = computed(() =>
  buttonLabel.value.toLowerCase().includes('unreserve') ? 'error' : 'primary'
);
const emit = defineEmits(['action', 'signin']);

const actionType = computed(() =>
  buttonLabel.value.toLowerCase().includes('unreserve') ? 'unreserve' : 'reserve'
);

function handleAction() {
  emit('action', { nodeId: props.node.nodeId, action: actionType.value });
}

const baseNodePrice = computed(() => {
  const base = Number(props.node.price_usd ?? 0);
  // Divide extra fee by 1000 as it's in musd
  const extra = Number(props.node.extraFee ?? 0) / 1000;
  const price = base + extra;
  return isNaN(price) ? null : price;
});

const monthlyPrice = computed(() => {
  return baseNodePrice.value == null ? 'N/A' : baseNodePrice.value.toFixed(2);
});

const hourlyPrice = computed(() => {
  return baseNodePrice.value == null ? 'N/A' : (baseNodePrice.value / 720).toFixed(2);
});

const resources = [
  {
    icon: 'mdi-cpu-64-bit',
    color: '#0ea5e9',
    label: 'CPU:',
    value: () => `${props.node.cpu} vCPU`
  },
  {
    icon: 'mdi-memory',
    color: '#10B981',
    label: 'RAM:',
    value: () => `${props.node.ram} GB`
  },
  {
    icon: 'mdi-harddisk',
    color: '#38bdf8',
    label: 'Storage:',
    value: () => `${props.node.storage} GB`
  }
];

const priceColor = '#10B981';
const priceLabelColor = '#a3a3a3';
</script>

<style scoped>
.node-card {
  border-radius: 16px;
  transition: box-shadow 0.2s, transform 0.2s;
}
.node-card:hover {
  transform: translateY(-3px) scale(1.015);
}
.price-area {
  background: rgba(16,185,129,0.07);
}
.v-card-text {
  padding-top: 0.5rem !important;
  padding-bottom: 0.5rem !important;
}
.resource-row {
  background: rgba(16,185,129,0.03);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
}
.resource-row:last-child {
  margin-bottom: 0;
}
.resource-icon {
  background: rgba(16,185,129,0.10);
  border-radius: 6px;
  padding: 4px;
  margin-right: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}
.font-weight-medium {
  font-weight: 500;
}
.text-white {
  color: #f8fafc;
}
</style>
