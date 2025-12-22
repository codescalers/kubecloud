<template>
  <StickySidebarLayout :cols="9">
    <template #items>
      <v-list-item link exact title="Overview" to="/dashboard" />
      <v-list-item link exact title="Clusters" to="/dashboard/clusters" />
      <v-list-item link exact title="My Nodes" to="/dashboard/nodes" />
      <v-list-item link exact title="SSH Keys" to="/dashboard/ssh" />
      <v-list-item link exact title="Add Funds" to="/dashboard/funds" />
      <v-list-item link exact title="Billing History" to="/dashboard/billing-history" />
      <v-list-item link exact title="Payments" to="/dashboard/payments" />
      <v-list-item link exact title="Vouchers" to="/dashboard/vouchers" />
    </template>

    <NuxtPage />
    <div :style="{ height: '2000px' }" />
  </StickySidebarLayout>
</template>
