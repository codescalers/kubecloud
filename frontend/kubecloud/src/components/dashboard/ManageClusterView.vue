<template>
  <div class="manage-cluster-container">
    <div class="container">
      <v-container fluid class="pa-0">
        <div v-if="loading" class="d-flex justify-center align-center" style="min-height: 60vh;">
          <v-progress-circular indeterminate color="primary" size="48" />
        </div>
        <div v-else-if="notFound" class="d-flex flex-column justify-center align-center" style="min-height: 60vh;">
          <h2>Cluster Not Found</h2>
          <v-btn color="primary" @click="goBack">Back to Dashboard</v-btn>
        </div>
        <div v-else-if="cluster" class="mb-6">
          <div>
            <h1 class="text-h3 font-weight-bold text-white mb-2">{{ cluster?.cluster?.name || '-' }}</h1>
            <p class="text-h6 text-medium-emphasis">Manage your Kubernetes cluster configuration and resources</p>
          </div>
        </div>
        <div v-if="!loading && !notFound && cluster" class="manage-content-wrapper">
          <div class="d-flex justify-end ga-3 mb-4">
            <v-btn variant="outlined" color="primary" @click="openKubeconfigModal">
              <v-icon icon="mdi-eye" class="me-2"></v-icon>
              Show Kubeconfig
            </v-btn>
            <v-btn variant="outlined" color="primary" @click="openEditClusterNodesDialog">
              <v-icon icon="mdi-pencil" class="me-2"></v-icon>
              Edit Cluster
            </v-btn>
            <v-btn variant="outlined" color="error" @click="openDeleteModal">
              <v-icon icon="mdi-delete" class="me-2"></v-icon>
              Delete
            </v-btn>
          </div>
          <v-card color="surface-variant" class="pa-6">
            <div class="mb-6">
              <v-row>
                <v-col cols="6" sm="3">
                  <div class="text-body-2 text-medium-emphasis mb-1">Project Name</div>
                  <div class="text-h6 font-weight-medium text-white">{{ cluster.cluster.name || '-' }}</div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="text-body-2 text-medium-emphasis mb-1">CPU</div>
                  <div class="text-h6 font-weight-medium text-white">{{ totalCPU }}</div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="text-body-2 text-medium-emphasis mb-1">Created</div>
                  <div class="text-body-1 text-white">{{ formatDate(cluster.created_at) }}</div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="text-body-2 text-medium-emphasis mb-1">Storage</div>
                  <div class="text-h6 font-weight-medium text-white">{{ Math.round(totalStorage / 1024) }} GB</div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="text-body-2 text-medium-emphasis mb-1">Last Updated</div>
                  <div class="text-body-1 text-white">{{ formatDate(cluster.updated_at) }}</div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="text-body-2 text-medium-emphasis mb-1">RAM</div>
                  <div class="text-h6 font-weight-medium text-white">{{ Math.round(totalRam / 1024) }} GB</div>
                </v-col>
              </v-row>
            </div>
            <div class="mt-6">
              <h3 class="text-h6 font-weight-medium mb-4 d-flex align-center text-white">
                <v-icon icon="mdi-lan" class="me-2" color="primary"></v-icon>
                Cluster Nodes
              </h3>
              <v-table v-if="filteredNodes.length" theme="dark">
                <thead>
                  <tr>
                    <th class="text-white">Name</th>
                    <th class="text-white">Type</th>
                    <th class="text-white">CPU</th>
                    <th class="text-white">RAM</th>
                    <th class="text-white">Storage</th>
                    <th class="text-white">IP</th>
                    <th class="text-white">Mycelium IP</th>
                    <th class="text-white">Planetary IP</th>
                    <th class="text-white">Contract ID</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="node in filteredNodes" :key="node.node_id">
                    <td class="text-white">{{ node.original_name }}</td>
                    <td class="text-white">{{ node.type }}</td>
                    <td class="text-white">{{ node.cpu }}</td>
                    <td class="text-white">{{ Math.round(node.memory / 1024) }} GB</td>
                    <td class="text-white">{{ Math.round((node.root_size + node.disk_size) / 1024) }} GB</td>
                    <td class="text-white">
                      <span class="truncate-cell">
                        {{ node.ip || '-' }}
                      </span>
                    </td>
                    <td class="text-white">
                      <span v-if="node.mycelium_ip" class="full-ip-cell">
                        {{ node.mycelium_ip }}
                      </span>
                      <span v-else>-</span>
                    </td>
                    <td class="text-white">
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
                    <td class="text-white">{{ node.contract_id || '-' }}</td>
                  </tr>
                </tbody>
              </v-table>
              <div v-else class="empty-message">No node details available.</div>
            </div>
          </v-card>
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
import { api } from '../../utils/api'
import { useNodeManagement, type RentedNode } from '../../composables/useNodeManagement'
import { useNotificationStore } from '../../stores/notifications'

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

async function showKubeconfig() {
  kubeconfigLoading.value = true
  kubeconfigError.value = ''
  kubeconfigContent.value = ''
  try {
    const response = await api.get(`/v1/deployments/${projectName.value}/kubeconfig`, { requiresAuth: true, showNotifications: false })
    const data = response.data as { kubeconfig?: string }
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
  const blob = new Blob([kubeconfigContent.value], { type: 'application/x-yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${projectName.value}-kubeconfig.yaml`
  document.body.appendChild(a)
  a.click()
  setTimeout(() => {
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }, 100)
}

const showDeleteModal = ref(false)
const deletingCluster = ref(false)

async function confirmDelete() {
  deletingCluster.value = true
  showDeleteModal.value = false

  if (cluster.value) {
    try {
      await clusterStore.deleteCluster(cluster.value.cluster.name)
      notificationStore.info('Cluster Removal Started', 'Cluster is being removed in the background. You will be notified when the operation completes.');
      goBack()
    } catch (e: any) {
      const errorMessage = e?.message || 'Failed to delete cluster';
      notificationStore.error('Delete Cluster Failed', errorMessage);
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
/* Most styling now handled by Vuetify utilities */
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
