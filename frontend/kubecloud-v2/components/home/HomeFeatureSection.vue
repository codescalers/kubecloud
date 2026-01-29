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
      <VCol v-for="feature in features" :key="feature.title" cols="4">
        <FeatureCard v-bind="feature" />
      </VCol>
    </VRow>
  </v-container>
</template>

<script setup lang="ts">
const features = [
  {
    icon: "mdi-cloud-outline",
    title: "Cloud-Native Architecture",
    description: "Built for the cloud with support for all major cloud providers and on-premise deployments.",
  },
  {
    icon: "mdi-shield-check-outline",
    title: "Enterprise Security",
    description: "Advanced security features including RBAC, network policies, and compliance monitoring.",
  },
  {
    icon: "mdi-chart-line",
    title: "Real-time Monitoring",
    description: "Comprehensive monitoring and alerting with detailed metrics and performance insights.",
  },
  {
    icon: "mdi-rocket-launch-outline",
    title: "One-Click Deployments",
    description: "Streamlined deployment process with automated CI/CD pipelines and rollback capabilities.",
  },
  {
    icon: "mdi-account-group-outline",
    title: "Team Collaboration",
    description: "Built-in collaboration tools for teams with role-based access and shared workspaces.",
  },
  {
    icon: "mdi-cog-outline",
    title: "Advanced Configuration",
    description: "Flexible configuration management with support for Helm charts and custom resources.",
  },
]
</script>
