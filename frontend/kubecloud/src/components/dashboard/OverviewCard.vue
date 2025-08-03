<template>
  <v-card color="surface-variant" class="pa-6">
    <div class="mb-8">
      <h3 class="text-h5 font-weight-semibold mb-2">Dashboard Overview</h3>
      <p class="text-body-1 text-medium-emphasis">Your KubeCloud platform at a glance</p>
    </div>
    <!-- Stats Grid -->
    <StatsGrid :stats="statsData" />

    <!-- Quick Actions -->
    <div class="mt-8">
      <h3 class="text-h6 font-weight-medium mb-4">Quick Actions</h3>
      <div class="d-flex flex-wrap ga-3">
        <v-btn
          v-for="(action, index) in quickActions"
          :key="index"
          variant="outlined"
          color="primary"
          @click="action.handler"
        >
          <v-icon :icon="action.icon" class="me-2"></v-icon>
          {{ action.label }}
        </v-btn>
      </div>
    </div>
  </v-card>
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
    label: 'Balance'
  },
  {
    icon: 'mdi-currency-usd',
    value: `$${props.totalSpent}`,
    label: 'Total Spent'
  },
  {
    icon: 'mdi-key',
    value: props.sshKeys.length,
    label: 'SSH Keys'
  }
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
    label: 'Add Payment',
    icon: 'mdi-credit-card-plus',
    color: 'secondary',
    variant: 'outlined' as const,
    handler: () => emit('navigate', 'payment')
  }
]
const emit = defineEmits(['navigate'])
</script>


