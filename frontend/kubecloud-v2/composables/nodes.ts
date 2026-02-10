import type { UseAsyncStateOptions } from "@vueuse/core"
import type { HandlersNodesWithDiscount } from "~/generated/api"
import { AxiosError } from "axios"

export function useNodeMonitoringUrl(node: () => HandlersNodesWithDiscount) {
  const api = useApi()

  return computedAsync(async () => {
    const { twinId, farmId } = node()
    const accountId = await api.helpers.getAccountId(twinId!)
    const params = new URLSearchParams({
      "orgId": "2",
      "refresh": "30s",
      "var-network": "dev",
      "var-farm": farmId!.toString(),
      "var-node": accountId,
      "var-diskdevices": "[a-z]+|nvme[0-9]+n[0-9]+|mmcblk[0-9]+",
    })
    return `https://metrics.grid.tf/d/rYdddlPWkfqwf/zos-host-metrics?${params.toString()}`
  })
}

export function useNodeReserve(node: () => HandlersNodesWithDiscount, options?: UseAsyncStateOptions<true, boolean>) {
  const api = useApi()
  const toast = useToast()
  return useAsyncState(
    async () => {
      const { nodeId } = node()
      const { data } = await api.nodes.reserveNode(nodeId!.toString())
      const workflowId = data.data?.workflow_id
      if (workflowId && await api.helpers.awaitWorkflowCompletion(workflowId)) {
        toast.success({ message: "Node reserved successfully" })
        return true
      }
      return false
    },
    false,
    {
      immediate: false,
      onError(e: unknown) {
        if (e instanceof AxiosError) {
          toast.error({ message: e.response?.data?.message ?? "An unknown error occurred" })
        }
      },
      ...options,
    },
  )
}
