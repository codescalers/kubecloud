<template>
  <div>
    <h1>Dashboard Overview {{ isLoading }}</h1>

    <pre>
      {{ balance }}
    </pre>

    <p>
      spent balance:
      {{ spent }}
    </p>

    <p>
      clusters Count:
      {{ deploymentsCount }}
    </p>

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

const { state: balance, isLoading } = useAsyncState(async () => {
  const { data } = await api.users.getUserBalance()
  return data
}, null)

const { state: spent } = useAsyncState(async () => {
  const { data } = await api.invoices.getInvoices()
  const { invoices } = data.data as unknown as { invoices: ModelsInvoice[] }
  const total = invoices.reduce((a, b) => a + (b.total ?? 0) * 1000, 0) ?? 0
  return toPrecision(total / 1000, 2)
}, null)

const { state: deploymentsCount } = useAsyncState(async () => {
  const { data } = await api.deployments.deploymentsGet()
  return data.data?.count ?? 0
}, 0)
</script>
