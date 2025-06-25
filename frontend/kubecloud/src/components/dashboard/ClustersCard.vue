<template>
  <v-card class="clusters-card card-enhanced">
    <div class="card-header">
      <div class="card-title-section">
        <v-icon icon="mdi-server-network" size="32" color="primary" class="card-icon"></v-icon>
        <div class="card-title-content">
          <h3 class="card-title">Kubernetes Clusters</h3>
          <p class="card-subtitle">Manage your cloud-native infrastructure</p>
        </div>
      </div>
      <v-btn 
        color="secondary" 
        variant="elevated" 
        size="small" 
        class="action-btn kubecloud-glow-orange"
        @click="goToDeployCluster"
      >
        <v-icon icon="mdi-plus" size="16" class="mr-1"></v-icon>
        New Cluster
      </v-btn>
    </div>

    <div class="card-content">
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-number kubecloud-gradient">12</div>
          <div class="stat-label">Active Clusters</div>
        </div>
        <div class="stat-item">
          <div class="stat-number kubecloud-gradient">156</div>
          <div class="stat-label">Total Nodes</div>
        </div>
        <div class="stat-item">
          <div class="stat-number kubecloud-gradient">99.9%</div>
          <div class="stat-label">Uptime</div>
        </div>
        <div class="stat-item">
          <div class="stat-number kubecloud-gradient">24/7</div>
          <div class="stat-label">Monitoring</div>
        </div>
      </div>

      <div class="recent-clusters">
        <h4 class="section-title">Recent Clusters</h4>
        <div class="cluster-list">
          <div 
            v-for="cluster in recentClusters" 
            :key="cluster.id" 
            class="cluster-item"
            @click="viewCluster(cluster.id)"
          >
            <div class="cluster-info">
              <div class="cluster-name">{{ cluster.name }}</div>
              <div class="cluster-details">
                <span class="cluster-region">{{ cluster.region }}</span>
                <span class="cluster-status" :class="cluster.status">
                  {{ cluster.status }}
                </span>
              </div>
            </div>
            <div class="cluster-actions">
              <v-btn 
                icon 
                variant="text" 
                size="small" 
                color="primary"
                class="cluster-action-btn kubecloud-hover-blue"
                @click.stop="editCluster(cluster.id)"
              >
                <v-icon icon="mdi-pencil" size="16"></v-icon>
              </v-btn>
              <v-btn 
                icon 
                variant="text" 
                size="small" 
                color="error"
                class="cluster-action-btn"
                @click.stop="deleteCluster(cluster.id)"
              >
                <v-icon icon="mdi-delete" size="16"></v-icon>
              </v-btn>
            </div>
          </div>
        </div>
      </div>

      <div class="quick-actions">
        <v-btn 
          variant="outlined" 
          color="primary" 
          size="small" 
          class="quick-action-btn kubecloud-hover-blue"
          @click="viewAllClusters"
        >
          View All Clusters
        </v-btn>
        <v-btn 
          variant="outlined" 
          color="primary" 
          size="small" 
          class="quick-action-btn kubecloud-hover-blue"
          @click="openMetrics"
        >
          View Metrics
        </v-btn>
      </div>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const recentClusters = ref([
  {
    id: 1,
    name: 'production-cluster-01',
    region: 'us-west-2',
    status: 'running'
  },
  {
    id: 2,
    name: 'staging-cluster-01',
    region: 'eu-west-1',
    status: 'running'
  },
  {
    id: 3,
    name: 'dev-cluster-01',
    region: 'us-east-1',
    status: 'stopped'
  }
])

const createCluster = () => {
  console.log('Create new cluster')
}

const viewCluster = (id: number) => {
  console.log('View cluster:', id)
}

const editCluster = (id: number) => {
  console.log('Edit cluster:', id)
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
.clusters-card {
  height: 100%;
  padding: 1.5rem;
  min-width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.card-title-section {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.card-icon {
  background: rgba(79, 70, 229, 0.1);
  border-radius: 12px;
  padding: 0.75rem;
  box-shadow: var(--shadow-kubecloud-blue);
}

.card-title-content {
  display: flex;
  flex-direction: column;
}

.card-title {
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--kubecloud-white);
  margin-bottom: 0.25rem;
  line-height: 1.2;
}

.card-subtitle {
  font-size: 0.95rem;
  color: var(--kubecloud-light-gray);
  font-weight: 400;
  line-height: 1.4;
}

.action-btn {
  font-weight: 600;
  text-transform: none;
  letter-spacing: 0.01em;
  padding: 0.625rem 1.25rem;
  height: 44px;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 1rem;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem;
  background: var(--glass-bg-light);
  border-radius: 12px;
  border: 1px solid rgba(79, 70, 229, 0.1);
  transition: all var(--transition-normal);
}

.stat-item:hover {
  border-color: rgba(79, 70, 229, 0.3);
  transform: translateY(-2px);
  box-shadow: var(--shadow-kubecloud-blue);
}

.stat-number {
  font-size: 1.75rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
  line-height: 1;
}

.stat-label {
  font-size: 0.85rem;
  color: var(--kubecloud-light-gray);
  font-weight: 500;
  text-align: center;
}

.recent-clusters {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--kubecloud-white);
  margin-bottom: 0.5rem;
}

.cluster-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.cluster-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--glass-bg-light);
  border-radius: 8px;
  border: 1px solid rgba(79, 70, 229, 0.1);
  transition: all var(--transition-normal);
  cursor: pointer;
}

.cluster-item:hover {
  border-color: rgba(79, 70, 229, 0.3);
  transform: translateX(4px);
  box-shadow: var(--shadow-sm);
}

.cluster-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.cluster-name {
  font-weight: 600;
  color: var(--kubecloud-white);
  font-size: 0.95rem;
}

.cluster-details {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.cluster-region {
  font-size: 0.8rem;
  color: var(--kubecloud-light-gray);
  font-weight: 500;
}

.cluster-status {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.125rem 0.5rem;
  border-radius: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.cluster-status.running {
  background: rgba(16, 185, 129, 0.2);
  color: #10B981;
}

.cluster-status.stopped {
  background: rgba(239, 68, 68, 0.2);
  color: #EF4444;
}

.cluster-status.pending {
  background: rgba(245, 158, 11, 0.2);
  color: #F59E0B;
}

.cluster-actions {
  display: flex;
  gap: 0.5rem;
}

.cluster-action-btn {
  opacity: 0.7;
  transition: all var(--transition-normal);
}

.cluster-action-btn:hover {
  opacity: 1;
  transform: scale(1.1);
}

.quick-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-start;
  margin-top: 0.5rem;
}

.quick-action-btn {
  font-weight: 600;
  text-transform: none;
  letter-spacing: 0.01em;
  padding: 0.5rem 1rem;
  height: 40px;
}

/* Responsive Design */
@media (max-width: 960px) {
  .clusters-card {
    padding: 1.25rem;
  }
  
  .card-header {
    margin-bottom: 1.25rem;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .quick-actions {
    flex-direction: column;
  }
}

@media (max-width: 600px) {
  .clusters-card {
    padding: 1rem;
  }
  
  .card-header {
    flex-direction: column;
    gap: 1rem;
    align-items: flex-start;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .stat-item {
    padding: 0.875rem;
  }
  
  .cluster-item {
    padding: 0.875rem;
  }
  
  .cluster-details {
    flex-direction: column;
    gap: 0.25rem;
    align-items: flex-start;
  }
}
</style>
