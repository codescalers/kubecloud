<template>
  <div class="stats-grid">
    <div v-for="stat in props.stats" :key="stat.label" class="stat-item">
      <v-icon :icon="stat.icon" size="24" :color="stat.color || 'var(--color-primary)'"></v-icon>
      <div class="stat-info">
        <div class="stat-number">{{ stat.value }}</div>
        <div v-if="stat.subvalue" class="stat-subvalue">{{ stat.subvalue }}</div>
        <div class="stat-label">{{ stat.label }}</div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue'

interface Stat {
  icon: string
  value: string | number
  subvalue?: string
  label: string
  color?: string
}

const props = defineProps<{ stats: Stat[] }>()
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: var(--space-4);
}

.stat-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: var(--space-2);
}

.stat-info {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  text-align: center;
  flex: 1;
}

.stat-number {
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-lg);
  color: var(--color-text);
  text-align: center;
}

.stat-subvalue {
  font-size: var(--font-size-xs, 0.75rem);
  text-align: center;
  color: rgba(203, 213, 225, 0.6);
  margin-top: -2px;
  margin-bottom: 2px;
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-primary);
}
</style>
