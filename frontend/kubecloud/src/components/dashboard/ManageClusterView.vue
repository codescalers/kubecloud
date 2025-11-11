<template>
  <div class="manage-cluster-container">
    <div class="container">
      <v-container fluid class="pa-0">
        <div v-if="loading" class="d-flex justify-center align-center" style="min-height: 60vh;">
          <v-progress-circular indeterminate color="primary" size="48" />
        </div>
        <div v-else-if="notFound" class="d-flex flex-column justify-center align-center" style="min-height: 60vh;">
          <h2>Cluster Not Found</h2>
          <v-btn variant="outlined" color="primary" @click="goBack">Back to Dashboard</v-btn>
        </div>
        <div v-else-if="cluster" class="manage-header mb-6">
          <div class="manage-header-content">
            <div class="header-top mb-3">
              <v-btn variant="text" color="primary" @click="goBack" class="back-button">
                <v-icon icon="mdi-arrow-left" class="mr-2"></v-icon>
                Back to Dashboard
              </v-btn>
            </div>
            <h1 class="manage-title">{{ cluster?.cluster?.name || '-' }}</h1>
            <p class="manage-subtitle">Manage your Kubernetes cluster configuration and resources</p>
          </div>
        </div>
        <div v-if="!loading && !notFound && cluster" class="manage-content-wrapper">
          <div class="status-actions d-flex flex-column flex-sm-row gap-2">
            <v-btn 
              variant="outlined" 
              class="btn btn-outline status-action-btn" 
              @click="openKubeconfigModal"
            >
              <v-icon icon="mdi-eye" class="mr-2"></v-icon>
              <span class="d-none d-sm-inline">Show Kubeconfig</span>
              <span class="d-sm-none">Kubeconfig</span>
            </v-btn>

            <v-tooltip location="top" :disabled="haveEnoughBalance">
              <template #activator="{ props }">
                <div v-bind="props" class="status-action-wrapper">
                  <v-btn 
                    variant="outlined" 
                    :disabled="!haveEnoughBalance" 
                    class="btn btn-outline status-action-btn" 
                    @click="openEditClusterNodesDialog"
                  >
                    <v-icon icon="mdi-pencil" class="mr-2"></v-icon>
                    Add Node
                  </v-btn>
                </div>
              </template>
              <span>Insufficient balance. Minimum 5 TFT required to add nodes.</span>
            </v-tooltip>

            <v-btn 
              variant="outlined" 
              class="btn btn-outline status-action-btn" 
              color="error" 
              @click="openDeleteModal"
            >
              <v-icon icon="mdi-delete" class="mr-2"></v-icon>
              Delete
            </v-btn>
          </div>
          <div class="main-content modern-cluster-card">
            <div class="modern-cluster-info">
              <div class="cluster-info-grid">
                <div class="info-item">
                  <div class="info-item-header">
                    <v-icon icon="mdi-folder-network" size="20" class="info-icon"></v-icon>
                    <span class="info-label">Project Name</span>
                  </div>
                  <div class="info-value">{{ cluster.cluster.name || '-' }}</div>
                </div>
                
                <div class="info-item">
                  <div class="info-item-header">
                    <v-icon icon="mdi-cpu-64-bit" size="20" class="info-icon"></v-icon>
                    <span class="info-label">CPU</span>
                  </div>
                  <div class="info-value">{{ totalCPU }}</div>
                </div>
                
                <div class="info-item">
                  <div class="info-item-header">
                    <v-icon icon="mdi-memory" size="20" class="info-icon"></v-icon>
                    <span class="info-label">RAM</span>
                  </div>
                  <div class="info-value">{{ Math.round(totalRam / 1024) }} GB</div>
                </div>
                
                <div class="info-item">
                  <div class="info-item-header">
                    <v-icon icon="mdi-harddisk" size="20" class="info-icon"></v-icon>
                    <span class="info-label">Storage</span>
                  </div>
                  <div class="info-value">{{ Math.round(totalStorage / 1024) }} GB</div>
                </div>
                
                <div class="info-item">
                  <div class="info-item-header">
                    <v-icon icon="mdi-calendar-plus" size="20" class="info-icon"></v-icon>
                    <span class="info-label">Created</span>
                  </div>
                  <div class="info-value">{{ formatDate(cluster.created_at) }}</div>
                </div>
                
                <div class="info-item">
                  <div class="info-item-header">
                    <v-icon icon="mdi-calendar-clock" size="20" class="info-icon"></v-icon>
                    <span class="info-label">Last Updated</span>
                  </div>
                  <div class="info-value">{{ formatDate(cluster.updated_at) }}</div>
                </div>
              </div>
            </div>
            <div class="nodes-section mt-8">
              <h3 class="dashboard-card-title mb-4">
                <v-icon icon="mdi-lan" class="mr-2"></v-icon>
                Cluster Nodes
              </h3>
              <template v-if="filteredNodes.length">
                <!-- Desktop Table View -->
                <v-table class="d-none d-lg-table">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Type</th>
                      <th>Node ID</th>
                      <th>CPU</th>
                      <th>RAM</th>
                      <th>Storage</th>
                      <th>IP</th>
                      <th>Mycelium IP</th>
                      <th>Contract ID</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="node in filteredNodes" :key="node.node_id">
                      <td>{{ node.original_name }}</td>
                      <td>{{ node.type }}</td>
                      <td>{{ node.node_id }}</td>
                      <td>{{ node.cpu }}</td>
                      <td>{{ Math.round(node.memory / 1024) }} GB</td>
                      <td>{{ Math.round((node.root_size + node.disk_size) / 1024) }} GB</td>
                      <td>
                        <span class="truncate-cell">
                          {{ node.ip || '-' }}
                        </span>
                      </td>
                      <td>
                        <span v-if="node.mycelium_ip" class="full-ip-cell">
                          {{ node.mycelium_ip }}
                        </span>
                        <span v-else>-</span>
                      </td>
                      <td>{{ node.contract_id || '-' }}</td>
                      <td>
                        <v-btn
                          icon
                          size="small"
                          color="error"
                          variant="text"
                          @click="showDeleteConfirmation(node.original_name)"
                          :disabled="node.type === 'leader'"
                          class="delete-node-btn"
                        >
                          <v-icon icon="mdi-delete" size="small" />
                          <v-tooltip activator="parent" location="top">
                            Delete Node
                          </v-tooltip>
                        </v-btn>
                      </td>
                    </tr>
                  </tbody>
                </v-table>
                <!-- Mobile/Tablet Card View -->
                <div class="d-lg-none">
                <v-card
                  v-for="node in filteredNodes"
                  :key="node.node_id"
                  variant="outlined"
                  class="mb-4"
                >
                  <v-card-title class="d-flex justify-space-between align-center">
                    <div class="d-flex align-center flex-wrap gap-2">
                      <span class="text-body-1 font-weight-bold">{{ node.original_name }}</span>
                      <v-chip size="small" :color="node.type === 'leader' ? 'primary' : 'default'">
                        {{ node.type }}
                      </v-chip>
                    </div>
                    <v-btn
                      icon
                      size="small"
                      color="error"
                      variant="text"
                      @click="showDeleteConfirmation(node.original_name)"
                      :disabled="node.type === 'leader'"
                    >
                      <v-icon icon="mdi-delete" size="small" />
                      <v-tooltip activator="parent" location="top">
                        Delete Node
                      </v-tooltip>
                    </v-btn>
                  </v-card-title>
                  <v-card-text>
                    <div class="d-flex flex-column gap-2">
                      <div class="d-flex justify-space-between">
                        <span class="text-medium-emphasis">Node ID:</span>
                        <span class="font-weight-medium text-right" style="word-break: break-all;">{{ node.node_id }}</span>
                      </div>
                      <div class="d-flex justify-space-between">
                        <span class="text-medium-emphasis">CPU:</span>
                        <span class="font-weight-medium">{{ node.cpu }}</span>
                      </div>
                      <div class="d-flex justify-space-between">
                        <span class="text-medium-emphasis">RAM:</span>
                        <span class="font-weight-medium">{{ Math.round(node.memory / 1024) }} GB</span>
                      </div>
                      <div class="d-flex justify-space-between">
                        <span class="text-medium-emphasis">Storage:</span>
                        <span class="font-weight-medium">{{ Math.round((node.root_size + node.disk_size) / 1024) }} GB</span>
                      </div>
                      <div class="d-flex justify-space-between">
                        <span class="text-medium-emphasis">IP:</span>
                        <span class="font-weight-medium text-right" style="word-break: break-all;">{{ node.ip || '-' }}</span>
                      </div>
                      <div v-if="node.mycelium_ip" class="d-flex justify-space-between">
                        <span class="text-medium-emphasis">Mycelium IP:</span>
                        <span class="font-weight-medium text-right" style="word-break: break-all;">{{ node.mycelium_ip }}</span>
                      </div>
                      <div class="d-flex justify-space-between">
                        <span class="text-medium-emphasis">Contract ID:</span>
                        <span class="font-weight-medium text-right" style="word-break: break-all;">{{ node.contract_id || '-' }}</span>
                      </div>
                    </div>
                  </v-card-text>
                </v-card>
                </div>
              </template>
              <div v-else class="empty-message">No node details available.</div>
            </div>
          </div>
        </div>
      </v-container>
    </div>


    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteConfirmDialog" :max-width="display.xs ? '90%' : '400'">
      <v-card>
        <v-card-title class="text-h6">
          Confirm Node Deletion
        </v-card-title>
        <v-card-text>
          Are you sure you want to delete the node <strong>{{ nodeToDelete }}</strong>? This action cannot be undone.
        </v-card-text>
        <v-card-actions class="flex-column flex-sm-row gap-2">
          <v-spacer class="d-none d-sm-flex"></v-spacer>
          <v-btn color="grey" variant="text" @click="deleteConfirmDialog = false" block class="d-sm-inline-block">
            Cancel
          </v-btn>
          <v-btn color="error" variant="text" @click="confirmDeleteNode" block class="d-sm-inline-block">
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <!-- Kubeconfig Modal -->
    <component :is="KubeconfigDialog"
      v-model="kubeconfigDialog"
      :projectName="projectName"
      :loading="kubeconfigLoading"
      :error="kubeconfigError"
      :content="kubeconfigContent"
      @copy="copyKubeconfig"
      @download="downloadKubeconfigFile"
    />

    <!-- Delete Confirmation Modal -->
    <component :is="DeleteClusterDialog"
      v-model="showDeleteModal"
      :loading="deletingCluster"
      @confirm="confirmDelete"
    />

    <!-- Edit Cluster Nodes Modal -->
    <component :is="EditClusterNodesDialog"
      v-model="editClusterNodesDialog"
      :cluster="cluster"
      @update:modelValue="editClusterNodesDialog = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, defineAsyncComponent } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useDisplay } from 'vuetify'
import { useClusterStore } from '../../stores/clusters'
import { useNotificationStore } from '../../stores/notifications'
import { useKubeconfig } from '../../composables/useKubeconfig'
import { api } from '../../utils/api'

import { formatDate } from '../../utils/dateUtils'
import { userService } from '@/utils/userService'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const display = useDisplay()

const haveEnoughBalance = computed(() => {
  return userStore.netBalance >= 5
})


// Import dialogs
const EditClusterNodesDialog = defineAsyncComponent(() => import('./EditClusterNodesDialog.vue'))
const KubeconfigDialog = defineAsyncComponent(() => import('./KubeconfigDialog.vue'))
const DeleteClusterDialog = defineAsyncComponent(() => import('./DeleteClusterDialog.vue'))

const router = useRouter()
const route = useRoute()
const clusterStore = useClusterStore()

const loading = ref(true)
const notFound = ref(false)
const deleteConfirmDialog = ref(false);
const nodeToDelete = ref<string>('');

const projectName = computed(() => route.params.id?.toString() || '')
const cluster = computed(() =>
  clusterStore.clusters.find(c => c.cluster.name === projectName.value)
)

const filteredNodes = computed(() => {
  if (Array.isArray(cluster.value?.cluster.nodes)) {
    return cluster.value.cluster.nodes.filter(node => typeof node === 'object' && node !== null)
  }
  return []
})

const totalCPU = computed(() => {
  return filteredNodes.value.length
    ? filteredNodes.value.reduce((sum, node) => sum + (typeof node.cpu === 'number' ? node.cpu : 0), 0)
    : '-'
})
const totalRam = computed(() => {
  return filteredNodes.value.length
    ? filteredNodes.value.reduce((sum, node) => sum + (typeof node.memory === 'number' ? node.memory : 0), 0)
    : '-'
})
const totalStorage = computed(() => {
  return filteredNodes.value.length
    ? filteredNodes.value.reduce((sum, node) => sum + ((typeof node.root_size === 'number' ? node.root_size : 0) + (typeof node.disk_size === 'number' ? node.disk_size : 0)), 0)
    : '-'
})

const kubeconfigDialog = ref(false)
const kubeconfigContent = ref('')
const kubeconfigLoading = ref(false)
const kubeconfigError = ref('')

const { downloadFile } = useKubeconfig()

async function showKubeconfig() {
  kubeconfigLoading.value = true
  kubeconfigError.value = ''
  kubeconfigContent.value = ''
  try {
    const response = await api.get(`/v1/deployments/${projectName.value}/kubeconfig`, {
      requiresAuth: true,
      showNotifications: false,
      timeout: 120000
    })
    const data = response.data as any
    kubeconfigContent.value = data.kubeconfig || ''
  } catch (err: any) {
    kubeconfigError.value = err?.message || 'Failed to fetch kubeconfig'
  } finally {
    kubeconfigLoading.value = false
  }
}

function openKubeconfigModal() {
  kubeconfigDialog.value = true
  if (!kubeconfigContent.value && !kubeconfigLoading.value) {
    showKubeconfig()
  }
}

async function copyKubeconfig() {
  if (kubeconfigContent.value) {
    try {
      await navigator.clipboard.writeText(kubeconfigContent.value)
      notificationStore.success('Copied!', 'Kubeconfig copied to clipboard')
    } catch (err) {
      notificationStore.error('Copy Failed', 'Failed to copy kubeconfig to clipboard')
    }
  }
}

function downloadKubeconfigFile() {
  if (!kubeconfigContent.value) return
  downloadFile(kubeconfigContent.value, `${projectName.value}-kubeconfig.yaml`)
}

const showDeleteModal = ref(false)
const deletingCluster = ref(false)

async function confirmDelete() {
  deletingCluster.value = true
  showDeleteModal.value = false

  if (cluster.value) {
    try {
      await clusterStore.deleteCluster(cluster.value.cluster.name)
      goBack()
    } catch (e: any) {
      notificationStore.error('Delete Cluster Failed', e?.message || 'Failed to delete cluster')
    }
  }
  deletingCluster.value = false
}

function openDeleteModal() {
  showDeleteModal.value = true
}

const loadCluster = async () => {
  loading.value = true
  notFound.value = false
  try {
    if (!clusterStore.clusters.length) {
      await clusterStore.fetchClusters()
    }
    if (!cluster.value) {
      notFound.value = true
    }
  } catch (e) {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadCluster)
const emit = defineEmits(['update:modelValue', 'nodes-updated', 'remove-node']);
function showDeleteConfirmation(nodeName: string) {
  nodeToDelete.value = nodeName;
  deleteConfirmDialog.value = true;
}
watch(() => projectName.value, loadCluster)

const goBack = () => {
  router.push('/dashboard')
}

const editClusterNodesDialog = ref(false)

async function openEditClusterNodesDialog() {
  editClusterNodesDialog.value = true
}

const deletingNodeId = ref<string | null>(null)


function confirmDeleteNode() {
  if (nodeToDelete.value) {
    handleRemoveNode(nodeToDelete.value);
    deleteConfirmDialog.value = false;
    nodeToDelete.value = '';
  }
}
async function handleRemoveNode(nodeName: string) {
  try {
    await removeNodeFromDeployment(cluster.value?.cluster?.name || '', nodeName)
  } catch (e: any) {
    const errorMessage = e?.message || 'Failed to remove node'
    notificationStore.error('Remove Node Failed', errorMessage)
  }
}

async function removeNodeFromDeployment(deploymentName: string, nodeName: string) {
  return await userService.removeNodeFromDeployment(deploymentName, nodeName)
}

const notificationStore = useNotificationStore()

</script>

<style scoped>
.manage-cluster-container {
  margin-top: 10rem;
  min-height: 100vh;
  background: var(--color-bg);
  padding: 0;
}
.manage-header {
  margin-bottom: var(--space-8);
}
.manage-title {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  margin: 0 0 var(--space-2) 0;
}
.manage-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-secondary);
  margin: 0;
}
.status-actions {
  padding: 1rem !important;
  display: flex;
  gap: var(--space-3);
  justify-content: flex-start;
  margin-bottom: var(--space-4);
}
.status-action-wrapper {
  width: 100%;
}
.status-action-btn {
  min-height: 44px;
  font-weight: 500;
  letter-spacing: 0.5px;
  width: 100%;
}
@media (min-width: 600px) {
  .status-actions {
    justify-content: flex-end;
  }
  .status-action-wrapper {
    width: auto;
  }
  .status-action-btn {
    width: auto;
  }
}
@media (max-width: 600px) {
  .status-actions {
    padding: 1rem !important;
    gap: 0.75rem;
  }
  .status-action-btn {
    min-height: 48px;
    font-size: 0.95rem;
  }
}
.main-content-card {
  overflow: hidden;
}
.modern-cluster-card {
  background: rgba(255,255,255,0.03);
  border-radius: 0.25rem;
  padding: 1.5rem 1.5rem 1rem 1.5rem !important;
}
.modern-cluster-info {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  margin-bottom: 2.5rem;
}
.cluster-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
}
.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem;
  border-left: 3px solid var(--v-primary-base);
  background: rgba(255, 255, 255, 0.02);
  border-radius: 4px;
  transition: all 0.2s ease;
}
.info-item:hover {
  background: rgba(255, 255, 255, 0.04);
  border-left-color: var(--v-primary-lighten1);
}
.info-item-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.info-icon {
  color: var(--v-primary-base);
  opacity: 0.8;
}
.info-label {
  color: var(--color-text-muted);
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.info-value {
  color: var(--color-text);
  font-size: 1.125rem;
  font-weight: 600;
  word-break: break-word;
  margin-top: 0.25rem;
}
.cluster-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 1.5rem;
}
.nodes-section {
  margin-top: 2rem;
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  display: flex;
  align-items: center;
}
.truncate-cell {
  display: inline-flex;
  align-items: center;
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}
.full-ip-cell {
  display: inline-flex;
  align-items: center;
  word-break: break-all;
  white-space: normal;
  vertical-align: bottom;
  max-width: 300px;
}
.empty-message {
  color: var(--color-text-muted);
  text-align: center;
  margin: 2rem 0;
}
.header-top {
  display: flex;
  align-items: center;
}
.back-button {
  padding: 0;
  font-size: 0.9rem;
  text-transform: none;
  letter-spacing: normal;
}
@media (max-width: 600px) {
  .manage-cluster-container {
    margin-top: 6rem;
  }
  .manage-header {
    padding: 0 1rem;
  }
  .manage-title {
    font-size: 1.5rem;
  }
  .manage-subtitle {
    font-size: 0.9rem;
  }
  .modern-cluster-card {
    padding: 1.2rem 0.5rem 1rem 0.5rem !important;
  }
  .cluster-title {
    font-size: 1.3rem;
  }
  .cluster-info-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }
  .info-item {
    padding: 0.875rem;
  }
  .info-value {
    font-size: 1rem;
  }
  .nodes-section {
    margin-top: 1.5rem;
  }
  .dashboard-card-title {
    font-size: 1rem;
  }
}
</style>
