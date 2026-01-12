<template>
  <StickySidebarLayout :is-open="isOpen" :is-fluid="isFluid">
    <template #sidebar>
      <v-list class="overflow-hidden w-100">
        <v-list-item link exact title="Overview" :to="ROUTES.Admin()" />
        <v-list-item link exact title="Users" :to="ROUTES.Admin.Users()" />
        <v-list-item link exact title="Vouchers" :to="ROUTES.Admin.Vouchers()" />
        <v-list-item link exact title="System" :to="ROUTES.Admin.System()" />
        <v-list-item link exact title="Invoices" :to="ROUTES.Admin.Invoices()" />
        <v-list-item link exact title="Payments" :to="ROUTES.Admin.Payments()" />
        <v-list-item link exact title="Emails" :to="ROUTES.Admin.Emails()" />
        <v-list-item link exact title="Workflows" :to="ROUTES.Admin.Workflows()" />
      </v-list>
    </template>

    <NuxtPage />
  </StickySidebarLayout>
</template>

<script setup lang="ts">
definePageMeta({ middleware: "admin" })

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
