<template>
  <v-container :fluid="isFluid" :style="{ transition: 'max-width 250ms ease-in-out' }">
    <div class="d-flex justify-start">
      <div
        class="sticky-top"
        :style="{
          willChange: 'width',
          transition: 'width 250ms ease-in-out',
          width: isOpen ? sidebarWidth + 24 + 'px' : 0,
        }"
      >
        <div class="pr-6">
          <slot name="sidebar" />
        </div>
      </div>

      <div
        :style="{
          willChange: 'width',
          transition: 'width 250ms ease-in-out',
          width: `calc(100% - ${isOpen ? sidebarWidth : 0}px)`,
        }"
      >
        <slot />
      </div>
    </div>
  </v-container>
</template>

<script setup lang="ts">
defineProps({
  sidebarWidth: { type: Number, default: 200 },
  isOpen: { type: Boolean, default: true },
  isFluid: { type: Boolean, default: false },
})
</script>
