<template>
  <div class="error-fallback" :class="variant">
    <div class="error-content">
      <div class="error-icon">
        <v-icon :icon="icon" :size="iconSize" :color="iconColor"></v-icon>
      </div>
      
      <h2 class="error-title" v-if="title">{{ title }}</h2>
      <p class="error-message" v-if="message">{{ message }}</p>
      
      <div class="error-details" v-if="showDetails && error">
        <v-expansion-panels>
          <v-expansion-panel>
            <v-expansion-panel-title>
              Error Details
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <pre class="error-stack">{{ error.stack || error.message }}</pre>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
      </div>
      
      <div class="error-actions" v-if="actions.length > 0">
        <v-btn
          v-for="action in actions"
          :key="action.label"
          :color="action.color || 'primary'"
          :variant="action.variant as any"
          :size="action.size || 'default'"
          @click="action.handler"
          class="action-btn"
        >
          <v-icon v-if="action.icon" :icon="action.icon" class="mr-2"></v-icon>
          {{ action.label }}
        </v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface ErrorAction {
  label: string
  handler: () => void
  color?: string
  variant?: 'flat' | 'text' | 'elevated' | 'tonal' | 'outlined' | 'plain'
  size?: string
  icon?: string
}

interface Props {
  error?: Error | null
  title?: string
  message?: string
  variant?: 'default' | 'compact' | 'fullscreen'
  showDetails?: boolean
  actions?: ErrorAction[]
}

const props = withDefaults(defineProps<Props>(), {
  error: null,
  title: 'Something went wrong',
  message: 'An unexpected error occurred. Please try again.',
  variant: 'default',
  showDetails: false,
  actions: () => [],
})

const icon = computed(() => {
  if (props.variant === 'compact') return 'mdi-alert'
  return 'mdi-alert-circle'
})

const iconSize = computed(() => {
  switch (props.variant) {
    case 'compact': return 24
    case 'fullscreen': return 64
    default: return 48
  }
})

const iconColor = computed(() => {
  return 'error'
})
</script>

<style scoped>
.error-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.error-fallback.fullscreen {
  min-height: 100vh;
  background: var(--kubecloud-navy);
}

.error-fallback.compact {
  padding: 1rem;
}

.error-content {
  max-width: 600px;
  text-align: center;
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border: 2px solid var(--kubecloud-blue);
  border-radius: 16px;
  padding: 2rem;
  box-shadow: var(--shadow-kubecloud-blue);
}

.error-fallback.compact .error-content {
  padding: 1rem;
  border-radius: 8px;
}

.error-fallback.fullscreen .error-content {
  padding: 3rem 2rem;
}

.error-icon {
  margin-bottom: 1.5rem;
}

.error-fallback.compact .error-icon {
  margin-bottom: 0.5rem;
}

.error-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--kubecloud-light-gray);
}

.error-fallback.compact .error-title {
  font-size: 1rem;
  margin-bottom: 0.5rem;
}

.error-fallback.fullscreen .error-title {
  font-size: 2rem;
}

.error-message {
  font-size: 1rem;
  color: var(--kubecloud-light-gray);
  line-height: 1.6;
  margin-bottom: 1.5rem;
}

.error-fallback.compact .error-message {
  font-size: 0.875rem;
  margin-bottom: 0.5rem;
}

.error-details {
  margin-bottom: 1.5rem;
}

.error-stack {
  background: var(--kubecloud-slate);
  border-radius: 8px;
  padding: 1rem;
  border: 1px solid rgba(37, 99, 235, 0.2);
  max-height: 200px;
  overflow-y: auto;
  margin: 0;
  color: var(--kubecloud-light-gray);
  font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
  font-size: 0.85rem;
  line-height: 1.4;
}

.error-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.error-fallback.compact .error-actions {
  gap: 0.5rem;
}

.action-btn {
  min-width: 120px;
}

.error-fallback.compact .action-btn {
  min-width: 80px;
  font-size: 0.875rem;
}

@media (max-width: 600px) {
  .error-content {
    padding: 1.5rem 1rem;
  }
  
  .error-actions {
    flex-direction: column;
    align-items: center;
  }
  
  .action-btn {
    width: 100%;
    max-width: 200px;
  }
}
</style> 