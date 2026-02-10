<template>
  <div>
    <h1>Dashboard Nodes {{ isLoading }}</h1>
    <v-btn :to="ROUTES.Dashboard.MyNodes.Explorer()">
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
  const { data } = await api.nodes.listReservedNodes()
  return data
}, null, { immediate: $meta.client })
</script>
