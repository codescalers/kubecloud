<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h4 font-weight-bold">
        Workflows
      </h1>
      <p class="text-body-1 mt-2 opacity-70">
        All platform workflows
      </p>
    </div>

    <div class="d-flex justify-space-between align-center mb-6">
      <v-select
        v-model="status"
        hide-details
        clearable
        max-width="300"
        label="Filter by status"
        :items="[
          { title: 'Pending', value: 'pending' },
          { title: 'Running', value: 'running' },
          { title: 'Completed', value: 'completed' },
          { title: 'Failed', value: 'failed' },
        ]"
        variant="outlined"
        density="compact"
      />

      <v-btn
        variant="text"
        class="border"
        prepend-icon="mdi-refresh"
        text="Refresh"
        :loading="isLoading"
        @click="loadWorkflows()"
      />
    </div>
    <v-data-table-server
      v-model:items-per-page="limit"
      v-model:page="page"
      :headers="[
        { title: 'name', key: 'name', align: 'center', sortable: false },
        { title: 'status', key: 'status', align: 'center', sortable: false },
        { title: 'progress', key: 'progress', align: 'center', sortable: false },
        { title: 'current step', key: 'step', align: 'center', sortable: false },
        { title: 'user id', key: 'userId', align: 'center', sortable: false },
        { title: 'created at', key: 'createdAt', align: 'center', sortable: false },
        { title: 'actions', key: 'actions', align: 'center', sortable: false },
      ]"
      :items="state?.workflows ?? []"
      :items-length="state?.total ?? 0"
      :loading="isLoading"
    >
      <template #item="{ item }">
        <WorkflowRow :workflow="item" @view="reveal(item)" />
      </template>
    </v-data-table-server>

    <v-dialog
      :model-value="isRevealed"
      max-width="700"
      max-height="90vh"
      scrollable
      @update:model-value="cancel()"
    >
      <WorkflowDialogCard :workflow="selectedWorkflow!" @cancel="cancel()" />
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { ServicesAdminWorkflow } from "../../generated/api"

const api = useApi()

const limit = ref(10)
const page = ref(1)
const status = ref<string>()

const {
  state,
  isLoading,
  execute: loadWorkflows,
} = useAsyncState(
  async () => {
    const { data } = await api.admin.listAllWorkflows(status.value, {
      params: {
        page: page.value,
        limit: limit.value,
      },
    })
    return data.data
  },
  null,
  { immediate: $meta.client, resetOnExecute: false },
)

watchDebounced([page, limit, status], () => loadWorkflows(), { debounce: 100 })

const { isRevealed, reveal, cancel, data: selectedWorkflow } = useDialog<ServicesAdminWorkflow>()
</script>
