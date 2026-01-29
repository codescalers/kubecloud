<template>
  <div ref="vantaDots" class="h-100">
    <v-container class="py-16">
      <div class="mb-16">
        <h1 class="text-h2 font-weight-bold text-center mb-4">
          Use Cases
        </h1>
        <p class="text-body-1 text-center text-primary mx-auto" :style="{ maxWidth: '600px' }">
          Explore how Mycelium Cloud empowers different industries and teams.
        </p>
      </div>

      <VRow>
        <VCol v-for="feature in features" :key="feature.title" cols="4">
          <FeatureCard v-bind="feature" />
        </VCol>
      </VRow>
    </v-container>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })

ensureThreeGlobal()

const vantaDots = ref<HTMLElement | null>(null)
useVanta("dots", () => vantaDots.value!)

const features = [
  {
    icon: "mdi-brain",
    title: "AI/ML Training",
    description: "Run GPU-accelerated workloads for deep learning and data science on demand.",
    tags: ["GPU Acceleration", "Scalable Compute", "Cost Optimization"],
  },
  {
    icon: "mdi-domain",
    title: "Enterprise Kubernetes",
    description: "Deploy production-grade clusters with full control and no vendor lock-in.",
    tags: ["High Availability", "Security", "Compliance"],
  },
  {
    icon: "mdi-router-wireless",
    title: "Edge & IoT",
    description: "Leverage global nodes for low-latency edge computing and IoT applications.",
    tags: ["Low Latency", "Global Distribution", "Real-time Processing"],
  },
]
</script>
