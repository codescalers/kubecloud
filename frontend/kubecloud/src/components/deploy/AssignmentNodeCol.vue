<template>
  <v-col cols="12">
    <div class="vm-assignment-card">
      <div class="vm-assignment-header">
        <div class="vm-avatar" :class="vm.name.includes('Master') ? 'master' : 'worker'">
          <v-icon
            :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'"
            color="white"
          ></v-icon>
        </div>
        <div class="vm-info">
          <h4 class="vm-title">{{ vm.name }}</h4>

          <v-chip v-if="vm.fullCapabilities" color="success" size="small" variant="outlined">
            <v-icon size="16" class="mr-1">mdi-check</v-icon>
            Use Full Node Capabilities
          </v-chip>
          <template v-else>
            <div class="vm-specs">
              <span class="spec-chip">{{ vm.vcpu }} vCPU</span>
              <span class="spec-chip">{{ vm.ram }}GB RAM</span>
              <span class="spec-chip">{{ vm.disk }}GB Disk</span>
              <span v-if="vm.gpu" class="spec-chip">GPU</span>
            </div>
          </template>
        </div>
      </div>
      <!-- @update:modelValue="onSelect" -->
      <NodeSelect
        ref="nodeSelectRef"
        :model-value="vm.node"
        @update:model-value="(val: any) => onNodeSelected(val, index)"
        :items="items"
        label="Select Node"
        clearable
        class="node-select"
        :get-node-resources="resources"
        cpu-label="vCPU"
        :loading="loading"
        :error-message="err"
        :error="!!err"
      />
      <!-- :loading="validatingNode"
          :error-message="validationError"
          :error="!!validationError" -->
    </div>
  </v-col>
</template>

<script setup lang="ts">
import type { VM } from '@/composables/useDeployCluster'
import type { NormalizedNode } from '@/types/normalizedNode'
import NodeSelect from '../ui/NodeSelect.vue'
import { ref } from 'vue'
import useNodeStoragePool from '@/composables/useNodeStoragePool'

const props = defineProps<{
  vm: VM
  items: NormalizedNode[]
  resources: any
  onAssignNode: any
  index: number
}>()

const nodeStoragePool = useNodeStoragePool()
const loading = ref(false)
const err = ref('')

const onNodeSelected = async (val: any, index: number) => {
  props?.onAssignNode(index, null)
  err.value = ''
  if (val) {
    loading.value = true
    const vm = props.vm
    const requiredStorage = (vm.fullCapabilities ? 0 : vm.disk || 0) + vm.rootfs
    try {
      const isValid = await nodeStoragePool.validateNodeStoragePool(requiredStorage, val)
      if (!isValid) {
        loading.value = false
        err.value = nodeStoragePool.createStoragePoolError(val)
        return
      }
      err.value = ''
      props?.onAssignNode(index, val)
    } catch (error) {
      console.error(error)
      err.value = nodeStoragePool.failedToCheckStoragePoolError().message
      return
    } finally {
      loading.value = false
    }
  }
}
</script>
