<template>
  <div class="manage-cluster-container">
    <div class="container">
      <v-container fluid class="pa-0">
        <div class="manage-header mb-6">
          <div class="manage-header-content">
            <div class="manage-navigation">
              <v-btn icon variant="text" class="back-btn mr-3" @click="goBack">
                <v-icon>mdi-chevron-left</v-icon>
              </v-btn>
              <div class="breadcrumb">
                <span class="breadcrumb-item">Clusters</span>
                <v-icon icon="mdi-chevron-right" class="breadcrumb-separator"></v-icon>
                <span class="breadcrumb-item active">{{ cluster.name }}</span>
              </div>
            </div>
            <h1 class="manage-title">{{ cluster.name }}</h1>
            <p class="manage-subtitle">Manage your Kubernetes cluster configuration and resources</p>
          </div>
        </div>
        
        <div class="manage-content-wrapper">
          <!-- Status Bar -->
          <div class="card status-bar">
            <div class="status-bar-content">
              <div class="status-indicator">
                <span class="status-dot" :class="cluster.status === 'Running' ? 'running' : 'stopped'"></span>
                <span class="status-text">{{ cluster.status }}</span>
              </div>
              <div class="status-actions">
                <v-btn variant="outlined" class="btn btn-outline">
                  <v-icon icon="mdi-download" class="mr-2"></v-icon>
                  Download Kubeconfig
                </v-btn>
                <v-btn variant="outlined" class="btn btn-outline">
                  <v-icon icon="mdi-view-dashboard" class="mr-2"></v-icon>
                  Open Dashboard
                </v-btn>
              </div>
            </div>
          </div>

          <!-- Main Content Card -->
          <div class="card main-content-card">
            <div class="tab-content">
              <div class="overview-grid">
                <div class="card overview-card">
                  <h3 class="dashboard-card-title">
                    <v-icon icon="mdi-server" class="mr-2"></v-icon>
                    Cluster Resources
                  </h3>
                  <div class="resource-list">
                    <div class="resource-item">
                      <span class="resource-label">Nodes:</span>
                      <span class="resource-value">3</span>
                    </div>
                    <div class="resource-item">
                      <span class="resource-label">vCPU:</span>
                      <span class="resource-value">12</span>
                    </div>
                    <div class="resource-item">
                      <span class="resource-label">RAM:</span>
                      <span class="resource-value">24 GB</span>
                    </div>
                    <div class="resource-item">
                      <span class="resource-label">Storage:</span>
                      <span class="resource-value">1 TB</span>
                    </div>
                  </div>
                </div>
                <div class="card overview-card">
                  <h3 class="dashboard-card-title">
                    <v-icon icon="mdi-chart-line" class="mr-2"></v-icon>
                    Live Usage
                  </h3>
                  <div class="usage-metrics">
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">CPU</span>
                        <span class="usage-value">70%</span>
                      </div>
                      <v-progress-linear 
                        :model-value="70" 
                        color="var(--color-primary)" 
                        height="8" 
                        rounded
                        class="usage-bar"
                      ></v-progress-linear>
                    </div>
                    <div class="usage-item">
                      <div class="usage-header">
                        <span class="usage-label">Memory</span>
                        <span class="usage-value">50%</span>
                      </div>
                      <v-progress-linear 
                        :model-value="50" 
                        color="var(--color-primary)" 
                        height="8" 
                        rounded
                        class="usage-bar"
                      ></v-progress-linear>
                    </div>
                  </div>
                </div>
                <div class="card overview-card details-card">
                  <h3 class="dashboard-card-title">
                    <v-icon icon="mdi-information" class="mr-2"></v-icon>
                    Cluster Details
                  </h3>
                  <div class="details-grid">
                    <div class="detail-item">
                      <span class="detail-label">API Endpoint:</span>
                      <span class="detail-value font-mono">ygg://[mycelium-ip-address]</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Location:</span>
                      <span class="detail-value">Ghent, Belgium</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Created:</span>
                      <span class="detail-value">2023-07-01</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Est. Cost:</span>
                      <span class="detail-value">$125/month</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </v-container>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// Mock cluster data
const cluster = ref({
  name: 'Production Cluster',
  status: 'Running'
})

const goBack = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
.manage-cluster-container {
  min-height: 100vh;
  background: var(--color-bg);
  padding: 0;
}

.manage-header {
  margin-bottom: var(--space-8);
}

.manage-navigation {
  display: flex;
  align-items: center;
  margin-bottom: var(--space-4);
}

.back-btn {
  color: var(--color-text-secondary) !important;
}

.back-btn:hover {
  color: var(--color-primary) !important;
  background: var(--color-primary-subtle) !important;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.breadcrumb-item {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
}

.breadcrumb-item.active {
  color: var(--color-primary);
  font-weight: var(--font-weight-medium);
}

.breadcrumb-separator {
  color: var(--color-text-muted) !important;
  font-size: var(--font-size-sm) !important;
}

.manage-title {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text);
  margin: 0 0 var(--space-2) 0;
}

.manage-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-secondary);
  margin: 0;
}

.status-bar {
  margin-bottom: var(--space-6);
}

.status-bar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.status-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.status-dot.running {
  background: var(--color-success);
}

.status-dot.stopped {
  background: var(--color-error);
}

.status-text {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
}

.status-actions {
  display: flex;
  gap: var(--space-3);
}

.main-content-card {
  padding: unset !important;
  overflow: hidden;
}

.tab-content {
  padding: var(--space-10) var(--space-8) var(--space-8) var(--space-8);
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: var(--space-8);
}

.overview-card {
  height: 100%;
}

.details-card {
  grid-column: 1 / -1;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.resource-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) 0;
  border-bottom: 1px solid var(--color-border);
}

.resource-item:last-child {
  border-bottom: none;
}

.resource-label {
  color: var(--color-text-muted);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-base);
}

.resource-value {
  color: var(--color-text);
  font-weight: var(--font-weight-semibold);
  font-family: 'Inter', monospace;
  font-size: var(--font-size-base);
}

.usage-metrics {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.usage-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.usage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.usage-label {
  color: var(--color-text-muted);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-base);
}

.usage-value {
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-base);
}

.usage-bar {
  border-radius: var(--radius-sm);
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-4);
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-3);
  background: var(--color-primary-subtle);
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-md);
}

.detail-label {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.detail-value {
  color: var(--color-text);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-base);
}

.font-mono {
  font-family: 'Inter', monospace;
  font-size: var(--font-size-sm);
}

@media (max-width: 1100px) {
  .overview-grid {
    grid-template-columns: 1fr;
  }
  .details-card {
    grid-column: auto;
  }
}

@media (max-width: 768px) {
  .manage-cluster-container {
    padding: var(--space-4);
  }
  
  .manage-title {
    font-size: var(--font-size-2xl);
  }
  
  .manage-subtitle {
    font-size: var(--font-size-base);
  }
  
  .status-bar-content {
    flex-direction: column;
    gap: var(--space-4);
    align-items: flex-start;
  }
  
  .status-actions {
    width: 100%;
    justify-content: flex-start;
  }
  
  .tab-content {
    padding: var(--space-6) var(--space-4) var(--space-4) var(--space-4);
  }
  
  .overview-grid {
    gap: var(--space-4);
  }
  
  .details-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .manage-cluster-container {
    padding: var(--space-3);
  }
  
  .manage-title {
    font-size: var(--font-size-xl);
  }
  
  .status-actions {
    flex-direction: column;
    gap: var(--space-2);
  }
  
  .tab-content {
    padding: var(--space-4) var(--space-3) var(--space-3) var(--space-3);
  }
}
</style>
