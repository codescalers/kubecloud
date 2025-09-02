<template>
  <div class="card vm-card">
    <div class="vm-header">
      <h3 class="vm-title">
        {{ vm.name }} <span class="vm-type">({{ type }})</span>
      </h3>
      <div class="vm-actions">
        <v-btn icon="mdi-pencil" size="small" @click="$emit('edit')" aria-label="Edit VM" />
        <v-btn icon="mdi-delete" size="small" @click="$emit('delete')" aria-label="Delete VM" />
      </div>
    </div>
    <div class="vm-specs">
      <v-chip color="primary" text-color="white" size="small" class="mr-2" variant="outlined">
        <v-icon size="16" class="mr-1">mdi-cpu-64-bit</v-icon>
        {{ vm.vcpu }} vCPU
      </v-chip>
      <v-chip color="success" text-color="white" size="small" class="mr-2" variant="outlined">
        <v-icon size="16" class="mr-1">mdi-memory</v-icon>
        {{ vm.ram }} GB RAM
      </v-chip>
      <v-chip color="info" text-color="white" size="small" class="mr-2" variant="outlined">
        <v-icon size="16" class="mr-1">mdi-harddisk</v-icon>
        {{ vm.disk }} GB Disk
      </v-chip>
      <v-chip v-if="vm.gpu" color="deep-purple-accent-2" text-color="white" size="small" class="mr-2" variant="outlined">
        <v-icon size="16" class="mr-1">mdi-expansion-card</v-icon>
        GPU
      </v-chip>
      <div class="spec-item" style="margin-top: 0.7em;">
          <span class="spec-label">SSH Keys:</span>
          <div class="ssh-keys-chips">
            <span v-for="id in vm.sshKeyIds" :key="id" class="ssh-key-chip">
              {{ availableSshKeys.find(k => k.ID === id)?.name }}
            </span>
          </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { VM, SshKey } from '../../composables/useDeployCluster';
import { defineProps, defineEmits } from 'vue';
const props = defineProps<{ vm: VM; type: 'master' | 'worker'; availableSshKeys: SshKey[] }>();
const emit = defineEmits(['edit', 'delete']);
</script>

<style scoped>
.card.vm-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 16px;
  padding: 1.5rem 1.2rem;
  margin-bottom: 1.2rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.vm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}
.vm-title {
  font-size: 1.1rem;
  font-weight: 600;
}
.vm-type {
  font-size: 0.95em;
  color: var(--color-text-muted, #7c7fa5);
  margin-left: 0.5em;
}
.vm-actions {
  display: flex;
  gap: 0.5rem;
}
.vm-specs {
  display: flex;
  flex-wrap: wrap;
  gap: 1.2rem;
  margin-top: 0.5rem;
}
.spec-item {
  font-size: 1em;
  color: var(--color-text, #cfd2fa);
  min-width: 110px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5em;
}

.ssh-keys-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4em;
  max-width: 100%;
}
.spec-label {
  color: var(--color-text-muted, #7c7fa5);
  margin-right: 0.5em;
}
.ssh-key-chip {
  background: var(--color-surface-2, #23243a);
  color: var(--color-primary, #6366f1);
  border-radius: 8px;
  padding: 0.1em 0.7em;
  margin-right: 0.4em;
  font-size: 0.97em;
}
</style>
