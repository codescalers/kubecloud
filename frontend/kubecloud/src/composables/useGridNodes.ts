import { ref } from 'vue'
import type { RawNode } from '../types/rawNode'
import { api } from '@/utils/api'

export interface GridNodeFilters {
  healthy?: boolean
  size?: number
  page?: number
}

export function useGridNodes() {
  const gridNodes = ref<RawNode[]>([])
  const total = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Fetch all grid nodes from the public endpoint
  async function fetchGridNodes(filters?: GridNodeFilters) {
    gridNodes.value = []
    loading.value = true
    error.value = null
    try {
      let endpoint = '/v1/nodes'
      if (filters && Object.keys(filters).length > 0) {
        const queryParams = new URLSearchParams()
        Object.entries(filters).forEach(([key, value]) => {
          if (value !== undefined && value !== null) {
            queryParams.append(key, String(value))
          }
        })
        endpoint += `?${queryParams.toString()}`
      }
      const response = await api.get(endpoint, {
        requiresAuth: false,
        showNotifications: false
      })
      const responseData = response.data as any
      if (responseData.data?.nodes) {
        gridNodes.value = responseData.data.nodes
        total.value = responseData.data.total || 0
      } else {
        gridNodes.value = []
        total.value = 0
      }
    } catch (err: any) {
      error.value = err?.message || 'Failed to fetch grid nodes'
      gridNodes.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  return {
    gridNodes,
    total,
    loading,
    error,
    fetchGridNodes
  }
}
