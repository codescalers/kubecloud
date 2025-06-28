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
            <div class="progress-header">
              <div class="progress-step" :class="{ active: step >= 1, completed: step > 1 }">
                <div class="step-number">1</div>
                <div class="step-label">Define VMs</div>
              </div>
              <div class="progress-line" :class="{ completed: step > 1 }"></div>
              <div class="progress-step" :class="{ active: step >= 2, completed: step > 2 }">
                <div class="step-number">2</div>
                <div class="step-label">Place VMs</div>
              </div>
              <div class="progress-line" :class="{ completed: step > 2 }"></div>
              <div class="progress-step" :class="{ active: step >= 3 }">
                <div class="step-number">3</div>
                <div class="step-label">Review</div>
              </div>
            </div>
          </div>

          <!-- Step Content -->
          <div class="step-content">
            <!-- Step 1: Define Your Cluster's Virtual Machines -->
            <div v-if="step === 1" class="step-section">
              <div class="section-header">
                <h3 class="section-title">
                  <v-icon icon="mdi-server" class="mr-2"></v-icon>
                  Define Virtual Machines
                </h3>
                <p class="section-subtitle">Configure the compute resources for your cluster</p>
              </div>
              
              <div class="vm-config-grid">
                <div class="vm-config-card">
                  <div class="card-header">
                    <h4 class="card-title">
                      <v-icon icon="mdi-server" class="mr-2"></v-icon>
                      Master Nodes
                    </h4>
                    <v-btn
                      color="primary"
                      :disabled="masters.length >= 3"
                      prepend-icon="mdi-plus"
                      size="small"
                      variant="outlined"
                      @click="addMaster"
                    >
                      Add Master
                    </v-btn>
                  </div>
                  <div class="vm-list">
                    <div v-for="(master, index) in masters" :key="index" class="vm-item">
                      <div class="vm-icon" data-type="master">
                        <v-icon icon="mdi-desktop-classic" color="var(--color-primary)"></v-icon>
                      </div>
                      <div class="vm-details">
                        <div class="vm-name">{{ master.name }}</div>
                        <div class="vm-specs">
                          <span class="spec-chip">{{ master.vcpu }} vCPU</span>
                          <span class="spec-chip">{{ master.ram }}GB RAM</span>
                        </div>
                      </div>
                      <v-btn 
                        icon="mdi-delete" 
                        variant="text" 
                        color="error" 
                        size="small" 
                        @click="removeMaster(index)"
                      ></v-btn>
                    </div>
                    <div v-if="!masters.length" class="empty-state">
                      <v-icon icon="mdi-plus-circle-outline" size="32" color="var(--color-text-muted)"></v-icon>
                      <p>No master nodes configured</p>
                    </div>
                  </div>
                </div>

                <div class="vm-config-card">
                  <div class="card-header">
                    <h4 class="card-title">
                      <v-icon icon="mdi-desktop-tower-monitor" class="mr-2"></v-icon>
                      Worker Nodes
                    </h4>
                    <v-btn
                      color="primary"
                      prepend-icon="mdi-plus"
                      size="small"
                      variant="outlined"
                      @click="addWorker"
                    >
                      Add Worker
                    </v-btn>
                  </div>
                  <div class="vm-list">
                    <div v-for="(worker, index) in workers" :key="index" class="vm-item">
                      <div class="vm-icon" data-type="worker">
                        <v-icon icon="mdi-desktop-tower" color="var(--color-success)"></v-icon>
                      </div>
                      <div class="vm-details">
                        <div class="vm-name">{{ worker.name }}</div>
                        <div class="vm-specs">
                          <span class="spec-chip">{{ worker.vcpu }} vCPU</span>
                          <span class="spec-chip">{{ worker.ram }}GB RAM</span>
                        </div>
                      </div>
                      <v-btn 
                        icon="mdi-delete" 
                        variant="text" 
                        color="error" 
                        size="small" 
                        @click="removeWorker(index)"
                      ></v-btn>
                    </div>
                    <div v-if="!workers.length" class="empty-state">
                      <v-icon icon="mdi-plus-circle-outline" size="32" color="var(--color-text-muted)"></v-icon>
                      <p>No worker nodes configured</p>
                    </div>
                  </div>
                </div>
              </div>

              <div class="step-actions">
                <v-btn
                  color="primary"
                  :disabled="!masters.length"
                  @click="nextStep"
                  class="btn-primary"
                >
                  Continue
                  <v-icon end icon="mdi-arrow-right"></v-icon>
                </v-btn>
              </div>
            </div>

            <!-- Step 2: Place Your VMs on Your Reserved Nodes -->
            <div v-if="step === 2" class="step-section">
              <div class="section-header">
                <h3 class="section-title">
                  <v-icon icon="mdi-server-network" class="mr-2"></v-icon>
                  Assign VMs to Reserved Nodes
                </h3>
                <p class="section-subtitle">Select which reserved nodes will host your cluster VMs</p>
              </div>
              
              <div class="vm-assignment-grid">
                <div 
                  v-for="(vm, index) in allVMs" 
                  :key="index"
                  class="vm-assignment-card"
                >
                  <div class="vm-assignment-header">
                    <div class="vm-avatar" :class="vm.name.includes('Master') ? 'master' : 'worker'">
                      <v-icon 
                        :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'" 
                        color="white"
                      ></v-icon>
                    </div>
                    <div class="vm-info">
                      <h4 class="vm-title">{{ vm.name }}</h4>
                      <div class="vm-specs">
                        <span class="spec-chip">{{ vm.vcpu }} vCPU</span>
                        <span class="spec-chip">{{ vm.ram }}GB RAM</span>
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
                    class="node-select"
                  ></v-select>
                </div>
              </div>

              <div class="step-actions">
                <v-btn variant="outlined" @click="prevStep" class="btn-outline">
                  <v-icon start icon="mdi-arrow-left"></v-icon>
                  Back
                </v-btn>
                <v-btn
                  color="primary"
                  :disabled="!allVMs.every(vm => vm.node)"
                  @click="nextStep"
                  class="btn-primary"
                >
                  Continue
                  <v-icon end icon="mdi-arrow-right"></v-icon>
                </v-btn>
              </div>
            </div>

            <!-- Step 3: Review and Deploy -->
            <div v-if="step === 3" class="step-section">
              <div class="section-header">
                <h3 class="section-title">
                  <v-icon icon="mdi-check-circle" class="mr-2"></v-icon>
                  Review Configuration
                </h3>
                <p class="section-subtitle">Review your cluster configuration before deployment</p>
              </div>
              
              <div class="review-grid">
                <div class="review-card">
                  <h4 class="review-title">Cluster Summary</h4>
                  <div class="review-stats">
                    <div class="review-stat">
                      <span class="stat-label">Master Nodes:</span>
                      <span class="stat-value">{{ masters.length }}</span>
                    </div>
                    <div class="review-stat">
                      <span class="stat-label">Worker Nodes:</span>
                      <span class="stat-value">{{ workers.length }}</span>
                    </div>
                    <div class="review-stat">
                      <span class="stat-label">Total vCPU:</span>
                      <span class="stat-value">{{ totalVcpu }}</span>
                    </div>
                    <div class="review-stat">
                      <span class="stat-label">Total RAM:</span>
                      <span class="stat-value">{{ totalRam }}GB</span>
                    </div>
                  </div>
                </div>

                <div class="review-card">
                  <h4 class="review-title">Node Assignment</h4>
                  <div class="node-assignments">
                    <div v-for="(vm, index) in allVMs" :key="index" class="node-assignment">
                      <div class="assignment-icon">
                        <v-icon 
                          :icon="vm.name.includes('Master') ? 'mdi-server' : 'mdi-desktop-tower'" 
                          :color="vm.name.includes('Master') ? 'var(--color-primary)' : 'var(--color-success)'"
                        ></v-icon>
                      </div>
                      <div class="assignment-details">
                        <div class="assignment-name">{{ vm.name }}</div>
                        <div class="assignment-node">{{ getNodeInfo(vm.node) }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="step-actions">
                <v-btn variant="outlined" @click="prevStep" class="btn-outline">
                  <v-icon start icon="mdi-arrow-left"></v-icon>
                  Back
                </v-btn>
                <v-btn
                  color="success"
                  @click="deployCluster"
                  class="btn-primary"
                  :loading="deploying"
                >
                  <v-icon start icon="mdi-rocket-launch"></v-icon>
                  Deploy Cluster
                </v-btn>
              </div>
            </div>
          </div>
        </div>
      </div>
    </v-container>
  </div>
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
const deploying = ref(false);

const allVMs = computed(() => [...masters.value, ...workers.value]);

// Computed properties for totals
const totalVcpu = computed(() => {
  return allVMs.value.reduce((total, vm) => total + vm.vcpu, 0);
});

const totalRam = computed(() => {
  return allVMs.value.reduce((total, vm) => total + vm.ram, 0);
});

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
function getNodeInfo(nodeId: number | null) {
  if (!nodeId) return '';
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

  deploying.value = true;
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
    deploying.value = false;
  }
}

// Initialize component
onMounted(() => {
  // Auto-generate cluster name on component mount
  generateClusterName();
});
</script>

<style scoped>
/* Deploy Container - matches dashboard layout */
.deploy-container {
  min-height: 100vh;
  background: var(--color-bg);
  color: var(--color-text);
  font-family: var(--font-family);
}

.deploy-header {
  text-align: center;
  max-width: 900px;
  margin: 7rem auto 3rem auto;
  position: relative;
  z-index: 2;
  padding: 0 1rem;
}

.hero-title {
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  margin-bottom: 1.5rem;
  line-height: 1.1;
  letter-spacing: -1px;
  color: var(--color-text);
}

.section-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-secondary);
  line-height: 1.5;
  opacity: 0.92;
  margin-bottom: 0;
  font-weight: var(--font-weight-normal);
}

.deploy-content-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
  position: relative;
  z-index: 2;
  margin-top: 4rem;
}

.deploy-card {
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: var(--space-8) var(--space-8) 0 var(--space-8);
  transition: all var(--transition-normal);
  position: relative;
  backdrop-filter: blur(8px);
  box-shadow: var(--shadow-lg);
}

.deploy-card:hover {
  border-color: var(--color-border-light);
}

/* Progress Section */
.progress-section {
  margin-bottom: var(--space-8);
  padding: var(--space-6);
  background: rgba(15, 30, 52, 0.5);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
}

.progress-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.progress-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  flex: 1;
  position: relative;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: transparent;
  border: 2px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
  transition: all var(--transition-normal);
}

.progress-step.active .step-number {
  background: transparent;
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.progress-step.completed .step-number {
  background: transparent;
  border-color: var(--color-success);
  color: var(--color-success);
}

.step-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-muted);
  text-align: center;
}

.progress-step.active .step-label {
  color: var(--color-primary);
}

.progress-step.completed .step-label {
  color: var(--color-success);
}

.progress-line {
  flex: 1;
  height: 2px;
  background: var(--color-border);
  transition: all var(--transition-normal);
}

.progress-line.completed {
  background: var(--color-success);
}

/* Step Content */
.step-content {
  padding: var(--space-8);
}

.step-section {
  margin-bottom: var(--space-8);
}

.section-header {
  margin-bottom: var(--space-6);
  text-align: center;
}

.section-title {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin-bottom: var(--space-2);
}

/* VM Configuration Grid */
.vm-config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--space-6);
  margin-bottom: var(--space-8);
}

.vm-config-card {
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  transition: all var(--transition-normal);
  backdrop-filter: blur(8px);
}

.vm-config-card:hover {
  border-color: var(--color-border-light);
  background: rgba(15, 30, 52, 0.75);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border);
}

.card-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin: 0;
}

.vm-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.vm-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  background: rgba(15, 30, 52, 0.5);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: all var(--transition-normal);
}

.vm-item:hover {
  border-color: var(--color-primary);
  background: rgba(59, 130, 246, 0.1);
}

.vm-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid var(--color-primary);
}

.vm-icon[data-type="master"] {
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid var(--color-primary);
}

.vm-icon[data-type="worker"] {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid var(--color-success);
}

.vm-details {
  flex: 1;
}

.vm-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin-bottom: var(--space-1);
}

.vm-specs {
  display: flex;
  gap: var(--space-2);
}

.spec-chip {
  background: rgba(59, 130, 246, 0.1);
  color: var(--color-primary);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  border: 1px solid var(--color-primary);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-8);
  color: var(--color-text-muted);
  text-align: center;
}

.empty-state p {
  margin: 0;
  font-size: var(--font-size-sm);
}

/* VM Assignment Grid */
.vm-assignment-grid {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  margin-bottom: var(--space-8);
}

.vm-assignment-card {
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  transition: all var(--transition-normal);
  backdrop-filter: blur(8px);
}

.vm-assignment-card:hover {
  border-color: var(--color-border-light);
  background: rgba(15, 30, 52, 0.75);
}

.vm-assignment-header {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.vm-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.vm-avatar.master {
  background: var(--color-primary);
}

.vm-avatar.worker {
  background: var(--color-secondary);
}

.vm-info {
  flex: 1;
}

.vm-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin-bottom: var(--space-2);
}

.node-select {
  margin-top: var(--space-4);
}

/* Review Grid */
.review-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--space-6);
  margin-bottom: var(--space-8);
}

.review-card {
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  transition: all var(--transition-normal);
  backdrop-filter: blur(8px);
}

.review-card:hover {
  border-color: var(--color-border-light);
  background: rgba(15, 30, 52, 0.75);
}

.review-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--color-border);
}

.review-stats {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.review-stat {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) 0;
}

.stat-label {
  color: var(--color-text-secondary);
  font-weight: var(--font-weight-medium);
}

.stat-value {
  color: var(--color-text);
  font-weight: var(--font-weight-semibold);
}

.node-assignments {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.node-assignment {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  background: rgba(15, 30, 52, 0.5);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.assignment-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: rgba(59, 130, 246, 0.1);
}

.assignment-details {
  flex: 1;
}

.assignment-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin-bottom: var(--space-1);
}

.assignment-node {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

/* Step Actions */
.step-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) 0;
}

/* Responsive Design */
@media (max-width: 960px) {
  .deploy-header {
    margin: 4rem auto 2rem auto;
  }
  
  .deploy-content-wrapper {
    margin-top: 2rem;
  }
  
  .vm-config-grid {
    grid-template-columns: 1fr;
  }
  
  .review-grid {
    grid-template-columns: 1fr;
  }
  
  .step-actions {
    flex-direction: column;
    gap: var(--space-4);
  }
}

@media (max-width: 600px) {
  .deploy-header {
    margin: 3rem auto 1.5rem auto;
    padding: 0 var(--space-4);
  }
  
  .hero-title {
    font-size: var(--font-size-3xl);
  }
  
  .deploy-content-wrapper {
    padding: 0 var(--space-4);
    margin-top: 1.5rem;
  }
  
  .deploy-card {
    padding: var(--space-4);
  }
  
  .progress-section {
    padding: var(--space-4);
  }
  
  .step-content {
    padding: var(--space-4);
  }
  
  .vm-config-card,
  .vm-assignment-card,
  .review-card {
    padding: var(--space-4);
  }
  
  .step-actions {
    padding: var(--space-4);
  }
}
</style>
