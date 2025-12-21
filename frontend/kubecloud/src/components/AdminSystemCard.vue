<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">System Stats</h2>
      <p class="section-subtitle">Platform health and performance metrics</p>
    </div>
    <div class="dashboard-card">
      <div class="empty-state-content">
        <v-icon icon="mdi-cog" size="64" color="var(--color-text-muted)" class="mb-4"></v-icon>
        <h3 class="empty-state-title">System Monitoring</h3>
        <p class="empty-state-message">Advanced system monitoring, logs, and platform status features coming soon. Monitor system health, performance metrics, and resource utilization.</p>
      </div>
    </div>

    <!-- Danger Zone Card -->
    <v-card variant="outlined">
      <v-card-title class="text-h6 font-weight-bold bg-error-lighten-4 text-error px-8 pt-8 pb-4">
        <v-icon icon="mdi-alert-circle" color="error" class="mr-2"></v-icon>
        Danger Zone
      </v-card-title>

      <v-card-text class=" px-9 pb-0" >
        <p class="text-body-2 mb-0">
          <span v-if="isMaintenanceMode">
            Maintenance mode is currently <strong>enabled</strong>. User dashboard access is restricted.
          </span>
          <span v-else>
            Enable maintenance mode to temporarily restrict access to the platform.
          </span>
        </p>
      </v-card-text>

      <v-card-actions class="px-9 pb-8">
        <v-btn 
          :color="isMaintenanceMode ? 'success' : 'error'" 
          variant="outlined" 
          :loading="isLoading" 
          @click="toggleMaintenanceMode"
          :prepend-icon="isMaintenanceMode ? 'mdi-check-circle' : 'mdi-wrench'">
          {{ isMaintenanceMode ? 'Disable Maintenance Mode' : 'Enable Maintenance Mode' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { adminService } from '@/utils/adminService'
import { useMaintenanceStore } from '@/stores/maintenance'

const maintenanceStore = useMaintenanceStore()
const isLoading = ref(false)

const isMaintenanceMode = computed(() => maintenanceStore.isMaintenanceMode)

// Check maintenance status on component mount
onMounted(async () => {
  await maintenanceStore.checkMaintenanceStatus()
})

// Methods
const toggleMaintenanceMode = async () => {
  isLoading.value = true
  try {
    const newStatus = !isMaintenanceMode.value
    await adminService.SetMaintenanceModeStatus(newStatus)
    await maintenanceStore.checkMaintenanceStatus()
  } catch (error) {
    console.error('Failed to toggle maintenance mode:', error)
  } finally {
    isLoading.value = false
  }
}
</script>