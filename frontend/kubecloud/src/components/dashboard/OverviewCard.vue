<template>
  <div class="overview-card card-enhanced">
    <div class="overview-header">
      <h2 class="overview-title kubecloud-gradient kubecloud-glow-blue">Dashboard Overview</h2>
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
          <div class="stat-number kubecloud-gradient">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions-section">
      <h3 class="section-title kubecloud-gradient">Quick Actions</h3>
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
      <h3 class="section-title kubecloud-gradient">Recent Activity</h3>
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
      <h3 class="section-title kubecloud-gradient">System Status</h3>
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
  padding: 1.5rem;
}

.overview-header {
  text-align: center;
  margin-bottom: 2rem;
}

.overview-title {
  font-size: 2.25rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.overview-subtitle {
  font-size: 1.1rem;
  color: var(--kubecloud-light-gray);
  line-height: 1.6;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 1.25rem;
  background: var(--glass-bg-light);
  border-radius: 12px;
  border: 1px solid rgba(79, 70, 229, 0.2);
  transition: all var(--transition-normal);
}

.stat-card:hover {
  border-color: var(--kubecloud-blue);
  transform: translateY(-2px);
  box-shadow: var(--shadow-kubecloud-blue);
}

.stat-icon {
  margin-right: 0.75rem;
}

.stat-number {
  font-size: 1.75rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
}

.stat-label {
  font-size: 0.9rem;
  color: var(--kubecloud-light-gray);
  font-weight: 500;
}

/* Quick Actions */
.quick-actions-section {
  margin-bottom: 2rem;
}

.section-title {
  font-size: 1.375rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.75rem;
}

.action-btn {
  height: 44px;
  font-weight: 600;
  text-transform: none;
  letter-spacing: 0.01em;
  padding: 0.625rem 1.25rem;
}

/* Recent Activity */
.recent-activity-section {
  margin-bottom: 2rem;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.activity-item {
  display: flex;
  align-items: center;
  padding: 0.875rem;
  background: var(--glass-bg-light);
  border-radius: 8px;
  border: 1px solid rgba(79, 70, 229, 0.1);
  transition: all var(--transition-normal);
}

.activity-item:hover {
  border-color: rgba(79, 70, 229, 0.3);
  transform: translateX(4px);
  box-shadow: var(--shadow-sm);
}

.activity-icon {
  margin-right: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: rgba(79, 70, 229, 0.1);
  transition: all var(--transition-normal);
}

.activity-item:hover .activity-icon {
  background: rgba(79, 70, 229, 0.2);
  transform: scale(1.05);
}

.activity-content {
  flex-grow: 1;
}

.activity-text {
  font-weight: 500;
  color: var(--kubecloud-white);
  margin-bottom: 0.25rem;
}

.activity-time {
  font-size: 0.85rem;
  color: var(--kubecloud-light-gray);
}

/* System Status */
.system-status-section {
  margin-bottom: 0.5rem;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.75rem;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 0.875rem;
  background: var(--glass-bg-light);
  border-radius: 8px;
  border: 1px solid rgba(79, 70, 229, 0.1);
  transition: all var(--transition-normal);
}

.status-item:hover {
  border-color: rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.status-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-right: 0.75rem;
  transition: all var(--transition-normal);
}

.status-indicator.running {
  background: var(--kubecloud-success);
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.status-item:hover .status-indicator.running {
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.7);
  transform: scale(1.1);
}

.status-label {
  font-size: 0.85rem;
  color: var(--kubecloud-light-gray);
  margin-bottom: 0.25rem;
}

.status-value {
  font-weight: 600;
  color: var(--kubecloud-white);
}

/* Responsive Design */
@media (max-width: 960px) {
  .overview-card {
    padding: 1.25rem;
  }
  
  .overview-title {
    font-size: 2rem;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .actions-grid {
    grid-template-columns: 1fr;
  }
  
  .status-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .overview-card {
    padding: 1rem;
  }
  
  .overview-title {
    font-size: 1.75rem;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .stat-card {
    padding: 1rem;
  }
  
  .stat-number {
    font-size: 1.5rem;
  }
  
  .status-grid {
    grid-template-columns: 1fr;
  }
}
</style> 