<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Kubernetes Clusters</h3>
          <p class="dashboard-card-subtitle">Manage your cloud-native infrastructure</p>
        </div>
      </div>
      <v-btn variant="outlined" class="btn btn-outline" @click="goToDeployCluster">
        <v-icon icon="mdi-plus" size="16" class="mr-1"></v-icon>
        New Cluster
      </v-btn>
    </div>
    <div class="card-content">
      <div class="clusters-list-toolbar">
        <v-text-field
          v-model="search"
          label="Search by name"
          prepend-inner-icon="mdi-magnify"
          clearable
          class="search-bar"
        />
        <v-select
          v-model="sortBy"
          :items="sortOptions"
          label="Sort by"
          class="filter-select"
        />
      </div>
      <v-divider class="mb-4" />
      <v-alert v-if="error" type="error" class="mb-4">{{ error }}</v-alert>
      <v-progress-linear v-if="isLoading" indeterminate color="primary" class="mb-4" />
      
      <div v-if="filteredClusters.length === 0 && !isLoading" class="empty-message">
        <v-icon icon="mdi-cloud-off-outline" size="48" class="mb-2" color="grey" />
        <div>No clusters found.</div>
      </div>
      
      <!-- Enhanced Clusters Grid -->
      <div v-else class="clusters-grid">
        <div 
          v-for="cluster in paginatedClusters" 
          :key="cluster.id" 
          class="cluster-card"
        >
          <div class="cluster-header">
            <div class="cluster-title-section">
              <h4 class="cluster-name">{{ cluster.cluster.name }}</h4>
              <div class="cluster-meta">
                <span class="cluster-type">
                  <v-icon icon="mdi-kubernetes" size="16" class="mr-1"></v-icon>
                  Kubernetes Cluster
                </span>
                <span class="cluster-date">{{ formatDate(cluster.created_at) }}</span>
              </div>
            </div>
            <div class="cluster-actions">
              <v-tooltip location="top">
                <template #activator="{ props }">
                  <v-btn 
                    icon size="small" 
                    variant="text"
                    v-bind="props" 
                    @click="viewCluster(cluster.cluster.name)"
                  >
                    <v-icon icon="mdi-cog" />
                  </v-btn>
                </template>
                <span>Manage cluster</span>
              </v-tooltip>
              
              <v-tooltip location="top">
                <template #activator="{ props }">
                  <v-btn 
                    icon size="small" 
                    variant="text"
                    v-bind="props" 
                    @click="download(cluster.cluster.name)" 
                    :loading="downloading === cluster.cluster.name" 
                    :disabled="downloading === cluster.cluster.name"
                  >
                    <v-icon icon="mdi-download" />
                  </v-btn>
                </template>
                <span>Download kubeconfig</span>
              </v-tooltip>
              
              <v-tooltip location="top">
                <template #activator="{ props }">
                  <v-btn 
                    icon size="small" 
                    variant="text"
                    color="error" 
                    v-bind="props" 
                    @click="deleteCluster(cluster.cluster.name)"
                  >
                    <v-icon icon="mdi-delete-outline" />
                  </v-btn>
                </template>
                <span>Delete cluster</span>
              </v-tooltip>
            </div>
          </div>
          
          <div class="cluster-specs">
            <!-- Node Count -->
            <v-chip color="primary" text-color="white" size="small" class="mr-2 mb-2" variant="outlined">
              <v-icon size="16" class="mr-1">mdi-lan</v-icon>
              {{ getNodeCount(cluster) }} Node{{ getNodeCount(cluster) !== 1 ? 's' : '' }}
            </v-chip>
            
            <!-- CPU Resources -->
            <v-chip color="success" text-color="white" size="small" class="mr-2 mb-2" variant="outlined">
              <v-icon size="16" class="mr-1">mdi-cpu-64-bit</v-icon>
              {{ getTotalCPU(cluster) }} vCPU
            </v-chip>
            
            <!-- RAM Resources -->
            <v-chip color="info" text-color="white" size="small" class="mr-2 mb-2" variant="outlined">
              <v-icon size="16" class="mr-1">mdi-memory</v-icon>
              {{ getTotalRAM(cluster) }} GB RAM
            </v-chip>
            
            <!-- Storage Resources -->
            <v-chip color="warning" text-color="white" size="small" class="mr-2 mb-2" variant="outlined">
              <v-icon size="16" class="mr-1">mdi-harddisk</v-icon>
              {{ getTotalStorage(cluster) }} GB Storage
            </v-chip>
            
            <!-- Status Badge -->
            <v-chip 
              :color="getStatusColor(cluster)" 
              text-color="white" 
              size="small" 
              class="mr-2 mb-2" 
              variant="outlined"
            >
              <v-icon size="16" class="mr-1">{{ getStatusIcon(cluster) }}</v-icon>
              {{ getStatusText(cluster) }}
            </v-chip>
          </div>
          
          <!-- Node Details -->
          <div v-if="getNodeCount(cluster) > 0" class="cluster-nodes">
            <div class="nodes-summary">
              <span class="nodes-label">Node Types:</span>
              <div class="node-type-chips">
                <span v-if="getMasterCount(cluster) > 0" class="node-type-chip master">
                  {{ getMasterCount(cluster) }} Master{{ getMasterCount(cluster) !== 1 ? 's' : '' }}
                </span>
                <span v-if="getWorkerCount(cluster) > 0" class="node-type-chip worker">
                  {{ getWorkerCount(cluster) }} Worker{{ getWorkerCount(cluster) !== 1 ? 's' : '' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <v-pagination
        v-model="page"
        :length="pageCount"
        circle
        total-visible="7"
        class="mt-4"
      />
    </div>
    
    <v-dialog v-model="showDeleteModal" max-width="400">
      <v-card>
        <v-card-title>Confirm Delete</v-card-title>
        <v-card-text>Are you sure you want to delete this cluster? This action cannot be undone.</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" color="primary" @click="showDeleteModal = false">Cancel</v-btn>
          <v-btn variant="outlined" color="error" @click="confirmDelete" :loading="deleting">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useClusterStore } from '../../stores/clusters'
import { useNotificationStore } from '../../stores/notifications'
import { useKubeconfig } from '../../composables/useKubeconfig'

const router = useRouter()
const clusterStore = useClusterStore()
const notificationStore = useNotificationStore()

const showDeleteModal = ref(false)
const deleting = ref(false)
const clusterToDelete = ref<string | null>(null)

const { download, downloading } = useKubeconfig()

const search = ref('')
const sortBy = ref('createdAt')
const page = ref(1)
const pageSize = 6 // Increased for grid layout

const sortOptions = [
  { value: 'name', title: 'Name' },
  { value: 'createdAt', title: 'Created' },
  { value: 'nodes', title: 'Nodes' },
]

const error = computed(() => clusterStore.error)
const isLoading = computed(() => clusterStore.isLoading)

function setSort(field: string) {
  sortBy.value = field
}

onMounted(() => {
  clusterStore.fetchClusters()
})

const filteredClusters = computed(() => {
  let clusters = [...clusterStore.clusters]
  if (search.value) {
    clusters = clusters.filter(c => c.project_name.toLowerCase().includes(search.value.toLowerCase()))
  }
  // Sorting
  clusters.sort((a, b) => {
    if (sortBy.value === 'name') return a.project_name.localeCompare(b.project_name)
    if (sortBy.value === 'createdAt') return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    if (sortBy.value === 'nodes') {
      const aNodes = getNodeCount(a)
      const bNodes = getNodeCount(b)
      return bNodes - aNodes
    }
    return 0
  })
  return clusters
})

const pageCount = computed(() => Math.ceil(filteredClusters.value.length / pageSize))

const paginatedClusters = computed(() => {
  const start = (page.value - 1) * pageSize
  return filteredClusters.value.slice(start, start + pageSize)
})

// Helper functions for cluster data
function getNodeCount(cluster: any): number {
  if (Array.isArray(cluster.cluster.nodes)) {
    return cluster.cluster.nodes.length
  }
  return typeof cluster.cluster.nodes === 'number' ? cluster.cluster.nodes : 0
}

function getMasterCount(cluster: any): number {
  if (Array.isArray(cluster.cluster.nodes)) {
    return cluster.cluster.nodes.filter((node: any) => 
      node.type === 'master' || node.type === 'leader'
    ).length
  }
  return 0
}

function getWorkerCount(cluster: any): number {
  if (Array.isArray(cluster.cluster.nodes)) {
    return cluster.cluster.nodes.filter((node: any) => 
      node.type === 'worker'
    ).length
  }
  return 0
}

function getTotalCPU(cluster: any): number {
  if (Array.isArray(cluster.cluster.nodes)) {
    return cluster.cluster.nodes.reduce((sum: number, node: any) => 
      sum + (typeof node.cpu === 'number' ? node.cpu : 0), 0
    )
  }
  return 0
}

function getTotalRAM(cluster: any): number {
  if (Array.isArray(cluster.cluster.nodes)) {
    const totalMB = cluster.cluster.nodes.reduce((sum: number, node: any) => 
      sum + (typeof node.memory === 'number' ? node.memory : 0), 0
    )
    return Math.round(totalMB / 1024) // Convert MB to GB
  }
  return 0
}

function getTotalStorage(cluster: any): number {
  if (Array.isArray(cluster.cluster.nodes)) {
    const totalMB = cluster.cluster.nodes.reduce((sum: number, node: any) => 
      sum + ((typeof node.root_size === 'number' ? node.root_size : 0) + 
             (typeof node.disk_size === 'number' ? node.disk_size : 0)), 0
    )
    return Math.round(totalMB / 1024) // Convert MB to GB
  }
  return 0
}

function getStatusText(cluster: any): string {
  // You can enhance this based on actual cluster status data
  return 'Active'
}

function getStatusColor(cluster: any): string {
  // You can enhance this based on actual cluster status data
  return 'success'
}

function getStatusIcon(cluster: any): string {
  // You can enhance this based on actual cluster status data
  return 'mdi-check-circle'
}

const viewCluster = (projectName: string) => {
  router.push(`/clusters/${projectName}`)
}

function deleteCluster(projectName: string) {
  clusterToDelete.value = projectName
  showDeleteModal.value = true
}

const goToDeployCluster = () => {
  router.push('/deploy')
}

async function confirmDelete() {
  if (!clusterToDelete.value) return
  
  deleting.value = true
  try {
    await clusterStore.deleteCluster(clusterToDelete.value)
    notificationStore.info('Cluster Removal Started', `Cluster is being removed in the background. You will be notified when the operation completes.`)
  } catch (error: any) {
  } finally {
    showDeleteModal.value = false
    deleting.value = false
    clusterToDelete.value = null
  }
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.clusters-list-toolbar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}
.search-bar {
  min-width: 220px;
  flex: 1 1 220px;
}
.filter-select {
  min-width: 160px;
}

/* Enhanced Grid Layout */
.clusters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.cluster-card {
  background: var(--color-bg-elevated, #1E293B);
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  border: 1px solid var(--color-border, #334155);
  transition: all 0.2s ease;
}

.cluster-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
}

.cluster-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.cluster-title-section {
  flex: 1;
}

.cluster-name {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-primary, #3B82F6);
  margin: 0 0 0.5rem 0;
}

.cluster-meta {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.cluster-type {
  font-size: 0.875rem;
  color: var(--color-text-muted, #7c7fa5);
  display: flex;
  align-items: center;
}

.cluster-date {
  font-size: 0.75rem;
  color: var(--color-text-muted, #7c7fa5);
}

.cluster-actions {
  display: flex;
  gap: 0.25rem;
  flex-shrink: 0;
}

.cluster-specs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.cluster-nodes {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-surface-2, #23243a);
}

.nodes-summary {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.nodes-label {
  font-size: 0.875rem;
  color: var(--color-text-muted, #7c7fa5);
  font-weight: 500;
}

.node-type-chips {
  display: flex;
  gap: 0.5rem;
}

.node-type-chip {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.node-type-chip.master {
  background: var(--color-primary, #3B82F6);
  color: white;
}

.node-type-chip.worker {
  background: var(--color-success, #10b981);
  color: white;
}

.empty-message {
  text-align: center;
  color: var(--color-text-muted, #7c7fa5);
  margin-top: 3rem;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .clusters-grid {
    grid-template-columns: 1fr;
  }
  
  .cluster-card {
    padding: 1rem;
  }
  
  .cluster-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .cluster-actions {
    align-self: flex-end;
  }
}
</style>
