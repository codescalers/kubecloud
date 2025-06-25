import './assets/main.css'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { aliases, mdi } from 'vuetify/iconsets/mdi'

import App from './App.vue'
import router from './router'

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

app.use(createPinia())
app.use(router)
app.use(vuetify)

app.mount('#app')
