<template>
  <v-container class="py-16">
    <h1 class="text-h2 font-weight-bold text-center mb-4">
      Powerful Features
    </h1>
    <p class="text-body-1 text-center text-primary mx-auto" :style="{ maxWidth: '600px' }">
      Everything you need to build, deploy, and scale your applications with confidence.
    </p>

    <FeatureMyceliumPanel />
    <FeatureWebGateway />
    <FeatureMultiMaster />
    <FeatureLoadBalancing />
  </v-container>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })
</script>
