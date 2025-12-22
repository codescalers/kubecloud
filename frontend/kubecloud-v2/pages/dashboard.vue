<template>
  <v-container>
    <v-layout class="bg-red">
      <v-navigation-drawer permanent sticky>
        <v-list-item link exact title="Overview" to="/dashboard" />
        <v-list-item link exact title="Clusters" to="/dashboard/clusters" />
        <v-list-item link exact title="My Nodes" to="/dashboard/nodes" />
        <!-- <v-list-item link exact title="Profile" to="/dashboard/profile" /> -->
        <v-list-item link exact title="SSH Keys" to="/dashboard/ssh" />
        <v-list-item link exact title="Add Funds" to="/dashboard/funds" />
        <v-list-item link exact title="Billing History" to="/dashboard/billing-history" />
        <v-list-item link exact title="Payments" to="/dashboard/payments" />
        <v-list-item link exact title="Vouchers" to="/dashboard/vouchers" />
      </v-navigation-drawer>

      <v-main>
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
        <NuxtPage />
      </v-main>
    </v-layout>
  </v-container>
</template>
