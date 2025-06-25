<template>
  <v-container class="manage-cluster-container" fluid>
    <v-row no-gutters>
      <!-- Main Content -->
      <v-col cols="12" class="main-content-area">
        <div class="manage-content-wrapper">
          <div class="manage-header d-flex align-center mb-6">
            <v-btn icon variant="text" class="mr-2" @click="goBack">
              <v-icon>mdi-chevron-left</v-icon>
            </v-btn>
            <span class="text-h5 font-weight-bold">Clusters</span>
            <span class="mx-2">/</span>
            <span class="text-h5 font-weight-bold">{{ cluster.name }}</span>
          </div>
          <div class="dashboard-card manage-status-card">
            <div class="d-flex flex-wrap align-center status-row">
              <span class="mr-4 status-label">Status:
                <span :class="cluster.status === 'Running' ? 'text-accent' : 'text-muted'">{{ cluster.status }}</span>
                <span class="status-dot ml-1" :class="cluster.status === 'Running' ? 'dot-running' : 'dot-stopped'">●</span>
              </span>
              <v-btn class="action-btn mr-2">Download Kubeconfig</v-btn>
              <v-btn class="action-btn">Open Dashboard</v-btn>
            </div>
          </div>
          <div class="dashboard-card manage-main-card pa-0">
            <v-tabs v-model="tab" class="manage-tabs px-6 pt-4" grow>
              <v-tab value="overview">Overview</v-tab>
              <v-tab value="nodes">Nodes</v-tab>
              <v-tab value="storage">Storage</v-tab>
              <v-tab value="networking">Networking</v-tab>
              <v-tab value="settings">Settings</v-tab>
            </v-tabs>
            <v-divider class="mb-0" />
            <v-window v-model="tab" class="px-6 pb-6 pt-4">
              <v-window-item value="overview">
                <v-row class="mb-0">
                  <v-col cols="12" md="6">
                    <div class="dashboard-card overview-card mb-4">
                      <h3 class="dashboard-card-title mb-2">Cluster Resources</h3>
                      <div>3 Nodes</div>
                      <div>12 vCPU</div>
                      <div>24 GB RAM</div>
                      <div>1 TB Storage</div>
                    </div>
                  </v-col>
                  <v-col cols="12" md="6">
                    <div class="dashboard-card overview-card mb-4">
                      <h3 class="dashboard-card-title mb-2">Live Usage</h3>
                      <div class="mb-2">CPU: <v-progress-linear :model-value="70" color="accent" height="12" rounded></v-progress-linear> <span class="ml-2">70%</span></div>
                      <div>MEM: <v-progress-linear :model-value="50" color="accent" height="12" rounded></v-progress-linear> <span class="ml-2">50%</span></div>
                    </div>
                  </v-col>
                </v-row>
                <div class="dashboard-card overview-card mb-0">
                  <h3 class="dashboard-card-title mb-2">Details</h3>
                  <div>API Endpoint: <span class="font-mono">ygg://[mycelium-ip-address]</span></div>
                  <div>Location: Ghent, Belgium</div>
                  <div>Created: 2023-07-01</div>
                  <div>Est. Cost: $125/month</div>
                </div>
              </v-window-item>
              <v-window-item value="nodes">
                <div class="dashboard-card overview-card">Nodes tab content (coming soon)</div>
              </v-window-item>
              <v-window-item value="storage">
                <div class="dashboard-card overview-card">Storage tab content (coming soon)</div>
              </v-window-item>
              <v-window-item value="networking">
                <div class="dashboard-card overview-card">Networking tab content (coming soon)</div>
              </v-window-item>
              <v-window-item value="settings">
                <div class="dashboard-card overview-card">Settings tab content (coming soon)</div>
              </v-window-item>
            </v-window>
          </div>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
const tab = ref('overview')
const cluster = ref({
  name: 'Production API',
  status: 'Running',
})
function goBack() {
  window.history.back()
}
function goToSection(section: string) {
  // Implement navigation logic if needed
}
function logout() {
  // Implement logout logic
}
</script>

<style scoped>
.manage-cluster-container {
  background: var(--color-bg);
  min-height: 100vh;
  padding: 0;
}
.main-content-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  min-height: 100vh;
}
.manage-content-wrapper {
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 1.5rem;
}
.manage-header {
  position: sticky;
  top: 0;
  z-index: 2;
  background: var(--color-bg);
  padding-bottom: 1rem;
  margin-bottom: 1.5rem;
  margin-top: 0.5rem;
  justify-content: flex-start;
  border-bottom: 1px solid var(--color-border);
}
.dashboard-card, .manage-status-card, .manage-main-card, .overview-card {
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  box-shadow: none;
  padding: 1.25rem;
  transition: border-color var(--transition);
}
.dashboard-card:hover, .manage-status-card:hover, .manage-main-card:hover, .overview-card:hover {
  /* Remove background and border color change on hover for minimalist look */
  background: var(--color-surface);
  border-color: var(--color-border);
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.5rem;
}
.status-label {
  font-size: 1rem;
  color: var(--color-text);
  font-weight: 600;
}
.status-dot.dot-running {
  color: var(--color-accent);
}
.status-dot.dot-stopped {
  color: var(--color-text-muted);
}
.action-btn, .v-btn {
  background: var(--color-accent);
  color: #fff;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none;
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  transition: background var(--transition);
}
.action-btn:hover, .v-btn:hover {
  background: var(--color-accent-hover);
}
.manage-tabs {
  background: var(--color-surface);
  border-radius: var(--rounded) var(--rounded) 0 0;
  border-bottom: 1px solid var(--color-border);
}
@media (max-width: 900px) {
  .main-content-area { padding: 1rem; }
  .manage-content-wrapper { max-width: 100%; gap: 1rem; }
  .manage-header { padding-bottom: 0.8rem; margin-bottom: 1rem; }
  .status-row { gap: 0.8rem; padding: 0.8rem; }
  .overview-card { padding: 1rem; }
}
</style>
