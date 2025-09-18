<template>
  <v-card class="node-card mb-4" elevation="0">
    <div class="price-area px-4 pt-4 pb-2 mb-3">
      <!-- Monthly Price -->
      <div class="d-flex align-center flex-wrap mb-1">
        <div class="d-flex align-center justify-space-between flex-wrap price-container">
          <div>
            <span :style="`color:${priceColor}; font-size:1.5rem; font-weight:700; letter-spacing:0.01em;`">${{
              monthlyPrice }}</span>
            <span class="text-caption ml-2"
              :style="`color:${priceLabelColor}; font-size:1.1rem; font-weight:500;`">/month</span>
          </div>
          <div v-if="hasDiscount">
            <span class="text-grey text-body-1 font-weight-medium mr-2 original-price-text">
              ${{ originalMonthlyPrice }}
            </span>
          </div>
        </div>
      </div>
      <div class="d-flex align-center justify-space-between flex-wrap">
        <div>
          <span :style="`color:${priceColor}; font-size:1.1rem; font-weight:600;`">${{ hourlyPrice }}</span>
          <span class="text-caption ml-1"
            :style="`color:${priceLabelColor}; font-size:1.05rem; font-weight:500;`">/hr</span>
        </div>
        <div v-if="hasDiscount">
          <v-chip color="success" variant="outlined" size="small" 
                  class="text-caption font-weight-bold pulse-chip">
            {{ discountPercentage }}% OFF
          </v-chip>
        </div>
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
    <v-card-actions class="pt-3 px-4 pb-4 d-flex flex-column">
      <v-btn
        variant="outlined"
        block
        @click="openMonitoring"
        aria-label="Check Node Health"
      >
        Check Node Health
      </v-btn>
      <v-btn
        :color="buttonColor"
        variant="outlined"
        block
        class="font-weight-bold"
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
import { defineProps, defineEmits, ref, computed, onMounted } from 'vue';
import { useNodes } from '../composables/useNodes';

const props = defineProps<{ node: NormalizedNode; loading?: boolean; disabled?: boolean; buttonLabel?: string }>();
const emit = defineEmits(['action', 'signin']);
const buttonLabel = computed(() => props.buttonLabel || 'Reserve Node');
const buttonColor = computed(() => buttonLabel.value.toLowerCase().includes('unreserve') ? 'error' : 'primary');
const actionType = computed(() => buttonLabel.value.toLowerCase().includes('unreserve') ? 'unreserve' : 'reserve');

function handleAction() {
  emit('action', { nodeId: props.node.nodeId, action: actionType.value });
}

const originalNodePrice = computed(() => {
  const base = Number(props.node.price_usd ?? 0);
  const extra = Number(props.node.extraFee ?? 0) / 1000;
  const price = base + extra;
  return isNaN(price) ? null : price;
});

/*
* Current price (with discount if available)
*/
  const baseNodePrice = computed(() => {
  const base = Number(props.node.discount_price ?? props.node.price_usd ?? 0);
  const extra = Number(props.node.extraFee ?? 0) / 1000;
  const price = base + extra;
  return isNaN(price) ? null : price;
});

// Check if there's a discount
const hasDiscount = computed(() => {
  return props.node.discount_price != null &&
    props.node.price_usd != null &&
    Number(props.node.discount_price) < Number(props.node.price_usd);
});

// Calculate discount percentage
const discountPercentage = computed(() => {
  if (!hasDiscount.value || !originalNodePrice.value || !baseNodePrice.value) return 0;
  const originalBase = Number(props.node.price_usd);
  const discountBase = Number(props.node.discount_price);
  return Math.round(((originalBase - discountBase) / originalBase) * 100);
});

const monthlyPrice = computed(() => baseNodePrice.value == null ? 'N/A' : baseNodePrice.value.toFixed(2));
const hourlyPrice = computed(() => baseNodePrice.value == null ? 'N/A' : (baseNodePrice.value / 720).toFixed(2));
const originalMonthlyPrice = computed(() => originalNodePrice.value == null ? 'N/A' : originalNodePrice.value.toFixed(2));
const resources = [
  { icon: 'mdi-cpu-64-bit', color: '#0ea5e9', label: 'CPU:', value: () => `${props.node.cpu} vCPU` },
  { icon: 'mdi-memory', color: '#10B981', label: 'RAM:', value: () => `${props.node.ram} GB` },
  { icon: 'mdi-harddisk', color: '#38bdf8', label: 'Storage:', value: () => `${props.node.storage} GB` }
];

const { fetchAccountId } = useNodes();
const monitoringUrl = ref('');

function getNetwork(env: string): string {
  switch (env) {
    case 'dev': return 'development';
    case 'qa': return 'qa';
    case 'test': return 'testing';
    case 'main': return 'production';
    default: return 'development';
  }
}

onMounted(async () => {
  let accountId = '';
  if (props.node.twinId) {
    accountId = await fetchAccountId(props.node.twinId);
  }
  const env = (typeof window !== 'undefined' && (window as any).__ENV__?.VITE_NETWORK) || (import.meta as any).env?.VITE_NETWORK;
  const network = getNetwork(env);
  const params = new URLSearchParams({
    orgId: '2',
    refresh: '30s',
    'var-network': network,
    'var-farm': props.node.farmId?.toString(),
    'var-node': accountId,
    'var-diskdevices': '[a-z]+|nvme[0-9]+n[0-9]+|mmcblk[0-9]+'
  });
  monitoringUrl.value = `https://metrics.grid.tf/d/rYdddlPWkfqwf/zos-host-metrics?${params.toString()}`;
});

function openMonitoring() {
  window.open(monitoringUrl.value, '_blank');
}

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
  background: rgba(16, 185, 129, 0.07);
  }

  /* Price styling - minimal CSS, using Vuetify classes in template */
  .price-container {
    flex: 1;
    min-width: 0;
  }

  .pulse-chip {
    animation: pulse-success 2s infinite;
  }

  .original-price-text {
    text-decoration: line-through;
    text-decoration-color: #ef4444;
    text-decoration-thickness: 2px;
  }

  @keyframes pulse-success {

    0%,
    100% {
      box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.4);
    }

    50% {
      box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.1);
    }
  }

  /* Responsive design */
  @media (max-width: 600px) {
    .price-area {
      padding-left: 1rem !important;
      padding-right: 1rem !important;
    }

    .original-price {
      font-size: 1rem;
    }

    .discount-chip {
      margin-left: 0.5rem !important;
      margin-top: 0.25rem;
    }

    /* Stack price elements on very small screens */
    @media (max-width: 400px) {
      .d-flex.align-center.flex-wrap {
        flex-direction: column;
        align-items: flex-start !important;
      }

      .discount-chip {
        margin-left: 0 !important;
        margin-top: 0.5rem;
        align-self: flex-start;
      }
    }
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
