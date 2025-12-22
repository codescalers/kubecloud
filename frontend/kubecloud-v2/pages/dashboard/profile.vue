<template>
  <div>
    <h1>Dashboard Profile {{ isLoading }}</h1>

    <pre>
      {{ state }}
    </pre>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const { state, isLoading } = useAsyncState(async () => {
  const { data } = await api.users.getUser()
  return data
}, null)
</script>
