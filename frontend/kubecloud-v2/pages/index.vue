<template>
  <v-container class="pt-16">
    <h1 class="text-h4 font-weight-bold text-center mb-4">Mycelium Cloud</h1>
    <p class="text-body-1 text-center text-primary mx-auto" :style="{ maxWidth: '600px' }">
      Revolutionary Kubernetes platform that transforms how teams deploy and manage cloud-native
      applications at scale
    </p>

    {{ isLoading }}
    {{ state }}
  </v-container>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "public" })

const api = useApi()
const { state, isLoading } = useAsyncState(async () => {
  const { data } = await api.admin.getStats({ unauthenticated: true })
  return data.data
}, null)
</script>
