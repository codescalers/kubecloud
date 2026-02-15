<template>
  <v-app>
    <app-navbar />

    <v-main>
      <NuxtPage />
    </v-main>
    <app-footer />
  </v-app>
</template>

<!-- <script setup lang="ts">
const api = useApi()

useAsyncState(
  () => api.notifications.notificationsGet(),
  null,
  {
    immediate: $meta.client,
    onSuccess(data) {
      console.log(data)
    },
  },
)
</script> -->
