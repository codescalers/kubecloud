<template>
  <div>
    <div class="section-header">
      <h3 class="section-title">
        <v-icon icon="mdi-check-circle" class="mr-2"></v-icon>
        Review Configuration
      </h3>
      <p class="section-subtitle">Review your cluster configuration before deployment</p>
    </div>
    <div class="review-grid">
      <div class="review-card cluster-info-card">
        <h4 class="review-title">Cluster Information</h4>
        <div class="cluster-info">
          <div class="cluster-name-display">
            <div class="cluster-name-label">Cluster Name</div>
            <div class="cluster-name-value">{{ clusterName }}</div>
          </div>
        </div>
      </div>
      <div class="review-card">
        <h4 class="review-title">Node Assignment</h4>
        <div class="node-assignments">
          <div v-for="(vm, index) in allVMs" :key="index" class="node-assignment">
            <div class="assignment-icon" :class="vm.name.includes('Master') ? 'master' : 'worker'">
              <v-icon :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'" color="white" />
            </div>
            <div class="assignment-details">
              <div class="assignment-name">{{ vm.name }}</div>
              <div class="assignment-node">Node: {{ vm.node ?? 'Unassigned' }}</div>
              <div class="assignment-specs">
                <v-chip color="primary" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-cpu-64-bit</v-icon>
                  {{ vm.vcpu }} vCPU
                </v-chip>
                <v-chip color="success" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-memory</v-icon>
                  {{ vm.ram }} GB RAM
                </v-chip>
                <v-chip color="info" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-harddisk</v-icon>
                  {{ vm.disk }} GB Disk
                </v-chip>
                <v-chip v-if="vm.gpu" color="deep-purple-accent-2" text-color="white" size="x-small" class="mr-1" variant="outlined">
                  <v-icon size="14" class="mr-1">mdi-expansion-card</v-icon>
                  GPU
                </v-chip>
                <span class="ssh-key-label">SSH Key: <b>{{ getSshKeyName(vm.sshKeyIds[0]) }}</b></span>
              </div>
              <div v-if="vm.node !== null && nodeResourceErrors[vm.node]" class="node-resource-warning">
                <v-alert type="error" dense>{{ nodeResourceErrors[vm.node].join(', ') }}</v-alert>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="step-actions">
      <v-btn variant="outlined" @click="$emit('prevStep')">
        <v-icon start icon="mdi-arrow-left"></v-icon>
        Back
      </v-btn>
      <v-btn variant="outlined" color="success" :loading="deploying" @click="$emit('onDeployCluster')">
        <v-icon start icon="mdi-rocket-launch"></v-icon>
        Deploy Cluster
      </v-btn>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { VM } from '../../composables/useDeployCluster';
import { defineProps, defineEmits } from 'vue';
defineProps<{
  allVMs: VM[];
  getNodeInfo: (id: number) => string;
  deploying: boolean;
  nodeResourceErrors: Record<number, string[]>;
  getSshKeyName: (id: number) => string;
  clusterName: string;
}>();
const emit = defineEmits(['onDeployCluster', 'prevStep']);
</script>
<style scoped>
.section-header {
  margin-bottom: 2rem;
  text-align: center;
}

.section-title {
  font-size: 1.4rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.section-subtitle {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1.1rem;
  margin-top: 0.25rem;
}
.review-grid {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}

.review-card,
.cluster-info-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}

.cluster-info-card {
  flex: 1 1 100%;
  min-width: 100%;
}

.review-card {
  flex: 1 1 320px;
  min-width: 320px;
}
.review-title {
  font-size: 1.2rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  color: var(--color-text, #cfd2fa);
  text-align: center;
}
.cluster-info {
  padding: 1.5rem;
}

.cluster-name-display {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
}

.cluster-name-label {
  font-weight: 500;
  color: var(--color-text-muted, #7c7fa5);
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.cluster-name-value {
  background: var(--color-surface-2, #23243a);
  border: 1px solid var(--color-surface-3, #334155);
  border-radius: 8px;
  padding: 1rem;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 1rem;
  color: var(--color-text, #cfd2fa);
  width: 100%;
  text-align: left;
  font-weight: 400;
}
.node-assignments {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}
.node-assignment {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--color-surface-2, #23243a);
}
.assignment-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary, #6366f1);
  flex-shrink: 0;
}

.assignment-icon.worker {
  background: var(--color-success, #22d3ee);
}

.assignment-icon .v-icon {
  font-size: 20px !important;
}
.assignment-details {
  display: flex;
  flex-direction: column;
}
.assignment-name {
  font-weight: 500;
  color: var(--color-text, #cfd2fa);
}
.assignment-node {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 0.98rem;
}
.assignment-specs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.ssh-key-label {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 0.98rem;
}
.node-resource-warning {
  margin-top: 0.5rem;
}
.step-actions {
  display: flex;
  justify-content: center;
  gap: 1.5rem;
  margin-top: 3rem;
}
</style>
