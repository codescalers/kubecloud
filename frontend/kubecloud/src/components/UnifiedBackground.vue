<template>
  <div class="unified-background">
    <!-- Floating cloud particles -->
    <div class="cloud-particle"></div>
    <div class="cloud-particle"></div>
    <div class="cloud-particle"></div>
    <div class="cloud-particle"></div>
    <div class="cloud-particle"></div>
    
    <!-- Cloud-themed background elements -->
    <div class="background-gradient"></div>
    <div class="background-pattern"></div>
    <div class="background-clouds"></div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  theme?: 'default' | 'home' | 'features' | 'pricing' | 'use-cases' | 'docs' | 'reserve' | 'dashboard'
}

withDefaults(defineProps<Props>(), {
  theme: 'default'
})
</script>

<style scoped>
.unified-background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: -1;
  overflow: hidden;
}

.background-gradient {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);
  pointer-events: none;
}

.background-pattern {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: 
    radial-gradient(circle at 20% 80%, rgba(59, 130, 246, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% 20%, rgba(96, 165, 250, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 40% 40%, rgba(59, 130, 246, 0.05) 0%, transparent 50%),
    radial-gradient(circle at 60% 60%, rgba(255, 255, 255, 0.02) 0%, transparent 50%);
  pointer-events: none;
}

.background-clouds {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: 
    radial-gradient(ellipse 200px 100px at 10% 20%, rgba(255, 255, 255, 0.03) 0%, transparent 70%),
    radial-gradient(ellipse 150px 75px at 90% 40%, rgba(255, 255, 255, 0.02) 0%, transparent 70%),
    radial-gradient(ellipse 180px 90px at 30% 70%, rgba(255, 255, 255, 0.03) 0%, transparent 70%),
    radial-gradient(ellipse 120px 60px at 70% 80%, rgba(255, 255, 255, 0.02) 0%, transparent 70%);
  pointer-events: none;
  opacity: 0.3;
}

/* Cloud particles are styled in main.css */
</style>
