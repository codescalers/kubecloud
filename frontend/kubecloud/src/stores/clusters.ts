import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../utils/api'

export interface Cluster {
  id: number
  project_name: string
  cluster: {
    name: string
    status?: string
    region?: string
    nodes?: number
    cpu?: string
    memory?: string
    storage?: string
    tags?: string[]
    [key: string]: any
  }
  created_at: string
  updated_at: string
}

export interface ClusterMetrics {
  cpuUsage: number
  memoryUsage: number
  storageUsage: number
  networkIn: number
  networkOut: number
  activeConnections: number
}

export const useClusterStore = defineStore('clusters', () => {
  // State
  const clusters = ref<Cluster[]>([])
  const selectedCluster = ref<Cluster | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed properties
  const clustersByRegion = computed(() => {
    const grouped: Record<string, Cluster[]> = {}
    clusters.value.forEach(cluster => {
      const region = cluster.cluster.region
      if (!region) return // skip clusters with undefined region
      if (!grouped[region]) {
        grouped[region] = []
      }
      grouped[region].push(cluster)
    })
    return grouped
  })

  // Actions
  const fetchClusters = async () => {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.get('/v1/deployments', { requiresAuth: true })
      const deployments = (response.data as { deployments: Cluster[] }).deployments
      clusters.value = Array.isArray(deployments) ? deployments : []
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch clusters'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteCluster = async (name: string) => {
    isLoading.value = true
    error.value = null

    try {
      await api.delete(`/v1/deployments/${name}`, { requiresAuth: true })
      clusters.value = clusters.value.filter(cluster => cluster.project_name !== name)
      
      if (selectedCluster.value?.project_name === name) {
        selectedCluster.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete cluster'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const getClusterMetrics = async (clusterId: string): Promise<ClusterMetrics> => {
    try {
      // Real API call (replace with actual endpoint if available)
      const response = await api.get(`/clusters/${clusterId}/metrics`)
      return response.data as ClusterMetrics
    } catch (err) {
      throw new Error('Failed to fetch cluster metrics')
    }
  }

  const getClusterByName = async (name: string): Promise<Cluster | null> => {
    isLoading.value = true
    error.value = null
    try {
      const response = await api.get(`/v1/deployments/${name}`, { requiresAuth: true })
      return response.data as Cluster
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch cluster'
      return null
    } finally {
      isLoading.value = false
    }
  }

  const addNodesToCluster = async (clusterName: string, clusterObject: any) => {
    return api.post(`/v1/deployments/${clusterName}/nodes`, clusterObject, { requiresAuth: true })
  }


  const removeNodeFromCluster = async (clusterName: string, nodeName: string) => {
    return api.delete(`/v1/deployments/${clusterName}/nodes/${nodeName}`, { requiresAuth: true })
  }

  const deleteAllDeployments = async () => {
    isLoading.value = true
    error.value = null

    try {
      await api.delete('/v1/deployments', { requiresAuth: true })
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete all deployments'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    // State
    clusters: computed(() => clusters.value),
    selectedCluster: computed(() => selectedCluster.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
    // Computed
    clustersByRegion,
    getClusterByName,
    // Actions
    fetchClusters,
    deleteCluster,
    deleteAllDeployments,
    getClusterMetrics,
    addNodesToCluster,
    removeNodeFromCluster,
  }
}) 