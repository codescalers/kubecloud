<template>
  <v-container class="fill-height">
    <v-card class="mx-auto deploy-card" width="100%" max-width="1200" elevation="8">
      <v-card-title class="text-h4 mb-2 text-center pa-6 bg-gradient">
        <v-icon icon="mdi-rocket-launch" class="mr-3" size="large"></v-icon>
        Deploy New Cluster
        <v-chip class="ml-3" color="primary" variant="outlined" size="small">
          {{ step }}/3 Steps
        </v-chip>
      </v-card-title>
      
      <!-- Progress Bar -->
      <v-progress-linear 
        :model-value="(step / 3) * 100" 
        color="primary" 
        height="4"
        class="mb-0"
      ></v-progress-linear>

      <v-stepper v-model="step" class="elevation-0">
        <v-stepper-header class="mb-4">
          <v-stepper-item 
            :complete="step > 1" 
            :value="1"
            title="Define VMs"
            @click="step = 1"
          >
            Virtual Machines
          </v-stepper-item>

          <v-divider></v-divider>

          <v-stepper-item 
            :complete="step > 2" 
            :value="2"
            title="Place VMs"
            @click="step = 2"
          >
            Node Placement
          </v-stepper-item>

          <v-divider></v-divider>

          <v-stepper-item 
            :complete="step > 3" 
            :value="3"
            title="Review"
            @click="step = 3"
          >
            Configure & Review
          </v-stepper-item>
        </v-stepper-header>

        <!-- Step 1: Define Your Cluster's Virtual Machines -->
        <v-stepper-window v-model="step">
          <v-stepper-window-item :value="1">
          <v-card>
            <v-card-text>
              <!-- Masters Section -->
              <v-row>
                <v-col cols="12" md="6">
                  <v-card variant="outlined" class="pa-4 mb-4">
                    <v-card-title class="d-flex align-center">
                      <v-icon icon="mdi-server" class="mr-2"></v-icon>
                      Masters
                      <v-spacer></v-spacer>
                      <v-btn
                        color="primary"
                        :disabled="masters.length >= 3"
                        prepend-icon="mdi-plus"
                        size="small"
                        @click="addMaster"
                      >
                        Add Master
                      </v-btn>
                    </v-card-title>
                    <v-list>
                      <v-list-item v-for="(master, index) in masters" :key="index">
                        <template v-slot:prepend>
                          <v-icon icon="mdi-desktop-classic"></v-icon>
                        </template>
                        <v-list-item-title class="font-weight-bold">{{ master.name }}</v-list-item-title>
                        <v-list-item-subtitle>
                          <v-chip color="primary" size="small" class="mr-2">vCPU: {{ master.vcpu }}</v-chip>
                          <v-chip color="primary" size="small">RAM: {{ master.ram }}GB</v-chip>
                        </v-list-item-subtitle>
                        <template v-slot:append>
                          <v-btn icon="mdi-delete" variant="text" color="error" size="small" @click="removeMaster(index)"></v-btn>
                        </template>
                      </v-list-item>
                    </v-list>
                  </v-card>
                </v-col>

                <v-col cols="12" md="6">
                  <v-card variant="outlined" class="pa-4 mb-4">
                    <v-card-title class="d-flex align-center">
                      <v-icon icon="mdi-desktop-tower-monitor" class="mr-2"></v-icon>
                      Workers
                      <v-spacer></v-spacer>
                      <v-btn
                        color="primary"
                        prepend-icon="mdi-plus"
                        size="small"
                        @click="addWorker"
                      >
                        Add Worker
                      </v-btn>
                    </v-card-title>
                    <v-list>
                      <v-list-item v-for="(worker, index) in workers" :key="index">
                        <template v-slot:prepend>
                          <v-icon icon="mdi-desktop-tower"></v-icon>
                        </template>
                        <v-list-item-title class="font-weight-bold">{{ worker.name }}</v-list-item-title>
                        <v-list-item-subtitle>
                          <v-chip color="primary" size="small" class="mr-2">vCPU: {{ worker.vcpu }}</v-chip>
                          <v-chip color="primary" size="small">RAM: {{ worker.ram }}GB</v-chip>
                        </v-list-item-subtitle>
                        <template v-slot:append>
                          <v-btn icon="mdi-delete" variant="text" color="error" size="small" @click="removeWorker(index)"></v-btn>
                        </template>
                      </v-list-item>
                    </v-list>
                  </v-card>
                </v-col>
              </v-row>
            </v-card-text>
            <v-divider></v-divider>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                :disabled="!masters.length"
                @click="nextStep"
              >
                Continue
                <v-icon end icon="mdi-arrow-right"></v-icon>
              </v-btn>
            </v-card-actions>
          </v-card>
          </v-stepper-window-item>

          <!-- Step 2: Place Your VMs on Your Reserved Nodes -->
          <v-stepper-window-item :value="2">
          <v-card>
            <v-card-text>
              <h3 class="text-h6 mb-4">
                <v-icon icon="mdi-server-network" class="mr-2"></v-icon>
                Assign VMs to Reserved Nodes
              </h3>
              <p class="text-body-2 mb-6 text-medium-emphasis">
                Select which reserved nodes will host your cluster VMs. Each VM will be deployed on the selected node.
              </p>
              
              <v-row>
                <v-col cols="12">
                  <div class="vm-assignment-grid">
                    <v-card 
                      v-for="(vm, index) in allVMs" 
                      :key="index"
                      variant="outlined" 
                      class="pa-4 mb-4 vm-card"
                    >
                      <div class="d-flex align-center mb-3">
                        <v-avatar 
                          :color="vm.name.includes('Master') ? 'primary' : 'secondary'" 
                          size="40" 
                          class="mr-3"
                        >
                          <v-icon 
                            :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'" 
                            color="white"
                          ></v-icon>
                        </v-avatar>
                        <div class="flex-grow-1">
                          <h4 class="text-h6 mb-1">{{ vm.name }}</h4>
                          <div class="d-flex gap-2">
                            <v-chip color="primary" size="small" variant="tonal">
                              {{ vm.vcpu }} vCPU
                            </v-chip>
                            <v-chip color="primary" size="small" variant="tonal">
                              {{ vm.ram }}GB RAM
                            </v-chip>
                          </div>
                        </div>
                      </div>
                      
                      <v-select
                        :items="availableNodes"
                        label="Select Reserved Node"
                        v-model="vm.node"
                        item-title="label"
                        item-value="id"
                        prepend-inner-icon="mdi-server-network"
                        variant="outlined"
                        :hint="vm.node ? getNodeInfo(vm.node) : 'Choose a node for this VM'"
                        persistent-hint
                        class="mt-2"
                      >
                        <template v-slot:item="{ props, item }">
                          <v-list-item v-bind="props" :title="item.raw.label">
                            <template v-slot:prepend>
                              <v-icon icon="mdi-server-network" class="mr-2"></v-icon>
                            </template>
                            <v-list-item-title>{{ item.raw.label }}</v-list-item-title>
                            <v-list-item-subtitle>
                              {{ item.raw.totalCPU }} vCPU, {{ item.raw.totalRAM }}GB RAM
                              <v-chip 
                                v-if="item.raw.hasGPU" 
                                color="success" 
                                size="x-small" 
                                class="ml-2"
                              >
                                GPU
                              </v-chip>
                            </v-list-item-subtitle>
                          </v-list-item>
                        </template>
                      </v-select>
                    </v-card>
                  </div>
                </v-col>
              </v-row>
            </v-card-text>
            <v-divider></v-divider>
            <v-card-actions>
              <v-btn text @click="prevStep">
                <v-icon start icon="mdi-arrow-left"></v-icon>
                Back
              </v-btn>
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                @click="nextStep"
                :disabled="!allVMsAssigned"
              >
                Continue
                <v-icon end icon="mdi-arrow-right"></v-icon>
              </v-btn>
            </v-card-actions>
          </v-card>
          </v-stepper-window-item>

          <!-- Step 3: Configure and Review -->
          <v-stepper-window-item :value="3">
          <v-card>
            <v-card-text>
              <!-- Configuration Section -->
              <div class="mb-6">
                <h3 class="text-h5 mb-4 d-flex align-center">
                  <v-icon icon="mdi-cog" class="mr-3" color="primary"></v-icon>
                  Cluster Configuration
                </h3>
                
                <!-- Cluster Name -->
                <v-card variant="outlined" class="pa-4 mb-4">
                  <v-card-title class="text-subtitle-1 pb-2">
                    <v-icon icon="mdi-kubernetes" class="mr-2" color="primary"></v-icon>
                    Cluster Name
                  </v-card-title>
                  <v-text-field
                    v-model="clusterName"
                    label="Enter cluster name"
                    :rules="[v => !!v || 'Cluster name is required']"
                    required
                    variant="outlined"
                    density="compact"
                    hide-details
                    class="mt-2"
                  ></v-text-field>
                </v-card>

                <!-- SSH Keys Selection -->
                <v-card variant="outlined" class="pa-4 mb-4">
                  <v-card-title class="text-subtitle-1 pb-2">
                    <v-icon icon="mdi-key" class="mr-2" color="primary"></v-icon>
                    SSH Keys
                    <v-chip color="error" size="x-small" class="ml-2" v-if="selectedSshKeys.length === 0">Required</v-chip>
                  </v-card-title>
                  <p class="text-body-2 text-medium-emphasis mb-3">
                    Select SSH keys to access your cluster nodes. You can select multiple keys.
                  </p>
                  
                  <v-chip-group
                    v-model="selectedSshKeys"
                    multiple
                    column
                    class="mb-3"
                  >
                    <v-chip
                      v-for="key in availableSshKeys"
                      :key="key.id"
                      :value="key.id"
                      filter
                      variant="outlined"
                      class="ssh-key-chip"
                    >
                      <template v-slot:prepend>
                        <v-icon icon="mdi-key-variant" size="small"></v-icon>
                      </template>
                      {{ key.name }}
                      <template v-slot:append>
                        <v-tooltip activator="parent" location="top">
                          <div class="text-caption">
                            <strong>{{ key.name }}</strong><br>
                            Fingerprint: {{ key.fingerprint }}<br>
                            Added: {{ key.createdAt }}
                          </div>
                        </v-tooltip>
                      </template>
                    </v-chip>
                  </v-chip-group>
                  
                  <div v-if="selectedSshKeys.length === 0" class="text-body-2 text-error mb-2">
                    <v-icon icon="mdi-alert" size="small" class="mr-1"></v-icon>
                    Please select at least one SSH key
                  </div>
                  
                  <v-btn
                    variant="text"
                    color="primary"
                    prepend-icon="mdi-plus"
                    size="small"
                    @click="navigateToSshKeys"
                  >
                    Add New SSH Key
                  </v-btn>
                </v-card>

                <!-- QSFS Configuration -->
                <v-card variant="outlined" class="pa-4">
                  <v-card-title class="text-subtitle-1 pb-2">
                    <v-icon icon="mdi-folder-network" class="mr-2" color="primary"></v-icon>
                    QSFS Configuration (Optional)
                  </v-card-title>
                  <p class="text-body-2 text-medium-emphasis mb-3">
                    Quantum Safe File System configuration for distributed storage.
                  </p>
                  <v-text-field
                    v-model="qsfsConfig"
                    label="QSFS configuration"
                    variant="outlined"
                    density="compact"
                    hide-details
                    placeholder="Enter QSFS configuration..."
                  ></v-text-field>
                </v-card>
              </div>

              <v-divider class="my-6"></v-divider>

              <!-- Review Section -->
              <div>
                <h3 class="text-h5 mb-4 d-flex align-center">
                  <v-icon icon="mdi-clipboard-check" class="mr-3" color="success"></v-icon>
                  Deployment Summary
                </h3>
                
                <v-row>
                  <!-- Cluster Details -->
                  <v-col cols="12" md="6">
                    <v-card variant="outlined" class="pa-4 h-100">
                      <v-card-title class="text-subtitle-1 pb-2">
                        <v-icon icon="mdi-kubernetes" class="mr-2" color="primary"></v-icon>
                        Cluster Details
                      </v-card-title>
                      <div class="space-y-3">
                        <div class="d-flex justify-space-between align-center">
                          <span class="text-body-2 text-medium-emphasis">Name:</span>
                          <span class="font-weight-medium">{{ clusterName || 'Not set' }}</span>
                        </div>
                        <div class="d-flex justify-space-between align-start">
                          <span class="text-body-2 text-medium-emphasis">SSH Keys:</span>
                          <div class="text-right">
                            <v-chip
                              v-for="keyId in selectedSshKeys"
                              :key="keyId"
                              size="x-small"
                              color="primary"
                              variant="tonal"
                              class="mb-1"
                            >
                              {{ getSshKeyName(keyId) }}
                            </v-chip>
                            <div v-if="selectedSshKeys.length === 0" class="text-medium-emphasis text-caption">
                              None selected
                            </div>
                          </div>
                        </div>
                        <div v-if="qsfsConfig" class="d-flex justify-space-between align-center">
                          <span class="text-body-2 text-medium-emphasis">QSFS:</span>
                          <span class="font-weight-medium text-truncate">{{ qsfsConfig }}</span>
                        </div>
                      </div>
                    </v-card>
                  </v-col>

                  <!-- VM Assignments -->
                  <v-col cols="12" md="6">
                    <v-card variant="outlined" class="pa-4 h-100">
                      <v-card-title class="text-subtitle-1 pb-2">
                        <v-icon icon="mdi-server-network" class="mr-2" color="primary"></v-icon>
                        VM Assignments
                        <v-chip 
                          :color="allVMsAssigned ? 'success' : 'warning'" 
                          size="x-small" 
                          class="ml-2"
                        >
                          {{ allVMsAssigned ? 'Complete' : 'Incomplete' }}
                        </v-chip>
                      </v-card-title>
                      <div class="space-y-2">
                        <div 
                          v-for="(vm, index) in allVMs" 
                          :key="index" 
                          class="d-flex align-center py-2 px-3 rounded"
                          :class="vm.node ? 'bg-success-subtle' : 'bg-warning-subtle'"
                        >
                          <v-icon 
                            :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'"
                            :color="vm.name.includes('Master') ? 'primary' : 'secondary'"
                            size="small"
                            class="mr-3"
                          ></v-icon>
                          <div class="flex-grow-1">
                            <div class="font-weight-medium text-body-2">{{ vm.name }}</div>
                            <div class="d-flex align-center mt-1">
                              <v-chip color="primary" size="x-small" class="mr-1">{{ vm.vcpu }}c</v-chip>
                              <v-chip color="primary" size="x-small" class="mr-2">{{ vm.ram }}GB</v-chip>
                              <span class="text-caption text-medium-emphasis">
                                → Node #{{ vm.node || 'Unassigned' }}
                              </span>
                            </div>
                          </div>
                          <v-icon 
                            :icon="vm.node ? 'mdi-check-circle' : 'mdi-alert'"
                            :color="vm.node ? 'success' : 'warning'"
                            size="small"
                          ></v-icon>
                        </div>
                      </div>
                    </v-card>
                  </v-col>
                </v-row>

                <!-- Validation Status -->
                <v-alert
                  v-if="!isFormValid"
                  type="warning"
                  variant="tonal"
                  class="mt-4"
                  :text="getValidationMessage()"
                ></v-alert>
              </div>
            </v-card-text>
            <v-divider></v-divider>
            <v-card-actions>
              <v-btn text @click="prevStep">
                <v-icon start icon="mdi-arrow-left"></v-icon>
                Back
              </v-btn>
              <v-spacer></v-spacer>
              <v-btn
                color="success"
                @click="deployCluster"
                :loading="isDeploying"
                :disabled="!isFormValid"
                size="large"
              >
                <v-icon start icon="mdi-cloud-upload"></v-icon>
                Deploy Cluster
              </v-btn>
            </v-card-actions>
          </v-card>
          </v-stepper-window-item>
        </v-stepper-window>
      </v-stepper>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';

// Type definitions
interface VM {
  name: string;
  vcpu: number;
  ram: number;
  node: number | null;
}

interface Node {
  id: number;
  label: string;
  totalCPU: number;
  totalRAM: number;
  hasGPU: boolean;
  location: string;
}

interface SshKey {
  id: number;
  name: string;
  fingerprint: string;
  createdAt: string;
}

const step = ref(1);
const masters = ref<VM[]>([]);
const workers = ref<VM[]>([]);
const isDeploying = ref(false);

const allVMs = computed(() => [...masters.value, ...workers.value]);

// Available nodes with better structure
const availableNodes = ref<Node[]>([
  { 
    id: 1234, 
    label: 'Node #1234 (GPU)', 
    totalCPU: 16, 
    totalRAM: 64, 
    hasGPU: true,
    location: 'us-east-1a'
  },
  { 
    id: 5678, 
    label: 'Node #5678', 
    totalCPU: 8, 
    totalRAM: 32, 
    hasGPU: false,
    location: 'us-east-1b'
  },
  { 
    id: 9012, 
    label: 'Node #9012', 
    totalCPU: 8, 
    totalRAM: 32, 
    hasGPU: false,
    location: 'us-east-1c'
  },
]);

const clusterName = ref('');
const selectedSshKeys = ref<number[]>([]);
const qsfsConfig = ref('');

// Available SSH keys (would come from dashboard/API)
const availableSshKeys = ref<SshKey[]>([
  {
    id: 1,
    name: 'my-laptop-key',
    fingerprint: 'SHA256:abc123...',
    createdAt: '2024-01-15'
  },
  {
    id: 2,
    name: 'production-key',
    fingerprint: 'SHA256:def456...',
    createdAt: '2024-01-10'
  },
  {
    id: 3,
    name: 'team-shared-key',
    fingerprint: 'SHA256:ghi789...',
    createdAt: '2024-01-05'
  }
]);

// Cluster name generator words
const adjectives = [
  'swift', 'bright', 'cosmic', 'quantum', 'stellar', 'azure', 'crimson', 'golden',
  'silver', 'emerald', 'sapphire', 'crystal', 'thunder', 'lightning', 'storm',
  'ocean', 'mountain', 'forest', 'desert', 'arctic', 'tropical', 'mystic'
];

const nouns = [
  'cluster', 'cloud', 'node', 'server', 'engine', 'core', 'hub', 'nexus',
  'forge', 'vault', 'tower', 'citadel', 'fortress', 'sanctuary', 'haven',
  'realm', 'domain', 'sphere', 'matrix', 'grid', 'network', 'system'
];

// Computed properties for validation
const allVMsAssigned = computed(() => {
  return allVMs.value.every(vm => vm.node);
});

const isFormValid = computed(() => {
  return clusterName.value && selectedSshKeys.value.length > 0 && allVMsAssigned.value;
});

// Helper function to get node info
function getNodeInfo(nodeId: number) {
  const node = availableNodes.value.find(n => n.id === nodeId);
  if (!node) return '';
  return `${node.totalCPU} vCPU, ${node.totalRAM}GB RAM${node.hasGPU ? ', GPU Available' : ''}`;
}

function addMaster() {
  if (masters.value.length < 3) {
    masters.value.push({
      name: `Master-${masters.value.length + 1}`,
      vcpu: 2,
      ram: 4,
      node: null
    });
  }
}

function removeMaster(index: number) {
  masters.value.splice(index, 1);
}

function addWorker() {
  workers.value.push({
    name: `Worker-${workers.value.length + 1}`,
    vcpu: 2,
    ram: 4,
    node: null
  });
}

function removeWorker(index: number) {
  workers.value.splice(index, 1);
}

// Cluster name generator
function generateClusterName() {
  const randomAdjective = adjectives[Math.floor(Math.random() * adjectives.length)];
  const randomNoun = nouns[Math.floor(Math.random() * nouns.length)];
  const randomNumber = Math.floor(Math.random() * 999) + 1;
  clusterName.value = `${randomAdjective}-${randomNoun}-${randomNumber}`;
}

// Navigate to SSH keys management
function navigateToSshKeys() {
  // This would navigate to the SSH keys page in the dashboard
  // For now, we'll just show an alert
  alert('This would navigate to the SSH Keys management page in the dashboard');
}

// Get SSH key name by ID
function getSshKeyName(keyId: number) {
  const key = availableSshKeys.value.find(k => k.id === keyId);
  return key ? key.name : 'Unknown';
}

// Get validation message for form errors
function getValidationMessage() {
  const errors = [];
  
  if (!clusterName.value) {
    errors.push('Cluster name is required');
  }
  
  if (selectedSshKeys.value.length === 0) {
    errors.push('At least one SSH key must be selected');
  }
  
  if (!allVMsAssigned.value) {
    errors.push('All VMs must be assigned to nodes');
  }
  
  return errors.join('. ');
}

function nextStep() {
  if (step.value < 3) {
    step.value++;
  }
}

function prevStep() {
  if (step.value > 1) {
    step.value--;
  }
}

async function deployCluster() {
  if (!isFormValid.value) return;

  isDeploying.value = true;
  try {
    // Simulating API call
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    const selectedKeys = selectedSshKeys.value.map(keyId => 
      availableSshKeys.value.find(k => k.id === keyId)
    );
    
    console.log('Deploying cluster with configuration:', {
      name: clusterName.value,
      vms: allVMs.value,
      sshKeys: selectedKeys,
      qsfsConfig: qsfsConfig.value
    });
    
    // Show success message
    alert(`Cluster "${clusterName.value}" deployment initiated successfully!`);
  } catch (error) {
    console.error('Failed to deploy cluster:', error);
    alert('Failed to deploy cluster. Please try again.');
  } finally {
    isDeploying.value = false;
  }
}

// Initialize component
onMounted(() => {
  // Auto-generate cluster name on component mount
  generateClusterName();
});
</script>

<style scoped>
.fill-height {
  min-height: calc(100vh - 64px);
  padding: 2rem;
  background: var(--v-theme-background, #0F172A);
}

.deploy-card {
  border-radius: 18px !important;
  overflow: hidden;
  box-shadow: 0 6px 32px rgba(60,60,60,0.10) !important;
  backdrop-filter: blur(10px);
  background: var(--v-theme-surface, #1E293B) !important;
  border: 1.5px solid var(--v-theme-outline, #e0e7ef);
}

.bg-gradient {
  background: linear-gradient(135deg, #3B82F6 0%, #60A5FA 100%);
  color: white !important;
  position: relative;
  overflow: hidden;
}

.bg-gradient::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(45deg, rgba(255,255,255,0.1) 0%, transparent 100%);
  pointer-events: none;
}

.v-stepper {
  background: transparent !important;
  padding: 2rem;
}

.v-stepper-header {
  box-shadow: none !important;
  background: var(--v-theme-surface, #1E293B) !important;
  border: 1px solid rgba(59, 130, 246, 0.2) !important;
  border-radius: 12px !important;
  padding: 1rem !important;
  backdrop-filter: blur(10px);
}

.v-stepper-item {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.v-stepper-item:hover {
  transform: translateY(-2px);
}

.v-stepper-window {
  margin-top: 2rem;
}

.v-stepper-window-item .v-card {
  border-radius: 18px !important;
  border: none !important;
  background: var(--v-theme-surface, #1E293B) !important;
  backdrop-filter: blur(10px);
  box-shadow: 0 6px 32px rgba(60,60,60,0.10) !important;
  overflow: hidden;
}

.v-card-text {
  padding: 2rem !important;
}

.v-card-actions {
  padding: 1.5rem 2rem !important;
  background: rgba(15, 23, 42, 0.5);
  backdrop-filter: blur(10px);
}

/* Enhanced VM Cards */
.v-card[variant="outlined"] {
  border: 1.5px solid rgba(59, 130, 246, 0.2) !important;
  border-radius: 18px !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: var(--v-theme-surface, #1E293B) !important;
  backdrop-filter: blur(5px);
}

.v-card[variant="outlined"]:hover {
  border-color: rgba(59, 130, 246, 0.4) !important;
  transform: translateY(-2px) scale(1.005);
  box-shadow: 0 0 15px rgba(59, 130, 246, 0.2), 0 0 30px rgba(59, 130, 246, 0.1) !important;
}

.v-card-title {
  font-weight: 600 !important;
  color: var(--v-theme-on-surface, #F1F5F9) !important;
  font-size: 1.125rem !important;
}

/* Enhanced List Items */
.v-list {
  background: transparent !important;
}

.v-list-item {
  margin-bottom: 0.75rem !important;
  border-radius: 8px !important;
  transition: all 0.2s ease;
  background: rgba(30, 41, 59, 0.6) !important;
  border: 1px solid rgba(59, 130, 246, 0.1);
  overflow: hidden !important;
}

.v-list-item:hover {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: rgba(59, 130, 246, 0.3);
  transform: translateY(-2px);
}

.v-list-item-title {
  font-weight: 600 !important;
  color: var(--v-theme-on-surface, #F1F5F9) !important;
}

.v-list-item-subtitle {
  color: rgba(241, 245, 249, 0.7) !important;
  margin-top: 0.5rem !important;
}

/* Enhanced Chips */
.v-chip {
  font-weight: 500 !important;
  border-radius: 6px !important;
  background: rgba(59, 130, 246, 0.15) !important;
  color: #FFFFFF !important;
  border: 1px solid rgba(59, 130, 246, 0.3) !important;
  transition: all 0.2s ease !important;
}

.v-chip:hover {
  background: rgba(59, 130, 246, 0.25) !important;
  border-color: rgba(59, 130, 246, 0.5) !important;
  transform: translateY(-1px) !important;
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.2) !important;
}

.v-chip[color="primary"] {
  background: rgba(59, 130, 246, 0.2) !important;
  color: #FFFFFF !important;
  border: 1px solid rgba(59, 130, 246, 0.4) !important;
}

.v-chip[color="primary"]:hover {
  background: rgba(59, 130, 246, 0.3) !important;
  border-color: rgba(59, 130, 246, 0.6) !important;
}

.v-chip[color="error"] {
  background: rgba(239, 68, 68, 0.2) !important;
  color: #FFFFFF !important;
  border: 1px solid rgba(239, 68, 68, 0.4) !important;
}

.v-chip[color="error"]:hover {
  background: rgba(239, 68, 68, 0.3) !important;
  border-color: rgba(239, 68, 68, 0.6) !important;
}

.v-chip[variant="outlined"] {
  background: transparent !important;
  color: #3B82F6 !important;
  border: 1px solid rgba(59, 130, 246, 0.5) !important;
}

.v-chip[variant="outlined"]:hover {
  background: rgba(59, 130, 246, 0.1) !important;
  color: #FFFFFF !important;
}

.v-chip[variant="tonal"] {
  background: rgba(59, 130, 246, 0.15) !important;
  color: #FFFFFF !important;
  border: 1px solid rgba(59, 130, 246, 0.3) !important;
}

.v-chip[variant="tonal"]:hover {
  background: rgba(59, 130, 246, 0.25) !important;
  border-color: rgba(59, 130, 246, 0.5) !important;
}

/* Enhanced Buttons */
.v-btn {
  border-radius: 8px !important;
  font-weight: 600 !important;
  text-transform: none !important;
  letter-spacing: 0.025em !important;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.v-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.v-btn[color="primary"] {
  background: linear-gradient(135deg, #3B82F6 0%, #60A5FA 100%) !important;
}

.v-btn[color="success"] {
  background: linear-gradient(135deg, #10B981 0%, #059669 100%) !important;
}

/* Enhanced Progress Bars */
.v-progress-linear {
  border-radius: 6px !important;
  overflow: hidden;
}

.v-progress-linear .v-progress-linear__background {
  background: rgba(226, 232, 240, 0.8) !important;
}

/* Enhanced Form Fields */
.v-text-field, .v-textarea, .v-select {
  margin-bottom: 1.5rem !important;
}

.v-field {
  border-radius: 8px !important;
  background: var(--v-theme-surface, #1E293B) !important;
  backdrop-filter: blur(5px);
}

.v-field--focused {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3) !important;
}

/* Step Headers Enhancement */
.text-h6 {
  color: var(--v-theme-on-surface, #F1F5F9) !important;
  font-weight: 600 !important;
  margin-bottom: 1.5rem !important;
}

/* Animations */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.v-stepper-window-item {
  animation: fadeInUp 0.4s ease-out;
}

/* VM Assignment Cards */
.vm-assignment-grid {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.vm-card {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.vm-card:hover {
  transform: translateY(-2px);
  border-color: rgba(59, 130, 246, 0.4) !important;
}

.v-avatar {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.text-medium-emphasis {
  opacity: 0.7;
}

/* SSH Key Chips */
.ssh-keys-container {
  border: 1px solid rgba(59, 130, 246, 0.1);
  border-radius: 8px;
  padding: 1rem;
  background: rgba(59, 130, 246, 0.02);
}

.ssh-key-chip {
  margin: 0.25rem;
  transition: all 0.2s ease;
}

.ssh-key-chip:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.ssh-key-chip.v-chip--selected {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: rgba(59, 130, 246, 0.3) !important;
}

/* Responsive Design */
@media (max-width: 960px) {
  .fill-height {
    padding: 1rem;
  }
  
  .v-stepper {
    padding: 1rem;
  }
  
  .v-card-text {
    padding: 1.5rem !important;
  }
  
  .v-card-actions {
    padding: 1rem 1.5rem !important;
  }
}

@media (max-width: 600px) {
  .fill-height {
    padding: 0.5rem;
  }
  
  .v-stepper {
    padding: 0.5rem;
  }
  
  .v-card-text {
    padding: 1rem !important;
  }
  
  .v-card-actions {
    padding: 1rem !important;
  }
  
  .bg-gradient {
    padding: 1rem !important;
  }
  
  .v-stepper-header {
    padding: 0.75rem !important;
  }
}

/* Light mode overrides for better contrast */
@media (prefers-color-scheme: light) {
  .fill-height {
    background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  }
  
  .deploy-card {
    background: rgba(255, 255, 255, 0.95) !important;
    border: 1.5px solid rgba(226, 232, 240, 0.8);
  }
  
  .v-stepper-header {
    background: rgba(255, 255, 255, 0.9) !important;
  }
  
  .v-stepper-window-item .v-card {
    background: rgba(255, 255, 255, 0.95) !important;
  }
  
  .v-card[variant="outlined"] {
    background: rgba(255, 255, 255, 0.9) !important;
  }
  
  .v-list-item {
    background: rgba(255, 255, 255, 0.8) !important;
  }
  
  .v-card-title {
    color: #1e293b !important;
  }
  
  .v-list-item-title {
    color: #1e293b !important;
  }
  
  .v-list-item-subtitle {
    color: #64748b !important;
  }
  
  .text-h6 {
    color: #1e293b !important;
  }
  
  .v-field {
    background: rgba(255, 255, 255, 0.9) !important;
  }
  
  .v-card-actions {
    background: rgba(248, 250, 252, 0.8);
  }
}
</style>
