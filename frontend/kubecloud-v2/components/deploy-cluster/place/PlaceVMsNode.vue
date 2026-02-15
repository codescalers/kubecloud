<template>
  <!-- class="border rounded-xl border-dashed" -->
  <div>
    <div class="pa-4">
      <div class="d-flex align-center ga-2">
        <h5 class="text-h6 font-weight-bold" v-text="modelValue.name" />
        <v-chip color="primary" size="x-small" :text="modelValue.type" class="rounded font-weight-bold text-caption text-capitalize" />
      </div>

      <v-chip
        v-if="modelValue.useFullNodeCapabilities"
        color="success"
        size="small"
        prepend-icon="mdi-check-decagram"
        class="font-weight-bold mt-2"
      >
        Use Full Node Capabilities
      </v-chip>

      <div v-else class="d-flex align-center flex-wrap ga-2 mt-2">
        <v-chip size="small" prepend-icon="mdi-cpu-64-bit" color="primary" class="font-weight-bold">
          CPU: {{ modelValue.cpu }} vCores
        </v-chip>
        <v-chip size="small" prepend-icon="mdi-memory" color="success" class="font-weight-bold">
          RAM: {{ modelValue.memory }} GiB
        </v-chip>
        <v-chip size="small" prepend-icon="mdi-server" color="secondary" class="font-weight-bold">
          Disk Size: {{ modelValue.disk }} GiB
        </v-chip>
      </div>
    </div>

    <v-expand-transition>
      <div v-if="activeNode">
        <NodeListItem class="mb-4" :active="modelValue.node?.valid" :deactive="!modelValue.node?.valid" :node="activeNode!" />
      </div>
    </v-expand-transition>

    <v-table class="border-t border-0 border-dashed" :class="{ 'table-hidden-overflow': !!loadingNode }" :style="{ maxHeight: '400px' }">
      <tbody>
        <tr v-if="filteredNodes.length === 0" class="text-caption text-accent text-center">
          <VListItem title="No available nodes for the requested configuration" />
        </tr>

        <tr
          v-for="(node, index) in filteredNodes"
          :key="node.nodeId"
          :class="{ 'border-t border-0 border-dashed': index !== 0 }"
        >
          <VDivider v-if="index !== 0" />

          <NodeListItem
            :disabled="pickedNodes.includes(node.nodeId!)"
            :node="node"
            @reserve="$emit('reserve', node.nodeId!)"
            @pick="getNodeStoragePool(undefined, $event)"
          />
        </tr>
      </tbody>
    </v-table>
  </div>
</template>

<script setup lang="ts">
import type { HandlersNodesWithDiscount } from "~/generated/api"

const props = defineProps<{ modelValue: ClusterNode, pickedNodes: number[], nodes: HandlersNodesWithDiscount[] }>()
const emit = defineEmits<{
  (e: "reserve", nodeId: number): void
  (e: "pick", node: null | { id: number, raw: HandlersNodesWithDiscount, valid: boolean }): void
}>()

const { loadingNode, setLoadingNode } = inject(NodePickCtxKey)!

const activeNode = computed(() => props.nodes.filter(n => n.nodeId === props.modelValue.node?.id)[0])
const filteredNodes = computed(() => {
  return props.nodes.filter((n) => {
    const { useFullNodeCapabilities, cpu, memory, disk } = props.modelValue

    if (props.pickedNodes.includes(n.nodeId!)) {
      return true
    }

    const cru = n.total_resources!.cru!
    const ram = (n.total_resources!.mru! * 0.98 - n.used_resources!.mru!) / 1024 ** 3
    const ssd = (n.total_resources!.sru! * 0.97 - n.used_resources!.sru!) / 1024 ** 3

    if (useFullNodeCapabilities) {
      return cru >= 0 && ram >= 0 && ssd >= 0
    }

    return cru >= cpu && ram >= memory && ssd >= disk
  })
})

const api = useApi()
const { execute: getNodeStoragePool } = useAsyncState(async (node: HandlersNodesWithDiscount) => {
  setLoadingNode(node.nodeId!)
  try {
    /* const { data } =  */ await api.nodes.getNodeStoragePool(node.nodeId!.toString())
    emit("pick", { id: node.nodeId!, raw: node, valid: true })
  }
  catch {
    emit("pick", { id: node.nodeId!, raw: node, valid: false })
  }
  finally {
    setLoadingNode(undefined)
  }
}, null, { immediate: false })
</script>
