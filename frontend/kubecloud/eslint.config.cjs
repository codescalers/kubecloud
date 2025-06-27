const pluginVue = require('eslint-plugin-vue')
const vueParser = require('vue-eslint-parser')
const tsParser = require('@typescript-eslint/parser')

module.exports = [
  {
    files: ['**/*.{js,ts,vue}'],
    plugins: { vue: pluginVue },
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tsParser,
        ecmaVersion: 2020,
        sourceType: 'module'
      }
    },
    rules: {
      'vue/multi-word-component-names': 'off'
    }
  }
]