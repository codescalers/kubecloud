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
        <v-col v-for="action in actions" :key="action.to" cols="3">
          <v-btn block variant="outlined" v-bind="action" />
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

const actions = [
  {
    to: "/dashboard/clusters/deploy",
    text: "Deploy Cluster",
    prependIcon: "mdi-plus",
  },
  {
    to: "/dashboard/nodes/explorer",
    text: "Reserve Node",
    prependIcon: "mdi-server-plus",
  },
  {
    to: "/dashboard/ssh-keys",
    text: "Add SSH Key",
    prependIcon: "mdi-key-plus",
  },
  {
    to: "/dashboard/funds",
    text: "Add Funds",
    prependIcon: "mdi-cash-plus",
  },
]
</script>
