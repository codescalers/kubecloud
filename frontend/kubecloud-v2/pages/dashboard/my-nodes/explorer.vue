<template>
  <div>
    <v-container fluid class="my-16">
      <div class="d-flex justify-space-between align-center">
        <v-btn
          variant="plain"
          prepend-icon="mdi-keyboard-backspace"
          text="My Nodes"
          :to="ROUTES.Dashboard.MyNodes()"
        />
        <h1 class="text-h4 font-weight-bold">
          Reserve Your Node
        </h1>
        <div :style="{ width: '122px' }" />
      </div>

      <p class="text-body-1 mt-2 text-center opacity-70">
        Choose and reserve your dedicated Kubernetes node from our global network
      </p>
    </v-container>

    <StickySidebarLayout :sidebar-width="400" :page-offset="300" is-fluid :mobile="down950">
      <template #sidebar>
        <v-card
          class="h-100 elevation-0 bg-transparent"
          :style="{ padding: '0 !important', border: 'none !important' }"
          :class="{ 'mb-8': down950 }"
        >
          <div
            class="d-flex justify-space-between align-center px-10 py-8 border"
            :style="{ borderRadius: 'var(--v-rounded-1) var(--v-rounded-1) 0 0' }"
          >
            <div class="d-flex align-center ga-2">
              <v-icon icon="mdi-tune" size="small" color="primary" />
              <span class="text-h6 font-weight-bold">Filter Nodes</span>
            </div>

            <v-btn variant="plain" class="border" text="Reset" @click="filterTick++" />
          </div>

          <v-card-text
            class="px-0 py-8 border border-t-0"
            :class="{ 'overflow-auto': !down950 }"
            :style="{
              maxHeight: down950 ? undefined : 'calc(100% - 104px)',
              borderRadius: '0 0 var(--v-rounded-1) var(--v-rounded-1)',
            }"
          >
            <NodeFilterPanel :key="filterTick" v-model="filters" :nodes="nodes" />
          </v-card-text>
        </v-card>
      </template>

      <div class="border rounded-1">
        <div class="pa-8">
          <div class="d-flex justify-space-between">
            <h4 class="text-h5 font-weight-bold">
              Available Nodes
            </h4>
            <v-btn
              variant="text"
              class="border"
              prepend-icon="mdi-refresh"
              text="Refresh"
              :loading="isLoading"
              @click="reloadNodes()"
            />
          </div>
          <p class="text-body-1 opacity-70 mt-2">
            Browse through our available nodes and select the one that best fits your requirements.
          </p>
        </div>

        <v-divider />

        <div class="pa-8">
          <p v-if="isLoading" class="text-subtitle-2 opacity-70 text-center">
            Loading nodes...
          </p>
          <p v-else-if="filteredNodes.length === 0" class="text-subtitle-2 opacity-70 text-center">
            No nodes found
          </p>

          <v-expand-transition v-else>
            <v-row>
              <v-col
                v-for="node in filteredNodes"
                :key="node.id"
                :cols="down1400 ? 12 : down1840 ? 6 : down2280 ? 4 : 3"
              >
                <NodeCard :node="node" />
              </v-col>

              <v-col
                v-for="node in filteredNodes"
                :key="node.id"
                :cols="12"
              >
                <NodeListItem :node="node" />
              </v-col>
            </v-row>
          </v-expand-transition>
        </div>
      </div>
    </StickySidebarLayout>
  </div>
</template>

<script setup lang="ts">
import type { NodeFilters } from "~/components/NodeFilterPanel.vue"

const filters = ref<NodeFilters>()
const api = useApi()

const filterTick = ref(0)

const down2280 = useMediaQuery("(max-width: 2280px)")
const down1840 = useMediaQuery("(max-width: 1840px)")
const down1400 = useMediaQuery("(max-width: 1400px)")
const down950 = useMediaQuery("(max-width: 950px)")

const {
  state: nodes,
  isLoading,
  execute: reloadNodes,
} = useAsyncState(async () => {
  const { data } = await api.nodes.listRentableNodes()
  return data.data?.nodes ?? []
}, [], { immediate: $meta.client })

const filteredNodes = computed(() => {
  const f = filters.value
  if (!f) {
    return nodes.value
  }

  const { cpu, ram, ssd, gpu, price, location } = f

  return nodes.value.filter((node) => {
    if (cpu[0] > node.total_resources!.cru! || cpu[1] < node.total_resources!.cru!) {
      return false
    }

    if (ram[0] > node.total_resources!.mru! || ram[1] < node.total_resources!.mru!) {
      return false
    }

    if (ssd[0] > node.total_resources!.sru! || ssd[1] < node.total_resources!.sru!) {
      return false
    }

    if (gpu && node.num_gpu! === 0) {
      return false
    }

    if (price[0] > node.discount_price! || price[1] < node.discount_price!) {
      return false
    }

    if (location && node.location!.country !== location) {
      return false
    }

    return true
  })
})

const { drawer, container } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onBeforeUnmount(drawer.open)

onMounted(container.fluidize)
onBeforeUnmount(container.containerize)
</script>
