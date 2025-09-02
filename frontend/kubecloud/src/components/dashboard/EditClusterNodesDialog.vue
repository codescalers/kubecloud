<template>
    <v-dialog v-model="props.modelValue" max-width="900">
      <BaseDialogCard>
        <template #title>
          Add Node
        </template>
        <template #content>
          <div>
            <v-form ref="form"  v-model="formValid">
              <div class="add-form-wrapper">
                <v-text-field
                  :rules="nameRules"
                  v-model="addFormName"
                  label="Name"
                />
                <v-text-field validate-on="eager" :rules="[RULES.cpu]" v-model.number="addFormCpu" label="CPU" type="number" min="1" />
                <v-text-field validate-on="eager" :rules="[RULES.ram]" v-model.number="addFormRam" label="RAM (GB)" type="number" min="1" />
                <v-text-field validate-on="eager" :rules="[RULES.storage]" v-model.number="addFormStorage" label="Storage (GB)" type="number" min="1" />
                <NodeSelect
                :loading="nodesLoading"
                v-model="addFormNodeId"
                :items="availableNodesWithName"
                label="Select Node"
                :get-node-resources="node => ({ cpu: getTotalCPU(node), ram: getAvailableRAM(node), storage: getAvailableStorage(node) })"
                :cpu-label="'CPU'"
                :gpu-icon="'mdi-nvidia'"
              />
              <v-select v-model="addFormRole" :items="['master', 'worker']" label="Role" />
              <div class="ssh-key-section" style="margin-top: 1.5rem; width: 100%;">
                <label class="ssh-key-label">SSH Keys</label>
                <div v-if="sshKeysLoading" class="mb-2"><v-progress-circular indeterminate size="24" color="primary" /></div>
                <v-chip-group
                  v-else
                  v-model="addFormSshKeys"
                  :multiple="true"
                  column
                  class="mb-2"
                >
                  <v-chip
                    v-for="key in sshKeys"
                    :key="key.ID"
                    :value="key.ID"
                    color="primary"
                    class="ma-1"
                    variant="elevated"
                  >
                    {{ key.name }}
                  </v-chip>
                </v-chip-group>
                <div v-if="!sshKeysLoading && sshKeys.length === 0" class="ssh-alert">
                  <v-icon color="error" class="mr-1">mdi-alert-circle</v-icon>
                  <span>No SSH keys found. Please add one in your dashboard.</span>
                </div>
              </div>
            </div>
          </v-form>
          </div>
        </template>
        <template #actions>
          <div >
            <v-btn variant="outlined" color="primary" :loading="submitting" :disabled="!canAssignToNode || submitting || !formValid" @click="confirmAddForm" class="mr-3">Add Node</v-btn>
            <v-btn variant="outlined" @click="emit('update:modelValue', false)">Cancel</v-btn>
          </div>
        </template>
      </BaseDialogCard>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { getAvailableCPU, getAvailableRAM, getAvailableStorage, getTotalCPU } from '../../utils/nodeNormalizer';
import type { RawNode } from '../../types/rawNode';
import BaseDialogCard from './BaseDialogCard.vue';
import { userService } from '../../utils/userService';
import { ROOTFS } from '../../composables/useDeployCluster';
import NodeSelect from '../ui/NodeSelect.vue';
import { RULES, createUniqueNodeNameRule } from "../../utils/validation";
import { useNodes } from '../../composables/useNodes';
import { useNotificationStore } from '../../stores/notifications';

const props = defineProps<{
  modelValue: boolean,
  cluster: any
}>();
const emit = defineEmits(['update:modelValue', 'nodes-updated', 'remove-node']);
// Initialize composables
const { nodes, loading: nodesLoading, fetchNodes } = useNodes()
const notificationStore = useNotificationStore()


const clusterNodes = computed(() => {
  if (!props.cluster?.cluster?.nodes) return []
  if (Array.isArray(props.cluster.cluster.nodes)) {
    return props.cluster.cluster.nodes.filter((node: any) => typeof node === 'object' && node !== null)
  }
  return []
})

const editNodes = ref<any[]>([])
watch(() => clusterNodes.value, (val) => { editNodes.value = val || [] }, { immediate: true })

// Add node form state
const addFormNodeId = ref<number|null>(null);
const addFormRole = ref('master');
const addFormCpu = ref(2);
const addFormRam = ref(4);
const addFormStorage = ref(25);
const addFormName = ref('');
const addFormSshKeys = ref<number[]>([]);
const sshKeys = ref<any[]>([]);
const sshKeysLoading = ref(false);
const sshKeysError = ref('');
const formValid = ref(false);
const submitting = ref(false);
const form = ref<any>(null)
const names = clusterNodes.value.map((node: any) => node.original_name);
const nameRules = computed(() => [
  createUniqueNodeNameRule(names, addFormName.value)
]);


// Node management functions
async function addNodeToDeployment(deploymentName: string, clusterPayload: { name: string, nodes: any[] }) {
  return await userService.addNodeToDeployment(deploymentName, clusterPayload)
}


async function handleAddNode(payload: any) {
  if (!payload || !payload.name || !Array.isArray(payload.nodes) || payload.nodes.length === 0) {
    notificationStore.error('Add Node Error', 'Invalid node payload.')
    return
  }
  try {
    await addNodeToDeployment(payload.name, payload)
    notificationStore.info('Deployment is being updated', 'Your node is being added in the background. You will be notified when it is ready.')
    emit('update:modelValue', false) // Close dialog on success
  } catch (e: any) {
    console.error(e)
    notificationStore.error('Add Node Failed', e?.message || 'Failed to add node')
  }
}




// Reset form fields to default values
function resetForm() {
  form.value?.reset()
  addFormRole.value = 'master'
  addFormCpu.value = 2
  addFormRam.value = 4
  addFormStorage.value = 25
  addFormSshKeys.value = sshKeys.value.length > 0 ? [sshKeys.value[0].ID] : []
}

watch(() => props.modelValue, async (isOpen) => {
  if (isOpen) {
    await fetchNodes()
  } else {
    resetForm()
  }
})

onMounted(async () => {
  sshKeysLoading.value = true;
  try {
    await fetchNodes()
    sshKeys.value = await userService.listSshKeys();
    if (sshKeys.value.length > 0) {
      addFormSshKeys.value = [sshKeys.value[0].ID];
    }
  } catch (e: any) {
    sshKeysError.value = e?.message || 'Failed to load SSH keys';
  } finally {
    sshKeysLoading.value = false;
  }
});
// Available nodes computation
const availableNodes = computed<RawNode[]>(() => {
  return nodes.value.filter((node: RawNode) => {
    const availRAM = getAvailableRAM(node)
    const availStorage = getAvailableStorage(node)
    return availRAM > 0 && availStorage > 0
  })
})

const addFormNode = computed<RawNode | undefined>(() => availableNodes.value.find((n: RawNode) => n.nodeId === addFormNodeId.value));
const canAssignToNode = computed(() => {
  const node = addFormNode.value;
  if (!node) return false;
  return (
    addFormCpu.value > 0 &&
    addFormRam.value > 0 &&
    addFormStorage.value > 0 &&
    addFormRam.value <= getAvailableRAM(node) &&
    addFormStorage.value <= getAvailableStorage(node)
  );
});
watch([addFormNodeId, addFormRam, addFormStorage], () => {
  const node = addFormNode.value;
  if (!node) {
    return;
  }
  if (
    addFormRam.value > getAvailableRAM(node) ||
    addFormStorage.value > getAvailableStorage(node)
  ) {
    addFormNodeId.value=null;
  }
});
async function confirmAddForm() {
  // Get all selected SSH keys and concatenate their public keys
  const sshKeyPublicKeys = addFormSshKeys.value
    .map(id => sshKeys.value.find((k: any) => k.ID === id)?.public_key)
    .filter(key => key) // Remove undefined values
    .join('\n'); // Join multiple keys with newlines
  const payload = {
    name: props.cluster.cluster.name,
    token: '',
    nodes: [
      {
        name: addFormName.value,
        type: addFormRole.value,
        node_id: addFormNodeId.value,
        cpu: addFormCpu.value,
        memory: addFormRam.value * 1024,
        root_size: ROOTFS * 1024,
        disk_size: addFormStorage.value * 1024,
        env_vars: sshKeyPublicKeys ? { SSH_KEY: sshKeyPublicKeys } : {},
      }
    ]
  };
  try {
    submitting.value = true;
    await handleAddNode(payload);
  } catch (e) {
  } finally {
    submitting.value = false;
  }
}
// Filter available nodes based on form requirements and add 'name' property for v-select display
const availableNodesWithName = computed(() =>
  availableNodes.value
    .filter(node => {
      // Only show nodes that have sufficient available resources
      const availableRAM = getAvailableRAM(node);
      const availableStorage = getAvailableStorage(node);
      const availableCPU = getAvailableCPU(node);
      return (
        addFormCpu.value <= availableCPU &&
        addFormRam.value <= availableRAM &&
        addFormStorage.value <= availableStorage
      );
    })
);
</script>
<style scoped>
.polished-error {
  color: red;
}
</style>
