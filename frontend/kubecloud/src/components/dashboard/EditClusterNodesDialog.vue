<template>
  <div>
    <v-dialog v-model="dialog" max-width="900">
      <BaseDialogCard>
        <template #title>
          Edit Cluster Nodes
        </template>
        <template #content>
          <v-tabs v-model="editTab" class="mb-4">
            <v-tab value="list">Node List</v-tab>
            <v-tab value="add">Add Node</v-tab>
          </v-tabs>
          <div v-if="editTab === 'list'">
            <v-table v-if="editNodesWithStorage.length">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Type</th>
                  <th>CPU</th>
                  <th>RAM</th>
                  <th>Storage</th>
                  <th>IP</th>
                  <th>Contract ID</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="node in editNodesWithStorage" :key="node.name">
                  <td>{{ node.original_name }}</td>
                  <td>{{ node.type }}</td>
                  <td>{{ node.cpu }}</td>
                  <td>{{ Math.round((node.memory ?? node.ram) / 1024) }} GB</td>
                  <td>{{ node.storage }} GB</td>
                  <td>{{ node.ip || '-' }}</td>
                  <td>{{ node.contract_id || '-' }}</td>
                  <td>
                    <v-btn @click="showDeleteConfirmation(node.original_name)" :disabled='node.type == "leader"'><v-icon>mdi-delete</v-icon></v-btn>
                  </td>
                </tr>
              </tbody>
            </v-table>
            <div v-else class="empty-list">No nodes in this cluster.</div>
          </div>
          <div v-else-if="editTab === 'add'">
            <v-form  v-model="formValid">
              <div class="add-form-wrapper">
                <v-text-field 
                  validate-on="eager" 
                  :rules="[RULES.nodeName]" 
                  v-model="addFormName" 
                  label="Name" 
                />
                <v-text-field validate-on="eager" :rules="[RULES.cpu]" v-model.number="addFormCpu" label="CPU" type="number" min="1" />
                <v-text-field validate-on="eager" :rules="[RULES.ram]" v-model.number="addFormRam" label="RAM (GB)" type="number" min="1" />
                <v-text-field validate-on="eager" :rules="[RULES.storage]" v-model.number="addFormStorage" label="Storage (GB)" type="number" min="1" />
              <NodeSelect
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
          <div v-if="editTab === 'add'" class="add-form-actions">
            <v-btn variant="outlined" color="primary" :loading="submitting" :disabled="!canAssignToNode || submitting || !formValid" @click="confirmAddForm" class="mr-3">Add Node</v-btn>
            <v-btn variant="outlined" @click="dialog = false">Cancel</v-btn>
          </div>
        </template>
      </BaseDialogCard>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteConfirmDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6">
          Confirm Node Deletion
        </v-card-title>
        <v-card-text>
          Are you sure you want to delete the node <strong>{{ nodeToDelete }}</strong>? This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" variant="text" @click="deleteConfirmDialog = false">
            Cancel
          </v-btn>
          <v-btn color="error" variant="text" @click="confirmDeleteNode">
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { getAvailableCPU, getAvailableRAM, getAvailableStorage, getTotalCPU } from '../../utils/nodeNormalizer';
import type { RawNode } from '../../types/rawNode';
import BaseDialogCard from './BaseDialogCard.vue';
import { userService } from '../../utils/userService';
import { ROOTFS } from '../../composables/useDeployCluster';
import NodeSelect from '../ui/NodeSelect.vue';
import { RULES } from "../../utils/validation";

const props = defineProps<{
  modelValue: boolean,
  cluster: any,
  nodes: any[],
  loading: boolean,
  availableNodes: RawNode[],
  onAddNode: (payload: any) => Promise<void>
}>();
const emit = defineEmits(['update:modelValue', 'nodes-updated', 'remove-node']);
const dialog = computed({
  get: () => props.modelValue,
  set: (val: boolean) => emit('update:modelValue', val)
});
const editTab = ref('list');
const editNodes = ref<any[]>(props.nodes || []);
watch(() => props.nodes, (val) => { editNodes.value = val || []; });
const editNodesWithStorage = computed(() =>
  (editNodes.value || []).map(node => ({
    ...node,
    storage: Math.round(((node.root_size || 0) + (node.disk_size || 0) + (node.storage || 0)) / 1024)
  }))
);
function removeNode(nodeName: string) {
  emit('remove-node', nodeName);
}

function showDeleteConfirmation(nodeName: string) {
  nodeToDelete.value = nodeName;
  deleteConfirmDialog.value = true;
}

function confirmDeleteNode() {
  if (nodeToDelete.value) {
    removeNode(nodeToDelete.value);
    deleteConfirmDialog.value = false;
    nodeToDelete.value = '';
  }
}
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

// Delete confirmation dialog state
const deleteConfirmDialog = ref(false);
const nodeToDelete = ref<string>('');

onMounted(async () => {
  sshKeysLoading.value = true;
  try {
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
const addFormNode = computed<RawNode | undefined>(() => (props.availableNodes || []).find((n: RawNode) => n.nodeId === addFormNodeId.value));
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
    await props.onAddNode(payload);
    editTab.value = 'list';
  } catch (e) {
  } finally {
    submitting.value = false;
  }
}
// Filter available nodes based on form requirements and add 'name' property for v-select display
const availableNodesWithName = computed(() =>
  (props.availableNodes || [])
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
