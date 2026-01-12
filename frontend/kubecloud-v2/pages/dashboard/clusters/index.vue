<template>
  <div>
    <h1>Dashboard Clusters {{ isLoading }}</h1>

    <VBtn :to="ROUTES.Dashboard.Clusters.Deploy()" text="Deploy Cluster" />
    <VBtn :to="ROUTES.Dashboard.Clusters('123')" text="Deploy Cluster" />

    <pre>
        {{ state }}
      </pre>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const { state, isLoading } = useAsyncState(async () => {
  const { data } = await api.deployments.deploymentsGet()
  return data
}, null)
</script>
