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
            <v-tabs v-model="tab" class="manage-tabs" grow>
              <v-tab value="overview" class="tab-item">
                <v-icon icon="mdi-view-dashboard" class="mr-2"></v-icon>
                Overview
              </v-tab>
              <v-tab value="nodes" class="tab-item">
                <v-icon icon="mdi-server" class="mr-2"></v-icon>
                Nodes
              </v-tab>
              <v-tab value="storage" class="tab-item">
                <v-icon icon="mdi-database" class="mr-2"></v-icon>
                Storage
              </v-tab>
              <v-tab value="networking" class="tab-item">
                <v-icon icon="mdi-network" class="mr-2"></v-icon>
                Networking
              </v-tab>
              <v-tab value="settings" class="tab-item">
                <v-icon icon="mdi-cog" class="mr-2"></v-icon>
                Settings
              </v-tab>
            </v-tabs>
            <v-divider class="tab-divider"></v-divider>
            <v-window v-model="tab" class="tab-content">
              <v-window-item value="overview">
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
              </v-window-item>
              <!-- Nodes Tab with Add Worker -->
              <v-window-item value="nodes">
                <div class="card nodes-card">
                  <div class="nodes-header">
                    <h3 class="dashboard-card-title">
                      <v-icon icon="mdi-server" class="mr-2"></v-icon>
                      Nodes Management
                    </h3>
                    <v-btn color="primary" class="btn btn-primary" @click="showAddWorker = true">
                      <v-icon icon="mdi-plus" class="mr-2"></v-icon>
                      Add Worker
                    </v-btn>
                  </div>
                  <div class="nodes-content">
                    <div class="coming-soon-content">
                      <v-icon icon="mdi-alert-circle-outline" size="40" color="var(--color-primary)" class="mb-2"></v-icon>
                      <h4 class="coming-soon-title">Worker node management coming soon!</h4>
                      <p class="coming-soon-text">You'll soon be able to add, remove, and manage worker nodes in your cluster.</p>
                    </div>
                  </div>
                  <!-- Add Worker Modal -->
                  <v-dialog v-model="showAddWorker" max-width="400">
                    <v-card class="add-worker-modal">
                      <v-card-title class="modal-title">
                        <v-icon icon="mdi-account-plus" class="mr-2"></v-icon>
                        Add Worker Node
                      </v-card-title>
                      <v-card-text>
                        <div class="modal-coming-soon">
                          <v-icon icon="mdi-timer-sand" size="32" color="var(--color-primary)" class="mb-2"></v-icon>
                          <div class="modal-coming-soon-title">Coming Soon</div>
                          <div class="modal-coming-soon-text">Adding worker nodes will be available in a future update.</div>
                        </div>
                      </v-card-text>
                      <v-card-actions>
                        <v-spacer></v-spacer>
                        <v-btn color="primary" text @click="showAddWorker = false">Close</v-btn>
                      </v-card-actions>
                    </v-card>
                  </v-dialog>
                </div>
              </v-window-item>
              <!-- Storage Tab -->
              <v-window-item value="storage">
                <div class="card placeholder-card">
                  <div class="placeholder-content">
                    <v-icon icon="mdi-database" size="48" color="var(--color-primary)" class="mb-4"></v-icon>
                    <h4 class="placeholder-title">Storage Management</h4>
                    <p class="placeholder-text">Storage configuration and management features are coming soon.</p>
                  </div>
                </div>
              </v-window-item>
              <!-- Networking Tab -->
              <v-window-item value="networking">
                <div class="card placeholder-card">
                  <div class="placeholder-content">
                    <v-icon icon="mdi-network" size="48" color="var(--color-primary)" class="mb-4"></v-icon>
                    <h4 class="placeholder-title">Network Configuration</h4>
                    <p class="placeholder-text">Network settings and configuration options will be available soon.</p>
                  </div>
                </div>
              </v-window-item>
              <!-- Settings Tab -->
              <v-window-item value="settings">
                <div class="card placeholder-card">
                  <div class="placeholder-content">
                    <v-icon icon="mdi-cog" size="48" color="var(--color-primary)" class="mb-4"></v-icon>
                    <h4 class="placeholder-title">Cluster Settings</h4>
                    <p class="placeholder-text">Advanced cluster configuration and settings will be available soon.</p>
                  </div>
                </div>
              </v-window-item>
            </v-window>
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
const tab = ref('overview')
const showAddWorker = ref(false)

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

.manage-header-content {
  /* Remove max-width and margin, handled by .container */
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

.manage-content-wrapper {
  /* Remove max-width and margin, handled by .container */
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

.manage-tabs {
  background: transparent;
  padding: 0 var(--space-8);
  padding-top: var(--space-6);
}

.tab-item {
  color: var(--color-text-muted) !important;
  font-weight: var(--font-weight-medium) !important;
  text-transform: none !important;
  letter-spacing: normal !important;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0 !important;
  transition: all var(--transition-normal) !important;
  font-size: var(--font-size-base) !important;
}

.tab-item:hover {
  color: var(--color-primary) !important;
  background: var(--color-primary-subtle) !important;
}

.v-tab--selected {
  color: var(--color-primary) !important;
  background: var(--color-primary-subtle) !important;
  border-bottom: 2px solid var(--color-primary) !important;
}

.tab-divider {
  border-color: var(--color-border);
  margin: 0;
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

.nodes-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-6);
}

.nodes-content {
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.coming-soon-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--space-2);
}

.coming-soon-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
}

.coming-soon-text {
  color: var(--color-text-muted);
  font-size: var(--font-size-base);
}

.placeholder-card {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--space-4);
}

.placeholder-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
}

.placeholder-text {
  color: var(--color-text-muted);
  font-size: var(--font-size-base);
}

.add-worker-modal {
  background: var(--color-bg-elevated);
  border-radius: var(--radius-xl);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}

.modal-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.modal-coming-soon {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  text-align: center;
}

.modal-coming-soon-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text);
}

.modal-coming-soon-text {
  color: var(--color-text-muted);
  font-size: var(--font-size-base);
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
  
  .manage-tabs {
    padding: 0 var(--space-4);
    padding-top: var(--space-4);
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
  
  .manage-tabs {
    padding: 0 var(--space-3);
    padding-top: var(--space-3);
  }
}
</style>
