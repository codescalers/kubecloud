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
        nodesStoragePool.value.set(nodeId, nodeStoragePoolResponse.data.data.pools)
        return nodeStoragePoolResponse.data.data.pools
      } catch (error) {
        console.error(error)
        throw new Error("Failed to verify node storage pool")
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

  return {
    validateNodeStoragePool
  }
}
