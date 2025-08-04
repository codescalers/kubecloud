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
      <v-table v-else class="clusters-table">
        <thead>
          <tr>
            <th @click="setSort('name')" :class="sortBy === 'name' ? 'active-sort' : ''">Name <v-icon v-if="sortBy === 'name'" size="14">mdi-arrow-up-down</v-icon></th>
            <th @click="setSort('nodes')" :class="sortBy === 'nodes' ? 'active-sort' : ''">Nodes <v-icon v-if="sortBy === 'nodes'" size="14">mdi-arrow-up-down</v-icon></th>
            <th @click="setSort('createdAt')" :class="sortBy === 'createdAt' ? 'active-sort' : ''">Created <v-icon v-if="sortBy === 'createdAt'" size="14">mdi-arrow-up-down</v-icon></th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cluster in paginatedClusters" :key="cluster.id">
            <td class="cluster-name-cell">
              <span class="cluster-name">{{ cluster.cluster.name }}</span>
            </td>
            <td>{{ Array.isArray(cluster.cluster.nodes) ? cluster.cluster.nodes.length : (typeof cluster.cluster.nodes === 'number' ? cluster.cluster.nodes : 0) }}</td>
            <td>{{ formatDate(cluster.created_at) }}</td>
            <td>
              <v-btn icon size="small" class="mr-1" @click="viewCluster(cluster.cluster.name)"><v-icon icon="mdi-eye-outline" /></v-btn>
              <v-btn icon size="small" class="ml-1" color="error" @click="deleteCluster(cluster.cluster.name)"><v-icon icon="mdi-delete-outline" /></v-btn>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useClusterStore } from '../../stores/clusters'
import { useNotificationStore } from '../../stores/notifications'

const router = useRouter()
const clusterStore = useClusterStore()
const notificationStore = useNotificationStore()

const showDeleteModal = ref(false)
const deleting = ref(false)
const clusterToDelete = ref<string | null>(null)

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
  await clusterStore.deleteCluster(clusterToDelete.value)
  notificationStore.info('Cluster Removal Started', `Cluster is being removed from the cluster in the background. You will be notified when the operation completes.`);
  showDeleteModal.value = false
  deleting.value = false
  clusterToDelete.value = null
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
.clusters-table {
  width: 100%;
  border-radius: 12px;
  overflow: hidden;
  background: var(--color-surface-1, #18192b);
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
th, td {
  padding: 0.75rem 1rem;
  text-align: left;
}
th {
  background: var(--color-surface-2, #23243a);
  font-weight: 600;
  cursor: pointer;
  user-select: none;
}
th.active-sort {
  color: var(--color-primary, #6366f1);
}
tr {
  border-bottom: 1px solid var(--color-surface-2, #23243a);
}
tr:last-child {
  border-bottom: none;
}
.cluster-name-cell {
  font-weight: 600;
  color: var(--color-primary, #6366f1);
}
.empty-message {
  text-align: center;
  color: var(--color-text-muted, #7c7fa5);
  margin-top: 3rem;
}
</style>
