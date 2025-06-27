<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="dashboard-card-title">Billing History</h3>
          <p class="dashboard-card-subtitle">View and manage your billing history and invoices</p>
        </div>
      </div>
    </div>
    <div class="billing-table-container">
      <v-table class="billing-table">
        <thead>
          <tr>
            <th class="table-header">Date</th>
            <th class="table-header">Description</th>
            <th class="table-header">Amount</th>
            <th class="table-header">Invoice</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="bill in billingHistory" :key="bill.id" class="table-row">
            <td class="table-cell">{{ bill.date }}</td>
            <td class="table-cell description-cell">{{ bill.description }}</td>
            <td class="table-cell amount-cell">
              {{ bill.amount > 0 ? '+' : '' }}${{ Math.abs(bill.amount).toFixed(2) }}
            </td>
            <td class="table-cell action-cell">
              <v-btn variant="outlined" size="small" class="action-btn">Download</v-btn>
            </td>
          </tr>
        </tbody>
      </v-table>
    </div>
    <div class="billing-pagination">
      <v-btn class="pagination-btn action-btn" :disabled="currentPage === 1" @click="prevPage">&lt; Prev</v-btn>
      <v-btn class="pagination-btn action-btn mx-2" :class="{ 'active-page': currentPage === 1 }" @click="goToPage(1)">1</v-btn>
      <v-btn class="pagination-btn action-btn mx-2" :class="{ 'active-page': currentPage === 2 }" @click="goToPage(2)">2</v-btn>
      <v-btn class="pagination-btn action-btn" :disabled="currentPage === 2" @click="nextPage">Next &gt;</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, ref } from 'vue'
interface Bill {
  id: string | number
  date: string
  description: string
  amount: number
}
const props = defineProps<{ billingHistory: Bill[] }>()
const currentPage = ref(1)
function prevPage() {
  if (currentPage.value > 1) currentPage.value--
}
function nextPage() {
  if (currentPage.value < 2) currentPage.value++
}
function goToPage(page: number) {
  currentPage.value = page
}
</script>

<style scoped>
.billing-table-container {
  margin-bottom: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.billing-table {
  background: transparent;
  width: 100%;
}

.table-header {
  background: rgba(59, 130, 246, 0.1);
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
  padding: var(--space-4) var(--space-3);
  border-bottom: 1px solid var(--color-border);
  text-align: left;
  white-space: nowrap;
}

.table-row {
  background: transparent;
  transition: all var(--transition-normal);
}

.table-row:hover {
  background: rgba(59, 130, 246, 0.05);
}

.table-cell {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  padding: var(--space-4) var(--space-3);
  border-bottom: 1px solid var(--color-border);
  text-align: left;
  vertical-align: middle;
}

.description-cell {
  max-width: 200px;
  word-wrap: break-word;
}

.amount-cell {
  text-align: right;
  font-weight: var(--font-weight-medium);
  white-space: nowrap;
}

.action-cell {
  text-align: center;
  white-space: nowrap;
}

.action-btn {
  background: transparent !important;
  border: 1px solid var(--color-border) !important;
  color: var(--color-text) !important;
  font-weight: var(--font-weight-medium);
  transition: all var(--transition-normal);
  white-space: nowrap;
}

.action-btn:hover {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}

.billing-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-6);
}

.pagination-btn {
  font-size: var(--font-size-sm);
  min-width: 60px;
}

.pagination-btn:disabled {
  background: transparent !important;
  border-color: var(--color-border) !important;
  color: var(--color-text-disabled) !important;
}

.pagination-btn.active-page {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}

@media (max-width: 768px) {
  .table-header,
  .table-cell {
    padding: var(--space-3) var(--space-2);
    font-size: var(--font-size-xs);
  }
  
  .description-cell {
    max-width: 120px;
  }
}

@media (max-width: 480px) {
  .billing-table-container {
    overflow-x: auto;
  }
  
  .billing-table {
    min-width: 400px;
  }

  .table-header,
  .table-cell {
    padding: var(--space-2) var(--space-1);
    font-size: var(--font-size-xs);
  }

  .billing-pagination {
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-1);
  }

  .pagination-btn {
    font-size: var(--font-size-xs);
    min-width: 50px;
  }
}
</style>
