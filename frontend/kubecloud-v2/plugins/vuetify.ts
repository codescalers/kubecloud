import { createVuetify } from "vuetify"

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(
    createVuetify({
      theme: {
        defaultTheme: "dark",
      },
      defaults: {
        VBtn: {
          style: "text-transform: none;",
        },
      },
    })
  )
})
