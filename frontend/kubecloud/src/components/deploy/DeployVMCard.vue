<template>
  <v-card color="surface-variant" class="pa-4 mb-3">
    <div class="d-flex justify-space-between align-center mb-3">
      <div>
        <h3 class="text-h6 font-weight-medium">
          {{ vm.name }} <v-chip size="small" color="primary" variant="outlined" class="ml-2">{{ type }}</v-chip>
        </h3>
      </div>
      <div class="d-flex">
        <v-btn icon="mdi-pencil" size="small" variant="text" @click="$emit('edit')" aria-label="Edit VM" />
        <v-btn icon="mdi-delete" size="small" variant="text" color="error" @click="$emit('delete')" aria-label="Delete VM" />
      </div>
    </div>

    <div class="d-flex flex-wrap ga-2 mb-3">
      <v-chip color="primary" size="small" variant="outlined">
        <v-icon size="16" class="me-1">mdi-cpu-64-bit</v-icon>
        {{ vm.vcpu }} vCPU
      </v-chip>
      <v-chip color="success" size="small" variant="outlined">
        <v-icon size="16" class="me-1">mdi-memory</v-icon>
        {{ vm.ram }} GB RAM
      </v-chip>
      <v-chip color="info" size="small" variant="outlined">
        <v-icon size="16" class="me-1">mdi-harddisk</v-icon>
        {{ vm.disk }} GB Disk
      </v-chip>
      <v-chip v-if="vm.gpu" color="deep-purple-accent-2" size="small" variant="outlined">
        <v-icon size="16" class="me-1">mdi-nvidia</v-icon>
        GPU
      </v-chip>
    </div>

    <div v-if="vm.sshKeyIds?.length">
      <div class="text-body-2 text-medium-emphasis mb-2">SSH Keys:</div>
      <div class="d-flex flex-wrap ga-1">
        <v-chip
          v-for="id in vm.sshKeyIds"
          :key="id"
          size="x-small"
          color="primary"
          variant="tonal"
        >
          {{ availableSshKeys.find(k => k.ID === id)?.name }}
        </v-chip>
      </div>
    </div>
  </v-card>
</template>
<script setup lang="ts">
import type { VM, SshKey } from '../../composables/useDeployCluster';
import { defineProps, defineEmits } from 'vue';
const props = defineProps<{ vm: VM; type: 'master' | 'worker'; availableSshKeys: SshKey[] }>();
const emit = defineEmits(['edit', 'delete']);
</script>


