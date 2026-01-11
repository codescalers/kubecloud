<template>
  <div>
    <h1>Dashboard Nodes {{ isLoading }}</h1>
    <v-btn to="/dashboard/nodes/explorer">
      Explorer
    </v-btn>

    <pre>
      {{ state?.data }}
    </pre>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const { state, isLoading } = useAsyncState(async () => {
  const { data } = await api.nodes.listRentableNodes()
  return data
}, null)
</script>
