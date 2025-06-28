<template>
  <div>
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Dashboard Overview</h3>
          <p class="dashboard-card-subtitle">Your KubeCloud platform at a glance</p>
        </div>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div
        v-for="(stat, index) in statsData"
        :key="index"
        class="stat-item"
      >
        <div class="stat-icon">
          <v-icon :icon="stat.icon" size="24" color="var(--color-primary)"></v-icon>
        </div>
        <div class="stat-content">
          <div class="stat-number">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>
    </div>

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

    <!-- Recent Activity -->
    <div class="recent-activity-section">
      <h3 class="section-title">Recent Activity</h3>
      <div class="activity-list">
        <div
          v-for="activity in recentActivity"
          :key="activity.id"
          class="list-item-interactive"
        >
          <div class="activity-icon">
            <v-icon :icon="activity.icon" size="20" :color="activity.iconColor"></v-icon>
          </div>
          <div class="activity-content">
            <div class="activity-text">{{ activity.text }}</div>
            <div class="activity-time">{{ activity.time }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- System Status -->
    <div class="system-status-section">
      <h3 class="section-title">System Status</h3>
      <div class="status-grid">
        <div
          v-for="(status, index) in systemStatus"
          :key="index"
          class="list-item-interactive"
        >
          <div class="status-dot running"></div>
          <div class="status-content">
            <div class="status-label">{{ status.label }}</div>
            <div class="status-value">{{ status.value }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

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
  recentActivity: Activity[]
  sshKeys: SshKey[]
  totalSpent: string
}

const props = defineProps<Props>()
const router = useRouter()

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
    icon: 'mdi-clock-outline',
    value: uptimeHours.value,
    label: 'Uptime Hours'
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
    handler: () => router.push('/reserve')
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

// System status data
const systemStatus = [
  {
    label: 'Platform',
    value: 'Operational'
  },
  {
    label: 'API',
    value: 'Healthy'
  },
  {
    label: 'Networking',
    value: 'Stable'
  },
  {
    label: 'Storage',
    value: 'Available'
  }
]

const emit = defineEmits(['navigate'])
</script>

<style scoped>
.dashboard-card-header {
  text-align: center;
  margin-bottom: var(--space-8);
}

.dashboard-card-title-section {
  display: flex;
  align-items: center;
  justify-content: center;
}

.dashboard-card-title-content {
  text-align: center;
}

.dashboard-card-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin: 0 0 var(--space-2) 0;
}

.dashboard-card-subtitle {
  font-size: var(--font-size-base);
  color: var(--color-primary);
  font-weight: var(--font-weight-medium);
  opacity: 0.9;
}

.quick-actions-section {
  margin-bottom: var(--space-8);
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

.recent-activity-section {
  margin-bottom: var(--space-8);
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.activity-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: var(--color-primary-subtle);
  border: 1px solid var(--color-primary);
}

.activity-content {
  flex: 1;
}

.activity-text {
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
  margin: 0 0 var(--space-1) 0;
}

.activity-time {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin: 0;
}

.system-status-section {
  margin-bottom: var(--space-4);
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: var(--space-4);
}

.status-content {
  flex: 1;
}

.status-label {
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
  margin: 0 0 var(--space-1) 0;
}

.status-value {
  font-size: var(--font-size-sm);
  color: var(--color-primary);
}

/* Responsive Design */
@media (max-width: 768px) {
  .dashboard-card-header {
    margin-bottom: var(--space-6);
  }

  .actions-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: var(--space-3);
  }

  .status-grid {
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: var(--space-3);
  }
}

@media (max-width: 480px) {
  .actions-grid {
    grid-template-columns: 1fr;
  }

  .status-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>

