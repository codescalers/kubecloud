import { onUnmounted, ref } from "vue";
import type { NodeStoragePool, StoragePool } from "../types/normalizedNode";
import { api } from "../utils/api";
import type { ApiResponse } from "../utils/api";

export default function useNodeStoragePool() {
  const nodesStoragePool = ref<Map<number, StoragePool[]>>(new Map());

  const getStoragePool = async (nodeId: number) => {
    const storagePool = nodesStoragePool.value.get(nodeId);
    if (!storagePool) {
      try {
        const nodeStoragePoolResponse: ApiResponse<ApiResponse<NodeStoragePool>> = await api.get(`/v1/nodes/${nodeId}/storage-pool`, {
          showNotifications: false
        })
        const pools = nodeStoragePoolResponse.data.data.pools.filter((pool) => pool.type === "ssd")
        nodesStoragePool.value.set(nodeId, pools)
        return pools
      } catch (error) {
        console.error(error)
        throw failedToCheckStoragePoolError()
      }
    }
    return storagePool;
  };

  const validateNodeStoragePool = async (requiredStorage: number, nodeId: number) => {
    const requiredStorageInBytes = requiredStorage * 1024 * 1024 * 1024;
    const storagePool = await getStoragePool(nodeId)
    if (storagePool) {
      return storagePool.some((pool) => pool.free >= requiredStorageInBytes)
    }
    return false
  };

  onUnmounted(() => {
    nodesStoragePool.value.clear()
  })

  const createStoragePoolError = (nodeId: number) => {
    return `Although node ${nodeId} appears to have sufficient storage capacity for your workload, it lacks a single internal partition capable of accommodating it. Please select a different node.`
  }

  const failedToCheckStoragePoolError = () => {
    return new Error('Something went wrong while checking status of the node. Please check your connection and try again.')
  }

  return {
    validateNodeStoragePool,
    createStoragePoolError,
    failedToCheckStoragePoolError
  }
}
