<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h4 font-weight-bold">
        Invoices
      </h1>
      <p class="text-body-1 mt-2 opacity-70">
        All platform invoices
      </p>
    </div>

    <v-data-table-server
      v-model:items-per-page="limit"
      v-model:page="page"
      :headers="[
        { title: 'id', key: 'id', align: 'center', sortable: false },
        { title: 'user id', key: 'userId', align: 'center', sortable: false },
        { title: 'total', key: 'total', align: 'center', sortable: false },
        { title: 'tax', key: 'tax', align: 'center', sortable: false },
        { title: 'created at', key: 'createdAt', align: 'center', sortable: false },
        { title: 'actions', key: 'actions', align: 'center', sortable: false },
      ]"
      :items="invoices"
      :items-length="state.length"
      :loading="isLoading"
    >
      <template #item="{ item }">
        <InvoiceRow :invoice="item" @view="reveal(item)" />
      </template>
    </v-data-table-server>

    <v-dialog
      :model-value="isRevealed"
      max-width="700"
      max-height="90vh"
      scrollable
      @update:model-value="cancel()"
    >
      <InvoiceDialogCard :invoice="selectedInvoice!" @cancel="cancel()" />
      <!-- <WorkflowDialogCard :workflow="selectedWorkflow!" @cancel="cancel()" /> -->
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import type { ModelsInvoice } from "../../generated/api"

const api = useApi()
const limit = ref(10)
const page = ref(1)

const { state, isLoading } = useAsyncState(async () => {
  const { data } = await api.admin.getAllInvoices()
  return (data.data as unknown as { invoices: ModelsInvoice[] }).invoices ?? []
}, [])

const invoices = computed(() => {
  const v = (page.value - 1) * limit.value
  return state.value.slice(v, v + limit.value)
})

const { isRevealed, reveal, cancel, data: selectedInvoice } = useDialog<ModelsInvoice>()
</script>
