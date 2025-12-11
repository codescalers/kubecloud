<template>
  <DefaultLayoutBase>
    <SidebarLayoutBase>
      <template #sidebar>
        <v-list-item link title="Overview" to="/dashboard" />
        <v-list-item link title="Clusters" to="/dashboard/clusters" />
        <v-list-item link title="My Nodes" to="/dashboard/nodes" />
        <v-list-item link title="Profile" to="/dashboard/profile" />
        <v-list-item link title="SSH Keys" to="/dashboard/ssh" />
        <v-list-item link title="Add Funds" to="/dashboard/funds" />
        <v-list-item link title="Billing History" to="/dashboard/billing-history" />
        <v-list-item link title="Payments" to="/dashboard/payments" />
        <v-list-item link title="Vouchers" to="/dashboard/vouchers" />
      </template>

      <slot />
    </SidebarLayoutBase>
  </DefaultLayoutBase>
</template>
