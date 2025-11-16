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
    defaultTheme: 'myceliumCloudTheme',
    themes: {
      myceliumCloudTheme: {
        dark: true,
        colors: {
          primary: '#3B82F6',      // mycelium cloud blue
          secondary: '#EA580C',    // mycelium cloud orange
          accent: '#60A5FA',       // mycelium cloud blue-light
          surface: '#1E293B',      // mycelium cloud slate
          background: '#0F172A',   // mycelium cloud navy
          success: '#10B981',      // mycelium cloud success
          warning: '#F59E0B',      // mycelium cloud warning
          error: '#EF4444',        // mycelium cloud error
          info: '#0891B2',         // mycelium cloud cyan
          'on-primary': '#FFFFFF',
          'on-secondary': '#FFFFFF',
          'on-surface': '#F1F5F9',
          'on-background': '#F1F5F9',
        },
      },
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
