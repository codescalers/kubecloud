<script setup lang="ts">
import { ref, computed, onMounted, defineAsyncComponent, type Ref } from 'vue'
import { adminService, type User, type Voucher, type GenerateVouchersRequest, type Invoice } from '../utils/adminService'
import AdminUsersTable from '../components/AdminUsersTable.vue'
import AdminStatsCards from '../components/AdminStatsCards.vue'
import AdminManualCredit from '../components/AdminManualCredit.vue'
import AdminVouchersSection from '../components/AdminVouchersTable.vue'
import AdminClustersSection from '../components/AdminClustersSection.vue'
import AdminSystemSection from '../components/AdminSystemCard.vue'
import AdminInvoicesTable from '../components/AdminInvoicesTable.vue'
import AdminPendingRecordsCard from '../components/dashboard/AdminPendingRecordsCard.vue'
import AdminEmailsCard from '../components/dashboard/AdminEmailsCard.vue'

const AdminSidebar = defineAsyncComponent(() => import('../components/AdminSidebar.vue'))

const selected = ref('overview')

const adminStats = ref([
  { label: 'Total Users', value: 0, icon: 'mdi-account-group', color: '#3B82F6' },
  { label: 'Active Clusters', value: 42, icon: 'mdi-server', color: '#3B82F6' },
])

// User management state
const users = ref<User[]>([])
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 5
const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value
  return users.value.filter(u =>
    u.username.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    u.email.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})
const paginatedUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredUsers.value.slice(start, start + pageSize)
})
const totalPages = computed(() => Math.ceil(filteredUsers.value.length / pageSize))

function deleteUser(userId: number) {
    adminService.deleteUser(userId)

    loadUsers()
}

async function loadUsers() {

    const rawUsers = await adminService.listUsers()
    users.value = rawUsers.map(u => ({ ...u, id: u.id ?? (u as any).ID }))

    adminStats.value[0].value = users.value.length
}

function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value) currentPage.value = page
}

// Voucher generation form state
const vouchers = ref<Voucher[]>([])

const creditDialog = ref(false)
const creditUserDialogObj = ref<User | null>(null)

function handleSidebarSelect(newSelected: string) {
  selected.value = newSelected
}

// Generate vouchers using real API
async function generateVouchers(voucherData: GenerateVouchersRequest) {
  await adminService.generateVouchers(voucherData)
  await loadVouchers()
}

async function loadVouchers() {
    vouchers.value = await adminService.listVouchers()
}

function openCreditDialog(user: User) {
  creditUserDialogObj.value = user
  creditDialog.value = true
}

function closeCreditDialog() {
  creditDialog.value = false
  creditUserDialogObj.value = null
}

function handleCreditApplied() {
  loadUsers()
}

const invoices: Ref<Invoice[]> = ref([])

onMounted(async () => {
  // Load initial data
  await loadUsers()
  await loadVouchers()
  await loadInvoices()
})

async function loadInvoices() {
  invoices.value = await adminService.listInvoices()
}
</script>

<template>
  <div class="dashboard-container">
    <div class="dashboard-content-wrapper">
      <div class="dashboard-layout">
        <div class="admin-sidebar">
          <AdminSidebar :selected="selected" @update:selected="handleSidebarSelect" />
        </div>
        <div class="dashboard-main" style="width: 100%;">
          <AdminStatsCards v-if="selected === 'overview'" :adminStats="adminStats" />
          <AdminUsersTable
            v-else-if="selected === 'users'"
            :users="paginatedUsers"
            :searchQuery="searchQuery"
            :currentPage="currentPage"
            :pageSize="pageSize"
            :totalPages="totalPages"
            @update:searchQuery="searchQuery = $event"
            @update:currentPage="goToPage($event)"
            @deleteUser="deleteUser"
            @creditUser="openCreditDialog"
          />
          <AdminClustersSection v-else-if="selected === 'clusters'" />
          <AdminSystemSection v-else-if="selected === 'system'" />
          <AdminVouchersSection
            v-else-if="selected === 'vouchers'"
            :vouchers="vouchers"
            @generateVouchers="generateVouchers"
          />
          <AdminInvoicesTable v-else-if="selected === 'invoices'" :invoices="invoices" />
          <AdminPendingRecordsCard v-else-if="selected === 'payments'" />
          <AdminEmailsCard v-else-if="selected === 'emails'" />
          <v-dialog v-model="creditDialog" max-width="600" persistent>
            <v-card class="pa-4" style="background: rgba(16,24,39,0.98); border-radius: 18px;">
              <v-card-title class="text-h6 font-weight-bold mb-2 text-center">Manual Credit</v-card-title>
              <v-card-subtitle class="mb-4 text-center">
                Apply credits to user: {{ creditUserDialogObj?.email }}
              </v-card-subtitle>

              <AdminManualCredit
                v-if="creditDialog && creditUserDialogObj"
                :userId="creditUserDialogObj.id"
                :userEmail="creditUserDialogObj.email"
                @creditApplied="handleCreditApplied"
                @close="closeCreditDialog"
              />
              <v-card-actions class="justify-end mt-2">
                <v-btn
                  text
                  color="grey-lighten-1"
                  @click="closeCreditDialog"
                >
                  Cancel
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>
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
  max-width: 80%;
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

  .dashboard-card {
    padding: 1.5rem;
  }
}

@media (max-width: 600px) {
  .dashboard-content-wrapper {
    padding: 0 0.5rem;
    margin-top: 2rem;
  }

  .dashboard-card {
    padding: 1rem;
  }
}
</style>
