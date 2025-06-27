<script setup lang="ts">
import { ref, computed, onMounted, defineAsyncComponent } from 'vue'

// Use defineAsyncComponent to avoid TypeScript issues
const AdminSidebar = defineAsyncComponent(() => import('../components/AdminSidebar.vue'))

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
const creditUserObj = ref<{ id: number; name: string; email: string } | null>(null)
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
  if (!creditUserObj.value) {
    showSnackbar('Please select a user.', 'error')
    return
  }
  creditResult.value = `Credited user ${creditUserObj.value.name} with $${creditAmount.value} for: ${creditReason.value}`
  showSnackbar('Manual credit applied!', 'success')
  // Optionally reset form
  creditUserObj.value = null
  creditAmount.value = 0
  creditReason.value = ''
}

function editUser(user: { id: number; name: string; email: string }) {
  alert('Edit user: ' + user.name)
}

const tabs = [
  { key: 'overview', label: 'Overview' },
  { key: 'users', label: 'Users' },
  { key: 'clusters', label: 'Clusters' },
  { key: 'system', label: 'System' },
  { key: 'vouchers', label: 'Vouchers' },
]

onMounted(() => {
  console.log('AdminDashboard mounted, selected:', selected.value)
})
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
          <!-- Overview Section -->
          <div v-if="selected === 'overview'" class="admin-section">
            <div class="section-header">
              <h2 class="dashboard-title">Admin Overview</h2>
              <p class="section-subtitle">Monitor platform health and key metrics</p>
            </div>
            <div class="stats-grid">
              <div v-for="stat in adminStats" :key="stat.label" class="stat-item">
                <div class="stat-icon">
                  <v-icon :icon="stat.icon" size="24" :color="stat.color"></v-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-number">{{ stat.value }}</div>
                  <div class="stat-label">{{ stat.label }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Users Section -->
          <div v-else-if="selected === 'users'" class="admin-section">
            <div class="section-header">
              <h2 class="dashboard-title">User Management</h2>
              <p class="section-subtitle">Manage user accounts and permissions</p>
            </div>
            
            <div class="dashboard-card">
              <div class="dashboard-card-header">
                <h3 class="dashboard-card-title">Create New User</h3>
                <p class="dashboard-card-subtitle">Add a new user to the platform</p>
              </div>
              <v-form @submit.prevent="createUser" class="create-user-form">
                <div class="form-row">
                  <v-text-field 
                    v-model="newUserName" 
                    label="Name" 
                    prepend-inner-icon="mdi-account" 
                    variant="outlined" 
                    density="comfortable"
                    required 
                    hide-details 
                    class="form-field"
                  />
                  <v-text-field 
                    v-model="newUserEmail" 
                    label="Email" 
                    prepend-inner-icon="mdi-email" 
                    variant="outlined" 
                    density="comfortable"
                    required 
                    hide-details 
                    class="form-field"
                  />
                  <v-btn type="submit" color="primary" variant="elevated" class="btn-primary">
                    <v-icon icon="mdi-plus" class="mr-2"></v-icon>
                    Create User
                  </v-btn>
                </div>
              </v-form>
            </div>

            <div class="dashboard-card">
              <div class="dashboard-card-header">
                <h3 class="dashboard-card-title">User Search</h3>
                <p class="dashboard-card-subtitle">Find and manage existing users</p>
              </div>
              <v-text-field 
                v-model="searchQuery" 
                label="Search users by name or email" 
                prepend-inner-icon="mdi-magnify" 
                variant="outlined" 
                density="comfortable"
                clearable 
                class="search-field"
              />
              
              <div class="table-container">
                <v-data-table
                  :headers="[
                    { title: 'ID', key: 'id', width: '80px' },
                    { title: 'Name', key: 'name' },
                    { title: 'Email', key: 'email' },
                    { title: 'Actions', key: 'actions', sortable: false, width: '160px' }
                  ]"
                  :items="filteredUsers"
                  :items-per-page="pageSize"
                  :page="currentPage"
                  @update:page="currentPage = $event"
                  class="admin-table"
                  hide-default-footer
                  density="comfortable"
                >
                  <template #item.actions="{ item }">
                    <div style="display: flex; gap: var(--space-4); align-items: center;">
                      <v-btn size="small" variant="outlined" class="action-btn" @click="editUser(item)">
                        <v-icon icon="mdi-pencil" size="16" class="mr-1"></v-icon>
                        Edit
                      </v-btn>
                      <v-btn size="small" variant="outlined" class="action-btn" @click="deleteUser(item.id)">
                        <v-icon icon="mdi-delete" size="16" class="mr-1"></v-icon>
                        Remove
                      </v-btn>
                    </div>
                  </template>
                </v-data-table>
              </div>
              
              <div class="pagination-container">
                <v-pagination
                  v-model="currentPage"
                  :length="totalPages"
                  color="primary"
                  circle
                  size="small"
                />
              </div>
            </div>
          </div>

          <!-- Clusters Section -->
          <div v-else-if="selected === 'clusters'" class="admin-section">
            <div class="section-header">
              <h2 class="dashboard-title">Cluster Management</h2>
              <p class="section-subtitle">Monitor and manage all platform clusters</p>
            </div>
            <div class="dashboard-card">
              <div class="empty-state-content">
                <v-icon icon="mdi-server-network" size="64" color="var(--color-text-muted)" class="mb-4"></v-icon>
                <h3 class="empty-state-title">Cluster Management</h3>
                <p class="empty-state-message">Advanced cluster management features coming soon. Monitor cluster health, performance metrics, and resource utilization.</p>
              </div>
            </div>
          </div>

          <!-- Credits Section -->
          <div v-else-if="selected === 'credits'" class="admin-section">
            <div class="section-header">
              <h2 class="dashboard-title">Credits History</h2>
              <p class="section-subtitle">Track all manual credit applications</p>
            </div>
            <div class="dashboard-card">
              <template v-if="manualCredits.length">
                <div class="dashboard-card-header">
                  <h3 class="dashboard-card-title">Recent Credits</h3>
                  <p class="dashboard-card-subtitle">Manual credits applied to user accounts</p>
                </div>
                <div class="table-container">
                  <v-data-table
                    :headers="[
                      { title: 'User', key: 'user' },
                      { title: 'Amount', key: 'amount' },
                      { title: 'Reason', key: 'reason' },
                      { title: 'Date', key: 'date' }
                    ]"
                    :items="manualCredits"
                    class="admin-table"
                    hide-default-footer
                    density="comfortable"
                  />
                </div>
              </template>
              <template v-else>
                <div class="empty-state-content">
                  <v-icon icon="mdi-cash-plus" size="64" color="var(--color-text-muted)" class="mb-4"></v-icon>
                  <h3 class="empty-state-title">No Credits Yet</h3>
                  <p class="empty-state-message">You haven't added any manual credits yet. Use the Users section to apply credits to a user.</p>
                </div>
              </template>
            </div>

            <div class="dashboard-card">
              <div class="dashboard-card-header">
                <h3 class="dashboard-card-title">Manual Credit</h3>
                <p class="section-subtitle">Apply credits to user accounts</p>
              </div>
              <v-form @submit.prevent="applyManualCredit" class="credit-form">
                <div class="form-row">
                  <v-select
                    v-model="creditUserObj"
                    :items="users"
                    item-title="name"
                    item-value="id"
                    label="User"
                    return-object
                    :menu-props="{ maxHeight: '300px' }"
                    prepend-inner-icon="mdi-account"
                    variant="outlined"
                    density="comfortable"
                    required
                    class="form-field"
                    :item-props="user => ({ title: user.name + ' (' + user.email + ')', value: user })"
                  />
                  <v-text-field 
                    v-model.number="creditAmount" 
                    label="Amount ($)" 
                    type="number" 
                    prepend-inner-icon="mdi-currency-usd" 
                    variant="outlined" 
                    min="0.01" 
                    step="0.01" 
                    density="comfortable"
                    required 
                    class="form-field"
                  />
                  <v-text-field 
                    v-model="creditReason" 
                    label="Reason/Memo" 
                    prepend-inner-icon="mdi-note-text" 
                    variant="outlined" 
                    density="comfortable"
                    required 
                    class="form-field"
                  />
                </div>
                <v-btn type="submit" color="primary" variant="elevated" class="btn-primary">
                  <v-icon icon="mdi-cash-plus" class="mr-2"></v-icon>
                  Apply Credit
                </v-btn>
              </v-form>
              <v-alert v-if="creditResult" type="success" variant="tonal" class="mt-4">{{ creditResult }}</v-alert>
            </div>
          </div>

          <!-- System Section -->
          <div v-else-if="selected === 'system'" class="admin-section">
            <div class="section-header">
              <h2 class="dashboard-title">System Stats</h2>
              <p class="section-subtitle">Platform health and performance metrics</p>
            </div>
            <div class="dashboard-card">
              <div class="empty-state-content">
                <v-icon icon="mdi-cog" size="64" color="var(--color-text-muted)" class="mb-4"></v-icon>
                <h3 class="empty-state-title">System Monitoring</h3>
                <p class="empty-state-message">Advanced system monitoring, logs, and platform status features coming soon. Monitor system health, performance metrics, and resource utilization.</p>
              </div>
            </div>
          </div>

          <!-- Vouchers Section -->
          <div v-else-if="selected === 'vouchers'" class="admin-section">
            <div class="section-header">
              <h2 class="dashboard-title">Voucher Management</h2>
              <p class="section-subtitle">Generate and manage platform vouchers</p>
            </div>
            <div class="dashboard-card">
              <div class="dashboard-card-header">
                <h3 class="dashboard-card-title">Generate Vouchers</h3>
                <p class="dashboard-card-subtitle">Create new vouchers for user promotions</p>
              </div>
              <v-form @submit.prevent="generateVouchers" class="voucher-form">
                <div class="form-row">
                  <v-text-field 
                    v-model.number="voucherValue" 
                    label="Voucher Value ($)" 
                    type="number" 
                    prepend-inner-icon="mdi-currency-usd" 
                    variant="outlined" 
                    min="1" 
                    density="comfortable"
                    required 
                    class="form-field"
                  />
                  <v-text-field 
                    v-model.number="voucherCount" 
                    label="Number of Vouchers" 
                    type="number" 
                    prepend-inner-icon="mdi-pound" 
                    variant="outlined" 
                    min="1" 
                    density="comfortable"
                    required 
                    class="form-field"
                  />
                  <v-text-field 
                    v-model="voucherExpiry" 
                    label="Expiry Date" 
                    type="date" 
                    prepend-inner-icon="mdi-calendar" 
                    variant="outlined" 
                    density="comfortable"
                    required 
                    class="form-field"
                  />
                </div>
                <v-btn type="submit" color="primary" variant="elevated" class="btn-primary">
                  <v-icon icon="mdi-ticket-percent" class="mr-2"></v-icon>
                  Generate Vouchers
                </v-btn>
              </v-form>
              <v-alert v-if="voucherResult" type="success" variant="tonal" class="mt-4">{{ voucherResult }}</v-alert>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-container {
  min-height: 100vh;
  background: var(--color-bg, #0F172A);
  position: relative;
  overflow-x: hidden;
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
  margin-top: 10rem;
}

.admin-sidebar {
  flex: 0 0 280px;
  display: flex;
  flex-direction: column;
  height: fit-content;
  position: sticky;
  top: 0;
  align-self: flex-start;
  margin-top: 0;
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border, #334155);
  border-radius: var(--radius-xl, 0.75rem);
  padding: 1.5rem;
  backdrop-filter: blur(8px);
}

.dashboard-main {
  flex: 1;
  min-width: 0;
}

.admin-section {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.section-header {
  text-align: left;
  margin-bottom: 1rem;
}

.dashboard-title {
  font-size: var(--font-size-3xl, 1.875rem);
  font-weight: var(--font-weight-bold, 700);
  margin-bottom: 0.5rem;
  line-height: 1.2;
  color: var(--color-text, #F8FAFC);
  letter-spacing: -0.5px;
}

.section-subtitle {
  font-size: var(--font-size-lg, 1.125rem);
  color: var(--color-text-secondary, #CBD5E1);
  line-height: 1.5;
  margin: 0;
  font-weight: var(--font-weight-normal, 400);
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.stat-item {
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border, #334155);
  border-radius: var(--radius-xl, 0.75rem);
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: all var(--transition-normal, 0.2s);
  backdrop-filter: blur(8px);
}

.stat-item:hover {
  border-color: var(--color-border-light, #475569);
  background: rgba(15, 30, 52, 0.75);
  transform: translateY(-1px);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-lg, 0.5rem);
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid var(--color-primary, #3B82F6);
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: var(--font-size-2xl, 1.5rem);
  font-weight: var(--font-weight-bold, 700);
  color: var(--color-text, #F8FAFC);
  line-height: 1.2;
}

.stat-label {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--color-text-secondary, #CBD5E1);
  margin-top: 0.25rem;
}

/* Dashboard Cards */
.dashboard-card {
  background: rgba(10, 25, 47, 0.65);
  border: 1px solid var(--color-border, #334155);
  border-radius: var(--radius-xl, 0.75rem);
  padding: 2rem;
  transition: all var(--transition-normal, 0.2s);
  backdrop-filter: blur(8px);
}

.dashboard-card:hover {
  border-color: var(--color-border-light, #475569);
  background: rgba(15, 30, 52, 0.75);
}

.dashboard-card-header {
  margin-bottom: 1.5rem;
}

.dashboard-card-title {
  font-size: var(--font-size-xl, 1.25rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text, #F8FAFC);
  margin: 0 0 0.5rem 0;
}

.dashboard-card-subtitle {
  font-size: var(--font-size-base, 1rem);
  color: var(--color-text-secondary, #CBD5E1);
  margin: 0;
}

/* Forms */
.create-user-form,
.credit-form,
.voucher-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.create-user-form .form-row {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.create-user-form .form-row > .form-field {
  width: 100%;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  align-items: end;
}

.form-field {
  flex: 1;
}

.search-field {
  margin-bottom: 1.5rem;
}

/* Buttons */
.btn-primary {
  background: var(--color-primary, #3B82F6);
  color: white;
  border-color: var(--color-primary, #3B82F6);
  font-weight: var(--font-weight-medium, 500);
  padding: 0.75rem 1.5rem;
  border-radius: var(--radius-md, 0.375rem);
  transition: all var(--transition-normal, 0.2s);
}

.btn-primary:hover {
  background: var(--color-primary-dark, #1E40AF);
  border-color: var(--color-primary-dark, #1E40AF);
  transform: translateY(-1px);
}

.btn-danger {
  border-color: var(--color-error, #EF4444);
  color: var(--color-error, #EF4444);
}

.btn-danger:hover {
  background: var(--color-error, #EF4444);
  color: white;
}

.btn-full {
  width: 100% !important;
  display: flex !important;
  justify-content: center;
  align-items: center;
}

/* Tables */
.table-container {
  margin: 1.5rem 0;
}

.admin-table {
  background: transparent !important;
}

.admin-table :deep(.v-data-table__wrapper) {
  background: transparent !important;
}

.admin-table :deep(.v-data-table-header) {
  background: rgba(30, 41, 59, 0.5) !important;
  border-bottom: 1px solid var(--color-border, #334155) !important;
}

.admin-table :deep(.v-data-table__tr) {
  background: transparent !important;
  border-bottom: 1px solid var(--color-border, #334155) !important;
}

.admin-table :deep(.v-data-table__tr:hover) {
  background: rgba(30, 41, 59, 0.3) !important;
}

.admin-table :deep(.v-data-table__td) {
  color: var(--color-text, #F8FAFC) !important;
  border-bottom: none !important;
}

.admin-table :deep(.v-data-table__th) {
  color: var(--color-text-secondary, #CBD5E1) !important;
  font-weight: var(--font-weight-medium, 500) !important;
  border-bottom: none !important;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 1.5rem;
}

/* Empty States */
.empty-state-content {
  text-align: center;
  padding: 3rem 2rem;
}

.empty-state-title {
  font-size: var(--font-size-xl, 1.25rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text, #F8FAFC);
  margin: 1rem 0 0.5rem 0;
}

.empty-state-message {
  font-size: var(--font-size-base, 1rem);
  color: var(--color-text-secondary, #CBD5E1);
  line-height: 1.6;
  max-width: 500px;
  margin: 0 auto;
}

/* Responsive Design */
@media (max-width: 900px) {
  .dashboard-layout {
    flex-direction: column;
    gap: 1.5rem;
  }
  
  .admin-sidebar {
    flex: none;
    width: 100%;
    position: static;
  }
  
  .stats-grid {
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .dashboard-card {
    padding: 1.5rem;
  }
}

@media (max-width: 600px) {
  .dashboard-content-wrapper {
    padding: 0 0.5rem;
    margin-top: 2rem;
  }
  
  .dashboard-title {
    font-size: var(--font-size-2xl, 1.5rem);
  }
  
  .section-subtitle {
    font-size: var(--font-size-base, 1rem);
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .dashboard-card {
    padding: 1rem;
  }
  
  .empty-state-content {
    padding: 2rem 1rem;
  }
}
</style> 