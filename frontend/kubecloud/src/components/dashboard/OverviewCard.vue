<template>
  <div class="dashboard-card">
    <div class="mb-8">
      <h3 class="dashboard-card-title">Dashboard Overview</h3>
      <p class="dashboard-card-subtitle">Your KubeCloud platform at a glance</p>
    </div>
    <!-- Stats Grid -->
    <StatsGrid :stats="statsData" />

    <!-- Quick Actions -->
    <div class="quick-actions-section">
      <h3 class="section-title">Quick Actions</h3>
      <div class="actions-grid">
        <v-btn
          v-for="(action, index) in quickActions"
          :key="index"
          variant="outlined"
          class="btn btn-outline"
          @click="action.handler"
        >
          <v-icon :icon="action.icon" class="mr-2"></v-icon>
          {{ action.label }}
        </v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../../stores/user'
import StatsGrid from '../StatsGrid.vue'

interface Cluster {
  id: number
  name: string
  status: string
  nodes: number
  region: string
}

interface Activity {
  id: number
  text: string
  time: string
  icon: string
  iconColor: string
}

interface SshKey {
  id: number
  name: string
  fingerprint: string
  addedDate: string
}

interface Props {
  clusters: Cluster[]
  sshKeys: SshKey[]
  totalSpent: string
  balance: number
}

const props = defineProps<Props>()
const router = useRouter()
const userStore = useUserStore()

const uptimeHours = computed(() => {
  return props.clusters
    .filter(cluster => cluster.status === 'running')
    .reduce((sum, cluster) => sum + cluster.nodes * 24, 0)
})

// Computed data for stats
const statsData = computed(() => [
  {
    icon: 'mdi-server',
    value: props.clusters.length,
    label: 'Active Clusters'
  },
  {
    icon: 'mdi-currency-usd',
    value: `$${userStore.netBalance.toFixed(2)}`,
    subvalue: userStore.pendingBalance > 0 ? `+$${userStore.pendingBalance.toFixed(2)} pending` : '',
    label: 'Balance'
  },
  {
    icon: 'mdi-currency-usd',
    value: `$${props.totalSpent}`,
    label: 'Total Spent'
  },
])

// Quick actions data
const quickActions = [
  {
    label: 'Deploy Cluster',
    icon: 'mdi-plus',
    color: 'primary',
    variant: 'elevated' as const,
    handler: () => router.push('/deploy')
  },
  {
    label: 'Reserve Node',
    icon: 'mdi-server-plus',
    color: 'secondary',
    variant: 'outlined' as const,
    handler: () => router.push('/nodes')
  },
  {
    label: 'Add SSH Key',
    icon: 'mdi-key-plus',
    color: 'primary',
    variant: 'outlined' as const,
    handler: () => emit('navigate', 'ssh')
  },
  {
    label: 'Add Funds',
    icon: 'mdi-credit-card-plus',
    color: 'secondary',
    variant: 'outlined' as const,
    handler: () => emit('navigate', 'add-funds')
  },
]
const emit = defineEmits(['navigate'])
</script>

<style scoped>
.dashboard-card-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
}

.dashboard-card-subtitle {
  font-size: var(--font-size-base);
  color: var(--color-primary);
  font-weight: var(--font-weight-medium);
  opacity: 0.9;
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin: 0 0 var(--space-4) 0;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4);
}

/* Responsive Design */
@media (max-width: 768px) {
  .actions-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: var(--space-3);
  }
}

@media (max-width: 480px) {
  .actions-grid {
    grid-template-columns: 1fr;
  }
}
</style>

