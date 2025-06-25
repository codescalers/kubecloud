<template>
  <div>
    <div v-if="error" class="error-boundary">
      <div class="error-container">
        <div class="error-icon">
          <v-icon icon="mdi-alert-circle" size="64" color="error"></v-icon>
        </div>
        <h2 class="error-title kubecloud-gradient">Something went wrong</h2>
        <p class="error-message">
          {{ error.message || 'An unexpected error occurred. Please try refreshing the page.' }}
        </p>
        <div class="error-actions">
          <v-btn 
            color="primary" 
            variant="elevated" 
            @click="handleRetry"
            class="btn-enhanced"
          >
            <v-icon icon="mdi-refresh" class="mr-2"></v-icon>
            Try Again
          </v-btn>
          <v-btn 
            color="secondary" 
            variant="outlined" 
            @click="goHome"
            class="btn-enhanced"
          >
            <v-icon icon="mdi-home" class="mr-2"></v-icon>
            Go Home
          </v-btn>
        </div>
        <div v-if="showDetails" class="error-details">
          <v-btn 
            variant="text" 
            size="small" 
            @click="showDetails = false"
            class="mb-2"
          >
            Hide Details
          </v-btn>
          <div class="error-stack">
            <pre>{{ error.stack }}</pre>
          </div>
        </div>
        <v-btn 
          v-else
          variant="text" 
          size="small" 
          @click="showDetails = true"
          class="mt-2"
        >
          Show Details
        </v-btn>
      </div>
    </div>
    <slot v-else />
  </div>
</template>

<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'
import { useRouter } from 'vue-router'

interface Props {
  fallback?: (error: Error) => void
}

const props = withDefaults(defineProps<Props>(), {
  fallback: undefined
})

const router = useRouter()
const error = ref<Error | null>(null)
const showDetails = ref(false)

onErrorCaptured((err: Error) => {
  error.value = err
  console.error('Error caught by boundary:', err)
  
  // Call custom fallback if provided
  if (props.fallback) {
    props.fallback(err)
  }
  
  // Prevent error from propagating
  return false
})

const handleRetry = () => {
  error.value = null
  showDetails.value = false
  window.location.reload()
}

const goHome = () => {
  error.value = null
  showDetails.value = false
  router.push('/')
}
</script>

<style scoped>
.error-boundary {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--kubecloud-navy);
  padding: 2rem;
}

.error-container {
  max-width: 600px;
  text-align: center;
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border: 2px solid var(--kubecloud-blue);
  border-radius: 16px;
  padding: 3rem 2rem;
  box-shadow: var(--shadow-kubecloud-blue);
}

.error-icon {
  margin-bottom: 2rem;
}

.error-title {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.error-message {
  font-size: 1.1rem;
  color: var(--kubecloud-light-gray);
  line-height: 1.6;
  margin-bottom: 2rem;
}

.error-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  margin-bottom: 2rem;
}

.error-details {
  margin-top: 2rem;
  text-align: left;
}

.error-stack {
  background: var(--kubecloud-slate);
  border-radius: 8px;
  padding: 1rem;
  border: 1px solid rgba(37, 99, 235, 0.2);
  max-height: 200px;
  overflow-y: auto;
}

.error-stack pre {
  margin: 0;
  color: var(--kubecloud-light-gray);
  font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
  font-size: 0.85rem;
  line-height: 1.4;
}

@media (max-width: 600px) {
  .error-container {
    padding: 2rem 1rem;
  }
  
  .error-title {
    font-size: 1.5rem;
  }
  
  .error-actions {
    flex-direction: column;
  }
}
</style> 