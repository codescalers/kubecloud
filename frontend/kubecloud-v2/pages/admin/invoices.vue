<template>
  <div>{{ state[0] }}</div>
</template>

<script setup lang="ts">
import type { ModelsInvoice } from "../../generated/api"

const api = useApi()

const { state } = useAsyncState(async () => {
  const { data } = await api.admin.getAllInvoices()
  return (data.data as unknown as { invoices: ModelsInvoice[] }).invoices ?? []
}, [])
</script>
