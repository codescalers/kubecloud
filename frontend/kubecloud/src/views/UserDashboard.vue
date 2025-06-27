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
  <div class="dashboard-container">
    <v-container fluid class="pa-0">
      <div class="dashboard-header mb-6">
        <h1 class="hero-title">Welcome back, {{ userName }}!</h1>
        <p class="section-subtitle">Manage your clusters, billing, and account settings from your dashboard.</p>
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
.dashboard-container {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  background: var(--kubecloud-bg);
}

.hero-title {
  font-size: var(--font-size-h1);
  font-weight: var(--font-weight-bold);
  margin-bottom: 1.5rem;
  line-height: 1.1;
  letter-spacing: -1px;
  color: var(--kubecloud-text);
}

.section-subtitle {
  font-size: var(--font-size-h3);
  color: var(--kubecloud-text-muted);
  line-height: 1.5;
  opacity: 0.92;
  margin-bottom: 0;
  font-weight: var(--font-weight-normal);
}

.dashboard-header {
  text-align: center;
  max-width: 900px;
  margin: 7rem auto 3rem auto;
  position: relative;
  z-index: 2;
  padding: 0 1rem;
}

.dashboard-content-wrapper {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 1rem;
  position: relative;
  z-index: 2;
  margin-top: 4rem;
}

.dashboard-layout {
  display: flex;
  min-height: 70vh;
  gap: 2.5rem;
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
  min-width: 0;
}

.dashboard-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
  gap: 2.2rem;
  width: 100%;
  align-items: stretch;
}

.dashboard-card {
  display: flex;
  flex-direction: column;
  width: 100%;
  background: var(--kubecloud-surface);
  border: 1px solid var(--kubecloud-border);
  border-radius: var(--kubecloud-radius);
  color: var(--kubecloud-text-secondary);
  padding: var(--kubecloud-spacing);
  transition: border-color 0.2s;
}

.dashboard-card:hover {
  border-color: var(--kubecloud-primary);
}

.dashboard-card-title {
  font-size: var(--font-size-h3);
  font-weight: var(--font-weight-bold);
  color: var(--kubecloud-text);
  margin-bottom: 0.5rem;
}

.dashboard-card-subtitle {
  font-size: 1.05rem;
  color: var(--kubecloud-text-muted);
  font-weight: var(--font-weight-bold);
  opacity: 0.9;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--kubecloud-surface-light);
  border: 1px solid var(--kubecloud-border);
  border-radius: var(--kubecloud-radius-md);
  padding: 1.2rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: border-color 0.2s;
}

.stat-card:hover {
  border-color: var(--kubecloud-primary);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--kubecloud-radius-sm);
  background: var(--kubecloud-surface);
  border: 1px solid var(--kubecloud-border);
}

.stat-number {
  font-size: 1.3rem;
  font-weight: var(--font-weight-bold);
  color: var(--kubecloud-text);
  margin-bottom: 0.15rem;
}

.stat-label {
  font-size: 0.9rem;
  color: var(--kubecloud-text-muted);
  font-weight: var(--font-weight-normal);
}

.quick-actions-section {
  margin-bottom: 2rem;
}

.section-title {
  font-size: 1.2rem;
  font-weight: var(--font-weight-bold);
  color: var(--kubecloud-text);
  margin-bottom: 1rem;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.action-btn {
  background: var(--kubecloud-surface-light) !important;
  border: 1px solid var(--kubecloud-primary) !important;
  color: var(--kubecloud-primary) !important;
  font-weight: var(--font-weight-bold);
  border-radius: var(--kubecloud-radius-md) !important;
  transition: background 0.2s, color 0.2s;
  font-size: 1rem;
  padding: 0.7rem 1.5rem;
}

.action-btn:hover {
  background: var(--kubecloud-primary) !important;
  color: #fff !important;
}

/* Floating Action Button for Add Cluster on mobile */
@media (max-width: 600px) {
  .fab-add-cluster {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    z-index: 1000;
    background: var(--kubecloud-primary) !important;
    color: #fff !important;
    border-radius: 50%;
    width: 56px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2rem;
  }
}
</style>


