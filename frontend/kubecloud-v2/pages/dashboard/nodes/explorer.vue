<template>
  <div>
    <v-container fluid class="my-16">
      <div class="d-flex justify-space-between align-center">
        <v-btn
          variant="plain"
          prepend-icon="mdi-keyboard-backspace"
          text="My Nodes"
          to="/dashboard/nodes"
        />
        <h1 class="text-h4 font-weight-bold">Reserve Your Node</h1>
        <div :style="{ width: '32px' }" />
      </div>

      <p class="text-body-1 mt-2 text-center opacity-70">
        Choose and reserve your dedicated Kubernetes node from our global network
      </p>
    </v-container>

    <StickySidebarLayout :sidebar-width="450" :page-offset="300" is-fluid>
      <template #sidebar>
        <v-card
          class="h-100 elevation-0 bg-transparent"
          :style="{ padding: '0 !important', border: 'none !important' }"
        >
          <div
            class="d-flex justify-space-between align-center px-10 py-8 border"
            :style="{ borderRadius: 'var(--v-rounded-1) var(--v-rounded-1) 0 0' }"
          >
            <div class="d-flex align-center ga-2">
              <v-icon icon="mdi-tune" size="small" color="primary" />
              <span class="text-h6 font-weight-bold">Filter Nodes</span>
            </div>

            <v-btn variant="plain" class="border" text="Reset" @click="reloadNodes()" />
          </div>

          <v-card-text
            class="overflow-auto px-0 py-8 border border-t-0"
            :style="{
              maxHeight: 'calc(100% - 104px)',
              borderRadius: '0 0 var(--v-rounded-1) var(--v-rounded-1)',
            }"
          >
            <NodeFilterPanel v-model="filters" :nodes="nodes" />
          </v-card-text>
        </v-card>
      </template>

      <v-progress-circular v-if="isLoading" indeterminate />

      <v-row v-else>
        <v-col v-for="node in nodes" :key="node.id" cols="12" md="6" lg="4">
          <NodeCard :node="node" />
        </v-col>
      </v-row>
    </StickySidebarLayout>
  </div>
</template>

<script setup lang="ts">
import type { NodeFilters } from "~/components/NodeFilterPanel.vue"

const filters = ref<NodeFilters>()
const api = useApi()

const {
  state: nodes,
  isLoading,
  execute: reloadNodes,
} = useAsyncState(async () => {
  const { data } = await api.nodes.listRentableNodes()
  return data.data?.nodes ?? []
}, [])

const { drawer, container } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onUnmounted(drawer.open)

onMounted(container.fluidize)
onUnmounted(container.containerize)
</script>
