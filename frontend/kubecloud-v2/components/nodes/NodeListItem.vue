<template>
  <v-card
    :loading="loadingNode === node.nodeId"
    :disabled="disabled || (!!loadingNode && loadingNode !== node.nodeId)"
    :class="{ 'opacity-50': disabled || (!!loadingNode && loadingNode !== node.nodeId) }"
    flat
    rounded="0"
    :style="{ padding: '16px !important', border: 'none !important' }"
    :color="active ? `rgba(var(--v-theme-success), 0.05)` : deactive ? `rgba(var(--v-theme-error), 0.05)` : undefined"
    v-bind="{
      onClick: active || deactive || loadingNode === node.nodeId ? undefined : () => $emit('pick', node),
    }"
  >
    <section class="d-flex justify-space-between align-center">
      <div class="d-flex align-center ga-2">
        <h4 class="text-h6 font-weight-bold" v-text="`Node ${node.nodeId}`" />
        <div class="d-flex align-center">
          <v-icon icon="mdi-map-marker-outline" size="x-small" color="primary" />
          <span class="text-caption text-accent text-capitalize" v-text="node.location?.country" />
        </div>
      </div>

      <div class="d-flex align-center ga-2">
        <DefineHeadChip #="{ color, $slots }">
          <p
            class="border rounded-lg px-2 py-1 text-caption text-accent font-weight-medium"
            :style="{
              color: `rgba(${color}, 1) !important`,
              borderColor: `rgba(${color}, 0.3) !important`,
              backgroundColor: `rgba(${color}, 0.05) !important`,
            }"
          >
            <component :is="$slots.default" />
          </p>
        </DefineHeadChip>

        <UseHeadChip v-if="node.rented" :color="getColor(0)" class="d-flex align-center ga-1">
          <v-icon icon="mdi-lock" size="small" />
          <span>Reserved (50% off)</span>
        </UseHeadChip>

        <UseHeadChip v-else :color="getColor(0)">
          {{ node.dedicated ? 'Dedicated' : 'Shared' }}
        </UseHeadChip>

        <UseHeadChip :color="getColor(2)" class="text-uppercase">
          {{ node.status }}
        </UseHeadChip>

        <UseHeadChip v-if="node.num_gpu" :color="getColor(3)">
          {{ node.num_gpu }} GPU
        </UseHeadChip>

        <UseHeadChip :color="getColor(5)" class="text-capitalize">
          {{ node.certificationType }}
        </UseHeadChip>

        <UseHeadChip v-if="node.speed?.upload && node.speed?.download" :color="getColor(6)">
          <span class="d-flex align-center ga-1">
            <VIcon icon="mdi-arrow-up" size="small" />{{ Math.round((node.speed?.upload ?? 0) / 1024 ** 2) }} mbps
          </span>

          <span class="d-flex align-center ga-1">
            <VIcon icon="mdi-arrow-down" size="small" />{{ Math.round((node.speed?.download ?? 0) / 1024 ** 2) }} mbps
          </span>
        </UseHeadChip>
      </div>
    </section>
    <section>
      <div class="d-flex align-center ga-2">
        <p v-if="node.farmName" class="text-caption text-accent">
          Farm: <strong>{{ node.farmName }}</strong>
        </p>

        <span v-if="node.farmName && node.uptime" class="text-caption text-accent">|</span>

        <p v-if="node.uptime" class="text-caption text-accent">
          Uptime: <strong>{{ uptime }}</strong>
        </p>
      </div>
    </section>

    <div class="d-flex align-center justify-space-between">
      <div class="d-flex align-center flex-wrap ga-2 mt-2">
        <v-chip size="small" prepend-icon="mdi-cpu-64-bit" color="primary" class="font-weight-bold">
          CPU: {{ node.total_resources!.cru! }} vCores
        </v-chip>
        <v-chip size="small" prepend-icon="mdi-memory" color="success" class="font-weight-bold">
          RAM: {{ ram }}
        </v-chip>
        <v-chip size="small" prepend-icon="mdi-server" color="secondary" class="font-weight-bold">
          Disk Size: {{ ssd }}
        </v-chip>
      </div>

      <div class="d-flex align-center ga-2">
        <v-btn
          text="Check Health"
          variant="plain"
          class="border"
          target="_blank"
          :href="monitoringUrl"
          @click.stop
        />
        <v-btn
          v-if="node.rentable || node.rented"
          class="btn-form"
          :text="reserved || node.rented ? 'Reserved' : 'Reserve'"
          :style="reserved || node.rented ? { borderColor: 'rgb(var(--v-theme-success)) !important', boxShadow: 'none !important' } : {}"
          :color="reserved || node.rented ? 'success' : undefined"
          variant="outlined"
          :readonly="reserved || node.rented"
          :loading="isLoading"
          @click.stop="reserveNode()"
        />
      </div>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import type { HandlersNodesWithDiscount } from "~/generated/api"
import humanizeDuration from "humanize-duration"
import prettyBytes from "pretty-bytes"

const props = defineProps<{ node: HandlersNodesWithDiscount, disabled?: boolean, active?: boolean, deactive?: boolean }>()
const emit = defineEmits<{ (e: "reserve", nodeId: number): void, (e: "pick", node: HandlersNodesWithDiscount): void }>()

const { loadingNode } = inject(NodePickCtxKey)!

const [DefineHeadChip, UseHeadChip] = createReusableTemplate({
  props: {
    color: String,
  },
})

const uptime = computed(() => humanizeDuration(
  (props.node.uptime ?? 0) * 1_000,
  { round: true, language: "en", units: ["y", "mo", "d"] },
))

const ram = computed(() => prettyBytes(
  props.node.total_resources!.mru! - props.node.used_resources!.mru!,
  { binary: true, maximumFractionDigits: 2 },
))

const ssd = computed(() => prettyBytes(
  props.node.total_resources!.sru! - props.node.used_resources!.sru!,
  { binary: true, maximumFractionDigits: 2 },
))

const monitoringUrl = useNodeMonitoringUrl(() => props.node)
const { isLoading, execute: reserveNode, state: reserved } = useNodeReserve(() => props.node, {
  onSuccess: () => emit("reserve", props.node.nodeId!),
})
</script>
