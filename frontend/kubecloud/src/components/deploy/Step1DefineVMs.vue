<template>
  <div>
    <div v-if="!props.sshKeysLoading && !availableSshKeys.length" class="ssh-key-warning">
      <v-alert type="warning" border="start" prominent>
        <div>No SSH keys found. <v-btn color="primary" variant="outlined" @click="$emit('navigateToSshKeys')">Add SSH Key</v-btn></div>
      </v-alert>
    </div>
    <div v-else-if="props.sshKeysLoading" class="ssh-key-loading">
      <v-skeleton-loader type="list-item" :loading="true" />
    </div>
    <div class="vm-config-grid">
      <div class="vm-config-card">
        <div class="card-header">
          <h4 class="card-title">
            <v-icon icon="mdi-server" class="mr-2"></v-icon>
            Master Nodes
          </h4>
          <v-btn color="primary" :disabled="masters.length >= 3" prepend-icon="mdi-plus" size="small" variant="outlined" @click="addMaster">Add Master</v-btn>
        </div>
        <DeployVMCard v-for="(master, masterIdx) in masters" :key="masterIdx" :vm="master" type="master" :availableSshKeys="availableSshKeys" @edit="() => openEditNodeModal('master', masterIdx)" @delete="() => removeMaster(masterIdx)" />
        <div v-if="!masters.length" class="empty-state">
          <v-icon icon="mdi-plus-circle-outline" size="32" color="var(--color-text-muted)"></v-icon>
          <p>No master nodes configured</p>
        </div>
      </div>
      <div class="vm-config-card">
        <div class="card-header">
          <h4 class="card-title">
            <v-icon icon="mdi-desktop-tower-monitor" class="mr-2"></v-icon>
            Worker Nodes
          </h4>
          <v-btn color="primary" prepend-icon="mdi-plus" size="small" variant="outlined" @click="addWorker" class="add-btn">Add Worker</v-btn>
        </div>
        <DeployVMCard v-for="(worker, workerIdx) in workers" :key="workerIdx" :vm="worker" type="worker" :availableSshKeys="availableSshKeys" @edit="() => openEditNodeModal('worker', workerIdx)" @delete="() => removeWorker(workerIdx)" />
        <div v-if="!workers.length" class="empty-state">
          <v-icon icon="mdi-plus-circle-outline" size="32" color="var(--color-text-muted)"></v-icon>
          <p>No worker nodes configured</p>
        </div>
      </div>
    </div>
    <div class="step-actions">
      <v-btn variant="outlined" color="primary" :disabled="!isStep1Valid" @click="$emit('nextStep')">
        Continue
        <v-icon end icon="mdi-arrow-right"></v-icon>
      </v-btn>
    </div>
  </div>
</template>
<script setup lang="ts">
import DeployVMCard from '../deploy/DeployVMCard.vue';
import type { VM, SshKey } from '../../composables/useDeployCluster';
import { defineProps, defineEmits, withDefaults } from 'vue';
const props = withDefaults(defineProps<{
  masters: VM[];
  workers: VM[];
  availableSshKeys: SshKey[];
  addMaster: () => void;
  addWorker: () => void;
  removeMaster: (idx: number) => void;
  removeWorker: (idx: number) => void;
  openEditNodeModal: (type: 'master' | 'worker', idx: number) => void;
  selectedSshKeys?: number[];
  setSelectedSshKeys: (keys: number[]) => void;
  isStep1Valid?: boolean;
  sshKeysLoading?: boolean;
}>(), {
  selectedSshKeys: () => [],
  isStep1Valid: false,
  sshKeysLoading: false
});
const emit = defineEmits(['navigateToSshKeys', 'nextStep']);
</script>
<style scoped>
.vm-config-grid {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
  margin-top: 2rem;
}
.vm-config-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 20px;
  padding: 2rem 1.5rem 1.5rem 1.5rem;
  margin-bottom: 2rem;
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
  flex: 1 1 320px;
  min-width: 320px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}
.card-title {
  display: flex;
  align-items: center;
  font-size: 1.1rem;
  font-weight: 500;
  gap: 0.5rem;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: var(--color-text-muted, #7c7fa5);
  margin-top: 1.5rem;
  font-size: 1.08rem;
}
.empty-state v-icon {
  margin-bottom: 0.5rem;
  font-size: 2.5rem !important;
  color: #7c7fa5 !important;
  opacity: 0.7;
}
.step-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 2rem;
}
</style>
