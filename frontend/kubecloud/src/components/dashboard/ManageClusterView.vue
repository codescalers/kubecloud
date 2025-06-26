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
          <v-card class="manage-status-card mb-6">
            <div class="d-flex flex-wrap align-center status-row">
              <span class="mr-4 status-label">Status:
                <span :class="cluster.status === 'Running' ? 'text-success' : 'text-error'">{{ cluster.status }}</span>
                <span class="status-dot ml-1" :class="cluster.status === 'Running' ? 'dot-running' : 'dot-stopped'">●</span>
              </span>
              <v-btn variant="text" color="primary" class="mr-2">Download Kubeconfig</v-btn>
              <v-btn variant="text" color="primary">Open Dashboard</v-btn>
            </div>
          </v-card>
          <v-card class="manage-main-card pa-0">
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
                    <v-card class="overview-card mb-4">
                      <h3 class="text-subtitle-1 mb-2">Cluster Resources</h3>
                      <div>3 Nodes</div>
                      <div>12 vCPU</div>
                      <div>24 GB RAM</div>
                      <div>1 TB Storage</div>
                    </v-card>
                  </v-col>
                  <v-col cols="12" md="6">
                    <v-card class="overview-card mb-4">
                      <h3 class="text-subtitle-1 mb-2">Live Usage</h3>
                      <div class="mb-2">CPU: <v-progress-linear :model-value="70" color="primary" height="12" rounded></v-progress-linear> <span class="ml-2">70%</span></div>
                      <div>MEM: <v-progress-linear :model-value="50" color="secondary" height="12" rounded></v-progress-linear> <span class="ml-2">50%</span></div>
                    </v-card>
                  </v-col>
                </v-row>
                <v-card class="overview-card mb-0">
                  <h3 class="text-subtitle-1 mb-2">Details</h3>
                  <div>API Endpoint: <span class="font-mono">ygg://[mycelium-ip-address]</span></div>
                  <div>Location: Ghent, Belgium</div>
                  <div>Created: 2023-07-01</div>
                  <div>Est. Cost: $125/month</div>
                </v-card>
              </v-window-item>
              <v-window-item value="nodes">
                <v-card class="overview-card">Nodes tab content (coming soon)</v-card>
              </v-window-item>
              <v-window-item value="storage">
                <v-card class="overview-card">Storage tab content (coming soon)</v-card>
              </v-window-item>
              <v-window-item value="networking">
                <v-card class="overview-card">Networking tab content (coming soon)</v-card>
              </v-window-item>
              <v-window-item value="settings">
                <v-card class="overview-card">Settings tab content (coming soon)</v-card>
              </v-window-item>
            </v-window>
          </v-card>
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
  background: var(--v-theme-background, #f5f7fa);
  min-height: 100vh;
  padding: 0;
}
.main-content-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem 2.5rem 2rem 2.5rem;
  min-height: 100vh;
}
.manage-content-wrapper {
  width: 100%;
  max-width: 700px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 2.5rem;
}
.manage-header {
  position: sticky;
  top: 0;
  z-index: 2;
  background: var(--v-theme-background, #f5f7fa);
  padding-bottom: 1.2rem;
  margin-bottom: 2.2rem;
  margin-top: 0.5rem;
  justify-content: flex-start;
  border-bottom: 1.5px solid var(--v-theme-outline, #e0e7ef);
}
.manage-status-card,
.manage-main-card {
  margin-left: auto;
  margin-right: auto;
}
.status-row {
  gap: 1.5rem;
}
.status-label {
  font-size: 1.08em;
  font-weight: 600;
  color: var(--v-theme-on-surface, #222B45);
  display: flex;
  align-items: center;
}
.status-dot {
  font-size: 1.1em;
  margin-left: 0.3em;
  vertical-align: middle;
  transition: color 0.2s;
}
.dot-running {
  color: var(--kubecloud-success);
  text-shadow: 0 0 1px rgba(16, 185, 129, 0.3);
}
.dot-stopped {
  color: var(--kubecloud-error);
  text-shadow: 0 0 1px rgba(239, 68, 68, 0.3);
}
.manage-main-card {
  background: var(--v-theme-surface, #fff);
  border-radius: 18px;
  box-shadow: 0 6px 32px var(--v-theme-shadow, rgba(60,60,60,0.10));
  border: 1.5px solid var(--v-theme-outline, #e0e7ef);
  overflow: visible;
}
.manage-tabs {
  background: transparent;
  border-radius: 18px 18px 0 0;
}
.overview-card {
  background: linear-gradient(120deg, var(--v-theme-background, #f5f7fa) 0%, var(--v-theme-surface, #e3eafc) 100%);
  border-radius: 14px;
  box-shadow: 0 2px 12px var(--v-theme-shadow, rgba(60,60,60,0.06));
  padding: 1.5rem 1.5rem 1.2rem 1.5rem;
  border: 1.5px solid var(--v-theme-outline, #e0e7ef);
  margin-bottom: 1.2rem;
}
.font-mono {
  font-family: 'Fira Mono', 'Menlo', 'Monaco', 'Consolas', monospace;
}
@media (prefers-color-scheme: dark) {
  .manage-main-card, .manage-status-card {
    background: var(--v-theme-surface, #23272f);
    border: 1.5px solid var(--v-theme-outline, #23272f);
  }
  .overview-card {
    background: linear-gradient(120deg, var(--v-theme-surface, #23272f) 0%, var(--v-theme-background, #181A20) 100%);
    border: 1.5px solid var(--v-theme-outline, #23272f);
  }
}
@media (max-width: 900px) {
  .main-content-area {
    padding: 1.2rem 0.5rem 1rem 0.5rem;
  }
  .manage-content-wrapper {
    max-width: 100%;
    gap: 1.2rem;
  }
}
</style>
