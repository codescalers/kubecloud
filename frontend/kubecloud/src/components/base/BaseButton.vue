<template>
  <v-btn
    :class="buttonClasses"
    v-bind="$attrs"
    tabindex="0"
    role="button"
    @focus="onFocus"
    @blur="onBlur"
    @keydown.space.prevent="onKeyboardClick"
    @keydown.enter.prevent="onKeyboardClick"
    :aria-pressed="$attrs['aria-pressed'] || undefined"
    :aria-label="$attrs['aria-label'] || undefined"
    ref="buttonRef"
  >
    <slot />
  </v-btn>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'outlined' | 'text'
  size?: 'small' | 'medium' | 'large'
  enhanced?: boolean
  glow?: 'blue' | 'orange' | 'cyan' | 'none'
  hover?: 'blue' | 'orange' | 'none'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'medium',
  enhanced: true,
  glow: 'none',
  hover: 'none'
})

const buttonRef = ref<HTMLElement | null>(null)

const isFocused = ref(false)
const isKeyboardFocus = ref(false)

function onFocus(e: FocusEvent) {
  isFocused.value = true
  // Only show focus ring if focus was via keyboard
  isKeyboardFocus.value = (e as any).sourceCapabilities === undefined
}
function onBlur() {
  isFocused.value = false
  isKeyboardFocus.value = false
}
function onKeyboardClick(e: KeyboardEvent) {
  // Only trigger click if not disabled
  if (!buttonRef.value?.hasAttribute('disabled')) {
    buttonRef.value?.click()
  }
}

const buttonClasses = computed(() => [
  'base-button',
  `base-button--${props.variant}`,
  `base-button--${props.size}`,
  {
    'btn-enhanced': props.enhanced,
    'focus-ring': isFocused.value && isKeyboardFocus.value
  }
])
</script>

<style scoped>
.base-button {
  background: var(--color-accent);
  color: #fff;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none;
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  transition: background var(--transition);
}
.base-button:hover {
  background: var(--color-accent-hover);
}
.base-button:focus-visible,
.base-button.focus-ring {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-accent);
}
</style>
