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
        Choose and reserve your dedicated Kubernetes node from our global network.
      </p>
    </v-container>

    <StickySidebarLayout :sidebar-width="450" :page-offset="300" is-fluid>
      <template #sidebar>
        <v-card class="h-100" :style="{ padding: '0 !important' }">
          <div class="d-flex justify-space-between align-center px-10 py-8 border-b">
            <div class="d-flex align-center ga-2">
              <v-icon icon="mdi-tune" size="small" color="primary" />
              <span class="text-h6 font-weight-bold">Filter Nodes</span>
            </div>

            <v-btn variant="plain" class="border" text="Reset" />
          </div>

          <v-card-text class="overflow-auto px-0 py-8" :style="{ maxHeight: 'calc(100% - 103px)' }">
            <div class="px-10">
              <div class="d-flex justify-space-between align-center mb-3">
                <p class="text-subtitle-1 font-weight-bold">CPU</p>
                <v-chip
                  color="primary"
                  size="small"
                  class="rounded-lg border border-primary font-weight-bold"
                  :style="{ '--v-border-opacity': 0.5 }"
                  :text="`${cpu.join(' - ')} vCores`"
                />
              </div>
              <v-range-slider
                v-model="cpu"
                thumb-size="13"
                thumb-color="white"
                track-size="1"
                hide-details
                color="primary"
                min="0"
                max="120"
                step="1"
              />
              <div class="d-flex justify-space-between align-center text-body-2 opacity-70 px-2">
                <p>0</p>
                <p>120</p>
              </div>
            </div>

            <v-divider class="my-6" />

            <div class="px-10">
              <div class="d-flex justify-space-between align-center mb-3">
                <p class="text-subtitle-1 font-weight-bold">RAM</p>
                <v-chip
                  color="primary"
                  size="small"
                  class="rounded-lg border border-primary font-weight-bold"
                  :style="{ '--v-border-opacity': 0.5 }"
                  :text="`${ram.join(' - ')} MB`"
                />
              </div>
              <v-range-slider
                v-model="ram"
                thumb-size="13"
                thumb-color="white"
                track-size="1"
                hide-details
                color="primary"
                min="200"
                max="4800"
                step="1"
              />
              <div class="d-flex justify-space-between align-center text-body-2 opacity-70 px-2">
                <p>0</p>
                <p>120</p>
              </div>
            </div>

            <v-divider class="my-6" />

            <div class="d-flex justify-space-between align-center px-10">
              <div>
                <p class="text-subtitle-1 font-weight-bold">GPU Required</p>
                <p class="text-caption opacity-70">Dedicated graphics card</p>
              </div>

              <div>
                <v-switch color="primary" inset hide-details />
              </div>
            </div>

            <v-divider class="my-6" />

            <div class="px-10">
              <div class="d-flex justify-space-between align-center mb-3">
                <p class="text-subtitle-1 font-weight-bold">Price / mo</p>
                <v-chip
                  color="secondary"
                  size="small"
                  class="rounded-lg border border-secondary font-weight-bold"
                  :style="{ '--v-border-opacity': 0.5 }"
                  :text="`${price.join(' - ')} USD`"
                />
              </div>
              <v-range-slider
                v-model="price"
                thumb-size="13"
                thumb-color="white"
                track-size="1"
                hide-details
                color="secondary"
                min="15.1"
                max="799.9"
                step="0.1"
              />
              <div class="d-flex justify-space-between align-center text-body-2 opacity-70 px-2">
                <p>15.1</p>
                <p>799.9</p>
              </div>
            </div>

            <v-divider class="my-6" />

            <div class="px-10">
              <p class="text-subtitle-1 font-weight-bold mb-3">Location</p>
              <!-- <v-select
                :items="[]"
                placeholder="Select Location"
              /> -->
              <v-select
                clearable
                placeholder="Pick a location"
                :items="['Egypt', 'Kenya', 'Nigeria']"
                variant="outlined"
                density="compact"
                hide-details
              />
            </div>
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
const api = useApi()

const cpu = ref<[number, number]>([0, 120])
const ram = ref<[number, number]>([200, 4800])
const price = ref<[number, number]>([15.1, 799.9])

const {
  state: nodes,
  isLoading,
  // execute: reloadNodes,
} = useAsyncState(
  async () => {
    const { data } = await api.nodes.listRentableNodes()
    return data.data?.nodes ?? []
  },
  [],
  { resetOnExecute: false }
)

const { drawer, container } = inject(DashboardLayoutCtxKey)!

onMounted(drawer.close)
onUnmounted(drawer.open)

onMounted(container.fluidize)
onUnmounted(container.containerize)
</script>
