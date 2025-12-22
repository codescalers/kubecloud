<template>
  <div>
    <h1>Dashboard Nodes {{ isLoading }}</h1>

    <pre>
      {{ state }}
    </pre>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const { state, isLoading } = useAsyncState(async () => {
  const { data } = await api.nodes.listReservedNodes()
  return data
}, null)
</script>
