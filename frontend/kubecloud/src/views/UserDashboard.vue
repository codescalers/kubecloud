<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useDisplay } from 'vuetify'
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
const { width } = useDisplay()
// Menu toggle should show at tablet size and below (768px and below, including 768px)
const isMobileView = computed(() => width.value <= 768)

// Initialize selected section from localStorage or default to 'overview'
const selected = ref('overview')

// Drawer state for mobile
const drawer = ref(false)

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
  // Close drawer on mobile when item is selected
  if (isMobileView.value) {
    drawer.value = false
  }
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
  <div class="dashboard-container mt-16">
    <v-container fluid class="pa-0">
      <div 
        class="dashboard-header mb-6 text-center"
        @click="isMobileView && drawer ? drawer = false : null"
      >
        <h1 class="hero-title">Welcome back, {{ userName }}!</h1>
        <p class="section-subtitle">Manage your clusters, billing, and account settings from your dashboard.</p>
      </div>
      <div v-if="isMobileView" class="menu-toggle-row mb-4">
        <div class="menu-toggle-wrapper">
          <v-btn
            icon
            variant="text"
            color="primary"
            class="menu-toggle-btn"
            @click.stop="drawer = !drawer"
          >
            <v-icon>mdi-menu</v-icon>
          </v-btn>
          <v-divider class="menu-divider"></v-divider>
        </div>
      </div>
      <div class="dashboard-content-wrapper">
        <div class="dashboard-layout">
          <!-- Sidebar: Permanent on desktop, drawer on mobile -->
          <v-navigation-drawer
            v-model="drawer"
            :permanent="!isMobileView"
            :temporary="isMobileView"
            location="left"
            class="dashboard-sidebar-drawer mt-16 mt-md-0"
            :width="280"
            :floating="!isMobileView"
            @click.stop
          >
            <div @click.stop>
              <DashboardSidebar :selected="selected" @update:selected="handleSidebarSelect" />
            </div>
          </v-navigation-drawer>
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
  font-size: 3rem;
  font-weight: 700;
  color: var(--color-text, #fff);
  margin-bottom: 0.5rem;
}

.section-subtitle {
  color: var(--color-text-muted, #7c7fa5);
  font-size: 1.1rem;
}

.dashboard-header {
  max-width: 900px;
  margin: 7rem auto 3rem auto;
  position: relative;
  z-index: 2;
  padding: 0 1rem;
}

.menu-toggle-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding: 0 1rem;
  max-width: 1400px;
  margin: 0 auto;
}

.menu-toggle-wrapper {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.menu-toggle-btn {
  margin: 0;
  width: 48px !important;
  height: 48px !important;
  border-radius: 50% !important;
  background: rgba(59, 130, 246, 0.1) !important;
  backdrop-filter: blur(8px);
  transition: all 0.3s ease;
}

.menu-toggle-btn:hover {
  background: rgba(59, 130, 246, 0.2) !important;
}

.menu-divider {
  border-color: rgba(96, 165, 250, 0.2) !important;
  opacity: 1;
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

/* Navigation Drawer Styling */
:deep(.dashboard-sidebar-drawer) {
  background: transparent !important;
  border: none !important;
}

:deep(.dashboard-sidebar-drawer .v-navigation-drawer__content) {
  background: transparent !important;
}

:deep(.dashboard-sidebar-drawer .v-navigation-drawer__scrim) {
  background: rgba(0, 0, 0, 0.5) !important;
  backdrop-filter: blur(4px);
  pointer-events: auto !important;
  cursor: pointer;
}

/* Desktop: make sidebar sticky */
@media (min-width: 769px) {
  :deep(.dashboard-sidebar-drawer) {
    position: sticky !important;
    top: 0 !important;
    height: fit-content !important;
  }
  
  :deep(.dashboard-sidebar-drawer .v-navigation-drawer__content) {
    overflow: visible !important;
  }
  
  :deep(.dashboard-sidebar-drawer) {
    transform: none !important;
  }
}

.dashboard-sidebar-drawer :deep(.v-list),
.dashboard-sidebar-drawer :deep(.v-list-item) {
  background: transparent !important;
  box-shadow: none !important;
  border: none !important;
  color: inherit !important;
}

.dashboard-sidebar-drawer :deep(.v-list-item) {
  margin-bottom: 0.25rem;
  min-height: 44px;
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.dashboard-sidebar-drawer :deep(.v-list-item:last-child) {
  margin-bottom: 0;
}

.dashboard-sidebar-drawer :deep(.v-list-item--active),
.dashboard-sidebar-drawer :deep(.sidebar-item--active) {
  background: transparent !important;
  border-left: 3px solid #3B82F6 !important;
  border-radius: 0 !important;
  color: #fff !important;
}

.dashboard-sidebar-drawer :deep(.v-list-item__prepend) {
  margin-right: 0.5rem !important;
  display: flex;
  align-items: center;
}

.dashboard-sidebar-drawer :deep(.v-list-item__prepend) .v-icon,
.dashboard-sidebar-drawer :deep(.sidebar-icon) {
  color: #f3f4f6 !important;
  background: none !important;
  filter: none !important;
}

.dashboard-sidebar-drawer :deep(.v-list-item--active) .v-list-item__prepend .v-icon,
.dashboard-sidebar-drawer :deep(.sidebar-item--active) .sidebar-icon {
  color: #3B82F6 !important;
}

.dashboard-sidebar-drawer :deep(.logout-item),
.dashboard-sidebar-drawer :deep(.v-list-item.logout-item) {
  color: #ef4444 !important;
  fill: #ef4444 !important;
}

.dashboard-sidebar-drawer :deep(.logout-item .v-icon),
.dashboard-sidebar-drawer :deep(.v-list-item.logout-item .v-icon) {
  color: #ef4444 !important;
  fill: #ef4444 !important;
}

.dashboard-main {
  flex: 1;
  min-width: 0;
}

/* Mobile/Tablet adjustments */
@media (max-width: 768px) {
  .dashboard-header {
    margin-top: 3rem;
    text-align: left;
  }
  
  .hero-title {
    font-size: var(--font-size-2xl);
  }
  
  .section-subtitle {
    font-size: var(--font-size-base);
  }
  
  .dashboard-content-wrapper {
    margin-top: 2rem;
  }
  
  .dashboard-layout {
    gap: 0;
  }
  
  .dashboard-main {
    width: 100%;
  }
  
  .dashboard-cards {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }
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


