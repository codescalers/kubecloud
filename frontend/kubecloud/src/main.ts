import './assets/main.css'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { aliases, mdi } from 'vuetify/iconsets/mdi'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

import App from './App.vue'
import router from './router'
import { useUserStore } from './stores/user'

const vuetify = createVuetify({
  theme: {
    defaultTheme: 'kubecloudTheme',
    themes: {
      kubecloudTheme: {
        dark: true,
        colors: {
          primary: '#3B82F6',      // kubecloud blue
          secondary: '#EA580C',    // kubecloud orange
          accent: '#60A5FA',       // kubecloud blue-light
          surface: '#1E293B',      // kubecloud slate
          background: '#0F172A',   // kubecloud navy
          success: '#10B981',      // kubecloud success
          warning: '#F59E0B',      // kubecloud warning
          error: '#EF4444',        // kubecloud error
          info: '#0891B2',         // kubecloud cyan
          'on-primary': '#FFFFFF',
          'on-secondary': '#FFFFFF',
          'on-surface': '#F1F5F9',
          'on-background': '#F1F5F9',
          // Enhanced surface variants with better contrast
          'surface-variant': 'rgba(30, 41, 59, 0.85)',
          'surface-bright': 'rgba(51, 65, 85, 0.9)',
          // White borders for better visibility
          'outline': '#FFFFFF',
          'outline-variant': 'rgba(255, 255, 255, 0.6)',
        },
        variables: {
          // Custom spacing that matches your design system
          'border-radius-root': '12px',
          'border-radius-lg': '16px',
          'border-radius-xl': '20px',
          // Enhanced shadows for better depth
          'shadow-key-umbra-opacity': '0.2',
          'shadow-key-penumbra-opacity': '0.1',
          'shadow-key-ambient-opacity': '0.05',
        }
      },
    },
  },
  defaults: {
    // Enhanced global component defaults for better visibility
    VCard: {
      elevation: 2,
      variant: 'outlined',
      style: 'backdrop-filter: blur(8px); border-color: white !important;',
    },
    VBtn: {
      style: 'text-transform: none; font-weight: 500;',
      rounded: 'lg',
    },
    VChip: {
      rounded: 'lg',
    },
    VTextField: {
      variant: 'outlined',
      rounded: 'lg',
      style: 'border-color: white;',
    },
    VSelect: {
      variant: 'outlined',
      rounded: 'lg',
      style: 'border-color: white;',
    },
    VDialog: {
      style: 'border: 2px solid white; border-radius: 16px;',
    },
  },
  components,
  directives,
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: { mdi },
  },
})

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)
app.use(pinia)
// Initialize auth state
const userStore = useUserStore()
userStore.initializeAuth()

app.use(router)
app.use(vuetify)

app.mount('#app')
