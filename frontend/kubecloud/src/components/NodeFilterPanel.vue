<template>
  <div class="filter-card">
    <h3 class="filter-title">Filter Nodes</h3>
    <div class="filter-section">
      <label class="filter-label">CPU Cores</label>
      <v-range-slider
        v-model="modelValue.cpu"
        :min="cpuMin"
        :max="cpuMax"
        :step="1"
        class="filter-slider"
      />
      <div class="slider-values">{{ modelValue.cpu[0] }} - {{ modelValue.cpu[1] }} vCPU</div>
    </div>
    <div class="filter-section">
      <label class="filter-label">RAM (GB)</label>
      <v-range-slider
        v-model="modelValue.ram"
        :min="ramMin"
        :max="ramMax"
        :step="1"
        class="filter-slider"
      />
      <div class="slider-values">{{ modelValue.ram[0] }} - {{ modelValue.ram[1] }} GB</div>
    </div>
    <div class="filter-section">
      <label class="filter-label">Storage (GB)</label>
      <v-range-slider
        v-model="modelValue.storage"
        :min="storageMin"
        :max="storageMax"
        :step="1"
        class="filter-slider"
      >
        <template #thumb-label="{ modelValue }">
          {{ formatStorage(modelValue) }}
        </template>
      </v-range-slider>
      <div class="slider-values">
        {{ formatStorage(modelValue.storage[0]) }} - {{ formatStorage(modelValue.storage[1]) }}
      </div>
    </div>
    <div class="filter-section">
      <label class="filter-label">GPU</label>
      <v-switch
        v-model="modelValue.gpu"
        :label="'Has GPU'"
        inset
        color="primary"
        hide-details
      />
    </div>
    <div class="filter-section">
      <label class="filter-label">Price Range ($/mo)</label>
      <v-range-slider
        v-model="modelValue.priceRange"
        :min="priceMin"
        :max="priceMax"
        :step="1"
        thumb-label
        class="filter-slider"
      />
      <div class="slider-values">${{ modelValue.priceRange[0] }} - ${{ modelValue.priceRange[1] }}</div>
    </div>
    <div class="filter-section">
      <label class="filter-label">Location</label>
      <v-select
        v-model="modelValue.location"
        :items="locationOptions"
        item-title="title"
        item-value="value"
        label="Location"
        clearable
        hide-details
        class="filter-select"
      />
    </div>
    <v-btn class="clear-filters-btn" color="primary" variant="outlined" @click="$emit('clear')">
      Clear All Filters
    </v-btn>
  </div>
</template>

<script setup lang="ts">
import { toRefs, watch, defineProps, defineEmits } from 'vue';
import type { NodeFilterState } from '../composables/useNodeFilters';

const props = defineProps<{
  modelValue: NodeFilterState;
  cpuMin: number;
  cpuMax: number;
  ramMin: number;
  ramMax: number;
  priceMin: number;
  priceMax: number;
  locationOptions: { title: string; value: string | null }[];
  storageMin: number;
  storageMax: number;
}>();
const emit = defineEmits(['update:modelValue', 'clear']);

const { modelValue } = toRefs(props);

watch(modelValue, (val) => {
  emit('update:modelValue', val);
}, { deep: true });

function formatStorage(val: number) {
  if (val >= 1024) {
    return (val / 1024).toLocaleString(undefined, { maximumFractionDigits: 1, minimumFractionDigits: 1 }) + ' TB';
  }
  return Math.round(val).toLocaleString() + ' GB';
}
</script>

<style scoped>
.filter-card {
  padding: 2rem;
  height: fit-content;
  position: sticky;
  top: 2rem;
  background: rgba(30, 41, 59, 0.98);
  border-radius: 1.25rem;
  box-shadow: 0 4px 24px 0 rgba(59, 130, 246, 0.10);
  border: 1.5px solid #334155;
}
.filter-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #fff;
  margin-bottom: 2rem;
  text-align: center;
}
.filter-section {
  margin-bottom: 2.25rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid rgba(96, 165, 250, 0.08);
}
.filter-section:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}
.clear-filters-btn {
  width: 100%;
  margin-top: 2.5rem;
  font-weight: 500;
  letter-spacing: 0.01em;
}
.slider-values {
  font-size: 0.95em;
  color: #CBD5E1;
  opacity: 0.85;
  margin-top: 0.25rem;
}
</style>