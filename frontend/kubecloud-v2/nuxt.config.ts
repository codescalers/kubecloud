// https://nuxt.com/docs/api/configuration/nuxt-config

const preCommit = process.env.PRE_COMMIT === "y"
const prod = process.env.NODE_ENV === "production" && !preCommit

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: {
    enabled: !prod,
  },
  ssr: prod,
  typescript: {
    strict: true,
  },
  modules: ["@nuxt/eslint"],
})
