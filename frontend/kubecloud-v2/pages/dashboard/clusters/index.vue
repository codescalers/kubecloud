<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h5 font-weight-bold">
        Kubernetes Clusters
      </h1>
      <p class="text-body-2 mt-1 text-accent">
        Manage your cloud-native infrastructure
      </p>
    </div>

    <div class="d-flex justify-end ga-4 mb-8">
      <v-btn
        prepend-icon="mdi-plus"
        text="New Cluster"
        variant="tonal"
        color="primary"
        :to="ROUTES.Dashboard.Clusters.Deploy()"
      />
      <v-btn
        prepend-icon="mdi-trash-can-outline"
        color="error"
        variant="text"
        text="Delete All"
        border
        class="border-error"
        :style="{ '--v-border-opacity': '0.3' }"
        :loading="isDeleting"
        :disabled="deployments.length === 0"
        @click="onDeleteAll()"
      />
    </div>

    <v-row class="mb-4">
      <v-col cols="9">
        <v-text-field
          label="Search by name"
          placeholder="Search by name"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
          density="compact"
          hide-details
        />
      </v-col>

      <v-col cols="3">
        <v-select
          label="Sort By"
          :items="sortByItems"
          variant="outlined"
          density="compact"
          hide-details
          clearable
        />
      </v-col>
    </v-row>

    <v-data-table
      :headers="[
        { title: 'Name', key: 'name', align: 'center', sortable: false },
        { title: 'Nodes', key: 'nodes', align: 'center', sortable: false },
        { title: 'Created At', key: 'createdAt', align: 'center', sortable: false },
        { title: 'Actions', key: 'actions', align: 'center', sortable: false },
      ]"
      :items="deployments"
      :loading="isLoading"
    >
      <template #item="{ item }">
        <DeploymentRow :deployment="item" :disabled="isDeleting" />
      </template>
    </v-data-table>

    <v-dialog
      :model-value="isRevealed"
      max-width="600"
      scrollable
      @update:model-value="cancel()"
    >
      <v-card>
        <v-btn @click="confirm()">
          Ok
        </v-btn>
        <v-btn @click="cancel()">
          Cancel
        </v-btn>
      </v-card>
    </v-dialog>

    <v-dialog
      :model-value="addNode.isRevealed.value"
      max-width="600"
      scrollable
      @update:model-value="addNode.cancel()"
    >
      <v-card>
        {{ addNode.data.value }}
      </v-card>
    </v-dialog>

    <v-dialog
      :model-value="deleteDeployment.isRevealed.value"
      max-width="600"
      scrollable
      @update:model-value="deleteDeployment.cancel()"
    >
      <v-card>
        <v-btn @click="deleteDeployment.confirm()">
          Ok
        </v-btn>
        <v-btn @click="deleteDeployment.cancel()">
          Cancel
        </v-btn>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { HandlersNodeInput, ServicesClusterData } from "~/generated/api"

const api = useApi()
const toast = useToast()

const { state: deployments, isLoading } = useAsyncState(async () => {
  const { data } = await api.deployments.deploymentsGet()
  return data.data?.deployments ?? []
}, [])

const sortByItems = [
  { title: "Name", value: "name" },
  { title: "Created At", value: "createdAt" },
  { title: "Nodes", value: "nodes" },
]

const { isRevealed, reveal, confirm, cancel } = useDialog()
const { execute: deleteDeployments, isLoading: isDeleting } = useAsyncState(async () => {
  const { data } = await api.deployments.deploymentsDelete()
  const id = data.data?.task_id
  if (id && await api.helpers.awaitWorkflowCompletion(id)) {
    deployments.value = []
    return toast.success({ message: "Deployments deleted successfully" })
  }
  toast.error({ message: data.message ?? "Failed to delete deployments" })
}, null, { immediate: false })

async function onDeleteAll() {
  const { isCanceled } = await reveal()
  if (!isCanceled) {
    deleteDeployments()
  }
}

const addNode = useDialog<ServicesClusterData, HandlersNodeInput>()
const deleteDeployment = useDialog<ServicesClusterData>()

provide(DeploymentDialogCtxKey, {
  addNode: d => addNode.reveal(d).then(d => d.data),
  delete: u => deleteDeployment.reveal(u).then(d => !d.isCanceled),
})
</script>
