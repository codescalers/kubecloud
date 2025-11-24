import { userService } from "@/utils/userService";

export default function useNodeStoragePool() {
  const failedToCheckStoragePoolError = () => {
    return new Error('Something went wrong while checking status of the node. Please check your connection and try again.')
  }

  const validateNodeStoragePool = async (requiredStorage: number, nodeId: number) => {
    const requiredStorageInBytes = requiredStorage * 1024 * 1024 * 1024;
    try {
      const storagePool = await userService.getStoragePool(nodeId)
      if (storagePool) {
        return storagePool.some((pool) => pool.free >= requiredStorageInBytes)
      }
      return false
    } catch (error) {
      console.error(error)
      throw failedToCheckStoragePoolError()
    }
  };

  const createStoragePoolError = (nodeId: number) => {
    return `Although node ${nodeId} appears to have sufficient storage capacity for your workload, it lacks a single internal partition capable of accommodating it. Please select a different node.`
  }

  return {
    validateNodeStoragePool,
    createStoragePoolError,
    failedToCheckStoragePoolError
  }
}
