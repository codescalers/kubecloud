<template>
  <div class="stats-grid">
    <div v-for="stat in stats" :key="stat.label" class="stat-item">
      <div class="stat-icon">
        <v-icon :icon="stat.icon" size="24" :color="stat.color || 'var(--color-primary)'"></v-icon>
      </div>
      <div class="stat-info">
        <div class="stat-number">{{ stat.value }}</div>
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
  label: string
  color?: string
}

const props = defineProps<{ stats: Stat[] }>()
const stats = props.stats
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.stat-item {
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border, #334155);
  border-radius: var(--radius-xl, 0.75rem);
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: all var(--transition-normal, 0.2s);
  backdrop-filter: blur(8px);
}

.stat-item:hover {
  border-color: var(--color-border-light, #475569);
  background: rgba(15, 30, 52, 0.75);
  transform: translateY(-1px);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-lg, 0.5rem);
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid var(--color-primary, #3B82F6);
  flex-shrink: 0;
}

.stat-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.stat-number {
  font-weight: var(--font-weight-bold, 700);
  font-size: var(--font-size-2xl, 1.5rem);
  color: var(--color-text, #F8FAFC);
  line-height: 1.2;
  margin-bottom: 0.25rem;
}

.stat-label {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--color-text-secondary, #CBD5E1);
  font-weight: var(--font-weight-medium, 500);
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
  }
  
  .stat-item {
    padding: 1rem;
  }
  
  .stat-icon {
    width: 40px;
    height: 40px;
  }
  
  .stat-number {
    font-size: var(--font-size-xl, 1.25rem);
  }
}
</style>
