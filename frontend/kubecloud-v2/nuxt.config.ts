// https://nuxt.com/docs/api/configuration/nuxt-config
import vuetify, { transformAssetUrls } from "vite-plugin-vuetify"

const prod = process.env.NODE_ENV === "production"

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: {
    enabled: !prod,
  },
  ssr: prod,
  typescript: {
    strict: true,
  },
  build: {
    transpile: ["vuetify"],
  },
  vite: {
    plugins: [vuetify({ autoImport: true })],
    vue: {
      template: {
        transformAssetUrls,
      },
    },
  },
  css: ["@mdi/font/css/materialdesignicons.css", "vuetify/styles", "~/assets/scss/global.scss"],
  modules: ["@nuxt/eslint", "@vueuse/nuxt", "@nuxtjs/google-fonts"],
  runtimeConfig: {
    public: {
      apiBasePath: process.env.API_BASE_PATH,
    },
  },
  googleFonts: {
    families: {
      Inter: [300, 400, 500, 600, 700],
    },
    display: "swap",
  },
})
