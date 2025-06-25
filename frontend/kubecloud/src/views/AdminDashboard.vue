<script setup lang="ts">
import { ref, computed } from 'vue'
import AdminSidebar from '../components/AdminSidebar.vue'

const selected = ref('overview')

const adminStats = ref([
  { label: 'Total Users', value: 128, icon: 'mdi-account-group', color: '#3B82F6' },
  { label: 'Active Clusters', value: 42, icon: 'mdi-server', color: '#3B82F6' },
  { label: 'Pending Approvals', value: 3, icon: 'mdi-alert', color: '#F59E0B' },
])

// User management state
const users = ref([
  { id: 1, name: 'Alice', email: 'alice@example.com' },
  { id: 2, name: 'Bob', email: 'bob@example.com' },
  { id: 3, name: 'Charlie', email: 'charlie@example.com' },
  // ... more mock users ...
])
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 5
const totalUsers = computed(() => users.value.length)
const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value
  return users.value.filter(u =>
    u.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    u.email.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})
const paginatedUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredUsers.value.slice(start, start + pageSize)
})
const totalPages = computed(() => Math.ceil(filteredUsers.value.length / pageSize))

// Create user form
const newUserName = ref('')
const newUserEmail = ref('')
const createUserResult = ref('')

const snackbar = ref(false)
const snackbarMsg = ref('')
const snackbarColor = ref('success')

function showSnackbar(msg: string, color: string = 'success') {
  snackbarMsg.value = msg
  snackbarColor.value = color
  snackbar.value = true
}

function createUser() {
  // TODO: Replace with real API call
  if (!newUserName.value || !newUserEmail.value) return
  const newId = users.value.length ? Math.max(...users.value.map(u => u.id)) + 1 : 1
  users.value.push({ id: newId, name: newUserName.value, email: newUserEmail.value })
  createUserResult.value = `User ${newUserName.value} created.`
  showSnackbar(`User ${newUserName.value} created.`, 'success')
  newUserName.value = ''
  newUserEmail.value = ''
}

function deleteUser(userId: number) {
  // TODO: Replace with real API call
  users.value = users.value.filter(u => u.id !== userId)
  showSnackbar('User deleted.', 'info')
}

function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value) currentPage.value = page
}

// Voucher generation form state
const voucherValue = ref(50)
const voucherCount = ref(10)
const voucherExpiry = ref('2024-12-31')
const voucherResult = ref('')

// Manual credit form state
const creditUserId = ref('')
const creditAmount = ref(0)
const creditReason = ref('')
const creditResult = ref('')

// Manual credits state for Credits section
const manualCredits = ref([
  // Example mock data
  { id: 1, user: 'Alice', amount: 50, reason: 'Support bonus', date: '2024-06-01' },
  { id: 2, user: 'Bob', amount: 25, reason: 'Promo credit', date: '2024-06-02' },
])

function handleSidebarSelect(newSelected: string) {
  selected.value = newSelected
}

// Placeholder for API call to generate vouchers
async function generateVouchers() {
  // TODO: Replace with real API call
  voucherResult.value = `Requested ${voucherCount.value} vouchers of $${voucherValue.value} expiring on ${voucherExpiry.value}`
  showSnackbar('Vouchers generated!', 'success')
}

// Placeholder for API call to apply manual credit
async function applyManualCredit() {
  // TODO: Replace with real API call
  creditResult.value = `Credited user ${creditUserId.value} with $${creditAmount.value} for: ${creditReason.value}`
  showSnackbar('Manual credit applied!', 'success')
}

const tabs = [
  { key: 'overview', label: 'Overview' },
  { key: 'users', label: 'Users' },
  { key: 'clusters', label: 'Clusters' },
  { key: 'system', label: 'System' },
  { key: 'vouchers', label: 'Vouchers' },
]
</script>

<template>
  <div class="dashboard-container">
    <div class="dashboard-content-wrapper">
      <div class="dashboard-layout">
        <div class="admin-sidebar">
          <AdminSidebar :selected="selected" @update:selected="handleSidebarSelect" />
        </div>
        <div class="dashboard-main">
          <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="2500" location="top right">{{ snackbarMsg }}</v-snackbar>
          <div v-if="selected === 'overview'">
            <div class="section-header">
              <h2 class="dashboard-title">Admin Overview</h2>
            </div>
            <div class="dashboard-cards">
              <v-card v-for="stat in adminStats" :key="stat.label" class="dashboard-card admin-stat-card cloud-card float-card elevation-2" color="surface-light">
                <v-icon :icon="stat.icon" :color="stat.color" size="36" class="stat-icon cloud-icon" />
                <div class="stat-label">{{ stat.label }}</div>
                <div class="stat-value">{{ stat.value }}</div>
              </v-card>
            </div>
          </div>
          <div v-else-if="selected === 'users'">
            <div class="section-header">
              <h2 class="dashboard-title">User Management</h2>
            </div>
            <v-card class="mb-6 pa-6 elevation-2" color="surface-light">
              <v-row class="mb-4" align="center" justify="space-between">
                <v-col cols="12" md="4">
                  <v-text-field v-model="searchQuery" label="Search users by name or email" prepend-inner-icon="mdi-magnify" variant="outlined" dense clearable />
                </v-col>
                <v-col cols="12" md="8" class="d-flex justify-end">
                  <v-form @submit.prevent="createUser" class="d-flex align-center" style="gap: 8px;">
                    <v-text-field v-model="newUserName" label="Name" prepend-inner-icon="mdi-account" variant="outlined" dense required hide-details />
                    <v-text-field v-model="newUserEmail" label="Email" prepend-inner-icon="mdi-email" variant="outlined" dense required hide-details />
                    <v-btn type="submit" color="primary" class="ml-2" variant="elevated">Create User</v-btn>
                  </v-form>
                </v-col>
              </v-row>
              <v-divider class="mb-4" />
              <v-data-table
                :headers="[
                  { title: 'ID', key: 'id' },
                  { title: 'Name', key: 'name' },
                  { title: 'Email', key: 'email' },
                  { title: 'Actions', key: 'actions', sortable: false }
                ]"
                :items="filteredUsers"
                :items-per-page="pageSize"
                :page.sync="currentPage"
                class="elevation-0 admin-table"
                hide-default-footer
                density="comfortable"
                color="surface-light"
                item-class="admin-table-row"
              >
                <template v-slot:item.actions="{ item }">
                  <v-btn color="error" size="small" @click="deleteUser(item.id)" variant="elevated">Delete</v-btn>
                </template>
              </v-data-table>
              <v-pagination
                v-model="currentPage"
                :length="totalPages"
                class="mt-4"
                color="primary"
                circle
                size="small"
              />
            </v-card>
            <v-card class="pa-6 elevation-2" color="surface-light">
              <div class="section-header mb-4">
                <v-icon icon="mdi-cash-plus" color="accent" size="28" class="mr-2" />
                <h3 class="mb-0">Apply Manual Credit</h3>
              </div>
              <v-form @submit.prevent="applyManualCredit">
                <v-row>
                  <v-col cols="12" md="4">
                    <v-text-field v-model="creditUserId" label="User ID" prepend-inner-icon="mdi-account" variant="outlined" dense required />
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-text-field v-model.number="creditAmount" label="Amount ($)" type="number" prepend-inner-icon="mdi-currency-usd" variant="outlined" min="0.01" step="0.01" dense required />
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-text-field v-model="creditReason" label="Reason/Memo" prepend-inner-icon="mdi-note-text" variant="outlined" dense required />
                  </v-col>
                </v-row>
                <v-btn type="submit" color="primary" class="mt-2" variant="elevated">Apply Credit</v-btn>
              </v-form>
              <v-alert v-if="creditResult" type="success" dense class="mt-4">{{ creditResult }}</v-alert>
            </v-card>
          </div>
          <div v-else-if="selected === 'clusters'">
            <h2 class="dashboard-title">Cluster Management</h2>
            <div class="dashboard-card">Cluster management table (coming soon)</div>
          </div>
          <div v-else-if="selected === 'credits'">
            <div class="section-header">
              <h2 class="dashboard-title">Credits</h2>
            </div>
            <v-card class="elevation-2 pa-6 cloud-card" color="surface-light">
              <template v-if="manualCredits.length">
                <v-data-table
                  :headers="[
                    { title: 'User', key: 'user' },
                    { title: 'Amount', key: 'amount' },
                    { title: 'Reason', key: 'reason' },
                    { title: 'Date', key: 'date' }
                  ]"
                  :items="manualCredits"
                  class="elevation-0 admin-table"
                  hide-default-footer
                  density="comfortable"
                  color="surface-light"
                />
              </template>
              <template v-else>
                <div class="empty-state-content py-12">
                  <div class="empty-state-title">No Credits Yet</div>
                  <div class="empty-state-message">You haven't added any manual credits yet. Use the Users section to apply credits to a user.</div>
                </div>
              </template>
            </v-card>
          </div>
          <div v-else-if="selected === 'system'">
            <div class="section-header">
              <h2 class="dashboard-title">System Stats</h2>
            </div>
            <v-card class="empty-state-card cloud-card elevation-2" color="surface-light">
              <div class="empty-state-content">
                <div class="empty-state-title">System stats and logs (coming soon)</div>
                <div class="empty-state-message">This section will show system health, logs, and platform status for admins.</div>
              </div>
            </v-card>
          </div>
          <div v-else-if="selected === 'vouchers'">
            <div class="section-header">
              <h2 class="dashboard-title">Voucher Management</h2>
            </div>
            <v-card class="dashboard-card admin-voucher-card elevation-2" color="surface-light">
              <h3 class="mb-4">Generate Vouchers</h3>
              <v-form @submit.prevent="generateVouchers">
                <v-row>
                  <v-col cols="12" md="4">
                    <v-text-field v-model.number="voucherValue" label="Voucher Value ($)" type="number" prepend-inner-icon="mdi-currency-usd" variant="outlined" min="1" dense required />
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-text-field v-model.number="voucherCount" label="Number of Vouchers" type="number" prepend-inner-icon="mdi-pound" variant="outlined" min="1" dense required />
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-text-field v-model="voucherExpiry" label="Expiry Date" type="date" prepend-inner-icon="mdi-calendar" variant="outlined" dense required />
                  </v-col>
                </v-row>
                <v-btn type="submit" color="primary" class="mt-2" variant="elevated">Generate</v-btn>
              </v-form>
              <v-alert v-if="voucherResult" type="success" dense class="mt-4">{{ voucherResult }}</v-alert>
            </v-card>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-container {
  min-height: 100vh;
  background: var(--color-bg);
  color: var(--color-text);
}

.dashboard-content-wrapper {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.dashboard-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 2rem;
  align-items: start;
}

.admin-sidebar {
  position: sticky;
  top: 2rem;
}

.dashboard-main {
  background: var(--glass-bg);
  border-radius: 16px;
  padding: 2rem;
  backdrop-filter: var(--glass-blur);
  border: 1px solid var(--color-border);
  box-shadow: var(--cloud-shadow);
}

.section-header {
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid var(--color-border);
}

.dashboard-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
}

.dashboard-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.dashboard-card {
  background: var(--glass-bg-light) !important;
  border: 1px solid var(--color-border) !important;
  border-radius: 16px !important;
  padding: 1.5rem !important;
  transition: all var(--transition-normal) !important;
  backdrop-filter: var(--glass-blur) !important;
}

.dashboard-card:hover {
  transform: translateY(-4px) !important;
  box-shadow: var(--shadow-cloud) !important;
  border-color: var(--color-accent) !important;
}

.admin-stat-card {
  text-align: center;
  position: relative;
  overflow: hidden;
}

.admin-stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--cloud-gradient);
  opacity: 0;
  transition: opacity var(--transition-normal);
  pointer-events: none;
}

.admin-stat-card:hover::before {
  opacity: 0.1;
}

.stat-icon {
  margin-bottom: 1rem;
  filter: drop-shadow(0 0 8px var(--color-accent));
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
}

.empty-state-card {
  text-align: center;
  padding: 4rem 2rem !important;
  background: var(--glass-bg-light) !important;
  border: 1px solid var(--color-border) !important;
}

.empty-state-content {
  max-width: 400px;
  margin: 0 auto;
}

.empty-state-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 1rem;
}

.empty-state-message {
  color: var(--color-text-muted);
  line-height: 1.6;
}

.admin-table {
  background: transparent !important;
}

.admin-table .v-data-table__wrapper {
  background: transparent !important;
}

.admin-table .v-data-table__table {
  background: transparent !important;
}

.admin-table .v-data-table__tr {
  background: transparent !important;
  border-bottom: 1px solid var(--color-border) !important;
}

.admin-table .v-data-table__td {
  color: var(--color-text) !important;
  border-bottom: 1px solid var(--color-border) !important;
}

.admin-table .v-data-table__th {
  color: var(--color-text-muted) !important;
  font-weight: 600 !important;
  border-bottom: 2px solid var(--color-border) !important;
}

@media (max-width: 1024px) {
  .dashboard-layout {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .admin-sidebar {
    position: static;
  }
  
  .dashboard-content-wrapper {
    padding: 1rem;
  }
  
  .dashboard-main {
    padding: 1.5rem;
  }
  
  .dashboard-cards {
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }
}

@media (max-width: 768px) {
  .dashboard-cards {
    grid-template-columns: 1fr;
  }
  
  .dashboard-title {
    font-size: 1.5rem;
  }
  
  .dashboard-main {
    padding: 1rem;
  }
}
</style> 