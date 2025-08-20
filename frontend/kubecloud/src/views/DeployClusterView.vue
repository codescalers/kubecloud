<template>
  <div class="deploy-container">
    <v-container fluid class="pa-0">
      <div class="deploy-header mb-6">
        <h1 class="hero-title">Deploy New Cluster</h1>
        <p class="section-subtitle">Create and configure your Kubernetes cluster in just a few steps</p>
      </div>

      <div class="deploy-content-wrapper">
        <div class="deploy-card">
          <!-- Progress Indicator -->
          <div class="progress-section">
            <div class="stepper">
              <div class="step" :class="{ active: step >= 1, completed: step > 1 }">
                <div class="step-circle">1</div>
                <div class="step-label">Define VMs</div>
              </div>
              <div class="step" :class="{ active: step >= 2, completed: step > 2 }">
                <div class="step-circle">2</div>
                <div class="step-label">Place VMs</div>
              </div>
              <div class="step" :class="{ active: step >= 3 }">
                <div class="step-circle">3</div>
                <div class="step-label">Review</div>
              </div>
            </div>
          </div>

          <!-- Step Content -->
          <div class="step-content">
            <Step1DefineVMs
              v-if="step === 1"
              :masters="masters"
              :workers="workers"
              :availableSshKeys="availableSshKeys"
              :addMaster="addMaster"
              :addWorker="addWorker"
              :removeMaster="removeMaster"
              :removeWorker="removeWorker"
              :openEditNodeModal="openEditNodeModal"
              :selectedSshKeys="selectedSshKeys"
              :setSelectedSshKeys="setSelectedSshKeys"
              :isStep1Valid="isStep1Valid"
              :sshKeysLoading="sshKeysLoading"
              :clusterName="clusterName"
              :onClusterNameChange="onClusterNameChange"
              @nextStep="nextStep"
              @navigateToSshKeys="navigateToSshKeys"
            />
            <Step2AssignNodes
              v-else-if="step === 2"
              :allVMs="allVMs"
              :availableNodes="availableNodes"
              :getNodeInfo="(id) => getNodeInfo(id, availableNodes)"
              :onAssignNode="onAssignNode"
              :isStep2Valid="isStep2Valid"
              :nodeResourceErrors="nodeResourceErrors"
              @nextStep="nextStep"
              @prevStep="prevStep"
            />
            <Step3Review
              v-else-if="step === 3"
              :allVMs="allVMs"
              :getNodeInfo="(id) => getNodeInfo(id, availableNodes)"
              :deploying="deploying"
              :nodeResourceErrors="nodeResourceErrors"
              :getSshKeyName="(id) => getSshKeyName(id, availableSshKeys)"
              :clusterName="clusterName"
              @onDeployCluster="onDeployCluster"
              @prevStep="prevStep"
            />
          </div>
        </div>
      </div>
    </v-container>
    <EditNodeModal v-if="editNodeModal.open && editNodeModal.node" :node="editNodeModal.node" :visible="editNodeModal.open" :availableSshKeys="availableSshKeys" @save="saveEditNode" @cancel="closeEditNodeModal" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import EditNodeModal from '../components/EditNodeModal.vue';
import { useDeployCluster } from '../composables/useDeployCluster';
import type { VM } from '../composables/useDeployCluster';
import Step1DefineVMs from '../components/deploy/Step1DefineVMs.vue';
import Step2AssignNodes from '../components/deploy/Step2AssignNodes.vue';
import Step3Review from '../components/deploy/Step3Review.vue';
import { api } from '../utils/api';
import type { ApiResponse } from '../utils/api';
import { useNotificationStore } from '../stores/notifications';
import { UserService } from '../utils/userService';
import { useNodeManagement } from '../composables/useNodeManagement';
import type { NormalizedNode } from '../types/normalizedNode';
import { validateClusterName, generateClusterName, getNodeInfo, getSshKeyName } from '../utils/clusterUtils';
import { useRouter } from 'vue-router';
import { normalizeNode } from '../utils/nodeNormalizer';
import type { Cluster, ClusterNode } from '../types/cluster';

const notificationStore = useNotificationStore();
const userService = new UserService();
const router = useRouter();

const step = ref(1);
const { masters, workers, availableSshKeys, addMaster, addWorker, removeMaster, removeWorker } = useDeployCluster();
const deploying = ref(false);

const allVMs = computed(() => [...masters.value, ...workers.value]);

const { rentedNodes, fetchRentedNodes } = useNodeManagement();
const availableNodes = ref<NormalizedNode[]>([]);

async function fetchAvailableNodes() {
  await fetchRentedNodes();
  availableNodes.value = rentedNodes.value.map(normalizeNode);
}

const clusterName = ref('');
const selectedSshKeys = ref<number[]>([]);

function onClusterNameChange(name: string) {
  clusterName.value = name;
}

// --- Step 1 Validation ---
const isStep1Valid = computed(() => {
  if (masters.value.length === 0) return false;
  if (!validateClusterName(clusterName.value).isValid) return false;
  // Every node (master/worker) must have at least one SSH key
  return allVMs.value.every(vm => Array.isArray(vm.sshKeyIds) && vm.sshKeyIds.length > 0);
});

// --- Step 2 Validation ---
const allVMsAssigned = computed(() => allVMs.value.length > 0 && allVMs.value.every((vm: any) => vm.node !== null && vm.node !== undefined));

// Resource validation for each node
const nodeResourceErrors = computed(() => {
  const errors: Record<number, string[]> = {};
  const nodeUsage: Record<number, {cpu: number, ram: number, disk: number, rootfs: number}> = {};
  allVMs.value.forEach(vm => {
    if (vm.node != null) {
      if (!nodeUsage[vm.node]) nodeUsage[vm.node] = {cpu: 0, ram: 0, disk: 0, rootfs: 0};
      nodeUsage[vm.node].cpu = Math.max(nodeUsage[vm.node].cpu, vm.vcpu);
      nodeUsage[vm.node].ram += vm.ram;
      nodeUsage[vm.node].disk += vm.disk;
      nodeUsage[vm.node].disk += vm.rootfs;
    }
  });
  availableNodes.value.forEach(node => {
    const usage = nodeUsage[node.nodeId] || {cpu: 0, ram: 0, disk: 0, rootfs: 0};
    const cpuOk = usage.cpu <= (node.cpu || 0);
    const ramOk = usage.ram <= (node.ram || 0);
    const diskOk = usage.disk <= (node.storage || 0);
    const errs: string[] = [];
    if (!cpuOk) errs.push(`CPU over-allocated (${usage.cpu}/${node.cpu})`);
    if (!ramOk) errs.push(`RAM over-allocated (${usage.ram}/${node.ram})`);
    if (!diskOk) errs.push(`Disk over-allocated (${usage.disk}/${node.storage})`);
    if (errs.length) errors[node.nodeId] = errs;
  });
  return errors;
});
const isStep2Valid = computed(() => allVMsAssigned.value && Object.keys(nodeResourceErrors.value).length === 0);

// --- Deploy Logic ---
const clusterToken = ref('securetoken');

function generateClusterNameLocal() {
  clusterName.value = generateClusterName();
}

// Navigate to SSH keys management
function navigateToSshKeys() {
  localStorage.setItem('dashboard-section', 'ssh')
  router.push('/dashboard');
}

// --- Navigation ---
function nextStep() {
  if ((step.value === 1 && isStep1Valid.value) || (step.value === 2 && isStep2Valid.value)) {
    step.value++;
  }
}
function prevStep() {
  if (step.value > 1) step.value--;
}

const clusterPayload = computed<Cluster>(() => {
  const token = clusterToken.value;

  function buildNode(vm: VM, type: 'master' | 'worker'): ClusterNode {
    const vmSshKeyObj = availableSshKeys.value.find(k => k.ID === vm.sshKeyIds[0]);
    return {
      name: vm.name,
      type: type === 'master' ? 'master' : 'worker',
      node_id: vm.node as number, // node is number | null, but must be number here
      cpu: vm.vcpu,
      memory: vm.ram * 1024, // GB to MB
      root_size: vm.rootfs * 1024, // GB to MB
      disk_size: vm.disk * 1024, // GB to MB
      env_vars: {
        SSH_KEY: vmSshKeyObj ? vmSshKeyObj.public_key : '',
        K3S_TOKEN: token,
      },
    };
  }

  const nodes: ClusterNode[] = [
    ...masters.value.map(vm => buildNode(vm, 'master')),
    ...workers.value.map(vm => buildNode(vm, 'worker')),
  ];

  return {
    name: clusterName.value,
    token: token,
    nodes: nodes,
  };
});

function navigateToDasgboard() {
  localStorage.setItem('dashboard-section', 'clusters')
  router.push('/dashboard');
}

async function onDeployCluster() {
  deploying.value = true;
  try {
    await api.post<ApiResponse<{ task_id: string }>>('/v1/deployments', clusterPayload.value, {
      showNotifications: false,
      loadingMessage: 'Deploying cluster...',
      errorMessage: 'Failed to deploy cluster',
      requiresAuth: true
    });
    notificationStore.info('Deployment started', 'Your cluster is being deployed in the background. You will be notified when it is ready.');
  } catch (err) {
    // Optionally handle error
  } finally {
    deploying.value = false;
    navigateToDasgboard();
  }
}

const sshKeysLoading = ref(true);
async function fetchSshKeys() {
  sshKeysLoading.value = true;
  try {
    const keys = await userService.listSshKeys();
    availableSshKeys.value = keys;
  } catch (err) {
    availableSshKeys.value = [];
    notificationStore.error('Error', 'Failed to load SSH keys');
  } finally {
    sshKeysLoading.value = false;
  }
}

// Initialize component
onMounted(() => {
  // Auto-generate cluster name on component mount
  generateClusterNameLocal();
  fetchAvailableNodes();
  fetchSshKeys();
});

const editNodeModal = ref({ open: false, type: '', idx: -1, node: null as null | VM });
function openEditNodeModal(type: 'master' | 'worker', idx: number) {
  const node = type === 'master' ? { ...masters.value[idx] } : { ...workers.value[idx] };
  editNodeModal.value = { open: true, type, idx, node };
}
function closeEditNodeModal() {
  editNodeModal.value = { open: false, type: '', idx: -1, node: null };
}
function saveEditNode(updatedNode: VM) {
  if (!editNodeModal.value.node) return;
  if (editNodeModal.value.type === 'master') {
    masters.value.splice(editNodeModal.value.idx, 1, { ...updatedNode });
  } else if (editNodeModal.value.type === 'worker') {
    workers.value.splice(editNodeModal.value.idx, 1, { ...updatedNode });
  }
  closeEditNodeModal();
}

function setSelectedSshKeys(keys: number[]) {
  selectedSshKeys.value = keys;
}
function onAssignNode(vmIdx: number, nodeId: number | null) {
  if (vmIdx < masters.value.length) {
    masters.value[vmIdx].node = nodeId != null ? nodeId : null;
  } else {
    const workerIdx = vmIdx - masters.value.length;
    workers.value[workerIdx].node = nodeId != null ? nodeId : null;
  }
}
</script>

<style>
.deploy-container {
  /* Enhanced palette for Deploy Cluster only */
  --color-bg-elevated: #20243a;
  --color-bg-hover: #23263b;
  --color-chip-bg: #23263b;
  --color-chip-border: #334155;
  --shadow-card: 0 6px 24px rgba(16, 24, 40, 0.12);
}
</style>

<style scoped>
.deploy-container {
  min-height: 100vh;
  background: var(--color-bg, #15162b);
  padding-top: 3.5rem;
  margin-top: 7rem;
}
.deploy-header {
  text-align: center;
  margin-bottom: 2.5rem;
}
.hero-title {
  font-size: 2.2rem;
  font-weight: 700;
  color: var(--color-text, #fff);
  margin-bottom: 0.5rem;
}
.section-subtitle {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1.1rem;
}
.deploy-content-wrapper {
  display: flex;
  justify-content: center;
}
.deploy-card {
  background: var(--color-surface-1, #18192b);
  border-radius: 22px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  padding: 3.5rem 3rem 2.5rem 3rem;
  width: 100%;
  max-width: 900px;
  margin-top: 2.5rem;
}
.progress-section {
  margin-bottom: 3rem;
}
.stepper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2.2rem;
  position: relative;
}
.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  position: relative;
  z-index: 1;
}
.step-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-surface-2, #23243a);
  color: var(--color-primary, #6366f1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.1rem;
  margin-bottom: 0.3rem;
  border: 2px solid var(--color-surface-2, #23243a);
  transition: background 0.2s, color 0.2s, border 0.2s;
  position: relative;
  z-index: 2;
}
.step.active .step-circle {
  background: var(--color-primary, #6366f1);
  color: #fff;
  border: 2px solid var(--color-primary, #6366f1);
}
.step.completed .step-circle {
  background: var(--color-success, #22d3ee);
  color: #fff;
  border: 2px solid var(--color-success, #22d3ee);
}

.step-label {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1rem;
  margin-top: 0.2rem;
  text-align: center;
  font-weight: 500;
  letter-spacing: 0.01em;
}
.step.active .step-label {
  color: var(--color-primary, #6366f1);
  font-weight: 600;
}
.step.completed .step-label {
  color: var(--color-success, #22d3ee);
  font-weight: 600;
}
.step:not(:last-child)::after {
  content: '';
  position: absolute;
  top: 18px;
  right: -50%;
  width: 100%;
  height: 4px;
  background: var(--color-surface-2, #23243a);
  z-index: 0;
}
.step.completed:not(:last-child)::after {
  background: var(--color-success, #22d3ee);
}
@media (max-width: 900px) {
  .deploy-card {
    padding: 1.2rem 0.5rem 1.2rem 0.5rem;
  }
  .stepper {
    flex-direction: column;
    gap: 1.2rem;
  }
  .step {
    flex-direction: row;
    align-items: center;
    gap: 0.7rem;
  }
  .step-label {
    margin-top: 0;
    margin-left: 0.7rem;
    text-align: left;
  }
}
</style>
