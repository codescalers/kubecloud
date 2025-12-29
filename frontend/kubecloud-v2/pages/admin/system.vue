<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h4 font-weight-bold">System Stats</h1>
      <p class="text-body-1 mt-2 opacity-70">Platform health and performance metrics</p>
    </div>

    <v-card class="mb-8 d-flex align-start ga-4">
      <v-avatar variant="tonal" size="80" class="rounded-xl">
        <v-icon icon="mdi-cog-outline" size="40" />
      </v-avatar>

      <div>
        <h3 class="text-h5 font-weight-bold">
          System Monitoring

          <v-chip
            color="primary"
            size="small"
            class="rounded-lg border border-primary font-weight-bold ml-4"
            :style="{ '--v-border-opacity': 0.5 }"
            text="Coming Soon"
          />
        </h3>
        <p class="text-body-1 mt-2 opacity-70">
          Advanced system monitoring, logs, and platform status features are currently in
          development. Soon you will be able to monitor system health, view detailed performance
          metrics, and track resource utilization in real-time.
        </p>
      </div>
    </v-card>

    <v-card
      color="rgba(var(--v-theme-error), 0.1)"
      class="border elevation-0 d-flex align-start ga-4"
      :style="{ borderColor: 'rgba(var(--v-theme-error), 0.2) !important' }"
    >
      <v-avatar color="error" variant="tonal" size="80" class="rounded-xl">
        <v-icon icon="mdi-alert-circle" size="40" />
      </v-avatar>

      <div>
        <h3 class="text-h5 font-weight-bold text-error">Danger Zone</h3>
        <p class="text-body-1 mt-2 opacity-70">
          Enable maintenance mode to temporarily restrict access to the platform.
        </p>
      </div>

      <v-spacer />

      <v-btn
        class="mt-4"
        variant="flat"
        :color="isEnabled ? 'success' : 'error'"
        prepend-icon="mdi-wrench"
        :text="isEnabled ? 'Disable Maintenance Mode' : 'Enable Maintenance Mode'"
        :loading="isLoading || isLoadingEnabled"
        @click="toggleMaintenanceMode()"
      />
    </v-card>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const { state: isEnabled, isLoading: isLoadingEnabled } = useAsyncState(async () => {
  const { data } = await api.admin.getMaintenanceMode()
  return data.data?.enabled
}, false)

const { execute: toggleMaintenanceMode, isLoading } = useAsyncState(
  async () => {
    await api.admin.setMaintenanceMode({ enabled: !isEnabled.value })
    isEnabled.value = !isEnabled.value
  },
  null,
  { immediate: false }
)
</script>
