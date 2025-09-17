import { ref } from 'vue'
import * as vm from '../utils/vm'
import { useNotificationStore } from '@/stores/notifications'
import type { VMInput, VM } from '@/types/vms'

/**
 * Vue composable for managing Virtual Machine operations
 *
 * Provides reactive state management and methods for VM CRUD operations
 * including deployment, fetching, deletion, and state tracking.
 *
 * @returns Object containing VM state and management methods
 *
 * @example
 * ```typescript
 * const { vmCount, isLoading, fetchVMs, deployVM, getVM } = useVMs()
 *
 * // Fetch all VMs
 * await fetchVMs()
 *
 * // Deploy a new VM
 * await deployVM({
 *   name: 'my-vm',
 *   node_id: 1,
 *   cpu: 2,
 *   memory: 4096,
 *   disk: 20
 * })
 * ```
 */
export default function useVMs() {
  /** Reactive count of total VMs */
  const vmCount = ref<number>(0)

  /** Reactive loading state for all VM operations */
  const isLoading = ref<boolean>(false)

  /** Notification store instance for error handling */
  const notificationStore = useNotificationStore()

  /** Reactive array of VM objects */
  const vms = ref<VM[]>([])

  /**
   * Fetches all Virtual Machines from the API
   *
   * Updates the vmCount and vms reactive references with the fetched data.
   * Shows error notifications if the request fails.
   *
   * @async
   * @throws Will show error notification if API request fails
   */
  const fetchVMs = async () => {
    isLoading.value = true
    try {
      const response = await vm.listVMs()
      vmCount.value = response.data.length
      vms.value = response.data
    } catch (err: any) {
      notificationStore.error('Failed to fetch VMs:', err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Deploys a new Virtual Machine
   *
   * @param {VMInput} vmData - The VM configuration data
   * @async
   * @throws Will show error notification if deployment fails
   */
  const deployVM = async (vmData: VMInput) => {
    isLoading.value = true
    try {
      await vm.deployVM(vmData)
    } catch (err: any) {
      notificationStore.error('Failed to deploy VM:', err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Retrieves a specific Virtual Machine by ID
   *
   * @param id - The unique identifier of the VM
   * @returns Promise that resolves to the VM data or undefined if error occurs
   * @async
   * @throws Will show error notification if fetch fails
   */
  const getVM = async (id: number) => {
    isLoading.value = true
    try {
      const response = await vm.getVM(id)
      return response.data
    } catch (err: any) {
      notificationStore.error('Failed to fetch VM:', err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Deletes a specific Virtual Machine by ID
   *
   * @param id - The unique identifier of the VM to delete
   * @async
   * @throws Will show error notification if deletion fails
   */
  const deleteVM = async (id: number) => {
    isLoading.value = true
    try {
      await vm.deleteVM(id)
    } catch (err: any) {
      notificationStore.error('Failed to delete VM:', err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Deletes all Virtual Machines
   *
   * ⚠️ **Warning**: This operation will permanently delete all VMs.
   * Use with caution as this action cannot be undone.
   *
   * @async
   * @throws Will show error notification if bulk deletion fails
   */
  const deleteAllVMs = async () => {
    isLoading.value = true
    try {
      await vm.deleteAllVMs()
    } catch (err: any) {
      notificationStore.error('Failed to delete all VMs:', err)
    } finally {
      isLoading.value = false
    }
  }

  return {
    /** Reactive count of total VMs */
    vmCount,
    /** Reactive loading state indicator */
    isLoading,
    /** Reactive array of VM objects */
    vms,
    /** Fetch all VMs from the API */
    fetchVMs,
    /** Deploy a new VM */
    deployVM,
    /** Get a specific VM by ID */
    getVM,
    /** Delete a specific VM by ID */
    deleteVM,
    /** Delete all VMs (use with caution) */
    deleteAllVMs,
  }
}
