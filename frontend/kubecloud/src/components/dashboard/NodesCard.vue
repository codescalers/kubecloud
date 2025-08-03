<template>
  <v-card color="surface-variant" class="pa-6">
    <!-- Header Section -->
    <div class="d-flex justify-space-between align-start mb-6">
      <div>
        <h2 class="text-h4 kubecloud-gradient kubecloud-glow-blue mb-2">
          My Nodes
        </h2>
        <p class="text-body-1 text-medium-emphasis">
          Manage your rented nodes and their resources.
        </p>
      </div>
      <v-btn
        color="primary"
        variant="elevated"
        prepend-icon="mdi-plus"
        @click="navigateToReserve"
      >
        Reserve New Node
      </v-btn>
    </div>

    <!-- Stats Cards -->
    <v-row class="mb-6">
      <v-col cols="12" sm="6" md="3">
        <v-card color="surface-variant" class="pa-4">
          <div class="d-flex align-center">
            <v-avatar color="primary" class="me-4">
              <v-icon size="24">mdi-server-network</v-icon>
            </v-avatar>
            <div>
              <div class="text-h5 font-weight-bold">{{ rentedNodes.length }}</div>
              <div class="text-body-2 text-medium-emphasis">Total Nodes</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="surface-variant" class="pa-4">
          <div class="d-flex align-center">
            <v-avatar color="success" class="me-4">
              <v-icon size="24">mdi-check-circle</v-icon>
            </v-avatar>
            <div>
              <div class="text-h5 font-weight-bold">{{ healthyNodes.length }}</div>
              <div class="text-body-2 text-medium-emphasis">Healthy</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="surface-variant" class="pa-4">
          <div class="d-flex align-center">
            <v-avatar color="warning" class="me-4">
              <v-icon size="24">mdi-alert-circle</v-icon>
            </v-avatar>
            <div>
              <div class="text-h5 font-weight-bold">{{ unhealthyNodes.length }}</div>
              <div class="text-body-2 text-medium-emphasis">Unhealthy</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="surface-variant" class="pa-4">
          <div class="d-flex align-center">
            <v-avatar color="info" class="me-4">
              <v-icon size="24">mdi-currency-usd</v-icon>
            </v-avatar>
            <div>
              <div class="text-h5 font-weight-bold">${{ totalMonthlyCost.toFixed(2) }}</div>
              <div class="text-body-2 text-medium-emphasis">Monthly Cost</div>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-16">
      <v-progress-circular
        indeterminate
        color="primary"
        size="64"
      />
      <p class="text-body-1 text-medium-emphasis mt-4">Loading your nodes...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="text-center py-16">
      <v-icon size="64" color="error" class="mb-4">mdi-alert-circle</v-icon>
      <h3 class="text-h5 mb-2">Failed to load nodes</h3>
      <p class="text-body-1 text-medium-emphasis mb-4">{{ error }}</p>
      <v-btn
        color="primary"
        variant="outlined"
        @click="fetchRentedNodes"
      >
        Try Again
      </v-btn>
    </div>

    <!-- Empty State -->
    <div v-else-if="rentedNodes.length === 0" class="text-center py-16">
      <v-icon size="64" color="primary" class="mb-4">mdi-server-network</v-icon>
      <h3 class="text-h5 mb-2">No nodes rented yet</h3>
      <p class="text-body-1 text-medium-emphasis mb-4">Start by reserving your first node to deploy your applications.</p>
      <v-btn
        color="primary"
        variant="elevated"
        @click="navigateToReserve"
      >
        Reserve Your First Node
      </v-btn>
    </div>

    <!-- Nodes Grid -->
    <v-row v-else>
      <v-col v-for="node in rentedNodes" :key="node.id" cols="12" sm="6" lg="4">
        <v-card color="surface-variant" class="pa-4 d-flex flex-column" height="100%">
          <div class="d-flex justify-space-between align-start mb-3">
            <h3 class="text-h6">Node {{ node.nodeId || node.id }}</h3>
            <div class="text-h6 text-success font-weight-bold">${{ node.price_usd?.toFixed(2) ?? 'N/A' }}/month</div>
          </div>

          <div v-if="node.country" class="d-flex align-center mb-3 text-medium-emphasis">
            <v-icon size="16" class="me-1">mdi-map-marker</v-icon>
            {{ node.country }}
          </div>

          <div class="mb-3">
            <div class="d-flex align-center mb-2">
              <v-icon size="18" class="me-2" color="primary">mdi-cpu-64-bit</v-icon>
              <span class="text-body-2 text-medium-emphasis me-2">CPU:</span>
              <span class="text-body-2">{{ Math.round(node.resources?.cpu ?? node.total_resources?.cru ?? 0) }} vCPU</span>
            </div>
            <div class="d-flex align-center mb-2">
              <v-icon size="18" class="me-2" color="success">mdi-memory</v-icon>
              <span class="text-body-2 text-medium-emphasis me-2">RAM:</span>
              <span class="text-body-2">{{ Math.round(node.resources?.memory ?? (node.total_resources?.mru ? node.total_resources.mru / (1024*1024*1024) : 0)) }} GB</span>
            </div>
            <div class="d-flex align-center mb-2">
              <v-icon size="18" class="me-2" color="info">mdi-harddisk</v-icon>
              <span class="text-body-2 text-medium-emphasis me-2">Storage:</span>
              <span class="text-body-2">{{ formatStorage(node.resources?.storage ?? (node.total_resources?.sru ? node.total_resources.sru / (1024*1024*1024) : 0)) }}</span>
            </div>
          </div>

          <div class="mb-4">
            <v-chip v-if="node.gpu || (node.gpus && node.gpus.length > 0)" color="deep-purple-accent-2" size="small" variant="elevated">
              <v-icon size="16" class="me-1">mdi-nvidia</v-icon>
              GPU
            </v-chip>
          </div>

          <v-spacer />

          <v-btn
            color="error"
            variant="outlined"
            @click="confirmUnreserve(node)"
            :loading="unreservingNode === node.rentContractId?.toString()"
            block
          >
            Unreserve
          </v-btn>
        </v-card>
      </v-col>
    </v-row>

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
  </v-card>
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
.kubecloud-gradient {
  background: linear-gradient(135deg, #3B82F6 0%, #60A5FA 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.kubecloud-glow-blue {
  text-shadow: 0 0 20px rgba(59, 130, 246, 0.3);
}
</style>
