<template>
  <v-card class="position-relative">
    <div class="position-absolute top-0 right-0">
      <p
        class="px-4 py-2 bg-primary text-uppercase text-caption font-weight-bold"
        :style="{ borderRadius: '0 18px' }"
      >
        <span class="text-background">{{ Math.round((node.discount_price! / node.price_usd!) * 100 * 100) / 100 }}% OFF</span>
      </p>
    </div>

    <div>
      <h4 class="text-h6 font-weight-bold" v-text="`Node ${node.nodeId}`" />
      <div
        class="d-flex justify-space-between align-end"
        :style="{ marginTop: node.num_gpu ? '-6px' : '0' }"
      >
        <p class="d-flex align-end ga-1">
          <v-icon icon="mdi-map-marker-outline" size="small" color="primary" />
          <span
            class="opacity-70 text-uppercase text-subtitle-2"
            :style="{ lineHeight: 1.3 }"
            v-text="node.location?.country"
          />
        </p>

        <v-chip
          v-if="node.num_gpu"
          color="primary"
          size="small"
          class="rounded-lg border border-primary font-weight-bold"
          :style="{ '--v-border-opacity': 0.5 }"
        >
          {{ node.num_gpu! }} GPU
        </v-chip>
      </div>
    </div>

    <div
      class="bg-2 border rounded-1 my-8 d-flex justify-space-around align-center rounded-1 overflow-hidden flex-wrap"
    >
      <NodeCardResource
        icon="mdi-cpu-64-bit"
        color="primary"
        title="CPU"
        :value="node.total_resources!.cru!"
        unit="vCores"
      />
      <NodeCardResource
        icon="mdi-memory"
        color="success"
        title="RAM"
        :value="Math.round((node.total_resources!.mru! / 1024 ** 3) * 100) / 100"
        unit="GB"
      />
      <NodeCardResource
        icon="mdi-server"
        color="secondary"
        title="SSD"
        :value="Math.round((node.total_resources!.sru! / 1024 ** 4) * 100) / 100"
        unit="TB"
      />
    </div>

    <div class="d-flex justify-space-between align-end flex-wrap">
      <div>
        <h6 class="opacity-70 text-uppercase text-caption font-weight-bold">
          Monthly Price
        </h6>
        <p class="text-h5">
          <span class="font-weight-bold">${{ Math.round(node.discount_price! * 100) / 100 }}</span>&nbsp;
          <span class="text-caption opacity-50">/mo</span>
        </p>
      </div>

      <div class="d-flex flex-column align-end">
        <p
          class="text-caption text-decoration-line-through"
          :style="{
            textDecorationColor: 'rgb(var(--v-theme-error)) !important',
            color: 'rgb(255, 255, 255, 0.5)',
          }"
        >
          &nbsp;&nbsp;{{ node.price_usd }} USD&nbsp;&nbsp;
        </p>
        <v-chip
          color="primary"
          size="small"
          class="rounded-lg border border-primary font-weight-bold"
          :style="{ '--v-border-opacity': 0.5 }"
        >
          ${{ Math.round((node.discount_price! / 30) * 100) / 100 }} /hr
        </v-chip>
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
        :text="reserved ? 'Node Reserved' : 'Reserve Node'"
        :style="reserved ? { borderColor: 'rgb(var(--v-theme-success)) !important', boxShadow: 'none !important' } : {}"
        :append-icon="reserved ? undefined : 'mdi-arrow-right'"
        :prepend-icon="reserved ? 'mdi-check' : undefined"
        :color="reserved ? 'success' : undefined"
        variant="outlined"
        :readonly="reserved"
        :loading="isLoading"
        @click="reserveNode()"
      />
    </div>
  </v-card>
</template>

<script setup lang="ts">
import type { HandlersNodesWithDiscount } from "../generated/api"

const props = defineProps<{ node: HandlersNodesWithDiscount }>()

const monitoringUrl = useNodeMonitoringUrl(() => props.node)
const { isLoading, execute: reserveNode, state: reserved } = useNodeReserve(() => props.node)
</script>
