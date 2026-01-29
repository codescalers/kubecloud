<template>
  <div>
    <HomeWelcomeSection />

    <HomeStatsSection />

    <HomeFeatureSection />

    <HomeExperienceSection />
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })
</script>
