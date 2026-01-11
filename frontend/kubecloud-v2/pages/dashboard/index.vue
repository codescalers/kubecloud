<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h5 font-weight-bold">
        Dashboard Overview
      </h1>
      <p class="text-body-2 mt-1 opacity-70">
        Your Mycelium Cloud platform at a glance
      </p>

      <v-row class="mt-4">
        <v-col
          v-for="stat in stats"
          :key="stat.id"
          cols="12"
          md="6"
          lg="4"
        >
          <StatsCard v-bind="stat" />
        </v-col>
      </v-row>
    </div>

    <div>
      <p class="text-h6 font-weight-bold mb-4">
        Quick Actions
      </p>

      <v-row>
        <v-col cols="3">
          <v-btn
            to="/dashboard/clusters/deploy"
            block
            variant="outlined"
            text="Deploy Cluster"
            prepend-icon="mdi-plus"
          />
        </v-col>

        <v-col cols="3">
          <v-btn
            to="/dashboard/nodes/explorer"
            block
            variant="outlined"
            text="Reserve Node"
            prepend-icon="mdi-server-plus"
          />
        </v-col>

        <v-col cols="3">
          <v-btn
            to="/dashboard/ssh"
            block
            variant="outlined"
            text="Add SSH Key"
            prepend-icon="mdi-key-plus"
          />
        </v-col>

        <v-col cols="3">
          <v-btn
            to="/dashboard/funds"
            block
            variant="outlined"
            text="Add Funds"
            prepend-icon="mdi-cash-plus"
          />
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ModelsInvoice } from "~/generated/api"

const api = useApi()

const { state: balance } = useAsyncState(async () => {
  const { data } = await api.users.getUserBalance()
  return toPrecision(data.data?.balance_usd ?? 0, 2)
}, 0)

const { state: spent } = useAsyncState(async () => {
  const { data } = await api.invoices.getInvoices()
  const { invoices } = data.data as unknown as { invoices: ModelsInvoice[] }
  const total = invoices.reduce((a, b) => a + (b.total ?? 0) * 1000, 0) ?? 0
  return toPrecision(total / 1000, 2)
}, 0)

const { state: deploymentsCount } = useAsyncState(async () => {
  const { data } = await api.deployments.deploymentsGet()
  return data.data?.count ?? 0
}, 0)

const stats = computed<StatsResource[]>(() => {
  return [
    {
      id: "clusters",
      title: "Active Clusters",
      icon: "mdi-server",
      color: "#359EFF",
      value: deploymentsCount.value,
    },
    {
      id: "balance",
      title: "Balance",
      icon: "mdi-wallet-bifold",
      color: "#607AFB",
      value: `$${balance.value}`,
    },
    {
      id: "spent",
      title: "Total Spent",
      icon: "mdi-currency-usd",
      color: "#39E079",
      value: `$${spent.value}`,
    },
  ]
})
</script>
