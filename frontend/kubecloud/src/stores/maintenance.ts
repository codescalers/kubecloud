import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { MaintenanceService, type MaintenanceStatus } from '@/utils/maintenanceService'

export const useMaintenanceStore = defineStore('maintenance', () => {
  const maintenanceStatus = ref<MaintenanceStatus>({ enabled: false })
  const isLoading = ref(false)
  const lastChecked = ref<Date | null>(null)

  const isMaintenanceMode = computed(() => maintenanceStatus.value.enabled)
  
  async function checkMaintenanceStatus(): Promise<void> {
    isLoading.value = true
    try {
      const status = await MaintenanceService.getMaintenanceStatus()
      maintenanceStatus.value = status
      lastChecked.value = new Date()
    } catch (error) {
      console.error('Failed to check maintenance status:', error)
    } finally {
      isLoading.value = false
    }
  }

  function isRouteAllowed(routePath: string): boolean {
    if (!isMaintenanceMode.value) return true
    return MaintenanceService.isRouteAllowedDuringMaintenance(routePath)
  }

 

  return {
    maintenanceStatus,
    isLoading,
    lastChecked,
    isMaintenanceMode,
    checkMaintenanceStatus,
    isRouteAllowed,
  }
})
