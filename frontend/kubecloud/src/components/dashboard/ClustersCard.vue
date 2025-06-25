<template>
  <div class="dashboard-card clusters-card">
    <div class="card-header">
      <div class="card-title-section">
        <v-icon icon="mdi-server-network" size="32" color="accent" class="card-icon"></v-icon>
        <div class="card-title-content">
          <h3 class="dashboard-card-title">Kubernetes Clusters</h3>
          <p class="card-subtitle text-muted">Manage your cloud-native infrastructure</p>
        </div>
      </div>
      <v-btn class="action-btn" @click="goToDeployCluster">
        <v-icon icon="mdi-plus" size="16" class="mr-1"></v-icon>
        New Cluster
      </v-btn>
    </div>
    <div class="card-content">
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-number">12</div>
          <div class="stat-label text-muted">Active Clusters</div>
        </div>
        <div class="stat-item">
          <div class="stat-number">156</div>
          <div class="stat-label text-muted">Total Nodes</div>
        </div>
        <div class="stat-item">
          <div class="stat-number">99.9%</div>
          <div class="stat-label text-muted">Uptime</div>
        </div>
        <div class="stat-item">
          <div class="stat-number">24/7</div>
          <div class="stat-label text-muted">Monitoring</div>
        </div>
      </div>
      <div class="recent-clusters">
        <h4 class="section-title">Recent Clusters</h4>
        <div class="cluster-list">
          <div v-for="cluster in recentClusters" :key="cluster.id" class="cluster-item" @click="viewCluster(cluster.id)">
            <div class="cluster-info">
              <div class="cluster-name">{{ cluster.name }}</div>
              <div class="cluster-details text-muted">
                <span class="cluster-region">{{ cluster.region }}</span>
                <span class="cluster-status" :class="cluster.status">
                  {{ cluster.status }}
                </span>
              </div>
            </div>
            <div class="cluster-actions">
              <v-btn icon class="action-btn" @click.stop="editCluster(cluster.id)">
                <v-icon icon="mdi-pencil" size="16"></v-icon>
              </v-btn>
              <v-btn icon class="action-btn" @click.stop="deleteCluster(cluster.id)">
                <v-icon icon="mdi-delete" size="16"></v-icon>
              </v-btn>
            </div>
          </div>
        </div>
      </div>
      <div class="quick-actions">
        <v-btn class="action-btn" @click="viewAllClusters">
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
.dashboard-card.clusters-card {
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  box-shadow: none;
  color: var(--color-text);
  padding: 1.25rem;
  min-width: 100%;
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.5rem;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}
.card-title-section {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.card-title-content {
  display: flex;
  flex-direction: column;
}
.card-subtitle {
  font-size: 1rem;
  color: var(--color-text-muted);
  font-weight: 500;
}
.card-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 1rem;
}
.stat-item {
  background: var(--color-surface);
  color: var(--color-text);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  box-shadow: none;
  padding: 1rem;
  transition: border-color var(--transition), background var(--transition);
}
.stat-item:hover {
  border-color: var(--color-accent);
}
.stat-number {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-text);
}
.stat-label {
  font-size: 0.9rem;
  color: var(--color-text-muted);
  font-weight: 500;
}
.section-title {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--color-text);
}
.recent-clusters {
  margin-top: 1.5rem;
}
.cluster-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.cluster-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--color-surface);
  color: var(--color-text);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  box-shadow: none;
  padding: 0.75rem 1rem;
  cursor: pointer;
  transition: border-color var(--transition), background var(--transition);
}
.cluster-item:hover {
  border-color: var(--color-accent);
}
.cluster-name {
  font-weight: 600;
  color: var(--color-text);
}
.cluster-details {
  font-size: 0.9rem;
  color: var(--color-text-muted);
  display: flex;
  gap: 0.5rem;
}
.cluster-status {
  font-weight: 600;
  color: var(--color-accent);
}
.cluster-actions {
  display: flex;
  gap: 0.5rem;
}
.action-btn {
  background: var(--color-accent);
  color: #fff;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none;
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  transition: background var(--transition);
}
.action-btn:hover {
  background: var(--color-accent-hover);
}
.quick-actions {
  margin-top: 1.5rem;
  display: flex;
  gap: 0.75rem;
}
</style>
