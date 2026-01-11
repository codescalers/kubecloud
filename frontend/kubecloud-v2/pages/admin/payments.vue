<template>
  <div>{{ state }}</div>
</template>

<script setup lang="ts">
const api = useApi()

const { state } = useAsyncState(async () => {
  const { data } = await api.admin.listPendingRecords()
  return data.data
}, null)
</script>
