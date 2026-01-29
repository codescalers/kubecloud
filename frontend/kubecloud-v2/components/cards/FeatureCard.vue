<template>
  <div ref="container" :style="{ transition: 'transform 0.1s ease-in-out' }" @mousemove="onMouseMove($event)" @mouseleave="onMouseLeave()">
    <v-card class="text-center">
      <v-icon color="primary" icon="mdi-cloud-outline" size="50" />

      <p class="text-h5 font-weight-bold mt-4 mb-1 text-no-wrap">
        Cloud-Native Architecture
      </p>
      <p class="text-body-1 text-accent">
        Built for the cloud with support for all major cloud providers and on-premise deployments.
      </p>
    </v-card>
  </div>
</template>

<script setup lang="ts">
const container = ref<HTMLElement | null>(null)

function onMouseMove(e: MouseEvent) {
  const rect = container.value!.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top

  const centerX = rect.width / 2
  const centerY = rect.height / 2

  const rotateX = (y - centerY) / 10
  const rotateY = (centerX - x) / 10

  container.value!.style.transform = `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateY(-10px)`
}

function onMouseLeave() {
  container.value!.style.transform = `perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(0px)`
}
</script>
