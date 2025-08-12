<template>
  <div v-if="visible && localNode">
    <div class="modal-backdrop" @click="$emit('cancel')"></div>
    <div class="edit-node-modal">
      <h3>Edit Node</h3>
      <div class="modal-fields">
        <label>Name</label>
        <input type="text" v-model="localNode.name" />
        <div v-if="errors.name" class="field-error">{{ errors.name }}</div>
        <label>vCPU</label>
        <input type="number" v-model.number="localNode.vcpu" />
        <div v-if="errors.vcpu" class="field-error">{{ errors.vcpu }}</div>
        <label>RAM (GB)</label>
        <input type="number" v-model.number="localNode.ram" />
        <div v-if="errors.ram" class="field-error">{{ errors.ram }}</div>
        <label>Disk Size (GB)</label>
        <input type="number" v-model.number="localNode.disk" />
        <div v-if="errors.disk" class="field-error">{{ errors.disk }}</div>
        <!-- <v-switch v-model="localNode.gpu" label="GPU" inset class="mt-2" color="primary" /> -->
        <div class="ssh-key-section" style="margin-top: 1.5rem;">
          <label class="ssh-key-label">SSH Key</label>
          <v-chip-group
            v-model="selectedSshKeyId"
            :multiple="false"
            column
          >
            <v-chip
              v-for="key in availableSshKeys"
              :key="key.ID"
              :value="key.ID"
              color="primary"
              class="ma-1"
              variant="elevated"
            >
              {{ key.name }}
            </v-chip>
          </v-chip-group>
          <div v-if="!selectedSshKeyId" class="ssh-alert">
            <v-icon color="error" class="mr-1">mdi-alert-circle</v-icon>
            <span>Please select an SSH key to proceed.</span>
          </div>
        </div>
        <div v-if="errors.ssh" class="field-error">{{ errors.ssh }}</div>
      </div>
      <div class="modal-actions">
        <button @click="onSave" :disabled="!valid">Save</button>
        <button @click="$emit('cancel')">Cancel</button>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { defineProps, defineEmits, watch, ref, computed } from 'vue';
import type { PropType } from 'vue';
import type { VM, SshKey } from '../composables/useDeployCluster';
import { required, min, isAlphanumeric, max } from '@/utils/validation';
const props = defineProps({
  node: { type: Object as PropType<VM>, required: true },
  visible: { type: Boolean, required: true },
  availableSshKeys: { type: Array as PropType<SshKey[]>, required: true }
});
const emit = defineEmits<{ (e: 'save', node: VM): void; (e: 'cancel'): void }>();
const localNode = ref<VM>({ ...props.node });
const selectedSshKeyId = ref<number | null>(null);

// When the modal opens or node changes, set the selected SSH key appropriately
watch(
  [() => props.node, () => props.availableSshKeys],
  ([node, keys]) => {
    localNode.value = { ...node };
    if (node.sshKeyIds && node.sshKeyIds.length > 0) {
      selectedSshKeyId.value = node.sshKeyIds[0];
    } else if (keys.length > 0) {
      selectedSshKeyId.value = keys[0].ID;
    } else {
      selectedSshKeyId.value = null;
    }
  },
  { immediate: true }
);

const errors = computed(() => {
  const node = localNode.value;
  const errs: Record<string, string> = {};
  errs.name = required('Name is required')(node.name) || isAlphanumeric('Node name can only contain letters, and numbers.')(node.name) || "";
  errs.vcpu = min('vCPU must be at least 1', 1)(node.vcpu)|| max('vCPU must be at most 32', 32)(node.vcpu) || "";
  errs.ram = min('RAM must be at least 0.5GB', 0.5)(node.ram)|| max('RAM must be at most 256GB', 256)(node.ram) || "";
  errs.disk = min('Disk must be at least 15GB', 15)(node.disk)|| max('Disk must be at most 10000GB', 10000)(node.disk) || "";
  // Only require SSH key if there are any available
  if (props.availableSshKeys.length > 0 && !selectedSshKeyId.value) errs.ssh = 'At least one SSH key must be selected.';
  return errs;
});

const valid = computed(() => {
  return Object.values(errors.value).every(e => !e);
});
function onSave() {
  if (valid.value) {
    emit('save', { ...localNode.value, sshKeyIds: selectedSshKeyId.value !== null ? [selectedSshKeyId.value] : [] });
  }
}

</script>
<style scoped>
.modal-backdrop { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; }
.edit-node-modal { position: fixed; top: 55%; left: 50%; transform: translate(-50%, -50%); background: #232946; color: #fff; border-radius: 16px; box-shadow: 0 8px 32px 0 rgba(0,0,0,0.25); padding: 2rem 2.5rem; z-index: 1001; width: 30vw; min-width: 320px; max-width: 90vw; }
.edit-node-modal h3 { margin-top: 0; margin-bottom: 1.2rem; font-size: 1.3rem; font-weight: 700; }
.modal-fields label { display: block; margin-top: 1rem; font-weight: 500; }
.modal-fields input[type="number"], .modal-fields input[type="text"] { width: 100%; margin-top: 0.3rem; padding: 0.4em 0.7em; border-radius: 6px; border: 1px solid #4f8cff; background: #1a1f2b; color: #fff; }
.modal-actions { display: flex; gap: 1rem; margin-top: 2rem; justify-content: flex-end; }
.modal-actions button { background: #4f8cff; color: #fff; border: none; border-radius: 8px; padding: 0.5em 1.2em; font-size: 1em; cursor: pointer; transition: background 0.2s; }
.modal-actions button:last-child { background: #232946; border: 1px solid #4f8cff; color: #4f8cff; }
.field-error { color: #ff5252; font-size: 0.97em; margin-top: 0.2em; }
.ssh-key-section {
  margin-top: 1.5rem;
}
.ssh-key-label {
  font-weight: 600;
  margin-bottom: 0.5rem;
  display: block;
}
.ssh-alert {
  color: #ff5252;
  background: #2d1a1a;
  border-radius: 6px;
  padding: 0.5em 1em;
  margin-top: 0.7em;
  display: flex;
  align-items: center;
  font-size: 1em;
}
</style> 