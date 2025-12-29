<template>
  <StickySidebarLayout :is-open="isOpen" :is-fluid="isFluid">
    <template #sidebar>
      <v-list class="overflow-hidden w-100">
        <v-list-item link exact title="Overview" to="/admin" />
        <v-list-item link exact title="Users" to="/admin/users" />
        <v-list-item link exact title="Vouchers" to="/admin/vouchers" />
        <v-list-item link exact title="System" to="/admin/system" />
        <v-list-item link exact title="Invoices" to="/admin/invoices" />
        <v-list-item link exact title="Payments" to="/admin/payments" />
        <v-list-item link exact title="Emails" to="/admin/emails" />
        <v-list-item link exact title="Workflows" to="/admin/workflows" />
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
