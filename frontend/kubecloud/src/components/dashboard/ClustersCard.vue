<template>
  <div class="dashboard-card">
    <div class="d-flex justify-space-between align-center mb-6">
      <div class="flex-grow-1">
        <h3 class="text-h5 font-weight-bold mb-1">Kubernetes Clusters</h3>
        <p class="text-body-2 text-medium-emphasis">Manage your cloud-native infrastructure</p>
      </div>
      <v-btn variant="outlined" class="mr-2" @click="goToDeployCluster">
        <v-icon icon="mdi-plus" size="16" class="mr-1"></v-icon>
        New Cluster
      </v-btn>
      <v-btn 
        v-if="filteredClusters.length > 0 && !isLoading" 
        variant="outlined" 
        color="error" 
        @click="showDeleteAllModal = true"
        :disabled="deletingAll"
      >
        <v-icon icon="mdi-delete-sweep" size="16" class="mr-1"></v-icon>
        Delete All
        <v-progress-circular v-if="deletingAll" indeterminate size="16" class="ml-2"></v-progress-circular>
      </v-btn>
    </div>
    <div class="card-content">
      <div class="d-flex gap-4 mb-6 flex-wrap">
        <v-text-field
          v-model="search"
          label="Search by name"
          prepend-inner-icon="mdi-magnify"
          clearable
          class="flex-grow-1"
          style="min-width: 220px;"
        />
        <v-select
          v-model="sortBy"
          :items="sortOptions"
          label="Sort by"
          style="min-width: 160px;"
        />
      </div>
      <v-divider class="mb-4" />
      <v-alert v-if="error" type="error" class="mb-4">{{ error }}</v-alert>
      <v-progress-linear v-if="isLoading" indeterminate color="primary" class="mb-4" />
      <div v-else-if="filteredClusters.length === 0 && !isLoading" class="text-center text-medium-emphasis mt-12">
        <v-icon icon="mdi-cloud-off-outline" size="48" class="mb-2" color="grey" />
        <div>No clusters found.</div>
      </div>
      <v-table v-else class="w-100 rounded-lg overflow-hidden">
        <thead>
          <tr>
            <th @click="setSort('name')" class="cursor-pointer">Name <v-icon v-if="sortBy === 'name'" size="14">mdi-arrow-up-down</v-icon></th>
            <th @click="setSort('nodes')" class="cursor-pointer">Nodes <v-icon v-if="sortBy === 'nodes'" size="14">mdi-arrow-up-down</v-icon></th>
            <th @click="setSort('createdAt')" class="cursor-pointer">Created <v-icon v-if="sortBy === 'createdAt'" size="14">mdi-arrow-up-down</v-icon></th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cluster in paginatedClusters" :key="cluster.id">
            <td>
              <span>{{ cluster.cluster.name }}</span>
              <v-chip v-if="deletingAll" size="small" color="warning" class="ml-2">
                <v-icon size="12" class="mr-1">mdi-clock</v-icon>
                Deleting...
              </v-chip>
            </td>
            <td>{{ Array.isArray(cluster.cluster.nodes) ? cluster.cluster.nodes.length : (typeof cluster.cluster.nodes === 'number' ? cluster.cluster.nodes : 0) }}</td>
            <td>{{ formatDate(cluster.created_at) }}</td>
            <td>
              <v-tooltip location="top">
                <template #activator="{ props }">
                  <v-btn icon size="small" class="mr-1" v-bind="props" @click="viewCluster(cluster.cluster.name)" :disabled="deletingAll">
                    <v-icon icon="mdi-cog" />
                  </v-btn>
                </template>
                <span>Edit cluster</span>
              </v-tooltip>
              
              <v-tooltip location="top">
                <template #activator="{ props }">
                  <v-btn icon size="small" class="mr-1" v-bind="props" @click="download(cluster.cluster.name)" :loading="downloading === cluster.cluster.name" :disabled="downloading === cluster.cluster.name || deletingAll">
                    <v-icon icon="mdi-download" />
                  </v-btn>
                </template>
                <span>Download kubeconfig file</span>
              </v-tooltip>
              
              <v-tooltip location="top">
                <template #activator="{ props }">
                  <v-btn icon size="small" class="ml-1" color="error" v-bind="props" @click="deleteCluster(cluster.cluster.name)" :disabled="deletingAll">
                    <v-icon icon="mdi-delete-outline" />
                  </v-btn>
                </template>
                <span>Delete cluster</span>
              </v-tooltip>
            </td>
          </tr>
        </tbody>
      </v-table>
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
    <v-dialog v-model="showDeleteAllModal" max-width="500">
      <v-card class="pa-3">
        <v-card-title>
          Delete All Deployments
        </v-card-title>
        <v-card-text>
          Are you sure you want to delete all your deployments? This action will permanently remove all your clusters and their resources.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" color="primary" @click="showDeleteAllModal = false">Cancel</v-btn>
          <v-btn 
            variant="outlined" 
            color="error" 
            @click="confirmDeleteAll" 
            :loading="deletingAll"
          >
            Delete All
          </v-btn>
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
const showDeleteAllModal = ref(false)
const deletingAll = ref(false)

const { download, downloading } = useKubeconfig()

const search = ref('')
const sortBy = ref('createdAt')
const page = ref(1)
const pageSize = 5

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
      const aNodes = Array.isArray(a.cluster.nodes) ? a.cluster.nodes.length : (typeof a.cluster.nodes === 'number' ? a.cluster.nodes : 0)
      const bNodes = Array.isArray(b.cluster.nodes) ? b.cluster.nodes.length : (typeof b.cluster.nodes === 'number' ? b.cluster.nodes : 0)
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

async function confirmDeleteAll() {
  if (filteredClusters.value.length === 0) return

  deletingAll.value = true
  try {
    await clusterStore.deleteAllDeployments()
    notificationStore.info('Delete All Started', `All your deployments are being deleted. The table will update automatically as deletions complete.`)
  } catch (error: any) {
    notificationStore.error('Delete All Failed', error?.message || 'Failed to delete all deployments')
    deletingAll.value = false // Reset on error
  } finally {
    showDeleteAllModal.value = false
  }
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// Watch for clusters to be removed and reset deletingAll state
watch(() => clusterStore.clusters.length, (newLength) => {
  if (newLength === 0 && deletingAll.value) {
    deletingAll.value = false
  }
})
</script>
