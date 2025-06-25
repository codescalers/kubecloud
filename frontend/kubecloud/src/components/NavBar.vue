<template>
  <v-app-bar app flat color="#0F172A" class="nav-app-bar minimalist-nav enhanced-nav">
    <v-row align="center" class="w-100">
      <v-col cols="2" class="d-flex align-center">
        <router-link to="/" class="logo-link minimalist-logo">
          <svg width="32" height="32" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <ellipse cx="24" cy="28" rx="16" ry="10" fill="#181A20" stroke="#3B82F6" stroke-width="2.5"/>
            <ellipse cx="18" cy="22" rx="7" ry="6" fill="#181A20" stroke="#60A5FA" stroke-width="2.5"/>
            <ellipse cx="30" cy="20" rx="6" ry="5" fill="#181A20" stroke="#93C5FD" stroke-width="2.5"/>
            <ellipse cx="24" cy="28" rx="10" ry="7" fill="#181A20" stroke="#3B82F6" stroke-width="2.5"/>
            <ellipse cx="24" cy="28" rx="5" ry="3.5" fill="#3B82F6" fill-opacity="0.7"/>
          </svg>
          <span class="logo-title ms-2">KubeCloud</span>
        </router-link>
      </v-col>
      <v-col cols="8" class="d-flex align-center justify-center nav-links minimalist-links">
        <router-link v-for="item in navigationItems" :key="item.path" :to="item.path" class="nav-link minimalist-link" :class="{ 'nav-link--active': $route.path === item.path }" style="color: #fff;" @mouseover="(e) => e.target.style.color = '#3B82F6'" @mouseleave="(e) => e.target.style.color = '#fff'">
          {{ item.label }}
        </router-link>
      </v-col>
      <v-col cols="2" class="d-flex justify-end align-center">
        <template v-if="isSignedIn">
          <v-menu offset-y>
            <template v-slot:activator="{ props }">
              <v-btn v-bind="props" variant="text" color="accent" class="user-menu-btn minimalist-user-btn">
                <v-icon icon="mdi-account-circle" color="accent" />
              </v-btn>
            </template>
            <v-list class="user-menu-list minimalist-user-list">
              <v-list-item @click="handleLogout" class="logout-item minimalist-logout-item">
                <v-list-item-title class="text-error">Logout</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </template>
        <router-link v-else to="/sign-up" class="nav-link minimalist-link">Sign Up</router-link>
      </v-col>
    </v-row>
  </v-app-bar>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()
const router = useRouter()

const isSignedIn = computed(() => userStore.isLoggedIn)

const handleLogout = () => {
  userStore.logout()
  router.push('/')
}

// Minimal navigation items
const navigationItems = [
  { label: 'Home', path: '/' },
  { label: 'Features', path: '/features' },
  { label: 'Pricing', path: '/pricing' },
  { label: 'Docs', path: '/docs' },
  { label: 'Dashboard', path: '/dashboard' }
]
</script>

<style scoped>
.nav-app-bar.minimalist-nav {
  background: #0F172A !important;
  border-bottom: 2px solid #3B82F6;
  box-shadow: none;
  border-radius: 0;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  position: sticky;
  top: 0;
  z-index: 1000;
  padding: 0.5rem 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-link.minimalist-logo {
  text-decoration: none;
  display: flex;
  align-items: center;
  color: #fff;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.3rem;
  font-weight: 700;
  letter-spacing: 0.01em;
}

.logo-title {
  color: #fff;
  font-weight: 700;
  letter-spacing: 0.01em;
}

.nav-links.minimalist-links {
  gap: 1.5rem;
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.nav-link.minimalist-link {
  color: #fff;
  text-decoration: none;
  font-size: 1.05rem;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-weight: 500;
  letter-spacing: 0.02em;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  position: relative;
  transition: color 0.2s, background 0.2s, transform 0.18s cubic-bezier(0.4,0,0.2,1);
  outline: none;
}

.nav-link.minimalist-link:focus-visible {
  box-shadow: 0 0 0 2px #3B82F6;
  border-radius: 6px;
}

.nav-link.minimalist-link::after {
  content: '';
  display: block;
  position: absolute;
  left: 0;
  right: 0;
  bottom: 2px;
  height: 2.5px;
  background: linear-gradient(90deg, #3B82F6 0%, #60A5FA 100%);
  border-radius: 2px;
  opacity: 0;
  transform: scaleX(0.5);
  transition: opacity 0.18s, transform 0.18s cubic-bezier(0.4,0,0.2,1);
}

.nav-link.minimalist-link:hover,
.nav-link--active {
  color: #3B82F6;
  background: rgba(59, 130, 246, 0.08);
  transform: scale(1.06);
}

.nav-link.minimalist-link:hover::after,
.nav-link--active::after {
  opacity: 1;
  transform: scaleX(1);
}

.user-menu-btn.minimalist-user-btn {
  color: #8B5CF6 !important;
  background: none;
  border: none;
  min-width: 36px;
  min-height: 36px;
  border-radius: 50%;
  padding: 0;
}

.user-menu-list.minimalist-user-list {
  background: #181A20;
  border-radius: 10px;
  box-shadow: none;
  min-width: 120px;
}

.logout-item.minimalist-logout-item {
  color: #e57373;
  font-size: 1rem;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  text-align: center;
  padding: 0.5rem 0.75rem;
}

.logout-item.minimalist-logout-item:hover {
  background: #2d223a;
}

.enhanced-nav {
  box-shadow: 0 6px 32px 0 rgba(59, 130, 246, 0.10);
}

@media (max-width: 900px) {
  .nav-links.minimalist-links {
    gap: 0.5rem;
  }
  .logo-link.minimalist-logo {
    font-size: 1.1rem;
  }
}
@media (max-width: 600px) {
  .nav-app-bar.minimalist-nav {
    padding: 0.25rem 0.5rem !important;
  }
  .nav-links.minimalist-links {
    flex-direction: column;
    gap: 0.25rem;
  }
  .nav-link.minimalist-link {
    font-size: 0.95rem;
    padding: 0.2rem 0.4rem;
  }
}
</style>
