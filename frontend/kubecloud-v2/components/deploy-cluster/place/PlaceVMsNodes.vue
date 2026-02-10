<template>
  <v-card>
    <div class="d-flex align-center justify-space-between mb-8">
      <h3 class="d-flex align-center ga-2">
        <v-icon icon="mdi-sitemap-outline" size="x-small" color="primary" />
        Cluster Nodes
      </h3>
    </div>

    <div class="d-flex border border-dashed rounded-lg">
      <v-tabs
        :model-value="nodeTab.join('|')"
        color="primary"
        direction="vertical"
        class="border-e border-0 border-dashed"
        mandatory
        @update:model-value="nodeTab = ($event.split('|') as ['masters' | 'workers', number])"
      >
        <DefineNodeTab #="{ text, value, active }">
          <v-tab
            :value="value"
            :text="text"
            :style="{ borderRadius: '0 4px 4px 0 !important', backgroundColor: active ? 'rgba(var(--v-theme-success), 0.12)' : nodeTab.join('|') === value ? 'rgba(var(--v-theme-primary), 0.12)' : undefined }"
            :class="{ 'text-success': active }"
          />
        </DefineNodeTab>

        <ReuseNodeTab
          v-for="(node, index) in cluster.masters"
          :key="node.id"
          :text="`${node.name} (${node.type})`"
          :value="`masters|${index}`"
          :loading="loadingNode && nodeTab.join('|') === `masters|${index}`"
          :disabled="loadingNode && nodeTab.join('|') !== `masters|${index}`"
          :active="!!node.nodeId"
        />

        <ReuseNodeTab
          v-for="(node, index) in cluster.workers"
          :key="node.id"
          :text="`${node.name} (${node.type})`"
          :value="`workers|${index}`"
          :loading="loadingNode && nodeTab.join('|') === `workers|${index}`"
          :disabled="loadingNode && nodeTab.join('|') !== `workers|${index}`"
          :active="!!node.nodeId"
        />
      </v-tabs>

      <div v-if="cluster[nodeTab[0]][nodeTab[1]]" class="flex-grow-1">
        <PlaceVMsNode
          v-model="cluster[nodeTab[0]][nodeTab[1]]!"
          :all-nodes="_nodes"
          :nodes="nodes"
          @reserve="onReserveNode"
          @pick="$props.cluster[nodeTab[0]][nodeTab[1]]!.nodeId = $event"
        />
      </div>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import type { HandlersNodesWithDiscount, ListRentableNodes200Response } from "~/generated/api"

const props = defineProps<{ cluster: ClusterForm }>()

const [DefineNodeTab, ReuseNodeTab] = createReusableTemplate({
  props: { text: String, value: String, active: Boolean },
})

const { loadingNode } = inject(NodePickCtxKey)!
const nodeTab = ref<["masters" | "workers", number]>(["masters", 0])

watchImmediate(() => {
  const [t, i] = nodeTab.value
  return props.cluster[t][i]
}, (v) => {
  if (!v) {
    nodeTab.value = ["masters", 0]
  }
})

const api = useApi()
function sortNodes(nodes: HandlersNodesWithDiscount[]): HandlersNodesWithDiscount[] {
  return nodes.toSorted((a, b) => {
    const _a = (a.rented ? 1 : 0) + (a.rentable ? 1 : 0)
    const _b = (b.rented ? 1 : 0) + (b.rentable ? 1 : 0)
    return _b - _a
  })
}

const { state: _nodes } = useAsyncState(async () => {
  const { data } = await api.nodes.listNodes()
  const nodes = (data as ListRentableNodes200Response).data?.nodes ?? []
  return sortNodes(nodes)
}, [], { resetOnExecute: false })

function onReserveNode(nodeId: number): void {
  // const __nodes = _nodes.value.map(n => n.nodeId === nodeId ? { ...n, rented: true } : n)
  // _nodes.value = sortNodes(__nodes)
  _nodes.value = _nodes.value.map(n => n.nodeId === nodeId ? { ...n, rented: true } : n)
}

const nodes = computed(() => {
  const pickedNodes = props.cluster.masters.concat(props.cluster.workers).map(n => n.nodeId)
  return _nodes.value.filter((n) => {
    const regionFilter = !props.cluster.region || n.location?.region === props.cluster.region
    return regionFilter && !pickedNodes.includes(n.nodeId!)
  })
})
</script>
