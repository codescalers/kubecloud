<template>
  <div class="nodes-card">
    <!-- Header Section -->
    <div class="card-header">
      <div class="header-content">
        <h2 class="card-title kubecloud-gradient kubecloud-glow-blue">
          My Nodes
        </h2>
        <p class="card-description">
          Manage your rented nodes and their resources.
        </p>
      </div>
      <div class="header-actions">
        <v-btn
          color="primary"
          variant="elevated"
          prepend-icon="mdi-plus"
          @click="navigateToReserve"
        >
          Reserve New Node
        </v-btn>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="stats-section">
      <v-row>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="primary">mdi-server-network</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ rentedNodes.length }}</div>
                <div class="stat-label">Total Nodes</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="success">mdi-check-circle</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ healthyNodes.length }}</div>
                <div class="stat-label">Healthy</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="warning">mdi-alert-circle</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ unhealthyNodes.length }}</div>
                <div class="stat-label">Unhealthy</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" color="info">mdi-currency-usd</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">${{ totalMonthlyCost.toFixed(2) }}</div>
                <div class="stat-label">Monthly Cost</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-section">
      <v-progress-circular
        indeterminate
        color="primary"
        size="64"
      />
      <p class="loading-text">Loading your nodes...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-section">
      <v-icon size="64" color="error" class="mb-4">mdi-alert-circle</v-icon>
      <h3>Failed to load nodes</h3>
      <p>{{ error }}</p>
      <v-btn
        color="primary"
        variant="outlined"
        @click="fetchRentedNodes"
      >
        Try Again
      </v-btn>
    </div>

    <!-- Empty State -->
    <div v-else-if="rentedNodes.length === 0" class="empty-section">
      <v-icon size="64" color="primary" class="mb-4">mdi-server-network</v-icon>
      <h3>No nodes rented yet</h3>
      <p>Start by reserving your first node to deploy your applications.</p>
      <v-btn
        color="primary"
        variant="elevated"
        @click="navigateToReserve"
      >
        Reserve Your First Node
      </v-btn>
    </div>

    <!-- Nodes Grid -->
    <div v-else class="nodes-section">
      <div class="nodes-grid">
        <div v-for="node in rentedNodes" :key="node.id" class="node-card">
          <div class="node-header">
            <h3 class="node-title">Node {{ node.nodeId || node.id }}</h3>
            <div class="node-price">${{ node.price_usd?.toFixed(2) ?? 'N/A' }}/month</div>
          </div>
          <div class="node-location" v-if="node.country">
            <v-icon size="16" class="mr-1">mdi-map-marker</v-icon>
            {{ node.country }}
          </div>
          <div class="node-specs">
            <div class="spec-item">
              <v-icon size="18" class="mr-1" color="primary">mdi-cpu-64-bit</v-icon>
              <span class="spec-label">CPU:</span>
              <span>{{ Math.round(node.resources?.cpu ?? node.total_resources?.cru ?? 0) }} vCPU</span>
            </div>
            <div class="spec-item">
              <v-icon size="18" class="mr-1" color="success">mdi-memory</v-icon>
              <span class="spec-label">RAM:</span>
              <span>{{ Math.round(node.resources?.memory ?? (node.total_resources?.mru ? node.total_resources.mru / (1024*1024*1024) : 0)) }} GB</span>
            </div>
            <div class="spec-item">
              <v-icon size="18" class="mr-1" color="info">mdi-harddisk</v-icon>
              <span class="spec-label">Storage:</span>
              <span>{{ formatStorage(node.resources?.storage ?? (node.total_resources?.sru ? node.total_resources.sru / (1024*1024*1024) : 0)) }}</span>
            </div>
            <div v-if="node.gpu || (node.gpus && node.gpus.length > 0)" class="spec-item">
              <v-icon size="18" class="mr-1" color="info">mdi-expansion-card</v-icon>
              <span class="spec-label">GPU:</span>
              <span>{{ node.gpus?.length }}</span>
            </div>
          </div>

          <v-btn
            color="error"
            variant="outlined"
            class="reserve-btn"
            @click="confirmUnreserve(node)"
            :loading="unreservingNode === node.rentContractId?.toString()"
            style="margin-top: auto; width: 100%;"
          >
            Unreserve
          </v-btn>
        </div>
      </div>
    </div>

    <!-- Unreserve Confirmation Dialog -->
    <v-dialog v-model="showUnreserveDialog" max-width="400">
      <v-card class="pa-3">
        <v-card-title>Confirm Unreservation</v-card-title>
        <v-card-text>
          Are you sure you want to unreserve this node? This action cannot be undone if there is an active cluster on the node.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            color="grey"
            variant="text"
            @click="showUnreserveDialog = false"
          >
            Cancel
          </v-btn>
          <v-btn
            color="error"
            variant="outlined"
            @click="handleUnreserve"
            :loading="unreservingNode === selectedNode?.rentContractId?.toString()"
          >
            Unreserve
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNodeManagement, type RentedNode } from '../../composables/useNodeManagement'
import { useNotificationStore } from '../../stores/notifications'

const router = useRouter()
const {
  rentedNodes,
  loading,
  error,
  fetchRentedNodes,
  unreserveNode,
  totalMonthlyCost,
  healthyNodes,
  unhealthyNodes
} = useNodeManagement()

const notificationStore = useNotificationStore()

// Dialog state
const showUnreserveDialog = ref(false)
const selectedNode = ref<RentedNode | null>(null)
const unreservingNode = ref<string | null>(null)

onMounted(() => {
  fetchRentedNodes()
})

const navigateToReserve = () => {
  router.push('/nodes')
}

const confirmUnreserve = (node: RentedNode) => {
  selectedNode.value = node
  showUnreserveDialog.value = true
}

const handleUnreserve = async () => {
  if (!selectedNode.value?.rentContractId) return
  unreservingNode.value = selectedNode.value.rentContractId.toString()
  try {
    await unreserveNode(selectedNode.value.rentContractId.toString())
    showUnreserveDialog.value = false
    selectedNode.value = null
  } catch (err) {
    console.error('Failed to unreserve node. Please try again.')
  } finally {
    unreservingNode.value = null
  }
}

// Resource calculation functions
function getTotalCPU(node: RentedNode) {
  return node.total_resources?.cru ?? node.resources?.cpu ?? 0
}

function getUsedCPU(node: RentedNode) {
  return node.used_resources?.cru ?? 0
}

function getAvailableCPU(node: RentedNode) {
  return Math.max(getTotalCPU(node) - getUsedCPU(node), 0)
}

function getTotalRAM(node: RentedNode) {
  return node.total_resources?.mru ? Math.round(node.total_resources.mru / (1024 * 1024 * 1024)) : (node.resources?.memory ?? 0)
}

function getUsedRAM(node: RentedNode) {
  return node.used_resources?.mru ? Math.round(node.used_resources.mru / (1024 * 1024 * 1024)) : 0
}

function getAvailableRAM(node: RentedNode) {
  return Math.max(getTotalRAM(node) - getUsedRAM(node), 0)
}

function getTotalStorage(node: RentedNode) {
  return node.total_resources?.sru ? Math.round(node.total_resources.sru / (1024 * 1024 * 1024)) : (node.resources?.storage ?? 0)
}

function getUsedStorage(node: RentedNode) {
  return node.used_resources?.sru ? Math.round(node.used_resources.sru / (1024 * 1024 * 1024)) : 0
}

function getAvailableStorage(node: RentedNode) {
  return Math.max(getTotalStorage(node) - getUsedStorage(node), 0)
}

function formatStorage(val: number) {
  if (val >= 1024) {
    return (val / 1024).toLocaleString(undefined, { maximumFractionDigits: 1, minimumFractionDigits: 1 }) + ' TB';
  }
  return Math.round(val).toLocaleString() + ' GB';
}
</script>

<style scoped>
.nodes-card {
  background: rgba(10, 25, 47, 0.85);
  border: 1px solid rgba(96, 165, 250, 0.15);
  border-radius: 1rem;
  padding: 2rem;
  color: #CBD5E1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  gap: 1rem;
}

.header-content {
  flex: 1;
}

.card-title {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  line-height: 1.2;
}

.card-description {
  color: #94A3B8;
  font-size: 1rem;
  line-height: 1.5;
  margin: 0;
}

.stats-section {
  margin-bottom: 2rem;
}

.stat-card {
  background: rgba(15, 23, 42, 0.6) !important;
  border: 1px solid rgba(96, 165, 250, 0.1) !important;
  border-radius: 0.75rem !important;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.stat-card:hover {
  border-color: rgba(96, 165, 250, 0.3) !important;
  transform: translateY(-2px);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  background: rgba(96, 165, 250, 0.1);
  border: 1px solid rgba(96, 165, 250, 0.2);
  border-radius: 0.5rem;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #F8FAFC;
  line-height: 1;
}

.stat-label {
  font-size: 0.875rem;
  color: #94A3B8;
  margin-top: 0.25rem;
}

.loading-section,
.error-section,
.empty-section {
  text-align: center;
  padding: 4rem 2rem;
}

.loading-text {
  margin-top: 1rem;
  color: #94A3B8;
}

.nodes-section {
  margin-top: 2rem;
}

.nodes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
  gap: 1.5rem;
}

.node-card {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(96, 165, 250, 0.1);
  border-radius: 0.75rem;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.node-card:hover {
  border-color: rgba(96, 165, 250, 0.3);
  transform: translateY(-2px);
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.node-title-section {
  flex: 1;
}

.node-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #F8FAFC;
  margin: 0 0 0.5rem 0;
  line-height: 1.2;
}

.node-status {
  margin-bottom: 0.5rem;
}

.node-price {
  font-size: 1.125rem;
  font-weight: 600;
  color: #10B981;
  text-align: right;
}

.node-location {
  display: flex;
  align-items: center;
  color: #94A3B8;
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.node-specs {
  margin-bottom: 1rem;
}

.spec-item {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;
}

.spec-label {
  font-size: 0.875rem;
  color: #94A3B8;
  min-width: 60px;
  margin-right: 0.5rem;
}

.node-chips {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.node-actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    align-self: stretch;
  }

  .nodes-grid {
    grid-template-columns: 1fr;
  }

  .node-header {
    flex-direction: column;
    align-items: stretch;
  }

  .node-price {
    text-align: left;
    margin-top: 0.5rem;
  }
}
</style>
