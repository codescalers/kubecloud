<template>
  <div>
    <div class="mb-16">
      <div class="d-flex justify-space-between align-center">
        <v-btn variant="plain" icon size="x-small" @click="$router.push('/dashboard/nodes')">
          <v-icon size="x-large">mdi-keyboard-backspace</v-icon>
        </v-btn>
        <h1 class="text-h4 font-weight-bold">Reserve Your Node</h1>
        <div :style="{ width: '32px' }" />
      </div>

      <p class="text-body-1 mt-2 text-center opacity-70">
        Choose and reserve your dedicated Kubernetes node from our global network.
      </p>
    </div>

    <StickySidebarLayout :sidebar-width="450" is-fluid>
      <template #sidebar>
        <v-card>
          {{ isLoading }}

          <v-btn color="primary" flat variant="tonal" :loading="isLoading" @click="reloadNodes()">
            Reload Nodes
          </v-btn>
        </v-card>
      </template>

      {{ nodes.length }}
      {{ nodes }}
    </StickySidebarLayout>
  </div>
</template>

<script setup lang="ts">
const api = useApi()

const {
  state: nodes,
  isLoading,
  execute: reloadNodes,
} = useAsyncState(
  async () => {
    const { data } = await api.nodes.listRentableNodes()
    return data.data?.nodes ?? []
  },
  [],
  { delay: 250, resetOnExecute: false }
)

const { drawer, container } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onUnmounted(drawer.open)

onMounted(container.fluidize)
onUnmounted(container.containerize)
</script>
