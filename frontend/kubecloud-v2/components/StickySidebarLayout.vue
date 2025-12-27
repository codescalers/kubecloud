<template>
  <v-container>
    <div class="d-flex justify-start">
      <div
        class="sticky-top overflow-hidden"
        :style="{
          willChange: 'width',
          transition: 'width 250ms ease-in-out',
          width: isOpen ? '224px' : '0px',
        }"
      >
        <div class="pr-6">
          <v-list class="overflow-hidden w-100">
            <slot name="items" />
          </v-list>
        </div>
      </div>

      <div
        :style="{
          willChange: 'width',
          transition: 'width 250ms ease-in-out',
          width: `calc(100% - ${isOpen ? 224 : 0}px)`,
        }"
      >
        <slot />
      </div>
    </div>
  </v-container>
</template>

<script setup lang="ts">
defineProps({
  cols: { type: Number, required: true },
  isOpen: { type: Boolean, default: true },
})
</script>
