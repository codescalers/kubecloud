<template>
  <nav class="navbar">
    <div class="navbar-content">
      <router-link to="/" class="navbar-logo">KubeCloud</router-link>
      <div class="navbar-main-links">
        <router-link v-for="link in links" :key="link.to" :to="link.to" class="navbar-link" active-class="active-link">
          {{ link.label }}
        </router-link>
      </div>
      <div class="navbar-signin">
        <router-link v-if="signInLink" :to="signInLink.to" custom v-slot="{ navigate, isActive }">
          <v-btn
            variant="outlined"
            color="white"
            @click="navigate"
            :class="{ 'active-link': isActive }"
          >
            {{ signInLink.label }}
          </v-btn>
        </router-link>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
const links = [
  { label: 'Home', to: '/' },
  { label: 'Features', to: '/features' },
  { label: 'Pricing', to: '/pricing' },
  { label: 'Docs', to: '/docs' },
  { label: 'Use Cases', to: '/use-cases' },
  { label: 'Dashboard', to: '/dashboard' },
  { label: 'Admin', to: '/admin' },
]
const signInLink = { label: 'Sign In', to: '/sign-in' }
</script>

<style scoped>
.navbar {
  width: 100%;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 100;
  background: rgba(10, 25, 47, 0.65); /* semi-transparent, blends with gradient */
  box-shadow: 0 2px 16px 0 rgba(33, 150, 243, 0.10); /* soft shadow for readability */
  border: none;
  backdrop-filter: blur(8px);
  transition: background 0.3s;
}
.navbar-content {
  max-width: 1300px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.2rem 2.5rem;
}
.navbar-logo {
  font-size: 2rem;
  font-weight: 500;
  color: #fff;
  letter-spacing: -1px;
  text-decoration: none;
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
.navbar-signin {
  margin-left: auto;
  display: flex;
  align-items: center;
}
.more-btn {
  font-size: 1rem;
  font-weight: 600;
  color: #e0e7ef;
  min-width: 0;
  padding: 0.2rem 0.8rem;
}
@media (max-width: 900px) {
  .navbar-content {
    padding: 1rem 1.2rem;
  }
  .navbar-main-links {
    gap: 0.7rem;
    margin-left: 1.2rem;
  }
  .navbar-logo {
    font-size: 1.4rem;
  }
}
.brand-title {
  color: #60a5fa;
  font-weight: 500;
  letter-spacing: 0.01em;
  margin-left: 0.5rem;
  font-size: 1.2rem;
  transition: color 0.2s;
}
</style>
