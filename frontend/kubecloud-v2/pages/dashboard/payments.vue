<template>
  <div>
    <h1>Dashboard Payments {{ isLoading }}</h1>

    <pre>
      {{ state }}
    </pre>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const { state, isLoading } = useAsyncState(async () => {
  const { data } = await api.users.listUserPendingRecords()
  return data
}, null, { immediate: $meta.client })
</script>
