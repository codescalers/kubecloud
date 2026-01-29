<template>
  <v-row>
    <v-col cols="6" class="position-relative" :style="{ height: '500px' }" :order="reversed ? 'last' : undefined">
      <div
        class="position-absolute h-100 w-100 inset-0"
        :style="{
          background: 'radial-gradient(circle, rgba(96, 165, 250, 0.18) 0%, transparent 80%)',
          filter: 'blur(32px)',
        }"
      />

      <slot :canvas-props="{ class: 'position-relative h-100 w-100', style: { zIndex: 1 } }" />

      <v-chip
        v-if="hoveredNode"
        class="position-absolute"
        color="primary"
        variant="tonal"
        :style="{ left: `${hoveredNode.pos.x + 12}px`, top: `${hoveredNode.pos.y}px`, transform: 'translate(-50%, -120%)' }"
        :text="hoveredNode.label"
      />
    </v-col>

    <v-col cols="6" class="d-flex align-center">
      <div>
        <h3 class="text-h4 font-weight-bold mb-2" v-text="title" />

        <div class="text-subtitle-1 text-accent" v-text="description" />

        <div class="d-flex ga-2 mt-2">
          <v-chip
            v-for="tag in tags"
            :key="tag"
            size="small"
            variant="tonal"
            color="primary"
            :text="tag"
          />
        </div>
      </div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description: string
  tags: string[]
  reversed?: boolean
  hoveredNode?: { label: string, pos: { x: number, y: number } }
}>()
</script>

<style scoped>
.feature-animation-glow {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 90%;
  height: 90%;
  transform: translate(-50%, -50%);
  background: radial-gradient(circle, rgba(96, 165, 250, 0.18) 0%, transparent 80%);
  z-index: 0;
  pointer-events: none;
  filter: blur(32px);
}
</style>
