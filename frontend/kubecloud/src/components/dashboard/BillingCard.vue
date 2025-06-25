<template>
  <v-card class="pa-8 profile-card billing-card card-enhanced">
    <h2 class="display-3 mb-6 neon-glow-violet">Billing History</h2>
    <v-table class="mb-6 billing-table">
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
          <td class="table-cell">{{ bill.description }}</td>
          <td :class="['table-cell', 'amount-cell', { 'text-success': bill.amount > 0, 'text-error': bill.amount < 0 }]">
            {{ bill.amount > 0 ? '+' : '' }}${{ Math.abs(bill.amount).toFixed(2) }}
          </td>
          <td class="table-cell">
            <v-btn variant="text" color="primary" size="large" class="download-btn neon-hover-sky">Download</v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>
    <div class="d-flex justify-center align-center mt-4 billing-pagination">
      <v-btn variant="text" size="large" :disabled="currentPage === 1" @click="prevPage" class="pagination-btn neon-hover-violet">&lt; Prev</v-btn>
      <v-btn variant="text" size="large" class="mx-2 pagination-btn" :class="currentPage === 1 ? 'neon-glow-violet' : 'neon-hover-violet'" @click="goToPage(1)">1</v-btn>
      <v-btn variant="text" size="large" class="mx-2 pagination-btn" :class="currentPage === 2 ? 'neon-glow-violet' : 'neon-hover-violet'" @click="goToPage(2)">2</v-btn>
      <v-btn variant="text" size="large" :disabled="currentPage === 2" @click="nextPage" class="pagination-btn neon-hover-violet">Next &gt;</v-btn>
    </div>
  </v-card>
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
.billing-card {
  border-radius: 24px;
  min-width: 100%;
}

.billing-table {
  background: transparent;
  border-radius: 16px;
  overflow: hidden;
}

.table-header {
  background: rgba(79, 70, 229, 0.1) !important;
  color: var(--kubecloud-blue) !important;
  font-weight: 600 !important;
  font-size: 1.125rem !important;
  padding: 1.25rem 1rem !important;
  border-bottom: 2px solid var(--kubecloud-blue) !important;
  text-align: left !important;
}

.table-row {
  background: transparent !important;
  transition: all var(--transition-normal);
}

.table-row:hover {
  background: rgba(79, 70, 229, 0.05) !important;
  transform: translateX(2px);
}

.table-cell {
  color: #fff !important;
  font-size: 1.125rem !important;
  padding: 1.25rem 1rem !important;
  border-bottom: 1px solid rgba(79, 70, 229, 0.2) !important;
  text-align: left !important;
}

.amount-cell {
  font-weight: 600 !important;
  font-family: 'Space Grotesk', 'Inter', sans-serif !important;
}

.text-success {
  color: var(--kubecloud-success) !important;
  text-shadow: 0 0 1px rgba(16, 185, 129, 0.3) !important;
}

.text-error {
  color: var(--kubecloud-error) !important;
  text-shadow: 0 0 1px rgba(239, 68, 68, 0.3) !important;
}

.download-btn {
  font-weight: 600;
  font-size: 1rem;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: transparent;
  border: 1px solid transparent;
  transition: all var(--transition-normal);
  color: #fff !important;
}

.download-btn:hover {
  background: rgba(79, 70, 229, 0.1);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
}

.billing-pagination {
  gap: 0.5rem;
}

.pagination-btn {
  font-weight: 600;
  font-size: 1rem;
  padding: 0.75rem 1rem;
  border-radius: 10px;
  text-transform: none;
  letter-spacing: 0.01em;
  background: transparent;
  border: 1px solid transparent;
  transition: all var(--transition-normal);
  color: #fff !important;
  min-width: 60px;
}

.pagination-btn:hover:not(:disabled) {
  background: rgba(79, 70, 229, 0.1);
  border: 1px solid var(--kubecloud-blue);
  box-shadow: var(--shadow-kubecloud-blue);
  text-shadow: 0 0 1px rgba(79, 70, 229, 0.3);
  transform: translateY(-1px);
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 960px) {
  .billing-card {
    padding: 2rem !important;
  }
  
  .table-header {
    font-size: 1rem !important;
    padding: 1rem 0.75rem !important;
  }
  
  .table-cell {
    font-size: 1rem !important;
    padding: 1rem 0.75rem !important;
  }
  
  .download-btn {
    font-size: 0.9rem;
    padding: 0.375rem 0.75rem;
  }
  
  .pagination-btn {
    font-size: 0.9rem;
    padding: 0.5rem 0.75rem;
    min-width: 50px;
  }
}

@media (max-width: 600px) {
  .billing-card {
    padding: 1.5rem !important;
  }
  
  .table-header {
    font-size: 0.9rem !important;
    padding: 0.75rem 0.5rem !important;
  }
  
  .table-cell {
    font-size: 0.9rem !important;
    padding: 0.75rem 0.5rem !important;
  }
  
  .download-btn {
    font-size: 0.8rem;
    padding: 0.25rem 0.5rem;
  }
  
  .pagination-btn {
    font-size: 0.8rem;
    padding: 0.375rem 0.5rem;
    min-width: 40px;
  }
}
</style>
