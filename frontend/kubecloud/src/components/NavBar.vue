<template>
  <v-app-bar app flat color="transparent" class="px-4 py-3 nav-app-bar kubecloud-nav">
    <v-row align="center" class="w-100">
      <v-col cols="3" class="d-flex align-center">
        <router-link to="/" class="logo-link">
          <span class="footer-logo kubecloud-glow-blue d-flex align-center" aria-label="KubeCloud logo">
            <svg width="32" height="32" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
              <ellipse cx="24" cy="28" rx="16" ry="10" fill="#1E293B" stroke="#2563EB" stroke-width="2.5"/>
              <ellipse cx="18" cy="22" rx="7" ry="6" fill="#1E293B" stroke="#3B82F6" stroke-width="2.5"/>
              <ellipse cx="30" cy="20" rx="6" ry="5" fill="#1E293B" stroke="#F97316" stroke-width="2.5"/>
              <ellipse cx="24" cy="28" rx="10" ry="7" fill="#1E293B" stroke="#3B82F6" stroke-width="2.5"/>
              <ellipse cx="24" cy="28" rx="5" ry="3.5" fill="#2563EB" fill-opacity="0.7"/>
            </svg>
            <span class="footer-title ms-2">KubeCloud</span>
          </span>
        </router-link>
      </v-col>
      <v-col cols="6" class="d-flex align-center justify-center nav-links">
        <v-btn 
          v-for="item in navigationItems" 
          :key="item.path"
          variant="text" 
          color="primary" 
          :to="item.path" 
          class="nav-btn kubecloud-hover-blue"
          :class="{ 'nav-btn--active': $route.path === item.path }"
        >
          {{ item.label }}
        </v-btn>
      </v-col>
      <v-col cols="3" class="d-flex justify-end align-center">
        <v-btn color="secondary" variant="elevated" to="/sign-up" class="sign-up-btn newsletter-btn btn-enhanced kubecloud-glow-orange kubecloud-hover-blue">Sign Up</v-btn>
      </v-col>
    </v-row>
  </v-app-bar>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const mobileMenuOpen = ref(false)

const toggleMobileMenu = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value
}

// Navigation items data
const navigationItems = [
  { label: 'Home', path: '/' },
  { label: 'Features', path: '/features' },
  { label: 'Pricing', path: '/pricing' },
  { label: 'Use Cases', path: '/usecases' },
  { label: 'Docs', path: '/docs' },
  { label: 'Dashboard', path: '/dashboard' }
]
</script>

<style scoped>
.nav-app-bar {
  background: var(--glass-bg) !important;
  backdrop-filter: var(--glass-blur);
  border-bottom: 2px solid var(--kubecloud-blue);
  box-shadow: 0 8px 32px rgba(37, 99, 235, 0.2), 0 2px 0 rgba(249, 115, 22, 0.1);
  border-radius: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: sticky;
  top: 0;
  z-index: 1000;
}

.logo-link {
  text-decoration: none;
  display: flex;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-link:hover {
  transform: scale(1.05);
}

.logo-link:hover .footer-logo {
  text-shadow: 0 0 8px var(--kubecloud-blue), 0 0 16px var(--kubecloud-blue);
}

.footer-logo {
  display: flex;
  align-items: center;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--kubecloud-blue);
  text-shadow: 0 0 4px var(--kubecloud-blue);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.footer-title {
  color: var(--kubecloud-blue);
  font-weight: 700;
}

.newsletter-btn, .sign-up-btn {
  color: var(--kubecloud-white) !important;
  height: 36px;
  line-height: 36px;
  border: none;
  box-shadow: none;
}

.newsletter-btn:hover, .sign-up-btn:hover {
  box-shadow: 0 0 15px rgba(249, 115, 22, 0.2), 0 0 30px rgba(249, 115, 22, 0.1);
  transform: translateY(-0.5px);
}

.nav-links {
  gap: 0.5rem;
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.nav-btn {
  margin: 0 0.25rem;
  min-width: 100px;
  font-weight: 500;
  color: var(--kubecloud-white) !important;
  align-items: center;
  height: 36px;
  display: flex;
}

.nav-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, var(--kubecloud-blue), var(--kubecloud-orange));
  opacity: 0;
  transition: opacity 0.3s ease;
  border-radius: 10px;
  transform: scale(1.02);
}

.nav-btn:focus, .nav-btn:hover {
  background: rgba(79, 70, 229, 0.08);
  border: 1px solid rgba(79, 70, 229, 0.3);
  box-shadow: 0 0 12px rgba(79, 70, 229, 0.15);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.2);
  transform: translateY(-1px);
}

.nav-btn--active {
  background: rgba(79, 70, 229, 0.12) !important;
  border: 1px solid rgba(79, 70, 229, 0.4) !important;
  box-shadow: 0 0 16px rgba(79, 70, 229, 0.2), inset 0 0 0 1px rgba(79, 70, 229, 0.1) !important;
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3) !important;
  color: var(--kubecloud-blue-light) !important;
  font-weight: 600 !important;
  transform: translateY(-1px);
}

.nav-btn--active::before {
  opacity: 0.05;
}

@media (max-width: 960px) {
  .nav-app-bar {
    padding: 0.5rem 1rem !important;
  }
  .nav-links {
    gap: 0.25rem;
  }
  .nav-btn {
    min-width: 80px;
    padding: 0.5rem 0.75rem;
    font-size: 0.9rem;
    height: 32px;
  }
  .footer-title {
    font-size: 1.25rem;
  }
  .sign-up-btn {
    padding: 0.5rem 1rem;
    font-size: 0.9rem;
    min-width: 100px;
    height: 32px;
    line-height: 32px;
  }
}

@media (max-width: 600px) {
  .nav-app-bar {
    padding: 0.25rem 0.5rem !important;
  }
  .nav-links {
    flex-direction: column;
    gap: 0.125rem;
  }
  .nav-btn {
    min-width: 60px;
    padding: 0.375rem 0.5rem;
    font-size: 0.8rem;
    height: 28px;
  }
  .footer-title {
    font-size: 1.125rem;
  }
  .sign-up-btn {
    padding: 0.375rem 0.75rem;
    font-size: 0.8rem;
    min-width: 80px;
    height: 28px;
    line-height: 28px;
  }
}
</style>
