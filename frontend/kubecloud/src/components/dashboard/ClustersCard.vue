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
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-icon">
            <v-icon icon="mdi-server" size="24" color="var(--color-primary)"></v-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">12</div>
            <div class="stat-label">Active Clusters</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon">
            <v-icon icon="mdi-cube-outline" size="24" color="var(--color-primary)"></v-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">156</div>
            <div class="stat-label">Total Nodes</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon">
            <v-icon icon="mdi-chart-line" size="24" color="var(--color-primary)"></v-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">99.9%</div>
            <div class="stat-label">Uptime</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon">
            <v-icon icon="mdi-eye" size="24" color="var(--color-primary)"></v-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">24/7</div>
            <div class="stat-label">Monitoring</div>
          </div>
        </div>
      </div>

      <!-- Recent Clusters -->
      <div class="recent-clusters">
        <h3 class="section-title">Recent Clusters</h3>
        <div class="cluster-list">
          <div
            v-for="cluster in recentClusters"
            :key="cluster.id"
            class="list-item-interactive"
            @click="viewCluster(cluster.id)"
          >
            <div class="cluster-info">
              <div class="cluster-name">{{ cluster.name }}</div>
              <div class="cluster-details">
                <span class="cluster-region">{{ cluster.region }}</span>
                <span>•</span>
                <span>{{ cluster.nodes }} nodes</span>
              </div>
            </div>
            <div class="cluster-status" :class="cluster.status.toLowerCase()">
              <span class="status-dot" :class="cluster.status.toLowerCase()"></span>
              {{ cluster.status }}
            </div>
            <div class="cluster-actions">
              <v-btn
                variant="outlined"
                size="small"
                class="btn btn-outline btn-sm"
                @click.stop="openMetrics()"
              >
                <v-icon icon="mdi-chart-line" size="16"></v-icon>
              </v-btn>
              <v-btn
                variant="outlined"
                size="small"
                class="btn btn-outline btn-sm"
                @click.stop="deleteCluster(cluster.id)"
              >
                <v-icon icon="mdi-delete" size="16"></v-icon>
              </v-btn>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="quick-actions">
        <v-btn
          variant="outlined"
          class="btn btn-outline"
          @click="viewAllClusters"
        >
          View All Clusters
        </v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// Mock data for recent clusters
const recentClusters = ref([
  {
    id: 1,
    name: 'Production API',
    status: 'Running',
    region: 'Ghent, Belgium',
    nodes: 3
  },
  {
    id: 2,
    name: 'Staging Environment',
    status: 'Running',
    region: 'Ghent, Belgium',
    nodes: 2
  },
  {
    id: 3,
    name: 'Development Cluster',
    status: 'Stopped',
    region: 'Ghent, Belgium',
    nodes: 1
  }
])

const viewCluster = (id: number) => {
  router.push(`/manage-cluster/${id}`)
}

const deleteCluster = (id: number) => {
  console.log('Delete cluster:', id)
}

const viewAllClusters = () => {
  console.log('View all clusters')
}

const openMetrics = () => {
  console.log('Open metrics')
}

const goToDeployCluster = () => {
  router.push('/deploy')
}
</script>

<style scoped>
.card-content {
  gap: var(--space-8);
  padding: unset !important;
}

.recent-clusters {
  margin-top: var(--space-4);
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  margin: 0 0 var(--space-4) 0;
}

.cluster-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.list-item-interactive {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) 0;
}

.cluster-info {
  flex: 1;
  min-width: 0;
}

.cluster-name {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
  font-size: var(--font-size-base);
}

.cluster-details {
  display: flex;
  gap: var(--space-4);
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.cluster-region {
  color: var(--color-primary);
}

.cluster-status {
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-left: var(--space-4);
}

.cluster-status.running {
  background: var(--color-success-subtle);
  color: var(--color-success);
  border: 1px solid var(--color-success);
}

.cluster-status.stopped {
  background: var(--color-error-subtle);
  color: var(--color-error);
  border: 1px solid var(--color-error);
}

.cluster-actions {
  margin-left: var(--space-4);
  display: flex;
  gap: var(--space-2);
}

.quick-actions {
  display: flex;
  justify-content: center;
  margin-top: var(--space-4);
}

/* Responsive Design */
@media (max-width: 768px) {
  .card-content {
    gap: var(--space-6);
  }

  .cluster-list {
    gap: var(--space-2);
  }

  .cluster-details {
    font-size: var(--font-size-xs);
  }

  .cluster-status {
    font-size: var(--font-size-xs);
  }

  .cluster-actions {
    gap: var(--space-1);
  }
}

@media (max-width: 480px) {
  .cluster-list {
    gap: var(--space-2);
  }

  .cluster-name {
    font-size: var(--font-size-sm);
  }

  .cluster-details {
    font-size: var(--font-size-xs);
  }

  .cluster-status {
    font-size: var(--font-size-xs);
  }

  .cluster-actions {
    gap: var(--space-1);
  }
}
</style>
