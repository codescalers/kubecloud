<template>
  <div>
    <div class="section-header">
      <h3 class="section-title">
        <v-icon icon="mdi-server-network" class="mr-2"></v-icon>
        Assign VMs to Nodes
      </h3>
      <p class="section-subtitle">Select nodes to host your cluster VMs</p>
      <v-alert
        type="info"
        variant="tonal"
        class="mb-7"
        icon="mdi-tag-outline"
      >
        <span class="text-h6 font-weight-bold">50% Discount Available!</span>
        <div class="d-flex align-center justify-space-between flex-wrap">
          <p class="text-body-1 flex-grow-1 me-4 mb-0">
            Reserve a node to get 50% discount and exclusive usage. Shared nodes are available at full price.
          </p>
          <v-btn
            variant="outlined"
            color="white"
            class="ms-auto"
            :to="{ name: 'reserve' }"
            prepend-icon="mdi-arrow-right"
          >
            Reserve Node
          </v-btn>
        </div>
      </v-alert>

      <div class="region-filter mb-4">
        <v-select
          v-model="selectedRegion"
          :items="availableRegions"
          label="Filter by region"
          prepend-inner-icon="mdi-earth"
          variant="outlined"
          density="compact"
          clearable
          :loading="loading"
          placeholder="All regions"
          item-title="label"
          item-value="value"
        />
      </div>
    </div>
    <v-row>
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
          <NodeSelect
            ref="nodeSelectRef"
            :model-value="vm.node"
            @update:modelValue="val => onNodeSelected(val, index)"
            :items="getAvailableNodesForVM(index)"
            label="Select Node"
            clearable
            class="node-select"
            :get-node-resources="getNodeAvailableResources"
            cpu-label="vCPU"
            :loading="validatingNode"
            :error-message="validationError"
            :error="!!validationError"

          />
        </div>
      </v-col>
    </v-row>
    <div class="step-actions">
      <v-btn variant="outlined" @click="$emit('prevStep')">
        <v-icon start icon="mdi-arrow-left"></v-icon>
        Back
      </v-btn>
      <v-btn variant="outlined" color="primary" :disabled="!isStep2Valid || validatingNode || !!validationError" @click="$emit('nextStep')">
        Continue
        <v-icon end icon="mdi-arrow-right"></v-icon>
      </v-btn>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { NormalizedNode } from '../../types/normalizedNode';
import { type VM } from '../../composables/useDeployCluster';
import { defineProps, withDefaults, defineEmits, onMounted, computed, ref, watch, useTemplateRef } from 'vue';
import NodeSelect from '../ui/NodeSelect.vue';
import useNodeStoragePool from '@/composables/useNodeStoragePool';

const props = withDefaults(defineProps<{
  allVMs: VM[];
  availableNodes: NormalizedNode[];
  getNodeInfo: (id: number) => string;
  onAssignNode: (vmIdx: number, nodeId: number | null) => void;
  onRegionFilter: (region?: string) => Promise<void>;
  loading?: boolean;
  isStep2Valid?: boolean;
}>(), {
  loading: false,
  isStep2Valid: false
});
const emit = defineEmits(['nextStep', 'prevStep']);
const nodeStoragePool = useNodeStoragePool();
const validationError = ref<string>('');


const selectedRegion = ref<string>('');
const validatingNode = ref<boolean>(false);

// All supported regions - show all regions, let backend handle filtering
const ALL_REGIONS = ['Africa', 'Asia', 'South America', 'North America', 'Europe', 'Oceania'];

const availableRegions = computed(() => [
  { label: 'All regions', value: '' },
  ...ALL_REGIONS.map(region => ({ label: region, value: region }))
]);

watch(selectedRegion, async (newRegion, oldRegion) => {
  if (newRegion !== oldRegion) {
    resetVMAssignments();
    await props.onRegionFilter(newRegion || undefined);
  }
});

// Validate and clear invalid VM assignments
const validateVMAssignments = () => {
  props.allVMs.forEach((vm, vmIndex) => {
    if (vm.node != null && !getAvailableNodesForVM(vmIndex).find(node => node.nodeId === vm.node)) {
      props.onAssignNode(vmIndex, null);
    }
  });
};

const resetVMAssignments = () => {
  props.allVMs.forEach((vm, index) => {
    if (vm.node !== null) props.onAssignNode(index, null);
  });
};

onMounted(() => {
  selectedRegion.value = '';
  validateVMAssignments();
});


const currentAllocations = computed(() => {
  const allocations: Record<number, { ram: number; storage: number }> = {};

  for (const vm of props.allVMs) {
    if (vm.node != null) {
      if (!allocations[vm.node]) {
        allocations[vm.node] = { ram: 0, storage: 0 };
      }
      allocations[vm.node].ram += vm.ram;
      allocations[vm.node].storage += (vm.disk || 0) + vm.rootfs;
    }
  }

  return allocations;
});

const getNodeResources = (node: NormalizedNode, excludeVM?: { ram: number; storage: number }) => {
  const used = currentAllocations.value[node.nodeId] || { ram: 0, storage: 0 };
  const excludeRam = excludeVM?.ram || 0;
  const excludeStorage = excludeVM?.storage || 0;

  return {
    cpu: node.cpu || 0,
    ram: (node.available_ram || 0) - used.ram + excludeRam,
    storage: (node.available_storage || 0) - used.storage + excludeStorage
  };
};

const getAvailableNodesForVM = (vmIndex: number) => {
  const vm = props.allVMs[vmIndex];
  if (!vm) return [];

  const requiredStorage = (vm.disk || 0) + vm.rootfs;

  return props.availableNodes.filter(node => {
    const nodeIsNotInCluster = !props.allVMs.some((vm: any) => vm.node === node.nodeId);
    if(!nodeIsNotInCluster &&  vm.node !== node.nodeId){
      return false;
    }
    const vmResources = vm.node === node.nodeId ? { ram: vm.ram, storage: requiredStorage } : undefined;
    const available = getNodeResources(node, vmResources);
    return available.cpu >= vm.vcpu && available.ram >= vm.ram && available.storage >= requiredStorage;
  });
};


const getNodeAvailableResources = (node: NormalizedNode) => {
  const available = getNodeResources(node);
  return {
    cpu: available.cpu,
    ram: Math.max(0, available.ram),
    storage: Math.max(0, available.storage)
  };
};

const onNodeSelected = async (val: any, index: number) => {
   props?.onAssignNode(index, val)
   validationError.value = ''
  if (val) {
    validatingNode.value = true
    console.log(val, index)
    const vm = props.allVMs[index];
    const requiredStorage = (vm.disk || 0) + vm.rootfs;
    try {
      const isValid = await nodeStoragePool.validateNodeStoragePool(requiredStorage, val);
      if (!isValid) {
        validatingNode.value = false
        validationError.value = nodeStoragePool.createStoragePoolError(val)
        return
      }
      validationError.value = ''
    } catch (error ) {
      console.error(error)
      validationError.value = nodeStoragePool.failedToCheckStoragePoolError().message
      return
    }finally {
      validatingNode.value = false
    }
  }
}
</script>
<style scoped>
.section-header {
  margin-bottom: 2rem;
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

</style>
