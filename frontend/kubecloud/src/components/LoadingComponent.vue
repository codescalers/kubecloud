<template>
  <div class="loading-component" :class="{ 'full-page': fullPage }">
    <div class="loading-content">
      <slot name="icon">
        <v-progress-circular 
          :size="size" 
          :width="width" 
          :color="color" 
          indeterminate
        />
      </slot>
      
      <p v-if="message && !$slots.default" class="loading-text">
        {{ message }}
      </p>
      
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  /** Loading message to display */
  message?: string;
  /** Size of the progress circular */
  size?: number | string;
  /** Width of the progress circular line */
  width?: number | string;
  /** Color of the progress circular */
  color?: string;
  /** Whether the loading component should take the full page */
  fullPage?: boolean;
}

// Default props
const props = withDefaults(defineProps<Props>(), {
  message: '',
  size: 48,
  width: 4,
  color: 'primary',
  fullPage: false
});
</script>

<style scoped>
.loading-component {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--mycelium-cloud-light-gray);
  padding: 1.5rem 0;
}

.full-page {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(10, 25, 47, 0.92);
  z-index: 9999;
  padding: 0;
}

.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.loading-text {
  font-size: 1.15rem;
  color: var(--mycelium-cloud-light-gray);
  opacity: 0.85;
  margin-top: 1rem;
}

@media (max-width: 600px) {
  .loading-text {
    font-size: 1rem;
  }
}
</style>
