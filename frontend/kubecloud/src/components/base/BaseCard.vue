<template>
  <component 
    :is="tag" 
    :class="cardClasses"
    v-bind="$attrs"
  >
    <slot />
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  tag?: string
  variant?: 'default' | 'feature' | 'dashboard' | 'auth'
  hover?: boolean
  padding?: 'none' | 'small' | 'medium' | 'large'
}

const props = withDefaults(defineProps<Props>(), {
  tag: 'div',
  variant: 'default',
  hover: true,
  padding: 'medium'
})

const cardClasses = computed(() => [
  'base-card',
  `base-card--${props.variant}`,
  `base-card--padding-${props.padding}`,
  {
    'base-card--hover': props.hover
  }
])
</script>

<style scoped>
.base-card {
  background: var(--glass-bg);
  backdrop-filter: var(--glass-blur);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  box-shadow: var(--shadow-md), var(--shadow-ambient);
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
  min-width: 100%;
}

.base-card--hover::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--kubecloud-blue-subtle), transparent);
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.base-card--hover:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl), var(--shadow-ambient);
  border-color: var(--kubecloud-blue-subtle);
}

.base-card--hover:hover::before {
  opacity: 1;
}

/* Variant-specific styles */
.base-card--feature {
  border-radius: 12px;
  text-align: center;
}

.base-card--dashboard {
  border-radius: 16px;
}

.base-card--auth {
  border-radius: 24px;
  max-width: 480px;
  width: 100%;
}

/* Padding variants */
.base-card--padding-none {
  padding: 0;
}

.base-card--padding-small {
  padding: 1rem;
}

.base-card--padding-medium {
  padding: 2rem;
}

.base-card--padding-large {
  padding: 2.5rem;
}

/* Responsive adjustments */
@media (max-width: 960px) {
  .base-card--padding-medium {
    padding: 1.5rem;
  }
  
  .base-card--padding-large {
    padding: 2rem;
  }
}

@media (max-width: 600px) {
  .base-card--padding-medium {
    padding: 1rem;
  }
  
  .base-card--padding-large {
    padding: 1.5rem;
  }
}
</style> 