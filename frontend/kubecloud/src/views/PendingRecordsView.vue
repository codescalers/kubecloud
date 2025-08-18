<template>
  <div class="dashboard-container">
    <v-container fluid class="pa-0">
      <div class="dashboard-header mb-6">
        <h1 class="hero-title">Pending Records</h1>
        <p class="section-subtitle">View your pending transfer records</p>
      </div>
      <div class="dashboard-content-wrapper">
        <div class="dashboard-layout">
          <div class="dashboard-sidebar">
            <DashboardSidebar :selected="'payments'" @update:selected="handleSidebarSelect" />
          </div>
          <div class="dashboard-main">
            <div class="dashboard-cards">
              <UserPendingRecordsCard />
            </div>
          </div>
        </div>
      </div>
    </v-container>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DashboardSidebar from '../components/DashboardSidebar.vue'
import UserPendingRecordsCard from '../components/dashboard/UserPendingRecordsCard.vue'

const selected = ref('overview')
function handleSidebarSelect(val: string) {
  selected.value = val
}
</script>

<style scoped>
.dashboard-container {
  min-height: 100vh;
  background: var(--color-background);
}

.dashboard-header {
  text-align: center;
  padding: var(--space-8) 0;
}

.hero-title {
  font-size: 2.5rem;
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-2);
}

.section-subtitle {
  font-size: 1.125rem;
  color: var(--color-text-secondary);
  max-width: 600px;
  margin: 0 auto;
}

.dashboard-content-wrapper {
  padding: 0 var(--space-6);
}

.dashboard-layout {
  display: flex;
  gap: var(--space-6);
  max-width: 1400px;
  margin: 0 auto;
}

.dashboard-sidebar {
  flex: 0 0 280px;
}

.dashboard-main {
  flex: 1;
  min-width: 0;
}

.dashboard-cards {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

@media (max-width: 1024px) {
  .dashboard-layout {
    flex-direction: column;
  }

  .dashboard-sidebar {
    flex: none;
  }
}
</style>
