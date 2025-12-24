import { createVuetify } from "vuetify"

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(
    createVuetify({
      theme: {
        defaultTheme: "dark",
        themes: {
          dark: {
            variables: {
              "bg-2": "30, 41, 59",
              "bg-2-opacity": "0.5",
            },
            colors: {
              background: "#0A192F",
              primary: "#60A5FA",
              surface: "rgba(24, 33, 52, 0.55)",
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
        },
        VTextField: {
          autocomplete: "off",
          spellcheck: false,
        },
      },
    })
  )
})
