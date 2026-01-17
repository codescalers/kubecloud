<template>
  <div>
    <div
      class="border-b d-flex justify-center align-center"
      :style="{
        minHeight: '500px',
        background: `
        url(${myceliumBg1}) no-repeat  bottom right/40%,
        url(${myceliumBg2}) no-repeat  -120px 80px/30%,
        linear-gradient(100.95deg, #0B1A30 71.46%, #225196 145.62%)
      `,
      }"
    >
      <v-container>
        <h1 class="text-h1 font-weight-bold text-center mb-4">
          Mycelium Cloud
        </h1>
        <p class="text-body-1 text-center text-primary mx-auto" :style="{ maxWidth: '600px' }">
          Revolutionary Kubernetes platform that transforms how teams deploy and manage cloud-native
          applications at scale
        </p>
      </v-container>
    </div>

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
  </div>
</template>

<script setup lang="ts">
import myceliumBg1 from "~/assets/images/mycelium_bg_1.svg"
import myceliumBg2 from "~/assets/images/mycelium_bg_2.svg"

definePageMeta({ middleware: "public" })

const { stats, isLoading } = useStats({ exclude: ["users", "clusters", "balance"] })
</script>
