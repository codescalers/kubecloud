<template>
  <div class="nodes-card">
    <!-- Header Section -->
    <div class="card-header">
      <div class="header-content">
        <h2 class="card-titl">
          My Nodes
        </h2>
        <p class="card-description">
          Manage your rented nodes and their resources.
        </p>
      </div>
      <div class="header-actions" style="display: flex; gap: 0.5rem; align-items: center;">
        <v-btn
          color="primary"
          variant="outlined"
          :disabled="loading"
          prepend-icon="mdi-refresh"
          @click="fetchRentedNodes"
          class="refresh-btn"
          style="min-width: 110px;"
        >
          Refresh
        </v-btn>
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
        <v-col
          v-for="(stat, index) in statCards"
          :key="index"
          cols="12"
          sm="6"
          md="3"
        >
          <v-card class="stat-card" flat>
            <div class="stat-content">
              <div class="stat-icon">
                <v-icon size="32" :color="stat.color">{{ stat.icon }}</v-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stat.value() }}</div>
                <div class="stat-label">{{ stat.label }}</div>
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

    <div v-else class="nodes-section">
      <v-row class="nodes-grid" align="stretch">
        <v-col
          class="node-col"
          v-for="(node, idx) in normalizedNodes"
          :key="node.id"
          cols="12"
          sm="6"
          md="4"
          lg="4"
        >
          <NodeCard
            :node="node"
            :isAuthenticated="true"
            :loading="unreservingNodes.includes(node.nodeId)"
            :disabled="false"
            :buttonLabel="'Unreserve Node'"
            @action="handleNodeAction(node, $event)"
          />
        </v-col>
      </v-row>
    </div>

    <!-- Unreserve Confirmation Dialog -->
    <v-dialog v-model="showUnreserveDialog" max-width="400">
      <v-card class="pa-3">
        <v-card-title>Confirm Unreservation</v-card-title>
        <v-card-text>
          Are you sure you want to unreserve this node?
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
            :loading="(!!selectedNode && unreservingNodes.includes(selectedNode.nodeId))"
          >
            Unreserve
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useNodeManagement } from '@/composables/useNodeManagement'
import type { RentedNode } from '@/composables/useNodeManagement'
import { useNotificationStore } from '../../stores/notifications'
import NodeCard from '../NodeCard.vue'

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

const listeningToNodeUpdate = ref(false)
const notificationStore = useNotificationStore()

// Dialog state
const showUnreserveDialog = ref(false)
const selectedNode = ref<RentedNode | null>(null)
const unreservingNodes = ref<number[]>([])

// Handle node update events
const handleNodeUpdate = () => {
  fetchRentedNodes();
  unreservingNodes.value = unreservingNodes.value.filter(id => !rentedNodes.value.some(node => node.nodeId === id))
  if (unreservingNodes.value.length === 0 || rentedNodes.value.length === 0) {
    window.removeEventListener('node-update', handleNodeUpdate)
    listeningToNodeUpdate.value = false
  }
}

onMounted(() => {
  fetchRentedNodes()
})

onBeforeUnmount(() => {
  listeningToNodeUpdate.value = false
  window.removeEventListener('node-update', handleNodeUpdate)
})

const navigateToReserve = () => {
  router.push('/nodes')
}

const confirmUnreserve = (node: RentedNode) => {
  selectedNode.value = node
  showUnreserveDialog.value = true
}

const handleUnreserve = async () => {
  const nodeId = selectedNode.value?.nodeId || 0
  const contractId = selectedNode.value?.rentContractId
  if (!contractId) return
  unreservingNodes.value.push(nodeId)
  try {
    listeningToNodeUpdate.value = true
    window.addEventListener('node-update', handleNodeUpdate)
    await unreserveNode(contractId.toString(), contractId)
    showUnreserveDialog.value = false
    selectedNode.value = null
  } catch (err) {
    console.error('Failed to unreserve node. Please try again.')
    unreservingNodes.value.splice(unreservingNodes.value.indexOf(nodeId), 1)
  }
}

function handleNodeAction(node: any, payload: { nodeId: number; action: string }) {
  if (payload.action === 'unreserve') {
    const found = rentedNodes.value.find(n => n.nodeId === payload.nodeId);
    if (found) confirmUnreserve(found);
  }
}

const statCards = [
  {
    icon: 'mdi-server-network',
    color: 'primary',
    value: () => rentedNodes.value.length,
    label: 'Total Nodes'
  },
  {
    icon: 'mdi-check-circle',
    color: 'success',
    value: () => healthyNodes.value.length,
    label: 'Healthy'
  },
  {
    icon: 'mdi-alert-circle',
    color: 'warning',
    value: () => unhealthyNodes.value.length,
    label: 'Unhealthy'
  },
  {
    icon: 'mdi-currency-usd',
    color: 'info',
    value: () => totalMonthlyCost.value.toFixed(2),
    label: 'Monthly Cost'
  }
]

const normalizedNodes = computed(() =>
  rentedNodes.value.map(node => ({
    nodeId: node.nodeId,
    farmId: node.farmId,
    twinId: node.twinId,
    price_usd: node.price_usd ?? 'N/A',
    cpu: Math.round(node.total_resources?.cru ?? 0),
    ram: Math.round(node.total_resources?.mru ? node.total_resources.mru / (1024*1024*1024) : 0),
    storage: Math.round(node.total_resources?.sru ? node.total_resources.sru / (1024*1024*1024) : 0),
    country: node.country,
    gpu: !!node.num_gpu,
    id: node.id,
    extraFee: node.extraFee || 0,
    locationString: node.country || '',
    city: node.city || '',
    status: node.status || '',
    healthy: node.healthy ?? true,
    rentable: false,
    rented: true,
    dedicated: false,
    certificationType: '',
    discount_price: node.discount_price
  }))
);
</script>
<style scoped>
.nodes-card {
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
  font-size: 2rem;
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
  background: transparent !important;
  border: 1px solid #23263a !important;
  border-radius: 0.75rem !important;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.stat-card:hover {
  border-color: #3b82f6 !important;
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


.node-col {
  flex: 1 1 250px; /* Allow growing and shrinking with a basis of 250px */
  min-width: 250px; /* Enforce the minimum width */
  max-width: 400px;
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
    gap: 1rem;
  }
}
</style>
