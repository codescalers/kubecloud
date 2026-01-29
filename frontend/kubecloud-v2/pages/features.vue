<template>
  <v-container>
    <div>Features</div>

    <FeatureMyceliumPanel />
    <FeatureWebGateway />
    <FeatureMultiMaster />
    <FeatureLoadBalancing />
  </v-container>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })
</script>
