<template>
  <div>
    <div class="section-header">
      <h3 class="section-title">
        <v-icon icon="mdi-server-network" class="mr-2"></v-icon>
        Assign VMs to Reserved Nodes
      </h3>
      <p class="section-subtitle">Select which reserved nodes will host your cluster VMs</p>
    </div>
    <div v-if="availableNodes.length === 0" class="nodes-empty">
      <div class="empty-card">
        <v-icon icon="mdi-server-network-off" size="38" color="primary" class="empty-icon mb-3"></v-icon>
        <div class="empty-title">No Reserved Nodes</div>
        <div class="empty-desc">You have not reserved any nodes yet.<br>To deploy a cluster, you must first reserve at least one node.</div>
        <v-btn color="primary" class="mt-4" variant="outlined" to="/nodes">
          <v-icon icon="mdi-plus" size="20" class="mr-2"></v-icon>
          Reserve Node
        </v-btn>
      </div>
    </div>
    <v-row v-else>
      <v-col cols="12" v-for="(vm, index) in allVMs" :key="index">
        <div class="vm-assignment-card">
          <div class="vm-assignment-header">
            <div class="vm-avatar" :class="vm.name.includes('Master') ? 'master' : 'worker'">
              <v-icon :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'" color="white"></v-icon>
            </div>
            <div class="vm-info">
              <h4 class="vm-title">{{ vm.name }}</h4>
              <div class="vm-specs">
                <span class="spec-chip">{{ vm.vcpu }} vCPU</span>
                <span class="spec-chip">{{ vm.ram }}GB RAM</span>
                <span class="spec-chip">{{ vm.disk }}GB Disk</span>
                <span v-if="vm.gpu" class="spec-chip">GPU</span>
              </div>
            </div>
          </div>
          <v-select
            v-model="vm.node"
            :items="getAvailableNodesForVM(index)"
            :item-title="nodeLabel"
            item-value="nodeId"
            label="Select Reserved Node"
            clearable
            class="node-select"
            @update:modelValue="val => props.onAssignNode(index, val)"
          >
            <template #item="{ item, index, props: itemProps }">
              <div>
                <div v-bind="itemProps" class="node-option-row">
                  <div class="node-id">Node {{ item.raw.nodeId }}</div>
                  <div class="chip-row">
                    <v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
                      {{ getNodeAvailableResources(item.raw).cpu }} vCPU
                    </v-chip>
                    <v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-memory</v-icon>
                      {{ getNodeAvailableResources(item.raw).ram }} GB RAM
                    </v-chip>
                    <v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
                      {{ getNodeAvailableResources(item.raw).storage }} GB Disk
                    </v-chip>
                    <v-chip v-if="item.raw.gpu" color="deep-purple-accent-2" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      <v-icon size="14" class="mr-1">mdi-expansion-card</v-icon>
                      GPU
                    </v-chip>
                    <v-chip color="secondary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                      {{ item.raw.country }}
                    </v-chip>
                  </div>
                </div>
                <v-divider v-if="index < availableNodes.length - 1" />
              </div>
            </template>
            <template #selection="{ item }">
              <div class="node-id">Node {{ item.raw.nodeId }}</div>
              <div class="chip-row">
                <v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
                  {{ getNodeAvailableResources(item.raw).cpu }} vCPU
                </v-chip>
                <v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-memory</v-icon>
                  {{ getNodeAvailableResources(item.raw).ram }} GB RAM
                </v-chip>
                <v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
                  {{ getNodeAvailableResources(item.raw).storage }} GB Disk
                </v-chip>
                <v-chip v-if="item.raw.gpu" color="deep-purple-accent-2" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-expansion-card</v-icon>
                  GPU
                </v-chip>
                <v-chip color="secondary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  {{ item.raw.country }}
                </v-chip>
              </div>
            </template>
          </v-select>
        </div>
      </v-col>
    </v-row>
    <div class="step-actions">
      <v-btn variant="outlined" @click="$emit('prevStep')">
        <v-icon start icon="mdi-arrow-left"></v-icon>
        Back
      </v-btn>
      <v-btn variant="outlined" color="primary" :disabled="!isStep2Valid" @click="$emit('nextStep')">
        Continue
        <v-icon end icon="mdi-arrow-right"></v-icon>
      </v-btn>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { NormalizedNode } from '@/types/normalizedNode';
import { type VM } from '../../composables/useDeployCluster';
import { defineProps, withDefaults, defineEmits, onMounted, computed } from 'vue';
const props = withDefaults(defineProps<{
  allVMs: VM[];
  availableNodes: NormalizedNode[];
  getNodeInfo: (id: number) => string;
  onAssignNode: (vmIdx: number, nodeId: number | null) => void;
  isStep2Valid?: boolean;
}>(), {
  isStep2Valid: false
});
const emit = defineEmits(['nextStep', 'prevStep']);

function nodeLabel(node: any) {
  if (!node) return '';
  return `Node ${node.nodeId}`;
}

const currentAllocations = computed(() =>
  props.allVMs.reduce((acc, vm) => {
    if (vm.node != null) {
      acc[vm.node] = {
        ram: (acc[vm.node]?.ram || 0) + vm.ram,
        storage: (acc[vm.node]?.storage || 0) + (vm.disk || 0) + vm.rootfs
      };
    }
    return acc;
  }, {} as Record<number, { ram: number; storage: number }>)
);

const getAvailableNodesForVM = (vmIndex: number) => {
  const vm = props.allVMs[vmIndex];
  if (!vm) return [];
  return props.availableNodes.filter(node => {
    const used = currentAllocations.value[node.nodeId] || { ram: 0, storage: 0 };
    const availableRam = (node.available_ram || 0) - used.ram + (vm.node === node.nodeId ? vm.ram : 0);
    const availableStorage = (node.available_storage || 0) - used.storage + (vm.node === node.nodeId ? (vm.disk || 0) + vm.rootfs : 0);
    return (node.cpu || 0) >= vm.vcpu && availableRam >= vm.ram && availableStorage >= (vm.disk || 0) + vm.rootfs;
  });
};


const getNodeAvailableResources = (node: NormalizedNode) => ({
  cpu: node.cpu || 0,
  ram: Math.max(0, (node.available_ram || 0) - (currentAllocations.value[node.nodeId]?.ram || 0)),
  storage: Math.max(0, (node.available_storage || 0) - (currentAllocations.value[node.nodeId]?.storage || 0))
});
// Validate existing VM assignments on mount and clear invalid ones
onMounted(() => {
  props.allVMs.forEach((vm, vmIndex) => {
    if (vm.node != null) {
      const availableNodes = getAvailableNodesForVM(vmIndex);
      const assignedNode = availableNodes.find(node => node.nodeId === vm.node);
      if (!assignedNode) {
        props.onAssignNode(vmIndex, null);
      }
    }
  });
});
</script>
<style scoped>
.section-header {
  margin-bottom: 4rem;
}
.section-title {
  font-size: 1.2rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.section-subtitle {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1rem;
  margin-top: 0.25rem;
}
.vm-assignment-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 12px;
  margin-bottom: 2rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  flex: 1 1 320px;
  min-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.vm-assignment-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}
.vm-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary, #6366f1);
}
.vm-avatar.worker {
  background: var(--color-success, #22d3ee);
}
.vm-title {
  font-size: 1.1rem;
  font-weight: 500;
}
.vm-specs {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.25rem;
}
.spec-chip {
  background: var(--color-surface-2, #23243a);
  color: var(--color-text, #cfd2fa);
  border-radius: 8px;
  padding: 0.2rem 0.7rem;
  font-size: 0.95rem;
}
.node-select {
  margin-top: 0.5rem;
}
.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 2rem;
}
.node-option-row {
  margin: .5rem;
  cursor: pointer;
}
.node-id {
  font-weight: 600;
  margin-bottom: 2px;
  margin-right: 1rem;
}
.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.nodes-empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 260px;
  width: 100%;
}
.empty-card {
  background: var(--color-surface-1, #18192b);
  border: 1.5px solid var(--color-border, #334155);
  border-radius: 16px;
  padding: 2.5rem 2rem 2rem 2rem;
  box-shadow: 0 2px 12px 0 rgba(0,0,0,0.10);
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 420px;
  width: 100%;
}
.empty-icon {
  margin-bottom: 0.5rem;
}
.empty-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text, #cfd2fa);
  margin-bottom: 0.5rem;
  text-align: center;
}
.empty-desc {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1.05rem;
  text-align: center;
  margin-bottom: 1.2rem;
  line-height: 1.5;
}
</style>
