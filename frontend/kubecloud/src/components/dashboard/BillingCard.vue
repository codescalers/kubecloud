<template>
  <v-card color="surface-variant" class="pa-6">
    <div class="mb-6">
      <h3 class="text-h5 font-weight-semibold mb-2">Billing History</h3>
      <p class="text-body-1 text-medium-emphasis">View and manage your billing history and invoices</p>
    </div>

    <v-data-table
      :headers="headers"
      :items="billingHistory"
      :items-per-page="5"
      :no-data-text="'No invoices found'"
      density="comfortable"
      :hide-default-footer="false"
    >
      <template v-slot:[`item.amount`]="{ item }">
        <span>{{ item.amount > 0 ? '+' : '' }}${{ Math.abs(item.amount).toFixed(2) }}</span>
      </template>
      <template v-slot:[`item.invoice`]="{ item }">
        <v-btn variant="outlined" size="small" color="primary" @click="$emit('downloadInvoice', item.id)">Download</v-btn>
      </template>
    </v-data-table>
  </v-card>
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


