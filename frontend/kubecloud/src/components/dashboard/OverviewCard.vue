<template>
  <div class="overview-card card-enhanced">
    <div class="overview-header">
      <h2 class="overview-title">Dashboard Overview</h2>
      <p class="overview-subtitle">Your KubeCloud platform at a glance</p>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div
        v-for="(stat, index) in statsData"
        :key="index"
        class="stat-card"
      >
        <div class="stat-icon">
          <v-icon :icon="stat.icon" size="32" color="primary"></v-icon>
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
          :color="action.color"
          :variant="action.variant"
          class="action-btn btn-enhanced"
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
          class="activity-item"
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
          class="status-item"
        >
          <div class="status-indicator running"></div>
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
.overview-card {
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  box-shadow: none;
  padding: 1.25rem;
  transition: border-color var(--transition);
}

.overview-card:hover {
  background: var(--color-surface);
  border-color: var(--color-border);
}

.overview-header {
  text-align: center;
  margin-bottom: 1.5rem; /* Adjusted margin */
}

.overview-title {
  font-size: var(--text-3xl); /* Use standardized font size */
  font-weight: var(--font-bold); /* Use standardized font weight */
  margin-bottom: 0.5rem; /* Adjusted margin */
  color: var(--kubecloud-text-primary); /* Use primary text color */
}

.overview-subtitle {
  font-size: var(--text-base); /* Use standardized font size */
  color: var(--kubecloud-text-secondary); /* Use secondary text color */
  line-height: 1.5; /* Adjusted line height */
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); /* Adjusted minmax */
  gap: 1rem; /* Adjusted gap */
  margin-bottom: 1.5rem; /* Adjusted margin */
}

.stat-card,
.activity-item,
.status-item {
  background: var(--kubecloud-surface-light, #2e2f35);
  color: var(--kubecloud-text-primary, #e0e0e0);
  border-radius: var(--rounded-md);
  border: 1.5px solid var(--kubecloud-border, #3a3b40);
  box-shadow: 0 2px 16px 0 rgba(59,130,246,0.10), 0 1px 8px 0 rgba(0,0,0,0.12);
  transition: box-shadow 0.18s, border-color 0.18s, background 0.18s;
}

.stat-card:hover,
.activity-item:hover,
.status-item:hover {
  border-color: var(--kubecloud-primary, #3B82F6);
  box-shadow: 0 2px 16px 0 rgba(59,130,246,0.10), 0 1px 8px 0 rgba(0,0,0,0.12);
  background: var(--kubecloud-surface-light-hover, #34353b);
}

.stat-icon {
  margin-right: 0.75rem;
}

.stat-number {
  font-size: var(--text-2xl); /* Use standardized font size */
  font-weight: var(--font-bold); /* Use standardized font weight */
  margin-bottom: 0.15rem; /* Adjusted margin */
  color: var(--kubecloud-text-primary); /* Use primary text color */
}

.stat-label {
  font-size: var(--text-sm); /* Use standardized font size */
  color: var(--kubecloud-text-secondary, #b0b0b0); /* Use secondary text color */
  font-weight: var(--font-medium); /* Use standardized font weight */
}

/* Quick Actions */
.quick-actions-section {
  margin-bottom: 1.5rem; /* Adjusted margin */
}

.section-title {
  font-size: var(--text-lg); /* Use standardized font size */
  font-weight: var(--font-semibold); /* Use standardized font weight */
  margin-bottom: 0.75rem; /* Adjusted margin */
  color: var(--kubecloud-text-primary); /* Use primary text color */
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); /* Adjusted minmax */
  gap: 0.75rem; /* Adjusted gap */
}

.action-btn {
  height: 40px; /* Adjusted height */
  font-weight: var(--font-medium); /* Use standardized font weight */
  text-transform: none;
  letter-spacing: 0;
  padding: 0.5rem 1rem; /* Adjusted padding */
}

/* Recent Activity */
.recent-activity-section {
  margin-bottom: 1.5rem; /* Adjusted margin */
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem; /* Adjusted gap */
}

.activity-item {
  display: flex;
  align-items: center;
  padding: 0.75rem; /* Adjusted padding */
  background: var(--kubecloud-surface-light, #2e2f35); /* Lighter than card for inner contrast */
  color: var(--kubecloud-text-primary, #e0e0e0); /* Always light text */
  border-radius: var(--rounded-md); /* Use standardized border radius */
  border: 1.5px solid var(--kubecloud-border, #3a3b40); /* Use border color */
  box-shadow: 0 2px 16px 0 rgba(59,130,246,0.10), 0 1px 8px 0 rgba(0,0,0,0.12);
  transition: box-shadow 0.18s, border-color 0.18s, background 0.18s;
}

.activity-item:hover {
  border-color: var(--kubecloud-primary, #3B82F6); /* Use primary color on hover */
  box-shadow: 0 2px 16px 0 rgba(59,130,246,0.10), 0 1px 8px 0 rgba(0,0,0,0.12);
  background: var(--kubecloud-surface-light-hover, #34353b);
}

.activity-icon {
  margin-right: 0.5rem; /* Adjusted margin */
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px; /* Adjusted size */
  height: 28px; /* Adjusted size */
  border-radius: var(--rounded-sm); /* Use standardized border radius */
  background: var(--kubecloud-primary-bg, rgba(59, 130, 246, 0.1));
  transition: all var(--transition-fast); /* Use faster transition */
}

.activity-item:hover .activity-icon {
  background: var(--kubecloud-primary-bg-hover, rgba(59, 130, 246, 0.2));
  transform: none; /* Remove scale effect */
}

.activity-content {
  flex-grow: 1;
}

.activity-text {
  font-weight: var(--font-medium); /* Use standardized font weight */
  color: var(--kubecloud-text-primary); /* Use primary text color */
  margin-bottom: 0.15rem; /* Adjusted margin */
}

.activity-time {
  font-size: var(--text-xs); /* Use standardized font size */
  color: var(--kubecloud-text-secondary, #b0b0b0); /* Use secondary text color */
}

/* System Status */
.system-status-section {
  margin-bottom: 0; /* Adjusted margin */
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); /* Adjusted minmax */
  gap: 0.75rem; /* Adjusted gap */
}

.status-item {
  display: flex;
  align-items: center;
  padding: 0.75rem; /* Adjusted padding */
  background: var(--kubecloud-surface-light, #2e2f35); /* Lighter than card for inner contrast */
  color: var(--kubecloud-text-primary, #e0e0e0); /* Always light text */
  border-radius: var(--rounded-md); /* Use standardized border radius */
  border: 1.5px solid var(--kubecloud-border, #3a3b40); /* Use border color */
  box-shadow: 0 2px 16px 0 rgba(59,130,246,0.10), 0 1px 8px 0 rgba(0,0,0,0.12);
  transition: box-shadow 0.18s, border-color 0.18s, background 0.18s;
}

.status-item:hover {
  border-color: var(--kubecloud-primary, #3B82F6); /* Use primary color on hover */
  box-shadow: 0 2px 16px 0 rgba(59,130,246,0.10), 0 1px 8px 0 rgba(0,0,0,0.12);
  background: var(--kubecloud-surface-light-hover, #34353b);
}

.status-indicator {
  width: 10px; /* Adjusted size */
  height: 10px; /* Adjusted size */
  border-radius: 50%;
  margin-right: 0.5rem; /* Adjusted margin */
  transition: all var(--transition-fast); /* Use faster transition */
}

.status-indicator.running {
  background: var(--kubecloud-success);
  box-shadow: var(--kubecloud-success-shadow, 0 0 6px rgba(76, 175, 80, 0.5));
}

.status-item:hover .status-indicator.running {
  box-shadow: var(--kubecloud-success-shadow-hover, 0 0 8px rgba(76, 175, 80, 0.7));
  transform: none; /* Remove scale effect */
}

.status-label {
  font-size: var(--text-xs); /* Use standardized font size */
  color: var(--kubecloud-text-secondary, #b0b0b0); /* Use secondary text color */
  margin-bottom: 0.15rem; /* Adjusted margin */
}

.status-value {
  font-weight: var(--font-semibold); /* Use standardized font weight */
  color: var(--kubecloud-text-primary); /* Use primary text color */
}

/* Responsive Design - Align with updated sizes and padding */
@media (max-width: 960px) {
  .overview-card {
    padding: 1.25rem;
  }

  .overview-title {
    font-size: var(--text-2xl); /* Adjusted font size */
  }

  .stats-grid {
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); /* Adjusted minmax */
  }

  .actions-grid {
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); /* Adjusted minmax */
  }

  .status-grid {
    grid-template-columns: repeat(auto-fit, minmax(100px, 1fr)); /* Adjusted minmax */
  }
}

@media (max-width: 600px) {
  .overview-card {
    padding: 1rem;
  }

  .overview-title {
    font-size: var(--text-xl); /* Adjusted font size */
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .stat-card {
    padding: 0.8rem;
  }

  .stat-number {
    font-size: var(--text-lg); /* Adjusted font size */
  }

  .actions-grid {
    grid-template-columns: 1fr;
  }

  .status-grid {
    grid-template-columns: 1fr;
  }
}
</style>
