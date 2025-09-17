<template>
  <nav class="navbar">
    <div class="navbar-content">
      <router-link to="/" class="navbar-logo">
        <img :src="logo" alt="Mycelium Cloud Logo" class="logo" width="200">
      </router-link>
      <div class="navbar-main-links">
        <router-link v-for="link in publicLinks" :key="link.to" :to="link.to" class="navbar-link" active-class="active-link">
          {{ link.label }}
        </router-link>
        <!-- Show authenticated-only links when user is logged in -->
        <template v-if="isLoggedIn">
          <router-link v-for="link in authenticatedLinks" :key="link.to" :to="link.to" class="navbar-link" active-class="active-link">
            {{ link.label }}
          </router-link>
        </template>
      </div>
      <div class="navbar-auth">
        <!-- Show notification bell and user menu when logged in -->
        <div v-if="isLoggedIn" class="user-section">
          <NotificationBell />
          <div class="user-menu">
            <v-menu>
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  variant="text"
                  color="white"
                  class="user-menu-btn"
                >
                  <span class="user-name">{{ userName }}</span>
                  <v-icon icon="mdi-chevron-down" class="ml-1"></v-icon>
                </v-btn>
              </template>
              <v-list>
                <v-list-item v-if="isAdmin" @click="goToAdmin">
                  <v-list-item-title>
                    <v-icon icon="mdi-shield-crown" class="mr-2"></v-icon>
                    Admin Panel
                  </v-list-item-title>
                </v-list-item>
                <v-list-item @click="goToDashboard">
                  <v-list-item-title>
                    <v-icon icon="mdi-view-dashboard" class="mr-2"></v-icon>
                    Dashboard
                  </v-list-item-title>
                </v-list-item>
                <v-divider></v-divider>
                <v-list-item @click="handleLogout">
                  <v-list-item-title>
                    <v-icon icon="mdi-logout" class="mr-2"></v-icon>
                    Sign Out
                  </v-list-item-title>
                </v-list-item>
              </v-list>
            </v-menu>
          </div>
        </div>
        <!-- Show sign in and sign up buttons when not logged in -->
        <div v-else>
          <router-link :to="'/sign-in'" custom v-slot="{ navigate, isActive }">
            <v-btn
              variant="outlined"
              color="white"
              @click="navigate"
              :class="{ 'active-link': isActive }"
            >
              Sign In
            </v-btn>
          </router-link>
          <router-link :to="'/sign-up'" custom v-slot="{ navigate, isActive }">
            <v-btn
              variant="outlined"
              color="white"
              class="ml-2"
              @click="navigate"
              :class="{ 'active-link': isActive }"
            >
              Sign Up
            </v-btn>
          </router-link>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'
import { computed, nextTick } from 'vue'
import logo from '../assets/logo.png'
import NotificationBell from './NotificationBell.vue'

const userStore = useUserStore()
const router = useRouter()

// Public links (visible to everyone)
const publicLinks = [
  { label: 'Home', to: '/' },
  { label: 'Features', to: '/features' },
  { label: 'Docs', to: '/docs' },
  { label: 'Use Cases', to: '/use-cases' },
]

// Authenticated-only links (visible when logged in)
const authenticatedLinks = [
  { label: 'Dashboard', to: '/dashboard' },
]

// Computed properties for better reactivity
const isLoggedIn = computed(() => userStore.isLoggedIn)
const userName = computed(() => {
  // If we have user data, use the username
  if (userStore.user?.username) {
    return userStore.user.username
  }

  // If we're logged in but don't have user data, try to extract from token
  if (userStore.isLoggedIn && userStore.token) {
    return userStore.user?.username
  }

  return 'User'
})
const isAdmin = computed(() => userStore.isAdmin)

const goToAdmin = () => {
  router.push('/admin')
}

const goToDashboard = () => {
  router.push('/dashboard')
}

const handleLogout = async () => {
  userStore.logout()
  await nextTick()
  router.push('/')
}
</script>

<style scoped>
.navbar {
  width: 100%;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 100;
  background: rgba(10, 25, 47, 0.65);
  box-shadow: 0 2px 16px 0 rgba(33, 150, 243, 0.10);
  backdrop-filter: blur(8px);
  transition: background 0.3s;
}

.navbar-content {
  max-width: 1300px;
  margin: 0 auto;
  padding: 1.2rem 2.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.navbar-logo {
  display: flex;
  align-items: center;
  text-decoration: none;
  transition: transform 0.2s ease;
}

.navbar-logo:hover {
  transform: scale(1.05);
}

.logo {
  height: auto;
  max-height: 40px;
}

.navbar-main-links {
  display: flex;
  gap: 1.2rem;
  flex: 1;
  justify-content: flex-start;
  margin-left: 6rem;
  align-items: center;
}

.navbar-link {
  color: #e0e7ef;
  font-size: 1.05rem;
  font-weight: 500;
  text-decoration: none;
  position: relative;
  padding: 0.2rem 0;
  transition: color 0.2s;
  min-width: 0;
}

.navbar-link:hover,
.navbar-link.active-link {
  color: #60a5fa;
}

.navbar-link::after {
  content: '';
  display: block;
  height: 2px;
  width: 0;
  background: linear-gradient(90deg, #60a5fa 0%, #38bdf8 100%);
  transition: width 0.3s;
  position: absolute;
  left: 0;
  bottom: -2px;
}

.navbar-link:hover::after,
.navbar-link.active-link::after {
  width: 100%;
}

.navbar-auth {
  margin-left: auto;
  display: flex;
  align-items: center;
}

.user-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-menu-btn {
  color: #e0e7ef !important;
  font-weight: 500;
  text-transform: none;
  letter-spacing: normal;
}

.user-menu-btn:hover {
  color: #60a5fa !important;
}

.user-name {
  font-weight: 500;
}

@media (max-width: 900px) {
  .navbar-content {
    padding: 1rem 1.2rem;
  }

  .navbar-main-links {
    gap: 0.7rem;
    margin-left: 1.2rem;
  }
}
</style>
