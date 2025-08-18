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
          <div class="status-actions align-end">
            <v-btn variant="outlined" class="btn btn-outline" @click="openKubeconfigModal">
              <v-icon icon="mdi-eye" class="mr-2"></v-icon>
              Show Kubeconfig
            </v-btn>
            <v-btn variant="outlined" class="btn btn-outline" @click="openEditClusterNodesDialog">
              <v-icon icon="mdi-pencil" class="mr-2"></v-icon>
              Edit Cluster
            </v-btn>
            <v-btn variant="outlined" class="btn btn-outline" color="error" @click="openDeleteModal">
              <v-icon icon="mdi-delete" class="mr-2"></v-icon>
              Delete
            </v-btn>
          </div>
          <div class="main-content-card modern-cluster-card">
            <div class="modern-cluster-info">
              <div class="cluster-info-grid">
                <div class="info-label">Project Name</div>
                <div>{{ cluster.cluster.name || '-' }}</div>
                <div class="info-label">CPU</div>
                <div>{{ totalCPU }}</div>
                <div class="info-label">Created</div>
                <div>{{ formatDate(cluster.created_at) }}</div>
                <div class="info-label">Storage</div>
                <div>{{ Math.round(totalStorage / 1024) }} GB</div>
                <div class="info-label">Last Updated</div>
                <div>{{ formatDate(cluster.updated_at) }}</div>

                <div class="info-label">RAM</div>
                <div>{{ Math.round(totalRam / 1024) }} GB</div>
              </div>
            </div>
            <div class="nodes-section mt-8">
              <h3 class="dashboard-card-title mb-4">
                <v-icon icon="mdi-lan" class="mr-2"></v-icon>
                Cluster Nodes
              </h3>
              <v-table v-if="filteredNodes.length">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>CPU</th>
                    <th>RAM</th>
                    <th>Storage</th>
                    <th>IP</th>
                    <th>Mycelium IP</th>
                    <th>Planetary IP</th>
                    <th>Contract ID</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="node in filteredNodes" :key="node.node_id">
                    <td>{{ node.original_name }}</td>
                    <td>{{ node.type }}</td>
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
                    <td>
                      <v-tooltip activator="parent" location="top" v-if="node.planetary_ip">
                        <template #activator="{ props }">
                          <span class="truncate-cell" v-bind="props">
                            {{ node.planetary_ip }}
                          </span>
                        </template>
                        <span>{{ node.planetary_ip }}</span>
                      </v-tooltip>
                      <span v-else>-</span>
                    </td>
                    <td>{{ node.contract_id || '-' }}</td>
                  </tr>
                </tbody>
              </v-table>
              <div v-else class="empty-message">No node details available.</div>
            </div>
          </div>
        </div>
      </v-container>
    </div>

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
      :nodes="filteredNodes"
      :loading="nodesLoading"
      :available-nodes="availableNodes"
      :add-form-error="addFormError"
      :add-form-node="addFormNode"
      :can-assign-to-node="canAssignToNode"
      :add-node-loading="addNodeLoading"
      :available-ssh-keys="sshKeys"
      @add-node="addNode"
      @remove-node="handleRemoveNode"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, defineAsyncComponent } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useClusterStore } from '../../stores/clusters'
import { useNodeManagement, type RentedNode } from '../../composables/useNodeManagement'
import { useNotificationStore } from '../../stores/notifications'
import { useKubeconfig } from '../../composables/useKubeconfig'
import { api } from '../../utils/api'

import { getAvailableCPU, getAvailableRAM, getAvailableStorage } from '../../utils/nodeNormalizer'

import { formatDate } from '../../utils/dateUtils'

// Import dialogs
const EditClusterNodesDialog = defineAsyncComponent(() => import('./EditClusterNodesDialog.vue'))
const KubeconfigDialog = defineAsyncComponent(() => import('./KubeconfigDialog.vue'))
const DeleteClusterDialog = defineAsyncComponent(() => import('./DeleteClusterDialog.vue'))

const router = useRouter()
const route = useRoute()
const clusterStore = useClusterStore()

const loading = ref(true)
const notFound = ref(false)

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

function copyKubeconfig() {
  if (kubeconfigContent.value) {
    navigator.clipboard.writeText(kubeconfigContent.value)
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
      notificationStore.info('Cluster Removal Started', 'Cluster is being removed in the background. You will be notified when the operation completes.')
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
watch(() => projectName.value, loadCluster)

// Watch for cluster updates and refresh data when needed
watch(() => clusterStore.clusters, (newClusters) => {
  // Update editNodes if dialog is open
  if (editClusterNodesDialog.value && cluster.value) {
    const updatedCluster = newClusters.find(c => c.project_name === cluster.value?.project_name)
    if (updatedCluster?.cluster?.nodes && Array.isArray(updatedCluster.cluster.nodes)) {
      editNodes.value = updatedCluster.cluster.nodes.map((n: any) => ({ ...n }))
    }
  }
}, { deep: true })

const goBack = () => {
  router.push('/dashboard')
}

const editClusterNodesDialog = ref(false)

// Dummy state for masters/workers (replace with real cluster data)
const editNodes = ref<any[]>([])

async function openEditClusterNodesDialog() {
  const nodesRaw = cluster.value?.cluster?.nodes;
  const nodes = Array.isArray(nodesRaw) ? nodesRaw : [];
  editNodes.value = nodes.map(n => ({ ...n }));
  editClusterNodesDialog.value = true;
  await fetchRentedNodes();
}

const addNodeLoading = ref(false)
const availableNodes = computed<RentedNode[]>(() => {
  return rentedNodes.value.filter((node: RentedNode) => {
    const availRAM = getAvailableRAM(node);
    const availStorage = getAvailableStorage(node);
    return availRAM > 0 && availStorage > 0;
  });
});

const { rentedNodes, loading: nodesLoading, fetchRentedNodes, addNodeToDeployment, removeNodeFromDeployment } = useNodeManagement()

// Notification store
const notificationStore = useNotificationStore()

// SSH keys state
const sshKeys = ref<any[]>([])
const addFormNodeId = ref(null);
const addFormRole = ref('master');
const addFormCpu = ref(1);
const addFormRam = ref(1);
const addFormStorage = ref(1);
const addFormError = ref('');

const addFormNode = computed<RentedNode | undefined>(() => availableNodes.value.find((n: RentedNode) => n.nodeId === addFormNodeId.value));

const canAssignToNode = computed(() => {
  const node = addFormNode.value;
  if (!node) return false;
  return (
    addFormCpu.value > 0 &&
    addFormRam.value > 0 &&
    addFormStorage.value > 0 &&
    addFormCpu.value <= getAvailableCPU(node) &&
    addFormRam.value <= getAvailableRAM(node) &&
    addFormStorage.value <= getAvailableStorage(node)
  );
});

watch([addFormNodeId, addFormCpu, addFormRam, addFormStorage], () => {
  const node = addFormNode.value;
  if (!node) {
    addFormError.value = '';
    return;
  }
  if (
    addFormCpu.value > getAvailableCPU(node) ||
    addFormRam.value > getAvailableRAM(node) ||
    addFormStorage.value > getAvailableStorage(node)
  ) {
    addFormError.value = 'Requested resources exceed available for the selected node.';
  } else {
    addFormError.value = '';
  }
});

async function addNode(payload: any) {
  // Accepts a cluster payload with a nodes array
  if (!payload || !payload.name || !Array.isArray(payload.nodes) || payload.nodes.length === 0) {
    addFormError.value = 'Invalid node payload.';
    notificationStore.error('Add Node Error', 'Invalid node payload.');
    return;
  }
  addNodeLoading.value = true;
  addFormError.value = '';
  try {
    await addNodeToDeployment(payload.name, payload);

    // Reset add form state
    addFormNodeId.value = null;
    addFormRole.value = 'master';
    addFormCpu.value = 1;
    addFormRam.value = 1;
    addFormStorage.value = 1;
  } catch (e: any) {
    const errorMessage = e?.message || 'Failed to add node';
    addFormError.value = errorMessage;
    notificationStore.error('Add Node Failed', errorMessage);
  } finally {
    addNodeLoading.value = false;
  }
}

async function handleRemoveNode(nodeName: string) {
  if (!cluster.value?.cluster?.name) return;
  try {
    await removeNodeFromDeployment(cluster.value.cluster.name, nodeName);
    notificationStore.info('Node Removal Started', `Node is being removed from the cluster in the background. You will be notified when the operation completes.`);
  } catch (e: any) {
    const errorMessage = e?.message || 'Failed to remove node';
    notificationStore.error('Remove Node Failed', errorMessage);
  }
}

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
  display: flex;
  gap: var(--space-3);
}
.status-actions.align-end {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}
.main-content-card {
  padding: unset !important;
  overflow: hidden;
}
.modern-cluster-card {
  background: rgba(255,255,255,0.03);
  border-radius: 0.25rem;
  padding: 2.5rem 2.5rem 2rem 2.5rem !important;
}
.modern-cluster-info {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  margin-bottom: 2.5rem;
}
.cluster-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 1.5rem;
}
.cluster-info-grid {
  display: grid;
  grid-template-columns: 1fr 1.5fr 1fr 1.5fr;
  gap: 0.7rem 2.5rem;
  align-items: center;
}
.info-label {
  color: var(--color-text-muted);
  font-size: 1rem;
  font-weight: 500;
  text-align: right;
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
@media (max-width: 900px) {
  .cluster-info-grid {
    grid-template-columns: 1fr 1.5fr;
  }
}
@media (max-width: 600px) {
  .modern-cluster-card {
    padding: 1.2rem 0.5rem 1rem 0.5rem;
  }
  .cluster-title {
    font-size: 1.3rem;
  }
  .cluster-info-grid {
    grid-template-columns: 1fr;
    gap: 0.5rem 1rem;
  }
  .info-label {
    font-size: 1rem;
    text-align: left;
  }
}
</style>
