<template>
  <StickySidebarLayout :is-open="isOpen" :is-fluid="isFluid">
    <template #sidebar>
      <v-list class="overflow-hidden w-100">
        <v-list-item link exact title="Overview" :to="ROUTES.Dashboard()" />
        <v-list-item link exact title="Clusters" :to="ROUTES.Dashboard.Clusters()" />
        <v-list-item link exact title="My Nodes" :to="ROUTES.Dashboard.MyNodes()" />
        <v-list-item link exact title="SSH Keys" :to="ROUTES.Dashboard.SshKeys()" />
        <v-list-item link exact title="Add Funds" :to="ROUTES.Dashboard.Funds()" />
        <v-list-item link exact title="Billing History" :to="ROUTES.Dashboard.BillingHistory()" />
        <v-list-item link exact title="Payments" :to="ROUTES.Dashboard.Payments()" />
        <v-list-item link exact title="Vouchers" :to="ROUTES.Dashboard.Vouchers()" />
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
