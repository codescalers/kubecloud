<template>
  <div>
    <HomeWelcomeSection />

    <v-container class="my-16">
      <div v-if="isLoading || stats.length === 0" class="d-flex justify-center align-center" :style="{ height: '198px' }">
        <v-progress-linear indeterminate color="primary" :style="{ maxWidth: '600px' }" />
      </div>

      <v-fade-transition>
        <v-row v-if="stats.length > 0 && !isLoading" class="justify-center">
          <v-col
            v-for="resource in stats"
            :key="resource.id"
            cols="12"
            sm="6"
            md="4"
            lg="3"
          >
            <StatsCard v-bind="resource" />
          </v-col>
        </v-row>
      </v-fade-transition>
    </v-container>

    <HomeExperienceSection />
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })

const { stats, isLoading } = useStats({ exclude: ["users", "clusters", "balance"] })
</script>
