<template>
  <v-app class="kubecloud-app">
    <NavBar v-if="!isAuthPage" />
    <v-main class="app-main">
      <RouterView />
    </v-main>
    <AppFooter v-if="!isAuthPage" />
    <NotificationToast />
  </v-app>
</template>

<script lang="ts" setup>
import { RouterView, useRoute } from 'vue-router'
import { computed, onMounted, onErrorCaptured } from 'vue'
import { useUserStore } from './stores/user'
import { useNotificationStore } from './stores/notifications'
import NavBar from './components/NavBar.vue'
import AppFooter from './components/AppFooter.vue'
import NotificationToast from './components/NotificationToast.vue'
import { useDeploymentEvents } from "./composables/useDeploymentEvents"
const route = useRoute()
const userStore = useUserStore()
const notificationStore = useNotificationStore()

// Global error handling
onErrorCaptured((error: Error) => {
  console.error('Global error caught:', error)

  // Show error as toast notification
  notificationStore.error(
    'Something went wrong',
    error.message || 'An unexpected error occurred. Please try refreshing the page.',
    { duration: 8000 }
  )

  // Prevent error from propagating and breaking the app
  return false
})

// Determine if current page is an authentication page
const isAuthPage = computed(() => {
  const authRoutes = ['/sign-in', '/sign-up', '/register/verify']
  return authRoutes.includes(route.path)
})

onMounted(async () => {
  try {
    await userStore.initializeAuth()
    useDeploymentEvents()
  } catch (error) {
    console.error('Failed to initialize authentication:', error)
  }

  // Global error handlers for unhandled errors
  window.addEventListener('error', (event) => {
    console.error('Unhandled error:', event.error)
    notificationStore.error(
      'Unexpected Error',
      event.error?.message || 'An unexpected error occurred',
      { duration: 8000 }
    )
  })
})

</script>

<style scoped>
.kubecloud-app {
  min-height: 100vh;
  background: var(--color-bg);
  color: var(--color-text);
  font-family: 'Inter', sans-serif;
}

.app-main {
  position: relative;
  z-index: 1;
  min-height: calc(100vh - 72px); /* Account for navbar height */
}


.loading-text {
  margin-top: 1rem;
  color: var(--color-text);
  font-size: 1.1rem;
  font-weight: 500;
}

/* Responsive adjustments */
@media (max-width: 960px) {
  .app-main {
    min-height: calc(100vh - 64px);
  }
}

@media (max-width: 600px) {
  .app-main {
    min-height: calc(100vh - 56px);
  }
}
</style>
