<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h4 font-weight-bold">
        Admin Overview
      </h1>
      <p class="text-body-1 mt-2 text-accent">
        Monitor platform health and key metrics
      </p>
    </div>

    <v-progress-circular v-if="isLoading" indeterminate color="primary" />

    <v-row v-else>
      <v-col
        v-for="resource in stats"
        :key="resource.title"
        cols="12"
        md="6"
        lg="4"
        xl="3"
      >
        <StatsCard v-bind="resource" />
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
const { stats, isLoading } = useStats()
</script>
