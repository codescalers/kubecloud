<template>
  <v-btn
    :class="buttonClasses"
    v-bind="$attrs"
  >
    <slot />
  </v-btn>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'outlined' | 'text'
  size?: 'small' | 'medium' | 'large'
  enhanced?: boolean
  glow?: 'blue' | 'orange' | 'cyan' | 'purple' | 'none'
  hover?: 'blue' | 'orange' | 'none'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'medium',
  enhanced: true,
  glow: 'none',
  hover: 'none'
})

const buttonClasses = computed(() => [
  'base-button',
  `base-button--${props.variant}`,
  `base-button--${props.size}`,
  {
    'btn-enhanced': props.enhanced,
    [`kubecloud-glow-${props.glow}`]: props.glow !== 'none',
    [`kubecloud-hover-${props.hover}`]: props.hover !== 'none'
  }
])
</script>

<style scoped>
.base-button {
  font-weight: 600;
  font-size: 1rem;
  padding: 0.75rem 1.25rem;
  text-transform: none;
  letter-spacing: 0.01em;
  background: transparent;
  border: 1px solid transparent;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Size variants */
.base-button--small {
  font-size: 0.9rem;
  padding: 0.5rem 1rem;
  height: 32px;
  line-height: 32px;
  border-radius: 6px;
}

.base-button--medium {
  font-size: 1rem;
  padding: 0.75rem 1.25rem;
  height: 36px;
  line-height: 36px;
  border-radius: 8px;
}

.base-button--large {
  font-size: 1.125rem;
  padding: 1rem 1.5rem;
  height: 44px;
  line-height: 44px;
  border-radius: 10px;
}

/* Variant styles */
.base-button--primary {
  color: var(--kubecloud-white) !important;
  background: var(--kubecloud-blue);
  border-color: var(--kubecloud-blue);
}

.base-button--secondary {
  color: var(--kubecloud-white) !important;
  background: var(--kubecloud-orange);
  border-color: var(--kubecloud-orange);
}

.base-button--outlined {
  color: var(--kubecloud-blue) !important;
  background: transparent;
  border-color: var(--kubecloud-blue);
}

.base-button--text {
  color: var(--kubecloud-blue) !important;
  background: transparent;
  border-color: transparent;
}

/* Enhanced button effects */
.btn-enhanced {
  position: relative;
  overflow: hidden;
}

.btn-enhanced::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, var(--kubecloud-blue), var(--kubecloud-orange));
  opacity: 0;
  transition: opacity 0.3s ease;
  transform: scale(1.02);
}

.base-button--small.btn-enhanced::before {
  border-radius: 6px;
}

.base-button--medium.btn-enhanced::before {
  border-radius: 8px;
}

.base-button--large.btn-enhanced::before {
  border-radius: 10px;
}

.btn-enhanced:hover::before {
  opacity: 0.1;
}

.btn-enhanced:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
}

.btn-enhanced:active {
  transform: translateY(0);
  box-shadow: 0 2px 6px rgba(37, 99, 235, 0.15);
}

/* Responsive adjustments */
@media (max-width: 960px) {
  .base-button--medium {
    font-size: 0.9rem;
    padding: 0.5rem 1rem;
    height: 32px;
    line-height: 32px;
  }
  
  .base-button--large {
    font-size: 1rem;
    padding: 0.75rem 1.25rem;
    height: 36px;
    line-height: 36px;
  }
}

@media (max-width: 600px) {
  .base-button--medium {
    font-size: 0.8rem;
    padding: 0.375rem 0.75rem;
    height: 28px;
    line-height: 28px;
  }
  
  .base-button--large {
    font-size: 0.9rem;
    padding: 0.5rem 1rem;
    height: 32px;
    line-height: 32px;
  }
}
</style> 