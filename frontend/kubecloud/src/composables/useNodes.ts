import { ref, computed } from 'vue'
import { userService } from '../utils/userService'
import type { RawNode } from '../types/rawNode'
import type { NormalizedNode } from '../types/normalizedNode'
import { normalizeNode } from '../utils/nodeNormalizer'

export interface NodeFilters {
  country?: string
  city?: string
  farm_id?: number
  free_hru?: number
  free_mru?: number
  free_sru?: number
  free_cru?: number
  status?: string
  healthy?: boolean
  rentable?: boolean
  dedicated?: boolean
  certification_type?: string
  grid_version?: number
  region?: string
}

export function useNodes() {
  const nodes = ref<RawNode[]>([])
  const total = ref<number>(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filters = ref<NodeFilters>({})

  // Fetch all available nodes (both reserved and shared)
  async function fetchNodes(nodeFilters?: NodeFilters) {
    nodes.value = []
    loading.value = true
    error.value = null
    try {
      const response = await userService.listNodes(nodeFilters || filters.value)
      const responseData = response.data as any
      if (responseData.data?.nodes) {
        nodes.value = responseData.data.nodes
        total.value = responseData.data.total || 0
      } else {
        nodes.value = []
        total.value = 0
      }
    } catch (err: any) {
      console.error('Failed to fetch nodes:', err)
      error.value = err?.message || 'Failed to fetch nodes'
      nodes.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  // Fetch only reserved nodes
  async function fetchReservedNodes() {
    loading.value = true
    error.value = null
    try {
      const response = await userService.listReservedNodes()
      const responseData = response.data as any
      if (responseData.data?.nodes) {
        nodes.value = responseData.data.nodes
        total.value = responseData.data.total || 0
      } else {
        nodes.value = []
        total.value = 0
      }
    } catch (err: any) {
      console.error('Failed to fetch reserved nodes:', err)
      error.value = err?.message || 'Failed to fetch reserved nodes'
      nodes.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  // Fetch only rentable/shared nodes
  async function fetchRentableNodes(nodeFilters?: NodeFilters) {
    const rentableFilters = { ...nodeFilters, rentable: true }
    await fetchNodes(rentableFilters)
  }

  // Update filters and refetch
  async function updateFilters(newFilters: NodeFilters) {
    filters.value = { ...filters.value, ...newFilters }
    await fetchNodes()
  }

  // Clear all filters
  async function clearFilters() {
    filters.value = {}
    await fetchNodes()
  }

  // Normalized nodes for UI
  const normalizedNodes = computed<NormalizedNode[]>(() => nodes.value.map(normalizeNode))

  // Computed properties to categorize nodes
  const reservedNodes = computed<RawNode[]>(() =>
    nodes.value.filter(node => node.rented && node.rentedByTwinId)
  )

  const sharedNodes = computed<RawNode[]>(() =>
    nodes.value.filter(node => !node.rented && node.rentable)
  )

  const availableNodes = computed<RawNode[]>(() =>
    nodes.value.filter(node => node.rentable || (node.rented && node.rentedByTwinId))
  )

  // Helper function to determine node type
  function getNodeType(node: RawNode): 'reserved' | 'shared' {
    if (node.rented && node.rentedByTwinId) {
      return 'reserved'  // 50% discount, rented by me
    } else {
      return 'shared'    // Everything else (rentable or not)
    }
  }

  // Helper function to check if a shared node is rentable
  function isNodeRentable(node: RawNode): boolean {
    return !node.rented && node.rentable
  }

  // Fetch account id for a twin id
  async function fetchAccountId(twinId: number): Promise<string> {
    try {
      const response = await userService.getTwinAccount(twinId)
      return response.account_id
    } catch (err: any) {
      console.error('Failed to fetch account id:', err)
      return ''
    }
  }

  return {
    nodes, // RawNode[]
    normalizedNodes, // NormalizedNode[]
    reservedNodes, // Reserved nodes only
    sharedNodes, // Shared/rentable nodes only
    availableNodes, // All available nodes (reserved + shared)
    total,
    loading,
    error,
    filters,
    fetchNodes,
    fetchReservedNodes,
    fetchRentableNodes,
    updateFilters,
    clearFilters,
    getNodeType,
    isNodeRentable,
    fetchAccountId
  }
}
