<template>
  <StickySidebarLayout :is-open="isOpen" :is-fluid="isFluid">
    <template #sidebar>
      <v-list class="overflow-hidden w-100">
        <v-list-item link exact title="Overview" to="/dashboard" />
        <v-list-item link exact title="Clusters" to="/dashboard/clusters" />
        <v-list-item link exact title="My Nodes" to="/dashboard/nodes" />
        <v-list-item link exact title="SSH Keys" to="/dashboard/ssh" />
        <v-list-item link exact title="Add Funds" to="/dashboard/funds" />
        <v-list-item link exact title="Billing History" to="/dashboard/billing-history" />
        <v-list-item link exact title="Payments" to="/dashboard/payments" />
        <v-list-item link exact title="Vouchers" to="/dashboard/vouchers" />
      </v-list>
    </template>

    <NuxtPage />
  </StickySidebarLayout>
</template>

<script setup lang="ts">
const isOpen = ref(true)
const isFluid = ref(false)

provide(DashboardLayoutCtxKey, {
  drawer: {
    isOpen,
    open: () => (isOpen.value = true),
    close: () => (isOpen.value = false),
  },
  container: {
    isFluid,
    fluidize: () => (isFluid.value = true),
    containerize: () => (isFluid.value = false),
  },
})
</script>
