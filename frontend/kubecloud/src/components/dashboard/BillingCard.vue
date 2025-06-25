<template>
  <div class="dashboard-card billing-card">
    <h2 class="dashboard-card-title mb-6">Billing History</h2>
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
          <td class="table-cell">
            {{ bill.amount > 0 ? '+' : '' }}${{ Math.abs(bill.amount).toFixed(2) }}
          </td>
          <td class="table-cell">
            <v-btn class="action-btn">Download</v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>
    <div class="d-flex justify-center align-center mt-4 billing-pagination">
      <v-btn class="action-btn pagination-btn" :disabled="currentPage === 1" @click="prevPage">&lt; Prev</v-btn>
      <v-btn class="action-btn pagination-btn mx-2" :class="{ 'active-page': currentPage === 1 }" @click="goToPage(1)">1</v-btn>
      <v-btn class="action-btn pagination-btn mx-2" :class="{ 'active-page': currentPage === 2 }" @click="goToPage(2)">2</v-btn>
      <v-btn class="action-btn pagination-btn" :disabled="currentPage === 2" @click="nextPage">Next &gt;</v-btn>
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
.dashboard-card.billing-card {
  background: var(--color-surface);
  border-radius: var(--rounded);
  border: 1px solid var(--color-border);
  box-shadow: none;
  color: var(--color-text);
  padding: 1.25rem;
  min-width: 100%;
}

.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 0.5rem;
}

.billing-table {
  background: transparent;
  border-radius: var(--rounded);
  overflow: hidden;
}
.table-header {
  background: var(--color-surface);
  color: var(--color-text-muted);
  font-weight: 600;
  font-size: 1rem;
  padding: 1rem;
  border-bottom: 1px solid var(--color-border);
  text-align: left;
}
.table-row {
  background: transparent;
  transition: background var(--transition);
}
.table-row:hover {
  background: var(--color-bg);
}
.table-cell {
  color: var(--color-text);
  font-size: 1rem;
  padding: 1rem;
  border-bottom: 1px solid var(--color-border);
  text-align: left;
}
.action-btn {
  background: var(--color-accent);
  color: #fff;
  border-radius: var(--rounded);
  border: none;
  box-shadow: none;
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  transition: background var(--transition);
}
.action-btn:hover {
  background: var(--color-accent-hover);
}
.billing-pagination {
  gap: 0.5rem;
}
.pagination-btn.active-page {
  background: var(--color-accent-hover);
  color: #fff;
}
</style>
