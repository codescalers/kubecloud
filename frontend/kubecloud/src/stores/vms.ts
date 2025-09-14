import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useNotificationStore } from './notifications'
import { api } from '../utils/api'

export interface VM {
  id: number
  project_name: string
  vm: {
    name: string
    node_id: number
    cpu: number
    memory: number
    root_size: number
    disk_size: number
    flist?: string
    entrypoint?: string
    env_vars?: Record<string, string>
    status?: string
  }
  created_at: string
}

export interface VMInput {
  name: string
  node_id: number
  cpu: number
  memory: number
  root_size: number
  disk_size: number
  flist?: string
  entrypoint?: string
  env_vars?: Record<string, string>
}

export const useVMStore = defineStore('vms', () => {
  const vms = ref<VM[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const notificationStore = useNotificationStore()

  const fetchVMs = async () => {
    isLoading.value = true
    error.value = null
    try {
      const response = await api.get('/v1/deployments/vms', { requiresAuth: true })
      vms.value = Array.isArray(response.data) ? response.data : []
    } catch (err: any) {
      const errorMsg = err.message || 'Failed to fetch VMs'
      error.value = errorMsg
      notificationStore.error('Error', errorMsg)
    } finally {
      isLoading.value = false
    }
  }

  const deployVM = async (vmData: VMInput) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await api.post('/v1/deployments/vms', vmData, { requiresAuth: true })
      notificationStore.success('VM Deployment Started', `VM "${vmData.name}" deployment has been initiated`)
      await fetchVMs() // Refresh the list
      return response.data
    } catch (err: any) {
      const errorMsg = err.message || 'Failed to deploy VM'
      error.value = errorMsg
      notificationStore.error('Deployment Failed', errorMsg)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const getVM = async (id: number) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await api.get(`/v1/deployments/vms/${id}`, { requiresAuth: true })
      return response.data
    } catch (err: any) {
      const errorMsg = err.message || 'Failed to fetch VM details'
      error.value = errorMsg
      notificationStore.error('Error', errorMsg)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteVM = async (id: number) => {
    isLoading.value = true
    error.value = null
    try {
      await api.delete(`/v1/deployments/vms/${id}`, { requiresAuth: true })
      notificationStore.success('VM Deletion Started', 'VM deletion has been initiated')
      await fetchVMs() // Refresh the list
    } catch (err: any) {
      const errorMsg = err.message || 'Failed to delete VM'
      error.value = errorMsg
      notificationStore.error('Deletion Failed', errorMsg)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteAllVMs = async () => {
    isLoading.value = true
    error.value = null
    try {
      // Delete all VMs one by one since there's no bulk delete endpoint
      const deletePromises = vms.value.map(vm => api.delete(`/v1/deployments/vms/${vm.id}`, { requiresAuth: true }))
      await Promise.all(deletePromises)
      notificationStore.success('All VMs Deletion Started', 'All VM deletions have been initiated')
      await fetchVMs() // Refresh the list
    } catch (err: any) {
      const errorMsg = err.message || 'Failed to delete all VMs'
      error.value = errorMsg
      notificationStore.error('Bulk Deletion Failed', errorMsg)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Computed properties
  const vmCount = computed(() => vms.value.length)
  const runningVMs = computed(() => vms.value.filter(vm => vm.vm.status === 'running').length)

  return {
    vms,
    isLoading,
    error,
    vmCount,
    runningVMs,
    fetchVMs,
    deployVM,
    getVM,
    deleteVM,
    deleteAllVMs
  }
})
