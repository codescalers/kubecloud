<template>
  <v-card class="position-relative">
    <div class="position-absolute top-0 right-0">
      <p
        class="px-4 py-2 bg-primary text-uppercase text-caption font-weight-bold"
        :style="{ borderRadius: '0 18px' }"
      >
        <span class="text-background">50% OFF</span>
      </p>
    </div>

    <div>
      <h4 class="text-h6 font-weight-bold" v-text="'Node ' + node.nodeId" />
      <p class="d-flex align-end ga-1">
        <v-icon icon="mdi-map-marker-outline" size="small" color="primary" />
        <span
          class="opacity-70 text-uppercase text-subtitle-2"
          :style="{ lineHeight: 1.3 }"
          v-text="node.location?.country"
        />
      </p>
    </div>

    <div class="my-8 d-flex align-center rounded-1 overflow-hidden border bg-2">
      <div class="flex-grow-1 pa-4 d-flex flex-column align-center">
        <v-icon icon="mdi-cpu-64-bit" color="primary" />
        <span class="text-caption font-weight-bold opacity-50 mt-2 mb-0">CPU</span>
        <span class="text-subtitle-2 font-weight-bold opacity-90">32 vCores</span>
      </div>
      <v-divider vertical />
      <div class="flex-grow-1 pa-4 d-flex flex-column align-center">
        <v-icon icon="mdi-memory" color="success" />
        <span class="text-caption font-weight-bold opacity-50 mt-2 mb-0">RAM</span>
        <span class="text-subtitle-2 font-weight-bold opacity-90">32 GB</span>
      </div>
      <v-divider vertical />
      <div class="flex-grow-1 pa-4 d-flex flex-column align-center">
        <v-icon icon="mdi-server" color="secondary" />
        <span class="text-caption font-weight-bold opacity-50 mt-2 mb-0">SSD</span>
        <span class="text-subtitle-2 font-weight-bold opacity-90">1.2 TB</span>
      </div>
    </div>

    <div class="d-flex justify-space-between align-end">
      <div>
        <h6 class="opacity-70 text-uppercase text-caption font-weight-bold">Monthly Price</h6>
        <p class="text-h5">
          <span class="font-weight-bold">$45.45</span>&nbsp;
          <span class="text-caption opacity-50">/mo</span>
        </p>
      </div>

      <div class="d-flex flex-column align-end">
        <p
          class="text-caption text-decoration-line-through opacity-50"
          :style="{ textDecorationColor: 'rgb(var(--v-theme-error)) !important' }"
        >
          &nbsp;{{ "$90.90" }}&nbsp;
        </p>
        <v-chip
          color="primary"
          size="small"
          class="rounded-lg border border-primary font-weight-bold"
          :style="{ '--v-border-opacity': 0.5 }"
          >$0.21/hr</v-chip
        >
      </div>
    </div>

    <div class="mt-8">
      <v-btn
        prepend-icon="mdi-resistor"
        block
        text="Check Node Health"
        variant="plain"
        class="border mb-4"
        size="x-large"
        target="_blank"
        :href="monitoringUrl"
      />
      <v-btn
        block
        size="x-large"
        class="btn-form"
        text="Reserve Node"
        append-icon="mdi-arrow-right"
        variant="outlined"
      />
    </div>
  </v-card>
</template>

<script setup lang="ts">
import type { HandlersNodesWithDiscount } from "../generated/api"

const props = defineProps<{ node: HandlersNodesWithDiscount }>()
const api = useApi()

const monitoringUrl = computedAsync(async () => {
  console.log("async computing")

  const accountId = await api.helpers.getAccountId(props.node.twinId!)
  const params = new URLSearchParams({
    orgId: "2",
    refresh: "30s",
    "var-network": "dev",
    "var-farm": props.node.farmId!.toString(),
    "var-node": accountId,
    "var-diskdevices": "[a-z]+|nvme[0-9]+n[0-9]+|mmcblk[0-9]+",
  })

  return `https://metrics.grid.tf/d/rYdddlPWkfqwf/zos-host-metrics?${params.toString()}`
})
</script>
