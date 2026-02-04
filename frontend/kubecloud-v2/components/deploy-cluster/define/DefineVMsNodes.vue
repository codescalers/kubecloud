<template>
  <v-card>
    <div class="d-flex align-center justify-space-between mb-8">
      <h3 class="d-flex align-center ga-2">
        <v-icon :icon="icon" size="x-small" color="primary" />
        {{ nodeType }} Nodes
      </h3>

      <v-btn
        variant="text"
        :text="`Add ${nodeType}`"
        color="primary"
        prepend-icon="mdi-plus"
        border
        @click="$emit('addNode')"
      />
    </div>

    <div v-if="nodes.length === 0" cols="12" class="border border-dashed rounded-lg d-flex justify-center align-center">
      <div class="text-center my-12" :style="{ color: 'rgba(var(--v-border-color), 0.5)' }">
        <v-icon icon="mdi-sitemap-outline" size="50" class="mb-4" />
        <p class="text-body-1">
          No {{ nodeType }} nodes configured
        </p>
      </div>
    </div>

    <div v-else>
      <v-tabs
        v-model="tab"
        align-tabs="center"
        color="primary"
        class="mb-8"
      >
        <v-tab v-for="node in nodes" :key="node.id" :value="node.id" :class="{ 'text-error': !isValidClusterNode(node) }">
          {{ node.name || '*Unnamed' }}

          <template v-if="node.type !== 'leader'" #append>
            <v-btn
              icon
              size="x-small"
              variant="plain"
              color="error"
              :style="{ borderRadius: '50% !important' }"
              @click.stop="onRemoveNode(node.id)"
            >
              <v-icon icon="mdi-close" size="small" />
            </v-btn>
          </template>
        </v-tab>
      </v-tabs>

      <DefineVMNodeForm v-if="activeNode" :ssh-keys="sshKeys" :node="activeNode" />
    </div>
  </v-card>
</template>

<script setup lang="ts">
import type { ModelsSSHKey } from "~/generated/api"

const props = defineProps<{
  icon: string
  nodeType: string
  sshKeys: ModelsSSHKey[]
  nodes: ClusterNode[]
}>()

const emit = defineEmits<{
  (e: "addNode"): void
  (e: "removeNode", id: string): void
}>()

const tab = ref<string>()
const activeNode = computed(() => props.nodes.find(node => node.id === tab.value))

function onRemoveNode(id: string) {
  if (tab.value === id) {
    const nodes = props.nodes.filter(node => node.id !== id)
    tab.value = nodes[0]?.id
  }

  emit("removeNode", id)
}
</script>
