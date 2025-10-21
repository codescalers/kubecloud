<template>
  <v-container fluid class="pa-0" style="background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%);">
    <!-- Hero Section -->
    <v-row no-gutters class="hero-section mx-auto" style="max-width: 1100px;">
      <v-col cols="12">
        <v-container class="py-16">
          <v-row justify="center">
            <v-col cols="12" md="10" lg="8" class="text-center">
              <h1 class="text-h4 text-md-h2 font-weight-medium mb-8 text-white">
                Powerful Features
              </h1>
              <p class="text-h6 text-md-h5 text-blue-lighten-2 line-height-1-7 opacity-92">
                Everything you need to build, deploy, and scale your applications with confidence.
              </p>
            </v-col>
          </v-row>
        </v-container>
      </v-col>
    </v-row>

    <FeatureMyceliumPanel />
    <FeatureWebGateway />
    <FeatureMultiMaster />
    <FeatureLoadBalancing />
  </v-container>
</template>

<script setup lang="ts">
import FeatureMyceliumPanel from '../components/features/FeatureMyceliumPanel.vue'
import FeatureWebGateway from '../components/features/FeatureWebGateway.vue'
import FeatureMultiMaster from '../components/features/FeatureMultiMaster.vue'
import FeatureLoadBalancing from '../components/features/FeatureLoadBalancing.vue'
</script>

<style scoped>
.hero-section {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 2;
}

@media (max-width: 900px) {
  .hero-section {
    min-height: 50vh;
  }
}

@media (max-width: 600px) {
  .hero-section {
    min-height: 40vh;
  }
}
</style>
