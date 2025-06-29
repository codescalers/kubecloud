<template>
  <div class="loading-container" :class="containerClass">
    <div class="loading-spinner" :class="spinnerClass">
      <v-progress-circular
        v-if="type === 'circular'"
        :size="size"
        :width="width"
        :color="color"
        indeterminate
      />
      <div v-else-if="type === 'dots'" class="loading-dots">
        <div class="dot" v-for="i in 3" :key="i"></div>
      </div>
      <div v-else-if="type === 'pulse'" class="loading-pulse"></div>
      <div v-else-if="type === 'skeleton'" class="loading-skeleton">
        <div class="skeleton-line" v-for="i in lines" :key="i"></div>
      </div>
    </div>
    
    <div v-if="message" class="loading-message">
      {{ message }}
    </div>
    
    <div v-if="progress !== null" class="loading-progress">
      <v-progress-linear
        :model-value="progress"
        :color="color"
        height="4"
        rounded
      />
      <span class="progress-text">{{ Math.round(progress) }}%</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
interface Props {
  type?: 'circular' | 'dots' | 'pulse' | 'skeleton'
  size?: number
  width?: number
  color?: string
  message?: string
  progress?: number | null
  lines?: number
  fullscreen?: boolean
  overlay?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'circular',
  size: 40,
  width: 4,
  color: 'primary',
  message: '',
  progress: null,
  lines: 3,
  fullscreen: false,
  overlay: false,
})

const containerClass = computed(() => ({
  'loading-fullscreen': props.fullscreen,
  'loading-overlay': props.overlay,
}))

const spinnerClass = computed(() => ({
  [`loading-${props.type}`]: true,
}))
</script>

<style scoped>
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 2rem;
}

.loading-fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(8px);
  z-index: 9999;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(4px);
  z-index: 100;
}

.loading-message {
  color: var(--v-primary-base);
  font-size: 1rem;
  font-weight: 500;
  text-align: center;
}

.loading-progress {
  width: 100%;
  max-width: 300px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.progress-text {
  color: var(--v-primary-base);
  font-size: 0.875rem;
  text-align: center;
}

/* Dots animation */
.loading-dots {
  display: flex;
  gap: 0.5rem;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--v-primary-base);
  animation: dot-pulse 1.4s ease-in-out infinite both;
}

.dot:nth-child(1) { animation-delay: -0.32s; }
.dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes dot-pulse {
  0%, 80%, 100% {
    transform: scale(0);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

/* Pulse animation */
.loading-pulse {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--v-primary-base);
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(1.1);
  }
}

/* Skeleton loading */
.loading-skeleton {
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.skeleton-line {
  height: 16px;
  background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
  background-size: 200% 100%;
  border-radius: 4px;
  animation: skeleton-loading 1.5s infinite;
}

.skeleton-line:nth-child(1) { width: 100%; }
.skeleton-line:nth-child(2) { width: 80%; }
.skeleton-line:nth-child(3) { width: 60%; }

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

/* Responsive adjustments */
@media (max-width: 600px) {
  .loading-container {
    padding: 1rem;
  }
  
  .loading-message {
    font-size: 0.875rem;
  }
  
  .loading-progress {
    max-width: 250px;
  }
}
</style> 