<template>
  <div class="floating-clouds-container">
    <div class="floating-cloud floating-cloud-1"></div>
    <div class="floating-cloud floating-cloud-2"></div>
    <div class="floating-cloud floating-cloud-3"></div>
    <div class="floating-cloud floating-cloud-4"></div>
    <div class="floating-cloud floating-cloud-5"></div>
    <div class="floating-cloud floating-cloud-6"></div>
    <div class="floating-cloud floating-cloud-7"></div>
  </div>
</template>

<script setup lang="ts">
// This component adds subtle floating cloud animations site-wide
// The animations are defined in main.css and respect accessibility preferences
</script>

<style scoped>
.floating-clouds-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: -1;
  overflow: hidden;
}
</style> 