<template>
  <div class="records-table-container">
    <v-data-table :loading="loading" :headers="headers" :items="pendingRecords" class="records-table"
      :items-per-page="5" :no-data-text="'No payments found'" density="comfortable">
      <template v-slot:[`item.user_id`]="{ item }">
        <span>{{ item.user_id }}</span>
      </template>
      <template  v-slot:[`item.username`]="{ item }">
        <span>{{ item.username }}</span>
      </template>
      <template v-slot:[`item.created_at`]="{ item }">
        <span>{{ formatDate(item.created_at) }}</span>
      </template>
      <template v-slot:[`item.usd_amount`]="{ item }">
        <span>${{ item.usd_amount.toFixed(2) }}</span>
      </template>
      <template v-slot:[`item.transferred_usd_amount`]="{ item }">
        <span>${{ item.transferred_usd_amount.toFixed(2) }}</span>
      </template>
            <template v-slot:[`item.transfer_mode`]="{ item }">
        <span>{{ item.transfer_mode.replace('_', ' ') }}</span>
      </template>
      <template v-slot:[`item.status`]="{ item }">
        <v-chip :color="getStatusColor(item)" size="small" class="status-chip">
          {{ getStatus(item) }}
        </v-chip>
      </template>

    </v-data-table>
  </div>
</template>

<script setup lang="ts">
import { type PendingRecord } from '../../utils/userService'
import { formatDate } from '../../utils/uiUtils'

const props = defineProps({
  pendingRecords: {
    type: Array as () => PendingRecord[],
    required: true,
    default: () => []
  },
  showUserID: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const headers = [
  { title: 'Record Date', key: 'created_at' },
  { title: 'Requested Amount', key: 'usd_amount' },
  { title: 'Transferred Amount', key: 'transferred_usd_amount' },
  { title: 'Transfer Mode', key: 'transfer_mode' },
  { title: 'Status', key: 'status' },
]

if (props.showUserID) {
  headers.unshift({ title: 'User ID', key: 'user_id' },
  { title: 'Username', key: 'username' })
}



function getStatus(item: PendingRecord): string {
  if (item.transferred_usd_amount === item.usd_amount) {
    return 'Completed'
  }
  return 'Pending'
}

function getStatusColor(item: PendingRecord): string {
  const status = getStatus(item)
  if (status === 'Completed') {
    return 'success'
  }
  return 'gray'
}
</script>

<style scoped>
.records-table-container {
  margin-bottom: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.records-table {
  background: transparent;
  width: 100%;
}

.status-chip {
  font-size: 0.75rem;
  font-weight: var(--font-weight-medium);
}
</style>
