<template>
  <v-container :fluid="isFluid" :style="{ transition: 'max-width 250ms ease-in-out' }">
    <div :class="{ 'd-flex justify-start': !mobile }">
      <div
        :class="{ 'sticky-top': !mobile }"
        :style="{
          maxHeight: mobile ? undefined : `calc(100vh - 100px)`,
          willChange: 'width',
          transition: 'width 250ms ease-in-out',
          width: mobile ? '100%' : isOpen ? sidebarWidth + 24 + 'px' : 0,
        }"
      >
        <div :class="{ 'pr-6 h-100': !mobile }">
          <slot name="sidebar" />
        </div>
      </div>

      <div
        :style="{
          willChange: 'width',
          transition: 'width 250ms ease-in-out',
          width: `calc(100% - ${isOpen && !mobile ? sidebarWidth : 0}px)`,
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
  pageOffset: { type: Number, default: 100 },
  mobile: { type: Boolean, default: false },
})
</script>
