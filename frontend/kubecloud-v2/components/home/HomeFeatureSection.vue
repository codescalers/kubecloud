<template>
  <v-container class="mb-16">
    <div class="mb-8">
      <h1 class="text-h3 font-weight-bold text-center mb-4">
        Everything You Need to Succeed
      </h1>
      <p class="text-body-1 text-center text-primary">
        Powerful tools and features designed for modern cloud-native applications
      </p>
    </div>

    <VRow>
      <VCol v-for="i in 6" :key="i" cols="4">
        <FeatureCard />
      </VCol>
    </VRow>
  </v-container>
</template>

<script setup lang="ts">
// defineProps<{
//   title: string
//   description: string
//   icon: string
// }>()
</script>
