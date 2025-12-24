import { createVuetify } from "vuetify"

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(
    createVuetify({
      theme: {
        defaultTheme: "dark",
        themes: {
          dark: {
            colors: {
              background: "#0A192F",
              primary: "#3B82F6",
            },
          },
        },
      },
      defaults: {
        global: {
          fontFamily: "Inter, sans-serif",
        },
        VBtn: {
          class: "button-style",
          // class: "button-style",
        },
      },
    })
  )
})
