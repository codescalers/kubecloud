// @ts-check

import antfu from "@antfu/eslint-config"
import withNuxt from "./.nuxt/eslint.config.mjs"

export default withNuxt(
  antfu({
    stylistic: {
      indent: 2,
      quotes: "double",
      semi: false,
    },
    rules: {
      "vue/block-order": [
        "error",
        { order: ["template", "script", "style"] },
      ],
      "vue/max-attributes-per-line": ["error", {
        singleline: { max: 4 },
        multiline: { max: 1 },
      }],
    },
  }),
)
