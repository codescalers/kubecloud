<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useUserStore } from '../stores/user'
import ClustersCard from '../components/dashboard/ClustersCard.vue'
import BillingCard from '../components/dashboard/BillingCard.vue'
import PaymentCard from '../components/dashboard/PaymentCard.vue'
import SshKeysCard from '../components/dashboard/SshKeysCard.vue'
import VouchersCard from '../components/dashboard/VouchersCard.vue'
import ProfileCard from '../components/dashboard/ProfileCard.vue'
import OverviewCard from '../components/dashboard/OverviewCard.vue'
import NodesCard from '../components/dashboard/NodesCard.vue'
import DashboardSidebar from '../components/DashboardSidebar.vue'
import { userService } from '../utils/userService'
import { useClusterStore } from '../stores/clusters'
import { useNotificationStore } from '../stores/notifications'
import UserPendingRecordsCard from '../components/dashboard/UserPendingRecordsCard.vue'

const userStore = useUserStore()
const userName = computed(() => userStore.user?.username || 'User')

// Initialize selected section from localStorage or default to 'overview'
const selected = ref('overview')

const clusterStore = useClusterStore()
const notificationStore = useNotificationStore()

const clusters = computed(() => clusterStore.clusters)
const clustersArray = computed(() =>
  Array.isArray(clusters.value)
    ? clusters.value.map((c, idx) => ({
        id: c.id ?? idx,
        name: c.cluster.name,
        status: c.cluster.status ?? '',
        nodes: typeof c.cluster.nodes === 'number' ? c.cluster.nodes : 0,
        region: c.cluster.region ?? ''
      }))
    : []
)

// Constants
const STORAGE_KEY_DASHBOARD_SECTION = 'dashboard-section'

// Note: Cluster events are handled globally by the useDeploymentEvents composable


onMounted(async () => {
  try {
    userStore.startBalanceRefresh()
    // Restore selected section from localStorage
    const savedSection = localStorage.getItem(STORAGE_KEY_DASHBOARD_SECTION)
    if (savedSection) {
      selected.value = savedSection
    }

    // Fetch initial data
    const [invoices] = await Promise.all([
      userService.listUserInvoices(),
      userStore.updateNetBalance(),
    ])

    // Process invoices
    billingHistory.value = invoices.map(inv => ({
      id: inv.id,
      date: inv.created_at,
      description: `Invoice ${inv.id}`,
      amount: inv.total
    }))
  } catch (error) {
    console.error(error);
  }
})

// Cleanup on unmount
onUnmounted(() => {
  userStore.stopBalanceRefresh()
})

interface Bill {
  id: string | number
  date: string
  description: string
  amount: number
}

const billingHistory = ref<Bill[]>([])

const sshKeys = ref([])
const vouchers = ref([])
const totalSpent = computed(() => {
  return billingHistory.value
    .filter(bill => bill.amount > 0)
    .reduce((sum, bill) => sum + bill.amount, 0)
    .toFixed(2)
})

function handleSidebarSelect(val: string) {
  selected.value = val
  // Save to localStorage for persistence
  localStorage.setItem(STORAGE_KEY_DASHBOARD_SECTION, val)
}

function handleNavigate(section: string) {
  selected.value = section
  // Save to localStorage for persistence
  localStorage.setItem('dashboard-section', section)
}
function handleNavigateToFund() {
  selected.value = 'add-funds'
  localStorage.setItem('dashboard-section', 'add-funds')
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
            <DashboardSidebar :selected="selected" @update:selected="handleSidebarSelect" />
          </div>
          <div class="dashboard-main">
            <div class="dashboard-cards">
              <OverviewCard
                v-if="selected === 'overview'"
                :clusters="clustersArray"
                :sshKeys="sshKeys"
                :totalSpent="totalSpent"
                @navigate="handleNavigate"
              />
              <ClustersCard v-if="selected === 'clusters'" :clusters="clusters" @navigateToFund="handleNavigateToFund" />
              <BillingCard v-if="selected === 'billing'" :billingHistory="billingHistory" />
              <PaymentCard v-if="selected === 'add-funds'" />
              <SshKeysCard v-if="selected === 'ssh'" :sshKeys="sshKeys" />
              <VouchersCard v-if="selected === 'vouchers'" :vouchers="vouchers" />
              <NodesCard v-if="selected === 'nodes'" />
              <ProfileCard v-if="selected === 'profile'" />
              <UserPendingRecordsCard v-if="selected === 'payments'"/>
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
  background: var(--mycelium-cloud-bg);
}

.hero-title {
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  margin-bottom: 1.5rem;
  line-height: 1.1;
  letter-spacing: -1px;
  color: var(--mycelium-cloud-text);
}

.section-subtitle {
  font-size: var(--font-size-xl);
  color: var(--mycelium-cloud-text-muted);
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
  background: var(--mycelium-cloud-surface);
  border: 1px solid var(--mycelium-cloud-border);
  border-radius: var(--mycelium-cloud-radius);
  color: var(--mycelium-cloud-text-secondary);
  padding: var(--mycelium-cloud-spacing);
  transition: border-color 0.2s;
}

.dashboard-card:hover {
  border-color: var(--mycelium-cloud-primary);
}

.dashboard-card-title {
  font-size: var(--font-size-h3);
  font-weight: var(--font-weight-bold);
  color: var(--mycelium-cloud-text);
  margin-bottom: 0.5rem;
}

.dashboard-card-subtitle {
  font-size: 1.05rem;
  color: var(--mycelium-cloud-text-muted);
  font-weight: var(--font-weight-bold);
  opacity: 0.9;
}


</style>


