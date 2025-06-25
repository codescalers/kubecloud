<script setup lang="ts">
import { ref, computed } from 'vue'
import ClustersCard from '../components/dashboard/ClustersCard.vue'
import BillingCard from '../components/dashboard/BillingCard.vue'
import PaymentCard from '../components/dashboard/PaymentCard.vue'
import SshKeysCard from '../components/dashboard/SshKeysCard.vue'
import VouchersCard from '../components/dashboard/VouchersCard.vue'
import ProfileCard from '../components/dashboard/ProfileCard.vue'
import OverviewCard from '../components/dashboard/OverviewCard.vue'
import DashboardSidebar from '../components/DashboardSidebar.vue'

const userName = ref('John Doe')
const selected = ref('overview')
const clusters = ref([
  { id: 1, name: 'Production Cluster', status: 'running', nodes: 3, region: 'US East' },
  { id: 2, name: 'Staging Cluster', status: 'stopped', nodes: 2, region: 'US West' },
  { id: 3, name: 'Development Cluster', status: 'running', nodes: 1, region: 'EU West' }
])
const billingHistory = ref([
  { id: 1, date: '2024-01-15', description: 'Production Cluster Usage', amount: 45.50 },
  { id: 2, date: '2024-01-10', description: 'Staging Cluster Usage', amount: 22.30 },
  { id: 3, date: '2024-01-05', description: 'Account Credit', amount: -50.00 }
])
const paymentMethods = ref([
  { id: 1, name: 'Visa ending in 4242', maskedNumber: '**** **** **** 4242', icon: 'mdi-credit-card', iconColor: '#1E40AF' },
  { id: 2, name: 'PayPal', maskedNumber: 'john.doe@example.com', icon: 'mdi-paypal', iconColor: '#0070BA' }
])
const sshKeys = ref([
  { id: 1, name: 'My Laptop', fingerprint: 'SHA256:Abc123...Xyz789', addedDate: '2024-01-01' },
  { id: 2, name: 'Work PC', fingerprint: 'SHA256:Def456...789', addedDate: '2024-01-05' }
])
const vouchers = ref([
  { id: 1, name: 'Welcome Bonus', description: 'New user welcome credit', amount: '$50.00', expiryDate: '2024-12-31', icon: 'mdi-gift', iconColor: '#F472B6' },
  { id: 2, name: 'Referral Credit', description: 'Friend referral bonus', amount: '$25.00', expiryDate: '2024-06-30', icon: 'mdi-account-multiple', iconColor: '#38BDF8' }
])
const recentActivity = ref([
  { id: 1, text: 'Production cluster scaled to 3 nodes', time: '2 hours ago', icon: 'mdi-server', iconColor: '#38BDF8' },
  { id: 2, text: 'Payment of $45.50 processed', time: '1 day ago', icon: 'mdi-credit-card', iconColor: '#3B82F6' },
  { id: 3, text: 'New SSH key added', time: '2 days ago', icon: 'mdi-key', iconColor: '#F472B6' },
  { id: 4, text: 'Staging cluster stopped', time: '3 days ago', icon: 'mdi-stop-circle', iconColor: '#F59E0B' }
])
const totalSpent = computed(() => {
  return billingHistory.value
    .filter(bill => bill.amount > 0)
    .reduce((sum, bill) => sum + bill.amount, 0)
    .toFixed(2)
})
const uptimeHours = computed(() => {
  return clusters.value
    .filter(cluster => cluster.status === 'running')
    .reduce((sum, cluster) => sum + cluster.nodes * 24, 0)
})
function handleSidebarSelect(val: string) {
  selected.value = val
}
function handleLogout() {
  // Implement logout logic here
}
function handleNavigate(section: string) {
  selected.value = section
}
</script>

<template>
  <div class="dashboard-container minimalist-bg">
    <v-container fluid class="pa-0">
      <div class="dashboard-header mb-6">
        <h1 class="display-2 kubecloud-gradient mb-3 minimalist-title">Welcome back, {{ userName }}!</h1>
        <p class="dashboard-subtitle minimalist-subtitle">Manage your clusters, billing, and account settings from your minimalist dashboard.</p>
      </div>
      <div class="dashboard-content-wrapper">
        <div class="dashboard-layout">
          <div class="dashboard-sidebar">
            <DashboardSidebar :selected="selected" @update:selected="handleSidebarSelect" @logout="handleLogout" />
          </div>
          <div class="dashboard-main">
            <div class="dashboard-cards">
              <OverviewCard
                v-if="selected === 'overview'"
                :clusters="clusters"
                :recentActivity="recentActivity"
                :sshKeys="sshKeys"
                :totalSpent="totalSpent"
                @navigate="handleNavigate"
              />
              <ClustersCard v-if="selected === 'clusters'" :clusters="clusters" />
              <BillingCard v-if="selected === 'billing'" :billingHistory="billingHistory" />
              <PaymentCard v-if="selected === 'payment'" :paymentMethods="paymentMethods" />
              <SshKeysCard v-if="selected === 'ssh'" :sshKeys="sshKeys" />
              <VouchersCard v-if="selected === 'vouchers'" :vouchers="vouchers" />
              <ProfileCard v-if="selected === 'profile'" />
            </div>
          </div>
        </div>
      </div>
    </v-container>
  </div>
</template>

<style scoped>
.dashboard-container.minimalist-bg {
  background: var(--color-background);
  position: relative;
}

.minimalist-title {
  font-size: 2rem;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--color-text);
  background: none;
}

.minimalist-subtitle {
  color: var(--color-text-muted);
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
}

.minimalist-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--rounded);
  box-shadow: none;
  color: var(--color-text);
  padding: 1.25rem;
  min-width: 100%;
  transition: border-color var(--transition), background var(--transition);
}

.minimalist-card:hover {
  border-color: var(--color-accent);
  background: var(--color-surface);
}

.dashboard-header {
  text-align: center;
  max-width: 900px;
  margin: 3rem auto 2rem auto;
  position: relative;
  z-index: 2;
  padding: 0 1rem;
}

.dashboard-subtitle {
  font-size: 1.25rem;
  color: #E0E7EF;
  font-weight: 400;
  opacity: 0.9;
  line-height: 1.6;
}

.dashboard-content-wrapper {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 1rem;
  position: relative;
  z-index: 2;
  margin-top: 3rem;
}

.dashboard-layout {
  display: flex;
  min-height: 70vh;
  gap: 1.5rem;
  position: relative;
  z-index: 2;
  align-items: flex-start;
  margin-top: 0;
}

.dashboard-sidebar {
  flex: 0 0 280px;
  display: flex;
  flex-direction: column;
  height: fit-content;
  background: var(--color-surface, #181f2a) !important;
  border: 1px solid var(--color-border, #232325) !important;
  border-radius: 12px;
  box-shadow: none !important;
  z-index: 3;
  overflow: hidden;
  position: sticky;
  top: 0;
  align-self: flex-start;
  margin-top: 0;
}

.dashboard-sidebar :deep(.v-list),
.dashboard-sidebar :deep(.v-list-item) {
  background: transparent !important;
  box-shadow: none !important;
  border: none !important;
  color: inherit !important;
}

.dashboard-sidebar :deep(.v-list-item) {
  margin-bottom: 0.25rem;
  min-height: 44px;
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.dashboard-sidebar :deep(.v-list-item:last-child) {
  margin-bottom: 0;
}

.dashboard-sidebar :deep(.v-list-item--active),
.dashboard-sidebar :deep(.sidebar-item--active) {
  background: transparent !important;
  border-left: 3px solid #3B82F6 !important;
  border-radius: 0 !important;
  color: #fff !important;
}

.dashboard-sidebar :deep(.v-list-item__prepend) {
  margin-right: 0.5rem !important;
  display: flex;
  align-items: center;
}

.dashboard-sidebar :deep(.v-list-item__prepend) .v-icon,
.dashboard-sidebar :deep(.sidebar-icon) {
  color: #f3f4f6 !important;
  background: none !important;
  filter: none !important;
}

.dashboard-sidebar :deep(.v-list-item--active) .v-list-item__prepend .v-icon,
.dashboard-sidebar :deep(.sidebar-item--active) .sidebar-icon {
  color: #3B82F6 !important;
}

.dashboard-sidebar :deep(.logout-item),
.dashboard-sidebar :deep(.v-list-item.logout-item) {
  color: #ef4444 !important;
  fill: #ef4444 !important;
}

.dashboard-sidebar :deep(.logout-item .v-icon),
.dashboard-sidebar :deep(.v-list-item.logout-item .v-icon) {
  color: #ef4444 !important;
  fill: #ef4444 !important;
}

.dashboard-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 70vh;
  max-width: calc(100% - 280px - 1.5rem);
  align-items: stretch;
  align-self: flex-start;
  margin-top: 0;
}

.dashboard-cards {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  align-items: stretch;
}

.dashboard-card {
  display: flex;
  flex-direction: column;
  width: 100%;
}

@media (max-width: 1200px) {
  .dashboard-content-wrapper {
    max-width: 1200px;
  }

  .dashboard-sidebar {
    flex: 0 0 240px;
  }

  .dashboard-main {
    max-width: calc(100% - 240px - 1.5rem);
  }
}

@media (max-width: 960px) {
  .dashboard-content-wrapper {
    max-width: 100%;
    padding: 0 0.5rem;
  }

  .dashboard-layout {
    flex-direction: column;
    gap: 1rem;
    margin-top: 0.5rem;
  }

  .dashboard-sidebar {
    flex: none;
    width: 100%;
    position: static;
    align-self: stretch;
  }

  .dashboard-main {
    max-width: 100%;
    align-self: stretch;
  }

  .dashboard-container {
    padding: 0.5rem 0;
  }

  .dashboard-header {
    margin: 1.5rem auto 1.5rem auto;
    padding: 0 0.5rem;
  }

  .dashboard-subtitle {
    font-size: 1.125rem;
  }

  .sidebar-content {
    padding: 1.5rem;
  }

  .sidebar-title {
    font-size: 1.25rem;
  }

  .stat-value {
    font-size: 1.25rem;
  }

  .action-btn {
    padding: 0.875rem 1.25rem;
    font-size: 0.9rem;
  }
}

@media (max-width: 600px) {
  .dashboard-content-wrapper {
    padding: 0 0.25rem;
  }

  .dashboard-layout {
    margin-top: 0.25rem;
  }

  .dashboard-container {
    padding: 0.25rem 0;
  }

  .dashboard-header {
    margin: 1rem auto 1rem auto;
    padding: 0 0.25rem;
  }

  .dashboard-subtitle {
    font-size: 1rem;
  }

  .sidebar-content {
    padding: 1rem;
  }

  .sidebar-title {
    font-size: 1.125rem;
  }

  .stat-item {
    padding: 0.75rem;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
  }

  .stat-value {
    font-size: 1.125rem;
  }

  .activity-item {
    padding: 0.5rem;
  }

  .action-btn {
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
  }
}
</style>
