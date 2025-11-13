<template>
  <div class="dashboard-card">
    <div class="dashboard-card-header">
      <div class="dashboard-card-title-section">
        <div class="dashboard-card-title-content">
          <h3 class="text-h5 font-weight-bold mb-1">Billing History</h3>
          <p class="text-body-1 dashboard-card-subtitle">View and manage your billing history and invoices</p>
        </div>
      </div>
    </div>
    <div class="billing-table-container">
      <v-data-table
        :headers="headers"
        :items="billingHistory"
        class="billing-table"
        :items-per-page="5"
        :no-data-text="'No invoices found'"
        density="comfortable"
        :hide-default-footer="false"
      >
        <template v-slot:[`item.amount`]="{ item }">
          <span>{{ item.amount > 0 ? '+' : '' }}${{ Math.abs(item.amount).toFixed(2) }}</span>
        </template>
        <template v-slot:[`item.invoice`]="{ item }">
          <v-btn variant="outlined" size="small" class="action-btn" @click="$emit('downloadInvoice', item.id)">Download</v-btn>
        </template>
      </v-data-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'
interface Bill {
  id: string | number
  date: string
  description: string
  amount: number
}
const props = defineProps<{ billingHistory: Bill[] }>()
const emit = defineEmits(['downloadInvoice'])

const headers = [
  { title: 'Date', key: 'date' },
  { title: 'Description', key: 'description' },
  { title: 'Amount', key: 'amount' },
  { title: 'Invoice', key: 'invoice', sortable: false },
]
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
</style>
