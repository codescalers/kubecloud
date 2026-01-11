<template>
  <v-expand-transition>
    <div v-if="nodes.length === 0">
      <p class="text-body-2 opacity-70 text-center">
        Please load some nodes
      </p>
    </div>

    <div v-else>
      <RangeFilter
        v-model="cpu"
        class="px-10"
        title="CPU"
        :range="filterRanges.cpu"
        unit="vCores"
      />

      <v-divider class="my-6" />

      <RangeFilter
        v-model="ram"
        class="px-10"
        title="RAM"
        :range="filterRanges.ram"
        unit="GB"
        :normalizer="(v: number) => Math.round((v / 1024 ** 3) * 100) / 100"
      />

      <v-divider class="my-6" />

      <RangeFilter
        v-model="ssd"
        class="px-10"
        title="SSD"
        :range="filterRanges.ssd"
        unit="TB"
        :normalizer="(v: number) => Math.round((v / 1024 ** 4) * 100) / 100"
      />

      <v-divider class="my-6" />

      <div class="d-flex justify-space-between align-center px-10">
        <div>
          <p class="text-subtitle-1 font-weight-bold">
            GPU Required
          </p>
          <p class="text-caption opacity-70">
            Dedicated graphics card
          </p>
        </div>

        <v-switch v-model="gpu" color="primary" inset hide-details />
      </div>

      <v-divider class="my-6" />

      <RangeFilter
        v-model="price"
        class="px-10"
        title="Price / mo"
        :range="filterRanges.price"
        color="secondary"
        unit="USD"
        step="0.1"
        :normalizer="(v: number) => Math.round(v * 100) / 100"
      />

      <v-divider class="my-6" />

      <div class="px-10">
        <p class="text-subtitle-1 font-weight-bold mb-3">
          Location
        </p>
        <v-select
          v-model="location"
          clearable
          placeholder="Pick a location"
          :items="filterRanges.location"
          variant="outlined"
          density="compact"
          hide-details
        />
      </div>
    </div>
  </v-expand-transition>
</template>

<script setup lang="ts">
import type { HandlersNodesWithDiscount } from "../generated/api"

export interface NodeFilters {
  cpu: [number, number]
  ram: [number, number]
  ssd: [number, number]
  gpu: boolean
  price: [number, number]
  location: string
}

const props = defineProps<{ modelValue?: NodeFilters, nodes: HandlersNodesWithDiscount[] }>()
const emit = defineEmits<{ (e: "update:model-value", value: NodeFilters): void }>()

const filterRanges = computed(() => {
  const cpu: number[] = []
  const ram: number[] = []
  const ssd: number[] = []
  const price: number[] = []
  const location: string[] = []

  for (const node of props.nodes) {
    cpu.push(node.total_resources!.cru!)
    ram.push(node.total_resources!.mru!)
    ssd.push(node.total_resources!.sru!)
    price.push(node.price_usd!)
    price.push(node.discount_price!)
    location.push(node.location!.country!)
  }

  cpu.sort((a, b) => a - b)
  ram.sort((a, b) => a - b)
  ssd.sort((a, b) => a - b)
  price.sort((a, b) => a - b)
  location.sort((a, b) => a.localeCompare(b))

  return {
    cpu: [cpu[0], cpu[cpu.length - 1]] as [number, number],
    ram: [ram[0], ram[ram.length - 1]] as [number, number],
    ssd: [ssd[0], ssd[ssd.length - 1]] as [number, number],
    price: [price[0], price[price.length - 1]] as [number, number],
    location: Array.from(new Set(location)) as string[],
  }
})

const cpu = ref<[number, number]>()
const ram = ref<[number, number]>()
const ssd = ref<[number, number]>()
const gpu = ref<boolean>(false)
const price = ref<[number, number]>()
const location = ref<string>()

const filters = computed(() => {
  return {
    cpu: cpu.value ?? [0, 0],
    ram: ram.value ?? [0, 0],
    ssd: ssd.value ?? [0, 0],
    gpu: gpu.value ?? false,
    price: price.value ?? [0, 0],
    location: location.value ?? "",
  } as NodeFilters
})

watchDebounced(filters, f => emit("update:model-value", f), { debounce: 100 })
</script>
