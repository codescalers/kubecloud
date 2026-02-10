<template>
  <tr class="text-no-wrap py-2">
    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ deployment.cluster?.name }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ deployment.cluster?.nodes.length }}</span>
    </td>

    <td class="text-subtitle-2 text-center">
      <span class="opacity-50">{{ createdAt }}</span>
    </td>

    <td>
      <div class="d-flex justify-center align-center ga-4">
        <v-btn
          variant="plain"
          border
          prepend-icon="mdi-eye-outline"
          size="small"
          text="View"
          :to="ROUTES.Dashboard.Clusters(deployment.cluster?.name)"
          :disabled="disabled"
        />

        <v-btn
          variant="text"
          border
          prepend-icon="mdi-download-outline"
          size="small"
          text="Download"
          :loading="downloading"
          :href="binary"
          :download="`${deployment.cluster?.name}-kubeconfig.yaml`"
          :disabled="disabled"
        />

        <v-btn
          variant="text"
          border
          prepend-icon="mdi-plus"
          size="small"
          text="Add Node"
          color="primary"
          :disabled="disabled"
          @click="onAddNode()"
        />

        <v-btn
          variant="text"
          border
          prepend-icon="mdi-trash-can-outline"
          size="small"
          text="Delete"
          color="error"
          :disabled="disabled"
          @click="onDelete()"
        />
      </div>
    </td>
  </tr>
</template>

<script setup lang="ts">
import type { ServicesClusterData } from "~/generated/api"

const props = defineProps<{ deployment: ServicesClusterData, disabled: boolean }>()
defineEmits<{ (e: "view"): void }>()

const createdAt = useDateFormat(() => props.deployment.created_at, DATE_FORMAT)

const api = useApi()

const { state: kubeconfig, isLoading: downloading } = useAsyncState(
  () => api.helpers.getKubeconfig(props.deployment.cluster?.name ?? ""),
  "",
  { immediate: $meta.client },
)

const binary = computed(() => {
  const data = kubeconfig.value

  if (!data) {
    return ""
  }

  const blob = new Blob([data], { type: "application/json" })
  return URL.createObjectURL(blob)
})

const toast = useToast()
const ctx = inject(DeploymentDialogCtxKey)!

async function onAddNode() {
  const result = await ctx.addNode(props.deployment)
  // TODO: add node
  if (result) {
    toast.success({ message: "Node added successfully" })
  }
}

async function onDelete() {
  const result = await ctx.delete(props.deployment)
  // TODO: delete cluster
  if (result) {
    toast.success({ message: "Deployment deleted successfully" })
  }
}
</script>
