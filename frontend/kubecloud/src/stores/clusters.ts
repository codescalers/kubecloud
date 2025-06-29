import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { mockApi, mockClusters, MOCK_CONFIG } from '@/utils/mockData'

export interface Cluster {
  id: string
  name: string
  status: 'running' | 'stopped' | 'starting' | 'stopping' | 'error'
  region: string
  nodes: number
  cpu: string
  memory: string
  storage: string
  createdAt: string
  lastUpdated: string
  cost: number
  tags: string[]
}

export interface ClusterMetrics {
  cpuUsage: number
  memoryUsage: number
  storageUsage: number
  networkIn: number
  networkOut: number
  activeConnections: number
}

export interface CreateClusterRequest {
  name: string
  region: string
  nodes: number
  nodeType: string
  tags?: string[]
}

export const useClusterStore = defineStore('clusters', () => {
  // State
  const clusters = ref<Cluster[]>([])
  const selectedCluster = ref<Cluster | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed properties
  const runningClusters = computed(() => 
    clusters.value.filter(cluster => cluster.status === 'running')
  )

  const stoppedClusters = computed(() => 
    clusters.value.filter(cluster => cluster.status === 'stopped')
  )

  const totalCost = computed(() => 
    clusters.value.reduce((sum, cluster) => sum + cluster.cost, 0)
  )

  const clustersByRegion = computed(() => {
    const grouped: Record<string, Cluster[]> = {}
    clusters.value.forEach(cluster => {
      if (!grouped[cluster.region]) {
        grouped[cluster.region] = []
      }
      grouped[cluster.region].push(cluster)
    })
    return grouped
  })

  // Actions
  const fetchClusters = async () => {
    isLoading.value = true
    error.value = null

    try {
      if (MOCK_CONFIG.enabled) {
        const response = await mockApi.get<{ data: Cluster[] }>('/clusters')
        clusters.value = response.data
      } else {
        // Real API call would go here
        // const response = await api.get('/clusters')
        // clusters.value = response.data
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch clusters'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const createCluster = async (clusterData: CreateClusterRequest) => {
    isLoading.value = true
    error.value = null

    try {
      if (MOCK_CONFIG.enabled) {
        const response = await mockApi.post<{ data: Cluster }>('/clusters', clusterData)
        const newCluster: Cluster = {
          ...response.data,
          status: 'starting',
          createdAt: new Date().toISOString(),
          lastUpdated: new Date().toISOString(),
          cost: 0, // Will be calculated once running
        }
        clusters.value.push(newCluster)
        return newCluster
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create cluster'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteCluster = async (clusterId: string) => {
    isLoading.value = true
    error.value = null

    try {
      if (MOCK_CONFIG.enabled) {
        await mockApi.delete(`/clusters/${clusterId}`)
        clusters.value = clusters.value.filter(cluster => cluster.id !== clusterId)
        
        if (selectedCluster.value?.id === clusterId) {
          selectedCluster.value = null
        }
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete cluster'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const startCluster = async (clusterId: string) => {
    const cluster = clusters.value.find(c => c.id === clusterId)
    if (!cluster) throw new Error('Cluster not found')

    cluster.status = 'starting'
    
    try {
      if (MOCK_CONFIG.enabled) {
        await mockApi.post(`/clusters/${clusterId}/start`, { action: 'start' })
        // Simulate startup delay
        setTimeout(() => {
          cluster.status = 'running'
          cluster.lastUpdated = new Date().toISOString()
        }, 3000)
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      cluster.status = 'error'
      throw err
    }
  }

  const stopCluster = async (clusterId: string) => {
    const cluster = clusters.value.find(c => c.id === clusterId)
    if (!cluster) throw new Error('Cluster not found')

    cluster.status = 'stopping'
    
    try {
      if (MOCK_CONFIG.enabled) {
        await mockApi.post(`/clusters/${clusterId}/stop`, { action: 'stop' })
        // Simulate shutdown delay
        setTimeout(() => {
          cluster.status = 'stopped'
          cluster.lastUpdated = new Date().toISOString()
        }, 2000)
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      cluster.status = 'error'
      throw err
    }
  }

  const updateCluster = async (clusterId: string, updates: Partial<Cluster>) => {
    const cluster = clusters.value.find(c => c.id === clusterId)
    if (!cluster) throw new Error('Cluster not found')

    try {
      if (MOCK_CONFIG.enabled) {
        await mockApi.put(`/clusters/${clusterId}`, updates)
        Object.assign(cluster, updates, { lastUpdated: new Date().toISOString() })
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update cluster'
      throw err
    }
  }

  const selectCluster = (cluster: Cluster | null) => {
    selectedCluster.value = cluster
  }

  const getClusterMetrics = async (clusterId: string): Promise<ClusterMetrics> => {
    try {
      if (MOCK_CONFIG.enabled) {
        // Mock metrics
        return {
          cpuUsage: Math.random() * 100,
          memoryUsage: Math.random() * 100,
          storageUsage: Math.random() * 100,
          networkIn: Math.random() * 1000,
          networkOut: Math.random() * 1000,
          activeConnections: Math.floor(Math.random() * 100),
        }
      } else {
        // Real API call would go here
        throw new Error('API not ready - use mock mode')
      }
    } catch (err) {
      throw new Error('Failed to fetch cluster metrics')
    }
  }

  const initializeClusters = () => {
    // Load initial mock data
    if (MOCK_CONFIG.enabled && clusters.value.length === 0) {
      clusters.value = [...mockClusters] as Cluster[]
    }
  }

  return {
    // State
    clusters: computed(() => clusters.value),
    selectedCluster: computed(() => selectedCluster.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),

    // Computed
    runningClusters,
    stoppedClusters,
    totalCost,
    clustersByRegion,

    // Actions
    fetchClusters,
    createCluster,
    deleteCluster,
    startCluster,
    stopCluster,
    updateCluster,
    selectCluster,
    getClusterMetrics,
    initializeClusters,
  }
}) 