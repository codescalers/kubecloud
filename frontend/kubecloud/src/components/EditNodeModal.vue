<template>
  <div v-if="visible && localNode">
    <div class="modal-backdrop" @click="$emit('cancel')"></div>
    <div class="edit-node-modal">
      <h3>Edit Node</h3>
      <v-form ref="formRef" @submit.prevent="onSave" v-model="isFormValid">
        <div class="modal-fields">
          <label>Name</label>
          <v-text-field
            v-model="localNode.name"
            variant="outlined"
            density="compact"
            :rules="nameRules"
            class="form-field"
          />
          <label>vCPU</label>
          <v-text-field
            v-model.number="localNode.vcpu"
            type="number"
            variant="outlined"
            density="compact"
            :rules="[RULES.cpu]"
            class="form-field"
          />
          <label>RAM (GB)</label>
          <v-text-field
            v-model.number="localNode.ram"
            type="number"
            variant="outlined"
            density="compact"
            :rules="[RULES.ram]"
            class="form-field"
          />
          <label>Disk Size (GB)</label>
          <v-text-field
            v-model.number="localNode.disk"
            type="number"
            variant="outlined"
            density="compact"
            :rules="[RULES.storage]"
            class="form-field"
          />
          <!-- <v-switch v-model="localNode.gpu" label="GPU" inset class="mt-2" color="primary" /> -->
          <div class="ssh-key-section" style="margin-top: 1.5rem;">
            <label class="ssh-key-label">SSH Keys</label>
            <v-chip-group
              v-model="selectedSshKeyIds"
              :multiple="true"
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
            <div v-if="selectedSshKeyIds.length === 0" class="ssh-alert">
              <v-icon color="error" class="mr-1">mdi-alert-circle</v-icon>
              <span>Please select at least one SSH key to proceed.</span>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <v-btn
            type="submit"
            color="primary"
            variant="elevated"
            :disabled="!isFormValid || selectedSshKeyIds.length === 0"
          >
            Save
          </v-btn>
          <v-btn
            variant="outlined"
            @click="$emit('cancel')"
          >
            Cancel
          </v-btn>
        </div>
      </v-form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { defineProps, defineEmits, watch, ref, computed } from 'vue';
import type { PropType } from 'vue';
import type { VM, SshKey } from '../composables/useDeployCluster';
import { RULES, createUniqueNodeNameRule } from '../utils/validation';

const props = defineProps({
  node: { type: Object as PropType<VM>, required: true },
  visible: { type: Boolean, required: true },
  availableSshKeys: { type: Array as PropType<SshKey[]>, required: true },
  existingNames: { type: Array as PropType<string[]>, required: true }
});

const emit = defineEmits<{ (e: 'save', node: VM): void; (e: 'cancel'): void }>();
const localNode = ref<VM>({ ...props.node });
const selectedSshKeyIds = ref<number[]>([]);
const formRef = ref();
const isFormValid = ref(false);
const nameRules = computed(() => [
  createUniqueNodeNameRule(props.existingNames, props.node.name)
]);

// When the modal opens or node changes, set the selected SSH key appropriately
watch(
  [() => props.node, () => props.availableSshKeys],
  ([node]) => {
    localNode.value = { ...node };
    if (node.sshKeyIds && node.sshKeyIds.length > 0) {
      selectedSshKeyIds.value = [...node.sshKeyIds];
    } else {
      selectedSshKeyIds.value = [];
    }
  },
  { immediate: true }
);

function onSave() {
  if (isFormValid.value && selectedSshKeyIds.value.length > 0) {
    emit('save', { ...localNode.value, sshKeyIds: [...selectedSshKeyIds.value] });
  }
}
</script>

<style scoped>
.modal-backdrop { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 1000; }
.edit-node-modal { position: fixed; top: 55%; left: 50%; transform: translate(-50%, -50%); background: #232946; color: #fff; border-radius: 16px; box-shadow: 0 8px 32px 0 rgba(0,0,0,0.25); padding: 2rem 2.5rem; z-index: 1001; width: 30vw; min-width: 320px; max-width: 90vw; }
.edit-node-modal h3 { margin-top: 0; margin-bottom: 1.2rem; font-size: 1.3rem; font-weight: 700; }
.modal-fields label { display: block; margin-top: 1rem; font-weight: 500; }
.modal-actions { display: flex; gap: 1rem; margin-top: 2rem; justify-content: flex-end; }
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
