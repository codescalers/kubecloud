<template>
  <div class="deploy-container">
    <v-container fluid class="pa-0">
      <!-- Header -->
      <div class="deploy-header">
        <h1 class="hero-title">Deploy Your Kubernetes Cluster</h1>
        <p class="hero-subtitle">Get your cluster running in minutes with smart defaults</p>
      </div>

      <div class="deploy-content">
        <!-- Quick Deploy Card -->
        <div class="quick-deploy-card">
          <div class="card-header">
            <v-icon icon="mdi-rocket-launch" size="32" color="primary" class="header-icon" />
            <h2 class="card-title">Quick Deploy</h2>
          </div>

          <div class="deploy-form">
            <!-- Cluster Name -->
            <div class="form-group">
              <label class="form-label">Cluster Name</label>
              <v-text-field
                v-model="clusterName"
                placeholder="app1"
                variant="outlined"
                density="comfortable"
                :rules="[rules.required, rules.alphaNum, rules.maxName]"
                color="primary"
                hint="Only letters and numbers allowed (max 30 chars)"
                persistent-hint
              >
                <template v-slot:append-inner>
                  <v-btn
                    icon="mdi-dice-5"
                    variant="text"
                    size="small"
                    @click="generateRandomName"
                    class="generate-btn"
                  />
                </template>
              </v-text-field>
            </div>

            <!-- Cluster Size -->
            <div class="form-group">
              <label class="form-label">Cluster Size</label>
              <div class="size-options">
                <div
                  v-for="size in CLUSTER_SIZES"
                  :key="size.id"
                  class="size-option"
                  :class="{ active: selectedSize === size.id }"
                  @click="selectedSize = size.id"
                >
                  <div class="size-header">
                    <v-icon :icon="size.icon" size="24" />
                    <span class="size-name">{{ size.name }}</span>
                    <v-chip v-if="size.recommended" color="success" size="x-small" class="recommended-chip">
                      Recommended
                    </v-chip>
                  </div>
                  <div class="size-specs">{{ size.specs }}</div>
                  <div class="size-description">{{ size.description }}</div>
                </div>

                <!-- Custom Size Option -->
                <div
                  class="size-option custom-option"
                  :class="{ active: selectedSize === 'custom' }"
                  @click="selectedSize = 'custom'"
                >
                  <div class="size-header">
                    <v-icon icon="mdi-tune" size="24" />
                    <span class="size-name">Custom</span>
                  </div>
                  <div class="size-specs">Configure your own</div>
                  <div class="size-description">Choose exact number of nodes</div>
                </div>
              </div>

              <!-- Custom Configuration -->
              <div v-if="selectedSize === 'custom'" class="custom-config">
                <div class="custom-inputs">
                  <div class="custom-input-group">
                    <label class="custom-label">Master Nodes</label>
                    <v-text-field
                      v-model.number="customConfig.masters"
                      type="number"
                      min="1"
                      max="5"
                      variant="outlined"
                      density="compact"
                      color="primary"
                      :rules="[rules.required, rules.minMasters]"
                    />
                  </div>
                  <div class="custom-input-group">
                    <label class="custom-label">Worker Nodes</label>
                    <v-text-field
                      v-model.number="customConfig.workers"
                      type="number"
                      min="0"
                      max="10"
                      variant="outlined"
                      density="compact"
                      color="primary"
                      :rules="[rules.minWorkers]"
                    />
                  </div>
                </div>
                <div class="custom-summary">
                  <v-icon icon="mdi-information" size="16" color="info" />
                  <span>Total nodes needed: {{ totalNodesNeeded }}</span>
                </div>
              </div>
            </div>

            <!-- SSH Key Selection -->
            <div class="form-group">
              <label class="form-label">
                SSH Access
                <v-tooltip text="Required to access your cluster nodes">
                  <template v-slot:activator="{ props }">
                    <v-icon v-bind="props" icon="mdi-help-circle" size="16" class="help-icon" />
                  </template>
                </v-tooltip>
              </label>

              <div v-if="sshKeysLoading" class="loading-state">
                <v-progress-circular indeterminate size="20" color="primary" />
                <span>Loading SSH keys...</span>
              </div>

              <div v-else-if="availableSshKeys.length === 0" class="no-ssh-keys">
                <v-icon icon="mdi-key-plus" size="24" color="warning" />
                <div>
                  <p class="no-keys-title">No SSH keys found</p>
                  <p class="no-keys-subtitle">You need at least one SSH key to deploy a cluster</p>
                  <v-btn
                    color="primary"
                    variant="outlined"
                    size="small"
                    @click="navigateToSshKeys"
                    prepend-icon="mdi-plus"
                  >
                    Add SSH Key
                  </v-btn>
                </div>
              </div>

              <v-select
                v-else
                v-model="selectedSshKey"
                :items="sshKeyOptions"
                item-title="name"
                item-value="id"
                variant="outlined"
                density="comfortable"
                placeholder="Select an SSH key"
                :rules="[rules.required]"
                color="primary"
              >
                <template v-slot:append-inner>
                  <v-btn
                    icon="mdi-plus"
                    variant="text"
                    size="small"
                    @click="navigateToSshKeys"
                    class="add-key-btn"
                  />
                </template>
              </v-select>
            </div>

            <!-- Available Nodes -->
            <div class="form-group">
              <label class="form-label">Available Resources</label>
              <div v-if="availableNodes.length === 0" class="no-nodes">
                <v-icon icon="mdi-server-off" size="24" color="warning" />
                <div>
                  <p class="no-nodes-title">No nodes available</p>
                  <p class="no-nodes-subtitle">You need to rent some nodes first</p>
                  <v-btn
                    color="primary"
                    variant="outlined"
                    size="small"
                    @click="navigateToNodes"
                    prepend-icon="mdi-plus"
                  >
                    Rent Nodes
                  </v-btn>
                </div>
              </div>
              <div v-else class="nodes-summary">
                <div class="nodes-count">
                  <v-icon icon="mdi-server" color="success" />
                  <span>{{ availableNodes.length }} nodes available</span>
                </div>
                <div class="resources-summary">
                  <div class="resource-item">
                    <v-icon icon="mdi-memory" size="16" />
                    <span>{{ totalCpu }} CPU cores</span>
                  </div>
                  <div class="resource-item">
                    <v-icon icon="mdi-chip" size="16" />
                    <span>{{ totalRam }} GB RAM</span>
                  </div>
                  <div class="resource-item">
                    <v-icon icon="mdi-harddisk" size="16" />
                    <span>{{ totalStorage }} GB Storage</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Deploy Button -->
            <div class="deploy-actions">
              <v-btn
                color="primary"
                size="large"
                :loading="deploying"
                :disabled="!canDeploy"
                @click="deployCluster"
                class="deploy-btn"
              >
                <v-icon icon="mdi-rocket-launch" class="mr-2" />
                Deploy Cluster
              </v-btn>

              <div v-if="!canDeploy" class="deploy-requirements">
                <v-icon icon="mdi-information" size="16" color="warning" />
                <span>{{ deploymentRequirements }}</span>
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
import { useRouter } from 'vue-router';
import { useNotificationStore } from '../stores/notifications';
import { UserService } from '../utils/userService';
import { useNodeManagement } from '../composables/useNodeManagement';
import { normalizeNode } from '../utils/nodeNormalizer';
import { generateClusterName } from '../utils/clusterUtils';
import { api } from '../utils/api';
import type { ApiResponse } from '../utils/api';
import type { NormalizedNode } from '../types/normalizedNode';
import { required as requiredRule, isAlphanumeric as isAlphanumericRule, min as minRule } from '../utils/validation';

// Constants
const CLUSTER_SIZES = [
  {
    id: 'small',
    name: 'Small',
    icon: 'mdi-desktop-tower',
    specs: '1 master, 1 worker',
    description: 'Great for small production workloads',
    recommended: true,
    masters: 1,
    workers: 1
  },
  {
    id: 'medium',
    name: 'Medium',
    icon: 'mdi-server',
    specs: '1 master, 2 workers',
    description: 'Balanced performance and resources',
    recommended: false,
    masters: 1,
    workers: 2
  },
  {
    id: 'large',
    name: 'Large',
    icon: 'mdi-server-network',
    specs: '3 masters, 3 workers',
    description: 'High availability for critical applications',
    recommended: false,
    masters: 3,
    workers: 3
  }
];

const DEFAULT_CONFIG = {
  ROOT_SIZE_GB: 20,
  MAX_CPU_PER_NODE: 2,
  MAX_RAM_GB_PER_NODE: 4,
  MAX_DISK_GB_PER_NODE: 50
};

const router = useRouter();
const notificationStore = useNotificationStore();
const userService = new UserService();

// Form data
const clusterName = ref('');
const selectedSize = ref('small');
const selectedSshKey = ref<number | null>(null);
const deploying = ref(false);
const customConfig = ref({ masters: 1, workers: 0 });

// Data loading
const sshKeysLoading = ref(true);
const availableSshKeys = ref<any[]>([]);
const { rentedNodes, fetchRentedNodes } = useNodeManagement();
const availableNodes = ref<NormalizedNode[]>([]);

// Validation rules
const rules = {
  required: (value: any) => requiredRule('This field is required')(value) ?? true,
  alphaNum: (value: string) => isAlphanumericRule('Only letters and numbers are allowed')(value) ?? true,
  minMasters: (value: number) => minRule('At least 1 master node is required', 1)(value) ?? true,
  minWorkers: (value: number) => minRule('Worker nodes cannot be negative', 0)(value) ?? true,
  maxName: (value: string) => !value || value.length <= 30 || 'Cluster name too long (max 30 characters)'
};

// Computed properties
const sshKeyOptions = computed(() =>
  availableSshKeys.value.map(key => ({
    id: key.ID,
    name: key.name || `Key ${key.ID}`
  }))
);

const totalCpu = computed(() => availableNodes.value.reduce((sum, node) => sum + (node.cpu || 0), 0));
const totalRam = computed(() => availableNodes.value.reduce((sum, node) => sum + (node.ram || 0), 0));
const totalStorage = computed(() => availableNodes.value.reduce((sum, node) => sum + (node.storage || 0), 0));

const currentClusterConfig = computed(() => {
  if (selectedSize.value === 'custom') {
    return customConfig.value;
  }
  const sizeConfig = CLUSTER_SIZES.find(s => s.id === selectedSize.value);
  return sizeConfig || { masters: 1, workers: 0 };
});

const totalNodesNeeded = computed(() =>
  currentClusterConfig.value.masters + currentClusterConfig.value.workers
);

const canDeploy = computed(() =>
  clusterName.value.trim() &&
  selectedSshKey.value &&
  availableNodes.value.length >= totalNodesNeeded.value &&
  totalNodesNeeded.value >= 1
);

const deploymentRequirements = computed(() => {
  const requirements = [];
  if (!clusterName.value.trim()) requirements.push('cluster name');
  if (!selectedSshKey.value) requirements.push('SSH key');
  if (availableNodes.value.length < totalNodesNeeded.value) {
    requirements.push(`at least ${totalNodesNeeded.value} node${totalNodesNeeded.value > 1 ? 's' : ''}`);
  }
  return `Please provide: ${requirements.join(', ')}`;
});

// Helper functions
function createNodePayload(name: string, type: string, node: NormalizedNode, sshKey: string, token: string) {
  return {
    name,
    type,
    node_id: node.nodeId,
    cpu: Math.min(DEFAULT_CONFIG.MAX_CPU_PER_NODE, node.cpu || DEFAULT_CONFIG.MAX_CPU_PER_NODE),
    memory: Math.min(DEFAULT_CONFIG.MAX_RAM_GB_PER_NODE * 1024, (node.ram || DEFAULT_CONFIG.MAX_RAM_GB_PER_NODE) * 1024),
    root_size: DEFAULT_CONFIG.ROOT_SIZE_GB * 1024,
    disk_size: Math.min(DEFAULT_CONFIG.MAX_DISK_GB_PER_NODE * 1024, (node.storage || DEFAULT_CONFIG.MAX_DISK_GB_PER_NODE) * 1024),
    env_vars: { SSH_KEY: sshKey, K3S_TOKEN: token }
  };
}

// Methods
function generateRandomName() {
  clusterName.value = generateClusterName();
}

function navigateToSshKeys() {
  localStorage.setItem('dashboard-section', 'ssh');
  router.push('/dashboard');
}

function navigateToNodes() {
  router.push('/nodes');
}

async function deployCluster() {
  if (!canDeploy.value) return;

  deploying.value = true;

  try {
    const config = currentClusterConfig.value;
    const sshKey = availableSshKeys.value.find(k => k.ID === selectedSshKey.value);
    const token = '';

    // Auto-assign nodes based on available resources
    const sortedNodes = [...availableNodes.value].sort((a, b) =>
      (b.cpu || 0) + (b.ram || 0) - (a.cpu || 0) - (a.ram || 0)
    );

    const assignedNodes = sortedNodes.slice(0, totalNodesNeeded.value);

    if (assignedNodes.length < totalNodesNeeded.value) {
      throw new Error(`Not enough nodes available. Need ${totalNodesNeeded.value}, have ${assignedNodes.length}`);
    }

    // Build cluster payload
    const clusterPayload = {
      name: clusterName.value,
      token: '',
      nodes: [
        // Masters
        ...Array.from({ length: config.masters }, (_, i) =>
          createNodePayload(`${clusterName.value}m${i + 1}`, 'master', assignedNodes[i], sshKey?.public_key || '', token)
        ),
        // Workers
        ...Array.from({ length: config.workers }, (_, i) =>
          createNodePayload(`${clusterName.value}w${i + 1}`, 'worker', assignedNodes[config.masters + i], sshKey?.public_key || '', token)
        )
      ]
    };

    await api.post<ApiResponse<{ task_id: string }>>('/v1/deployments', clusterPayload, {
      showNotifications: false,
      loadingMessage: 'Deploying cluster...',
      errorMessage: 'Failed to deploy cluster',
      requiresAuth: true
    });

    notificationStore.success(
      'Deployment Started!',
      `Your cluster "${clusterName.value}" is being deployed. You'll be notified when it's ready.`
    );

    localStorage.setItem('dashboard-section', 'clusters');
    router.push('/dashboard');

  } catch (error: any) {
    notificationStore.error(
      'Deployment Failed',
      error.message || 'Failed to deploy cluster. Please try again.'
    );
  } finally {
    deploying.value = false;
  }
}

async function fetchSshKeys() {
  sshKeysLoading.value = true;
  try {
    const keys = await userService.listSshKeys();
    availableSshKeys.value = keys;
    if (keys.length === 1) {
      selectedSshKey.value = keys[0].ID;
    }
  } catch (err) {
    availableSshKeys.value = [];
    notificationStore.error('Error', 'Failed to load SSH keys');
  } finally {
    sshKeysLoading.value = false;
  }
}

async function fetchAvailableNodes() {
  try {
    await fetchRentedNodes();
    availableNodes.value = rentedNodes.value.map(normalizeNode);
  } catch (err) {
    notificationStore.error('Error', 'Failed to load available nodes');
  }
}

// Initialize
onMounted(async () => {
  generateRandomName();
  await Promise.all([fetchSshKeys(), fetchAvailableNodes()]);
});
</script>

<style scoped>
.deploy-container {
  min-height: 100vh;
  background: var(--color-bg);
  padding: var(--space-8) 0;
  padding-top: var(--space-20);
}

.deploy-header {
  text-align: center;
  margin-bottom: var(--space-12);
  padding-top: var(--space-8);
}

.hero-title {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  margin-bottom: var(--space-4);
}

.hero-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-muted);
  font-weight: var(--font-weight-normal);
}

.deploy-content {
  max-width: 60%;
  margin: 0 auto;
  padding: 0 var(--space-8);
}

.quick-deploy-card {
  background: var(--color-bg-elevated);
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-lg);
  padding: var(--space-10);
  border: 1px solid var(--color-border);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-bottom: var(--space-8);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--color-border);
}

.header-icon {
  background: var(--color-primary-subtle);
  padding: var(--space-2);
  border-radius: var(--radius-xl);
}

.card-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin: 0;
}

.form-group {
  margin-bottom: var(--space-3);
}

.form-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin-bottom: var(--space-3);
  font-size: var(--font-size-sm);
}

.help-icon {
  opacity: 0.6;
  cursor: help;
}

.generate-btn {
  opacity: 0.7;
  color: var(--color-text-muted);
}

.generate-btn:hover {
  opacity: 1;
  color: var(--color-primary);
}

.size-options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4);
}

.size-option {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: var(--space-6);
  cursor: pointer;
  transition: all var(--transition-normal);
  background: var(--color-bg);
}

.size-option:hover {
  border-color: var(--color-primary);
  background: var(--color-bg-hover);
}

.size-option.active {
  border-color: var(--color-primary);
  background: var(--color-primary-subtle);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.size-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
}

.size-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
}

.recommended-chip {
  margin-left: auto;
}

.size-specs {
  font-weight: var(--font-weight-medium);
  color: var(--color-primary);
  font-size: var(--font-size-sm);
  margin-bottom: var(--space-1);
}

.size-description {
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
}

.custom-config {
  margin-top: var(--space-4);
  padding: var(--space-4);
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
}

.custom-inputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.custom-input-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.custom-label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
}

.custom-summary {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
}

.loading-state {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  color: var(--color-text-muted);
  border: 1px solid var(--color-border);
}

.no-ssh-keys, .no-nodes {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-6);
  background: var(--color-bg);
  border: 1px solid var(--color-warning);
  border-radius: var(--radius-lg);
}

.no-keys-title, .no-nodes-title {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin: 0 0 var(--space-1) 0;
}

.no-keys-subtitle, .no-nodes-subtitle {
  color: var(--color-text-muted);
  margin: 0 0 var(--space-3) 0;
  font-size: var(--font-size-sm);
}

.nodes-summary {
  background: var(--color-bg);
  border: 1px solid var(--color-success);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
}

.nodes-count {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin-bottom: var(--space-3);
}

.resources-summary {
  display: flex;
  gap: var(--space-6);
  flex-wrap: wrap;
}

.resource-item {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
}

.deploy-actions {
  text-align: center;
  padding-top: var(--space-4);
}

.deploy-btn {
  min-width: 200px;
  height: 48px;
  border-radius: var(--radius-xl);
  font-weight: var(--font-weight-semibold);
  text-transform: none;
  letter-spacing: 0;
}

.deploy-requirements {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  margin-top: var(--space-4);
  color: var(--color-warning);
  font-size: var(--font-size-sm);
}

.add-key-btn {
  opacity: 0.7;
  color: var(--color-text-muted);
}

.add-key-btn:hover {
  opacity: 1;
  color: var(--color-primary);
}

/* Form field styling */
:deep(.v-field) {
  background: var(--color-bg) !important;
  border-color: var(--color-border) !important;
}

:deep(.v-field--focused) {
  border-color: var(--color-primary) !important;
}

:deep(.v-field__input) {
  color: var(--color-text) !important;
}

:deep(.v-field__input::placeholder) {
  color: var(--color-text-muted) !important;
}

/* Responsive */
@media (max-width: 768px) {
  .deploy-content {
    padding: 0 var(--space-4);
  }

  .quick-deploy-card {
    padding: var(--space-6);
  }

  .hero-title {
    font-size: var(--font-size-2xl);
  }

  .size-options {
    grid-template-columns: 1fr;
  }

  .resources-summary {
    flex-direction: column;
    gap: var(--space-2);
  }

  .custom-inputs {
    grid-template-columns: 1fr;
  }
}
</style>
