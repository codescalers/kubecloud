<template>
  <ErrorBoundary>
    <v-app class="kubecloud-app">
      <!-- Shared Moving Background - Persists across route changes -->
      <UnifiedBackground :theme="currentTheme" />
      
      <!-- Navigation Bar -->
      <NavBar v-if="!isAuthPage" />
      
      <!-- Main Content Area -->
      <v-main class="app-main">
        <RouterView />
      </v-main>
      
      <!-- Footer -->
      <AppFooter v-if="!isAuthPage" />
      
      <!-- Global Notifications -->
      <NotificationToast />
    </v-app>
  </ErrorBoundary>
</template>

<script lang="ts" setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'
import NavBar from './components/NavBar.vue'
import AppFooter from './components/AppFooter.vue'
import ErrorBoundary from './components/ErrorBoundary.vue'
import NotificationToast from './components/NotificationToast.vue'
import UnifiedBackground from './components/UnifiedBackground.vue'

const route = useRoute()

// Determine if current page is an authentication page
const isAuthPage = computed(() => {
  const authRoutes = ['/sign-in', '/sign-up']
  return authRoutes.includes(route.path)
})

// Dynamic theme based on current route
const currentTheme = computed(() => {
  const path = route.path
  
  // Theme mapping for different routes
  const themeMap: Record<string, 'default' | 'home' | 'features' | 'pricing' | 'use-cases' | 'docs' | 'reserve' | 'dashboard'> = {
    '/': 'home',
    '/features': 'features',
    '/pricing': 'pricing',
    '/usecases': 'use-cases',
    '/docs': 'docs',
    '/reserve': 'reserve',
    '/deploy': 'dashboard'
  }
  
  // Check for dashboard routes
  if (path.startsWith('/dashboard')) {
    return 'dashboard'
  }
  
  return themeMap[path] || 'default'
})
</script>

<style scoped>
.kubecloud-app {
  min-height: 100vh;
  background: var(--kubecloud-bg);
  color: var(--kubecloud-white);
  font-family: 'Space Grotesk', 'Inter', sans-serif;
}

.app-main {
  position: relative;
  z-index: 1;
  min-height: calc(100vh - 72px); /* Account for navbar height */
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
