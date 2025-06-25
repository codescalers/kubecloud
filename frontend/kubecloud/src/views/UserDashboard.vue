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
  { id: 2, text: 'Payment of $45.50 processed', time: '1 day ago', icon: 'mdi-credit-card', iconColor: '#7C3AED' },
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
    <!-- Animated Background -->
    <div class="dashboard-background">
      <div class="dashboard-particles">
        <div class="particle particle-1"></div>
        <div class="particle particle-2"></div>
        <div class="particle particle-3"></div>
        <div class="particle particle-4"></div>
        <div class="particle particle-5"></div>
        <div class="particle particle-6"></div>
      </div>
      <div class="dashboard-grid-lines"></div>
      <div class="dashboard-connections">
        <div class="connection connection-1"></div>
        <div class="connection connection-2"></div>
        <div class="connection connection-3"></div>
      </div>
    </div>
    
    <v-container fluid class="pa-0">
      <div class="dashboard-header mb-6">
        <h1 class="display-2 kubecloud-gradient mb-3">Welcome back, {{ userName }}!</h1>
        <p class="dashboard-subtitle">Manage your clusters, billing, and account settings from your personalized dashboard.</p>
      </div>
      <div class="dashboard-content-wrapper">
        <div class="dashboard-layout">
          <div class="dashboard-sidebar">
            <DashboardSidebar :selected="selected" @update:selected="handleSidebarSelect" @logout="handleLogout" />
          </div>
          <div class="dashboard-main">
            <div class="dashboard-cards">
              <div v-if="selected === 'overview'" class="dashboard-card">
                <OverviewCard 
                  :clusters="clusters"
                  :recentActivity="recentActivity"
                  :sshKeys="sshKeys"
                  :totalSpent="totalSpent"
                  @navigate="handleNavigate"
                />
              </div>
              <div v-if="selected === 'clusters'" class="dashboard-card">
                <ClustersCard :clusters="clusters" />
              </div>
              <div v-if="selected === 'billing'" class="dashboard-card">
                <BillingCard :billingHistory="billingHistory" />
              </div>
              <div v-if="selected === 'payment'" class="dashboard-card">
                <PaymentCard :paymentMethods="paymentMethods" />
              </div>
              <div v-if="selected === 'ssh'" class="dashboard-card">
                <SshKeysCard :sshKeys="sshKeys" />
              </div>
              <div v-if="selected === 'vouchers'" class="dashboard-card">
                <VouchersCard :vouchers="vouchers" />
              </div>
              <div v-if="selected === 'profile'" class="dashboard-card">
                <ProfileCard />
              </div>
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
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
  padding: 1rem 0;
  position: relative;
  overflow: hidden;
}

/* Animated Background */
.dashboard-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 0;
}

/* Floating Particles */
.dashboard-particles {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.particle {
  position: absolute;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(79, 70, 229, 0.3) 0%, rgba(79, 70, 229, 0.1) 50%, transparent 100%);
  animation: float-particle 15s ease-in-out infinite;
}

.particle-1 {
  width: 120px;
  height: 120px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.particle-2 {
  width: 80px;
  height: 80px;
  top: 20%;
  right: 15%;
  background: radial-gradient(circle, rgba(234, 88, 12, 0.2) 0%, rgba(234, 88, 12, 0.1) 50%, transparent 100%);
  animation-delay: -3s;
}

.particle-3 {
  width: 100px;
  height: 100px;
  bottom: 20%;
  left: 20%;
  background: radial-gradient(circle, rgba(8, 145, 178, 0.25) 0%, rgba(8, 145, 178, 0.1) 50%, transparent 100%);
  animation-delay: -6s;
}

.particle-4 {
  width: 60px;
  height: 60px;
  top: 60%;
  right: 25%;
  background: radial-gradient(circle, rgba(79, 70, 229, 0.2) 0%, rgba(79, 70, 229, 0.05) 50%, transparent 100%);
  animation-delay: -9s;
}

.particle-5 {
  width: 90px;
  height: 90px;
  bottom: 10%;
  right: 10%;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.2) 0%, rgba(139, 92, 246, 0.1) 50%, transparent 100%);
  animation-delay: -12s;
}

.particle-6 {
  width: 70px;
  height: 70px;
  top: 40%;
  left: 5%;
  background: radial-gradient(circle, rgba(236, 72, 153, 0.2) 0%, rgba(236, 72, 153, 0.1) 50%, transparent 100%);
  animation-delay: -15s;
}

@keyframes float-particle {
  0%, 100% { 
    transform: translateY(0px) translateX(0px) scale(1);
    opacity: 0.3;
  }
  25% { 
    transform: translateY(-20px) translateX(10px) scale(1.1);
    opacity: 0.5;
  }
  50% { 
    transform: translateY(-10px) translateX(-15px) scale(0.9);
    opacity: 0.4;
  }
  75% { 
    transform: translateY(-30px) translateX(5px) scale(1.05);
    opacity: 0.6;
  }
}

/* Grid Lines */
.dashboard-grid-lines {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: 
    linear-gradient(rgba(79, 70, 229, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(79, 70, 229, 0.03) 1px, transparent 1px);
  background-size: 50px 50px;
  animation: grid-move 20s linear infinite;
}

@keyframes grid-move {
  0% { transform: translate(0, 0); }
  100% { transform: translate(50px, 50px); }
}

/* Network Connections */
.dashboard-connections {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.connection {
  position: absolute;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(79, 70, 229, 0.2), transparent);
  animation: connection-pulse 8s ease-in-out infinite;
}

.connection-1 {
  top: 30%;
  left: 20%;
  width: 200px;
  transform: rotate(45deg);
  animation-delay: 0s;
}

.connection-2 {
  top: 60%;
  right: 30%;
  width: 150px;
  transform: rotate(-30deg);
  animation-delay: -3s;
}

.connection-3 {
  bottom: 40%;
  left: 50%;
  width: 180px;
  transform: rotate(15deg);
  animation-delay: -6s;
}

@keyframes connection-pulse {
  0%, 100% { 
    opacity: 0.1;
    transform: scaleX(0.8);
  }
  50% { 
    opacity: 0.4;
    transform: scaleX(1.2);
  }
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
  background: var(--glass-bg, #181f2a);
  box-shadow: 2px 0 24px 0 rgba(44, 62, 80, 0.12);
  z-index: 3;
  border-radius: 12px;
  overflow: hidden;
  position: sticky;
  top: 0;
  align-self: flex-start;
  margin-top: 0;
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
