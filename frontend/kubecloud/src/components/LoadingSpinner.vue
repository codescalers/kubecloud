<template>
  <div class="loading-container" :class="{ overlay: overlay }">
    <div class="loading-spinner">
      <div class="spinner-ring">
        <div class="spinner-ring-inner"></div>
      </div>
      <div class="loading-text">
        <h3 class="loading-title kubecloud-gradient">{{ title }}</h3>
        <p class="loading-subtitle">{{ subtitle }}</p>
      </div>
      <div v-if="progress !== null" class="loading-progress">
        <div class="progress-bar">
          <div 
            class="progress-fill" 
            :style="{ width: `${progress}%` }"
          ></div>
        </div>
        <span class="progress-text">{{ Math.round(progress) }}%</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title?: string
  subtitle?: string
  overlay?: boolean
  progress?: number | null
}

withDefaults(defineProps<Props>(), {
  title: 'Loading...',
  subtitle: 'Please wait while we process your request',
  overlay: false,
  progress: null
})
</script>

<style scoped>
.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  padding: 2rem;
}

.loading-container.overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(10px);
  z-index: 9999;
  min-height: 100vh;
}

.loading-spinner {
  text-align: center;
  max-width: 400px;
}

.spinner-ring {
  position: relative;
  width: 80px;
  height: 80px;
  margin: 0 auto 2rem auto;
}

.spinner-ring-inner {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  border: 4px solid transparent;
  border-top: 4px solid var(--kubecloud-blue);
  border-right: 4px solid var(--kubecloud-orange);
  border-radius: 50%;
  animation: spin 1.5s linear infinite;
}

.spinner-ring-inner::before {
  content: '';
  position: absolute;
  top: 8px;
  left: 8px;
  right: 8px;
  bottom: 8px;
  border: 2px solid transparent;
  border-top: 2px solid var(--kubecloud-light-blue);
  border-radius: 50%;
  animation: spin 1s linear infinite reverse;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-text {
  margin-bottom: 1.5rem;
}

.loading-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.loading-subtitle {
  font-size: 1rem;
  color: var(--kubecloud-light-gray);
  line-height: 1.5;
}

.loading-progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: var(--kubecloud-slate);
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid rgba(37, 99, 235, 0.2);
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--kubecloud-blue), var(--kubecloud-orange));
  border-radius: 4px;
  transition: width 0.3s ease;
  position: relative;
}

.progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0%, 100% { transform: translateX(-100%); }
  50% { transform: translateX(100%); }
}

.progress-text {
  font-size: 0.9rem;
  color: var(--kubecloud-light-gray);
  font-weight: 500;
}

@media (max-width: 600px) {
  .loading-container {
    padding: 1rem;
  }
  
  .loading-title {
    font-size: 1.25rem;
  }
  
  .loading-subtitle {
    font-size: 0.9rem;
  }
  
  .spinner-ring {
    width: 60px;
    height: 60px;
  }
}
</style> 