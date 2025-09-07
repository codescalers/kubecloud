<template>
  <div class="admin-section">
    <div class="section-header">
      <h2 class="dashboard-title">Invoices</h2>
      <p class="section-subtitle">All platform invoices</p>
    </div>
    <div class="dashboard-card">
      <div class="dashboard-card-header">
        <h3 class="dashboard-card-title">All Invoices</h3>
        <p class="dashboard-card-subtitle">List of all generated invoices</p>
      </div>
      <div class="table-container">
        <v-data-table
          :headers="headers"
          :items="invoices"
          class="admin-table"
          :page="page"
          :items-per-page="itemsPerPage"
          density="comfortable"
        >
          <template v-slot:item.created_at="{ item }">
            <span>{{ formatDate(item.created_at) }}</span>
          </template>
          <template v-slot:item.total="{ item }">
            ${{ item.total.toFixed(2) }}
          </template>
          <template v-slot:item.tax="{ item }">
            ${{ item.tax.toFixed(2) }}
          </template>
          <template v-slot:item.actions="{ item }">
            <v-btn size="small" variant="outlined" class="action-btn" @click="viewInvoice(item)">
              <v-icon icon="mdi-file-pdf" size="16" class="mr-1"></v-icon>
              View
            </v-btn>
          </template>
        </v-data-table>
      </div>
    </div>
    <v-dialog v-model="showInvoiceModal" max-width="500" persistent>
      <v-card v-if="selectedInvoice" class="pa-4" style="background: rgba(16,24,39,0.98); border-radius: 18px;">
        <v-card-title class="text-h6 font-weight-bold mb-2 text-center">Invoice Details</v-card-title>
        <v-card-text>
          <div><strong>ID:</strong> {{ selectedInvoice.id }}</div>
          <div><strong>User ID:</strong> {{ selectedInvoice.user_id }}</div>
          <div><strong>Total:</strong> ${{ selectedInvoice.total.toFixed(2) }}</div>
          <div><strong>Tax:</strong> ${{ selectedInvoice.tax.toFixed(2) }}</div>
          <div><strong>Created At:</strong> {{ formatDate(selectedInvoice.created_at) }}</div>
          <!-- Add more fields as needed -->
        </v-card-text>
        <v-card-actions class="justify-end mt-2">
          <v-btn text color="grey-lighten-1" @click="closeInvoiceModal">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps } from 'vue'
import {formatDate} from '../utils/uiUtils.ts'
const props = defineProps<{ invoices: any[] }>()
const page = ref(1)
const itemsPerPage = ref(10)

const headers = [
  { title: 'ID', key: 'id', width: '80px' },
  { title: 'User ID', key: 'user_id', width: '100px' },
  { title: 'Total', key: 'total', width: '120px' },
  { title: 'Tax', key: 'tax', width: '100px' },
  { title: 'Created At', key: 'created_at', width: '180px' },
  { title: 'Actions', key: 'actions', sortable: false, width: '120px' }
]

const showInvoiceModal = ref(false)
const selectedInvoice = ref<any | null>(null)

function viewInvoice(invoice: any) {
  selectedInvoice.value = invoice
  showInvoiceModal.value = true
}

function closeInvoiceModal() {
  showInvoiceModal.value = false
  selectedInvoice.value = null
}
</script>

<style scoped>
.admin-section {
  margin-bottom: 2rem;
}
.section-header {
  margin-bottom: 1.5rem;
}
.dashboard-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
}
.section-subtitle {
  color: #94a3b8;
  font-size: 1rem;
}
.dashboard-card {
  background: rgba(10, 25, 47, 0.85);
  border: 1px solid var(--color-border, #334155);
  border-radius: 1rem;
  padding: 1.5rem;
  margin-bottom: 2rem;
}
.dashboard-card-header {
  margin-bottom: 1rem;
}
.dashboard-card-title {
  font-size: 1.2rem;
  font-weight: 500;
}
.dashboard-card-subtitle {
  color: #64748b;
  font-size: 0.95rem;
}
.table-container {
  margin-top: 1rem;
}
.action-btn {
  background: transparent !important;
  border: 1px solid var(--color-border) !important;
  color: var(--color-text) !important;
  font-weight: 500;
  transition: all 0.2s;
}
.action-btn:hover {
  background: rgba(59, 130, 246, 0.1) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}
</style>
