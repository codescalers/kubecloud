// https://nuxt.com/docs/api/configuration/nuxt-config
import process from "node:process"
import eslint from "vite-plugin-eslint2"
import vuetify, { transformAssetUrls } from "vite-plugin-vuetify"

const prod = process.env.NODE_ENV === "production"

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: {
    enabled: !prod,
  },
  ssr: prod,
  components: [
    {
      path: "~/components",
      pathPrefix: false,
    },
  ],
  typescript: {
    strict: true,
  },
  build: {
    transpile: ["vuetify"],
  },
  vite: {
    plugins: [eslint(), vuetify({ autoImport: true })],
    vue: {
      template: {
        transformAssetUrls,
      },
    },
    css: {
      preprocessorOptions: {
        scss: {
          silenceDeprecations: ["if-function"],
          quietDeps: true,
        },
      },
    },
  },
  css: ["@mdi/font/css/materialdesignicons.css", "~/assets/scss/vuetify.scss", "~/assets/scss/global.scss"],
  modules: ["@nuxt/hints", "@nuxt/eslint", "@vueuse/nuxt", "@nuxtjs/google-fonts", "nuxt-toast"],
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
  eslint: {
    checker: true,
    config: {
      standalone: false,
    },
  },
})
